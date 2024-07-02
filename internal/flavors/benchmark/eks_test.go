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
	"errors"
	"testing"

	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/config"
	k8sprovider "github.com/elastic/cloudbeat/internal/dataprovider/providers/k8s"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestEKS_Initialize(t *testing.T) {
	testhelper.SkipLong(t)

	t.Setenv("NODE_NAME", "node-name")
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
		name                   string
		cfg                    config.Config
		awsIdentityProvider    awslib.IdentityProviderGetter
		awsMetadataProvider    awslib.MetadataProvider
		eksClusterNameProvider awslib.EKSClusterNameProviderAPI
		clientProvider         k8sprovider.ClientGetterAPI
		want                   []string
		wantErr                string
	}{
		{
			name:    "nothing initialized",
			wantErr: "uninitialized",
		},
		{
			name:                   "aws identity provider error",
			awsIdentityProvider:    mockAwsIdentityProvider(errors.New("some error")),
			clientProvider:         mockKubeClient(nil),
			awsMetadataProvider:    mockMetadataProvider(nil),
			eksClusterNameProvider: mockEksClusterNameProvider(errors.New("not this error")), // ignored
			cfg:                    awsCfg,
			wantErr:                "some error",
		},
		{
			// TODO
			name:                   "kubernetes provider error",
			awsIdentityProvider:    mockAwsIdentityProvider(errors.New("not this error")), // ineffectual
			clientProvider:         mockKubeClient(errors.New("some error")),
			awsMetadataProvider:    mockMetadataProvider(errors.New("not this error")),       // ignored
			eksClusterNameProvider: mockEksClusterNameProvider(errors.New("not this error")), // ignored
			wantErr:                "some error",
		},
		{
			name:                   "no error without AWS-related fetchers",
			awsIdentityProvider:    mockAwsIdentityProvider(errors.New("some error")),    // ineffectual
			awsMetadataProvider:    mockMetadataProvider(errors.New("some error")),       // ignored
			eksClusterNameProvider: mockEksClusterNameProvider(errors.New("some error")), // ignored
			clientProvider:         mockKubeClient(nil),
			want: []string{
				fetching.FileSystemType,
				fetching.KubeAPIType,
				fetching.ProcessType,
			},
		},
		{
			name:                   "no error with AWS-related fetchers",
			cfg:                    awsCfg,
			awsIdentityProvider:    mockAwsIdentityProvider(nil),
			awsMetadataProvider:    mockMetadataProvider(nil),
			eksClusterNameProvider: mockEksClusterNameProvider(nil),
			clientProvider:         mockKubeClient(nil),
			want: []string{
				fetching.FileSystemType,
				fetching.KubeAPIType,
				fetching.ProcessType,
				fetching.EcrType,
				fetching.ElbType,
			},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testInitialize(t, &EKS{
				AWSIdentityProvider:    tt.awsIdentityProvider,
				AWSMetadataProvider:    tt.awsMetadataProvider,
				EKSClusterNameProvider: tt.eksClusterNameProvider,
				ClientProvider:         tt.clientProvider,
				leaderElector:          nil,
			}, &tt.cfg, tt.wantErr, tt.want)
		})
	}
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
