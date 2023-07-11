// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package benchmark

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider/providers/k8s"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
)

func TestNewBenchmark(t *testing.T) {
	t.Setenv("NODE_NAME", "node-name")
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
		want    []string
	}{
		{
			name: "Get k8s benchmark",
			cfg: &config.Config{
				Benchmark: config.CIS_K8S,
			},
			want: []string{
				fetching.FileSystemType,
				fetching.KubeAPIType,
				fetching.ProcessType,
			},
		},
		{
			name: "Get CIS AWS benchmark",
			cfg: &config.Config{
				Benchmark: config.CIS_AWS,
				CloudConfig: config.CloudConfig{
					Aws: config.AwsConfig{
						Cred: aws.ConfigAWS{
							AccessKeyID: "test",
						},
					},
				},
			},
			want: []string{
				fetching.IAMType,
				fetching.KmsType,
				fetching.TrailType,
				fetching.MonitoringType,
				fetching.EC2NetworkingType,
				fetching.RdsType,
				fetching.S3Type,
			},
		},
		{
			name: "Get CIS AWS org benchmark",
			cfg: &config.Config{
				Benchmark: config.CIS_AWS,
				CloudConfig: config.CloudConfig{
					Aws: config.AwsConfig{
						Cred: aws.ConfigAWS{
							AccessKeyID: "test",
						},
						AccountType: "organization_account",
					},
				},
			},
			want: []string{
				"123-" + fetching.IAMType,
				"123-" + fetching.KmsType,
				"123-" + fetching.TrailType,
				"123-" + fetching.MonitoringType,
				"123-" + fetching.EC2NetworkingType,
				"123-" + fetching.RdsType,
				"123-" + fetching.S3Type,
				"test-account-" + fetching.IAMType,
				"test-account-" + fetching.KmsType,
				"test-account-" + fetching.TrailType,
				"test-account-" + fetching.MonitoringType,
				"test-account-" + fetching.EC2NetworkingType,
				"test-account-" + fetching.RdsType,
				"test-account-" + fetching.S3Type,
			},
		},
		{
			name: "Get CIS EKS benchmark without the aws related fetchers",
			cfg: &config.Config{
				Benchmark: config.CIS_EKS,
			},
			want: []string{
				fetching.FileSystemType,
				fetching.KubeAPIType,
				fetching.ProcessType,
			},
		},
		{
			name: "Get CIS EKS benchmark with aws related fetchers",
			cfg: &config.Config{
				Benchmark: config.CIS_EKS,
				CloudConfig: config.CloudConfig{
					Aws: config.AwsConfig{
						Cred: aws.ConfigAWS{
							AccessKeyID: "test",
						},
					},
				},
			},
			want: []string{
				fetching.FileSystemType,
				fetching.KubeAPIType,
				fetching.ProcessType,
				fetching.EcrType,
				fetching.ElbType,
			},
		},
		{
			name: "Non supported benchmark fail",
			cfg: &config.Config{
				Benchmark: "Non existing benchmark",
			},
			want:    []string{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b, err := NewBenchmark(tt.cfg)
			if tt.wantErr {
				if b == nil {
					require.Error(t, err)
					return
				}
			} else {
				require.NoError(t, err)
			}
			fetchersMap, _, err := b.Initialize(
				context.Background(),
				testhelper.NewLogger(t),
				tt.cfg,
				make(chan fetching.ResourceInfo),
				&Dependencies{
					AwsCfgProvider:           mockAwsCfg(nil),
					AwsIdentityProvider:      mockIdentityProvider(nil),
					AwsAccountProvider:       mockAccountProvider(nil),
					KubernetesClientProvider: mockKubeClient(nil),
					AwsMetadataProvider:      mockMetadataProvider(nil),
					EksClusterNameProvider:   mockEksClusterNameProvider(nil),
				},
			)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, fetchersMap.Keys(), len(tt.want))

			require.NoError(t, b.Run(context.Background()))
			defer b.Stop()
			for _, fetcher := range tt.want {
				ok := fetchersMap.ShouldRun(fetcher)
				assert.Truef(t, ok, "fetcher %s enabled", fetcher)
			}
		})
	}
}

