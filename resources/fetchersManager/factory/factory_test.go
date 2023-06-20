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

package factory

import (
	"context"
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/uniqueness"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"testing"
)

type expectedFetchers struct {
	names []string
	count int
}

func TestNewFactory(t *testing.T) {
	logger := logp.NewLogger("test new factory")
	ch := make(chan fetching.ResourceInfo)
	le := &uniqueness.DefaultUniqueManager{}
	kubeClient := k8sfake.NewSimpleClientset()

	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
		want    expectedFetchers
	}{
		{
			name: "Get k8s factory",
			cfg: &config.Config{
				Benchmark: config.CIS_K8S,
			},
			want: expectedFetchers{
				names: []string{
					fetching.FileSystemType,
					fetching.KubeAPIType,
					fetching.ProcessType,
				},
				count: 3,
			},
		},
		{
			name: "Get CIS AWS factory",
			cfg: &config.Config{
				Benchmark: config.CIS_AWS,
				CloudConfig: config.CloudConfig{
					AwsCred: aws.ConfigAWS{
						AccessKeyID: "test",
					},
				},
			},
			want: expectedFetchers{
				names: []string{
					fetching.IAMType,
					fetching.KmsType,
					fetching.TrailType,
					fetching.MonitoringType,
					fetching.EC2NetworkingType,
					fetching.RdsType,
					fetching.S3Type,
				},
				count: 7,
			},
		},
		{
			name: "No AWS credentials - unable to get CIS AWS factory",
			cfg: &config.Config{
				Benchmark: config.CIS_AWS,
			},
			want: expectedFetchers{
				names: []string{},
				count: 0,
			},
			wantErr: true,
		},
		{
			name: "Get CIS EKS factory without the aws related fetchers",
			cfg: &config.Config{
				Benchmark: config.CIS_EKS,
			},
			want: expectedFetchers{
				names: []string{
					fetching.FileSystemType,
					fetching.KubeAPIType,
					fetching.ProcessType,
					fetching.EcrType,
					fetching.ElbType,
				},
				count: 3,
			},
		},
		{
			name: "Get CIS EKS factory with aws related fetchers",
			cfg: &config.Config{
				Benchmark: config.CIS_EKS,
				CloudConfig: config.CloudConfig{
					AwsCred: aws.ConfigAWS{
						AccessKeyID: "test",
					},
				},
			},
			want: expectedFetchers{
				names: []string{
					fetching.FileSystemType,
					fetching.KubeAPIType,
					fetching.ProcessType,
					fetching.EcrType,
					fetching.ElbType,
				},
				count: 5,
			},
		},
		{
			name: "Non supported benchmark fail to get factory",
			cfg: &config.Config{
				Benchmark: "Non existing benchmark",
			},
			want: expectedFetchers{
				names: []string{},
				count: 0,
			},
			wantErr: true,
		},
		{
			name: "No config fail to get factory",
			cfg:  nil,
			want: expectedFetchers{
				names: []string{},
				count: 0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		identity := &awslib.MockIdentityProviderGetter{}
		identity.EXPECT().GetIdentity(mock.Anything).Return(&awslib.Identity{
			Account: awssdk.String("test-account"),
		}, nil)

		identityProvider := func(cfg awssdk.Config) awslib.IdentityProviderGetter {
			return identity
		}

		t.Run(tt.name, func(t *testing.T) {

			fetchersMap, err := NewFactory(context.TODO(), logger, tt.cfg, ch, le, kubeClient, identityProvider)
			assert.Equal(t, tt.want.count, len(fetchersMap))
			for fetcher := range fetchersMap {
				if _, ok := fetchersMap[fetcher]; !ok {
					t.Errorf("NewFactory() fetchersMap = %v, want %v", fetchersMap, tt.want.names)
				}
			}

			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}