func Test_Initialize(t *testing.T) {
	awsCfg := config.Config{
		CloudConfig: config.CloudConfig{
			Aws: config.AwsConfig{
				Cred: aws.ConfigAWS{
					AccessKeyID: "some-key",
				},
			},
		},
	}

	tests := []struct {
		name         string
		benchmark    Benchmark
		dependencies Dependencies
		cfg          config.Config
		wantErr      string
	}{
		// AWS test
		{
			name:      "nothing initialized",
			benchmark: &AWS{},
			wantErr:   "aws identity provider is uninitialized",
		},
		{
			name:      "identity provider error",
			benchmark: &AWS{},
			dependencies: Dependencies{
				AwsCfgProvider:           nil,
				AwsIdentityProvider:      mockIdentityProvider(errors.New("some error")),
				AwsAccountProvider:       nil,
				KubernetesClientProvider: nil,
				AwsMetadataProvider:      nil,
				EksClusterNameProvider:   nil,
			},
			wantErr: "some error",
		},
		{
			// TODO: this doesn't finish instantly because there is code in MultiRegionClientFactory that is not initialized lazily
			name:      "no error",
			benchmark: &AWS{},
			dependencies: Dependencies{
				AwsCfgProvider:           nil,
				AwsIdentityProvider:      mockIdentityProvider(nil),
				AwsAccountProvider:       nil,
				KubernetesClientProvider: mockKubeClient(errors.New("some error")), // ineffectual
				AwsMetadataProvider:      nil,
				EksClusterNameProvider:   nil,
			},
		},
		// AWS org tests
		{
			name:      "nothing initialized",
			benchmark: &AWSOrg{},
			wantErr:   "aws identity provider is uninitialized",
		},
		{
			name:      "account provider uninitialized",
			benchmark: &AWSOrg{},
			dependencies: Dependencies{
				AwsCfgProvider:           nil,
				AwsIdentityProvider:      mockIdentityProvider(nil),
				AwsAccountProvider:       nil,
				KubernetesClientProvider: nil,
				AwsMetadataProvider:      nil,
				EksClusterNameProvider:   nil,
			},
			wantErr: "account provider is uninitialized",
		},
		{
			name:      "identity provider error",
			benchmark: &AWSOrg{},
			dependencies: Dependencies{
				AwsCfgProvider:           nil,
				AwsIdentityProvider:      mockIdentityProvider(errors.New("some error")),
				AwsAccountProvider:       nil,
				KubernetesClientProvider: nil,
				AwsMetadataProvider:      nil,
				EksClusterNameProvider:   nil,
			},
			wantErr: "some error",
		},
		{
			name:      "no error",
			benchmark: &AWSOrg{},
			dependencies: Dependencies{
				AwsCfgProvider:           nil,
				AwsIdentityProvider:      mockIdentityProvider(nil),
				AwsAccountProvider:       mockAccountProvider(nil),
				KubernetesClientProvider: mockKubeClient(errors.New("some error")), // ineffectual
				AwsMetadataProvider:      nil,
				EksClusterNameProvider:   nil,
			},
		},
		// K8S tests
		{
			name:      "nothing initialized",
			benchmark: &K8S{},
			wantErr:   "k8s provider is uninitialized",
		},
		{
			name:      "kubernetes provider error",
			benchmark: &K8S{},
			dependencies: Dependencies{
				AwsCfgProvider:           nil,
				AwsIdentityProvider:      nil,
				AwsAccountProvider:       nil,
				KubernetesClientProvider: mockKubeClient(errors.New("some error")),
				AwsMetadataProvider:      nil,
				EksClusterNameProvider:   nil,
			},
			wantErr: "some error",
		},
		{
			name:      "ignored uninitialized aws provider",
			benchmark: &K8S{},
			dependencies: Dependencies{
				AwsCfgProvider:           nil,
				AwsIdentityProvider:      nil,
				AwsAccountProvider:       nil,
				KubernetesClientProvider: mockKubeClient(nil),
				AwsMetadataProvider:      nil,
				EksClusterNameProvider:   nil,
			},
			cfg: awsCfg,
		},
		{
			name:      "no error",
			benchmark: &K8S{},
			dependencies: Dependencies{
				AwsCfgProvider:           nil,
				AwsIdentityProvider:      mockIdentityProvider(errors.New("some error")), // ineffectual
				AwsAccountProvider:       nil,
				KubernetesClientProvider: mockKubeClient(nil),
				AwsMetadataProvider:      nil,
				EksClusterNameProvider:   nil,
			},
		},
		// EKS tests
		{
			name:      "nothing initialized",
			benchmark: &EKS{},
			wantErr:   "k8s provider is uninitialized",
		},
		{
			name:      "kubernetes provider error",
			benchmark: &EKS{},
			dependencies: Dependencies{
				AwsCfgProvider:           nil,
				AwsIdentityProvider:      nil,
				AwsAccountProvider:       nil,
				KubernetesClientProvider: mockKubeClient(errors.New("some error")),
				AwsMetadataProvider:      nil,
				EksClusterNameProvider:   nil,
			},
			wantErr: "some error",
		},
		{
			name:      "uninitialized aws provider",
			benchmark: &EKS{},
			dependencies: Dependencies{
				AwsCfgProvider:           nil,
				AwsIdentityProvider:      nil,
				AwsAccountProvider:       nil,
				KubernetesClientProvider: mockKubeClient(nil),
				AwsMetadataProvider:      nil,
				EksClusterNameProvider:   nil,
			},
			cfg:     awsCfg,
			wantErr: "aws config provider is uninitialized",
		},
		{
			name:      "aws error",
			benchmark: &EKS{},
			dependencies: Dependencies{
				AwsCfgProvider:           mockAwsCfg(errors.New("some error")),
				AwsIdentityProvider:      nil,
				AwsAccountProvider:       nil,
				KubernetesClientProvider: mockKubeClient(nil),
				AwsMetadataProvider:      nil,
				EksClusterNameProvider:   nil,
			},
			cfg:     awsCfg,
			wantErr: "some error",
		},
		{
			name:      "aws identity provider error",
			benchmark: &EKS{},
			dependencies: Dependencies{
				AwsCfgProvider:           mockAwsCfg(nil),
				AwsIdentityProvider:      mockIdentityProvider(errors.New("some error")),
				AwsAccountProvider:       nil,
				KubernetesClientProvider: mockKubeClient(nil),
				AwsMetadataProvider:      nil,
				EksClusterNameProvider:   nil,
			},
			cfg:     awsCfg,
			wantErr: "some error",
		},
		{
			name:      "dependencies uninitialized",
			benchmark: &EKS{},
			dependencies: Dependencies{
				AwsCfgProvider:           nil,
				AwsIdentityProvider:      nil,
				AwsAccountProvider:       nil,
				KubernetesClientProvider: mockKubeClient(nil),
				AwsMetadataProvider:      nil,
				EksClusterNameProvider:   nil,
			},
			wantErr: "EKS dependencies uninitialized",
		},
		{
			name:      "no error",
			benchmark: &EKS{},
			dependencies: Dependencies{
				AwsCfgProvider:           mockAwsCfg(nil),
				AwsIdentityProvider:      mockIdentityProvider(errors.New("some error")), // ineffectual
				AwsAccountProvider:       nil,
				KubernetesClientProvider: mockKubeClient(nil),
				AwsMetadataProvider:      mockMetadataProvider(errors.New("some error")),       // ignored
				EksClusterNameProvider:   mockEksClusterNameProvider(errors.New("some error")), // ignored
			},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T: %s", tt.benchmark, tt.name), func(t *testing.T) {
			t.Setenv("NODE_NAME", "node-name")

			reg, dp, err := tt.benchmark.Initialize(
				context.Background(),
				testhelper.NewLogger(t),
				&tt.cfg,
				make(chan fetching.ResourceInfo),
				&tt.dependencies,
			)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, reg)
			assert.NotNil(t, dp)
		})
	}
}

func mockAwsCfg(err error) *awslib.MockConfigProviderAPI {
	awsCfg := awslib.MockConfigProviderAPI{}
	awsCfg.EXPECT().InitializeAWSConfig(mock.Anything, mock.Anything).
		Call.
		Return(
			func(ctx context.Context, config aws.ConfigAWS) *awssdk.Config {
				if err != nil {
					return nil
				}

				awsConfig := awssdk.NewConfig()
				awsCredentials := awssdk.Credentials{
					AccessKeyID:     config.AccessKeyID,
					SecretAccessKey: config.SecretAccessKey,
					SessionToken:    config.SessionToken,
				}

				awsConfig.Credentials = credentials.StaticCredentialsProvider{
					Value: awsCredentials,
				}
				awsConfig.Region = "us1-east"
				return awsConfig
			},
			func(ctx context.Context, config aws.ConfigAWS) error {
				return err
			},
		)
	return &awsCfg
}

func mockKubeClient(err error) k8s.ClientGetterAPI {
	kube := k8s.MockClientGetterAPI{}
	on := kube.EXPECT().GetClient(mock.Anything, mock.Anything, mock.Anything)
	if err == nil {
		on.Return(
			k8sfake.NewSimpleClientset(
				&core_v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-name",
					},
				},
				&core_v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "kube-system",
					},
				},
			), nil)
	} else {
		on.Return(nil, err)
	}
	return &kube
}

func mockIdentityProvider(err error) *awslib.MockIdentityProviderGetter {
	identityProvider := &awslib.MockIdentityProviderGetter{}
	on := identityProvider.EXPECT().GetIdentity(mock.Anything, mock.Anything)
	if err == nil {
		on.Return(
			&awslib.Identity{
				Account: "test-account",
			},
			nil,
		)
	} else {
		on.Return(nil, err)
	}
	return identityProvider
}

func mockMetadataProvider(err error) *awslib.MockMetadataProvider {
	provider := awslib.MockMetadataProvider{}
	on := provider.EXPECT().GetMetadata(mock.Anything, mock.Anything)
	if err == nil {
		on.Return(&awslib.Ec2Metadata{
			InstanceID: "instance-id",
		}, nil)
	} else {
		on.Return(nil, err)
	}

	return &provider
}

func mockEksClusterNameProvider(err error) *awslib.MockEKSClusterNameProviderAPI {
	provider := awslib.MockEKSClusterNameProviderAPI{}
	on := provider.EXPECT().GetClusterName(mock.Anything, mock.Anything, mock.Anything)
	if err == nil {
		on.Return("cluster-name", nil)
	} else {
		on.Return("", err)
	}

	return &provider
}

func mockAccountProvider(err error) *awslib.MockAccountProviderAPI {
	provider := awslib.MockAccountProviderAPI{}
	on := provider.EXPECT().ListAccounts(mock.Anything, mock.Anything)
	if err == nil {
		on.Return([]awslib.Identity{
			{
				Account: "123",
				Alias:   "some-name",
			},
		}, nil)
	} else {
		on.Return(nil, err)
	}
	return &provider
}

func mockAccountProviderWithIdentities(identities []awslib.Identity) *awslib.MockAccountProviderAPI {
	provider := awslib.MockAccountProviderAPI{}
	provider.EXPECT().ListAccounts(mock.Anything, mock.Anything).Return(identities, nil)
	return &provider
}
