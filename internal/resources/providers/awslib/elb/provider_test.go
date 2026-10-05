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

package elb

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

var onlyDefaultRegion = []string{awslib.DefaultRegion}

func TestProvider_DescribeLoadBalancers(t *testing.T) {
	tests := []struct {
		name            string
		client          func() Client
		expectedResults int
		wantErr         bool
		regions         []string
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeLoadBalancers", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
		{
			name: "with resources",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeLoadBalancers", mock.Anything, mock.Anything, mock.Anything).
					Return(&elasticloadbalancing.DescribeLoadBalancersOutput{
						LoadBalancerDescriptions: []types.LoadBalancerDescription{
							{
								AvailabilityZones:         []string{"us-east-1a"},
								CanonicalHostedZoneName:   pointers.Ref("HZ-NAME"),
								CanonicalHostedZoneNameID: pointers.Ref("HZ-ID"),
								CreatedTime:               pointers.Ref(time.Now()),
								DNSName:                   pointers.Ref("internal-my-elb-v1.us-east-1.elb.amazonaws.com"),
								ListenerDescriptions: []types.ListenerDescription{
									{
										Listener: &types.Listener{
											Protocol:         pointers.Ref("HTTP"),
											LoadBalancerPort: 80,
											InstanceProtocol: pointers.Ref("HTTP"),
											InstancePort:     pointers.Ref(int32(80)),
										},
									},
								},
								LoadBalancerName: pointers.Ref("my-elb-v1"),
								Scheme:           pointers.Ref("internal"),
								SecurityGroups:   []string{"sg-123"},
								SourceSecurityGroup: &types.SourceSecurityGroup{
									OwnerAlias: pointers.Ref("123"),
									GroupName:  pointers.Ref("default"),
								},
								Subnets: []string{"subnet-123"},
								VPCId:   pointers.Ref("vpc-id"),
							},
						},
					}, nil)
				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			var client Client
			for _, r := range tt.regions {
				clients[r] = tt.client()
				client = tt.client()
			}
			p := &Provider{
				log:     testhelper.NewLogger(t),
				clients: clients,
				client:  client,
			}
			got, err := p.DescribeLoadBalancers(t.Context(), []string{})
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.expectedResults)
		})
	}
}

// elbV1ClientWithResources builds a mock Client that returns one classic ELB with a DNSName.
func elbV1ClientWithResources() Client {
	m := &MockClient{}
	m.On("DescribeLoadBalancers", mock.Anything, mock.Anything).
		Return(&elasticloadbalancing.DescribeLoadBalancersOutput{
			LoadBalancerDescriptions: []types.LoadBalancerDescription{
				{
					AvailabilityZones:         []string{"us-east-1a"},
					CanonicalHostedZoneName:   pointers.Ref("HZ-NAME"),
					CanonicalHostedZoneNameID: pointers.Ref("HZ-ID"),
					CreatedTime:               pointers.Ref(time.Now()),
					DNSName:                   pointers.Ref("internal-my-elb-v1.us-east-1.elb.amazonaws.com"),
					ListenerDescriptions: []types.ListenerDescription{
						{
							Listener: &types.Listener{
								Protocol:         pointers.Ref("HTTP"),
								LoadBalancerPort: 80,
								InstanceProtocol: pointers.Ref("HTTP"),
								InstancePort:     pointers.Ref(int32(80)),
							},
						},
					},
					LoadBalancerName: pointers.Ref("my-elb-v1"),
					Scheme:           pointers.Ref("internal"),
					SecurityGroups:   []string{"sg-123"},
					SourceSecurityGroup: &types.SourceSecurityGroup{
						OwnerAlias: pointers.Ref("123"),
						GroupName:  pointers.Ref("default"),
					},
					Subnets: []string{"subnet-123"},
					VPCId:   pointers.Ref("vpc-id"),
				},
			},
		}, nil)
	m.On("DescribeTags", mock.Anything, mock.Anything, mock.Anything).
		Return(&elasticloadbalancing.DescribeTagsOutput{
			TagDescriptions: []types.TagDescription{
				{
					LoadBalancerName: pointers.Ref("my-elb-v1"),
					Tags: []types.Tag{
						{Key: pointers.Ref("Owner"), Value: pointers.Ref("team-infra")},
					},
				},
			},
		}, nil)
	return m
}

func TestProvider_DescribeAllLoadBalancers(t *testing.T) {
	tests := []struct {
		name            string
		client          func() Client
		resolver        func(t *testing.T) hostResolver
		expectedResults int
		wantErr         bool
		regions         []string
		checkResult     func(t *testing.T, got []awslib.AwsResource)
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeLoadBalancers", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			resolver: func(t *testing.T) hostResolver {
				t.Helper()
				// LookupHost is never reached: DescribeLoadBalancers fails first.
				return newMockHostResolver(t)
			},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
		{
			name:   "with resources and DNS IPs",
			client: elbV1ClientWithResources,
			resolver: func(t *testing.T) hostResolver {
				t.Helper()
				m := newMockHostResolver(t)
				// unsorted on purpose: the provider is expected to sort the IPs
				m.EXPECT().LookupHost(mock.Anything, mock.Anything).Return([]string{"10.0.0.2", "10.0.0.1"}, nil)
				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
			checkResult: func(t *testing.T, got []awslib.AwsResource) {
				t.Helper()
				lb, ok := got[0].(*ElasticLoadBalancerInfo)
				require.True(t, ok)
				assert.Equal(t, "team-infra", lb.GetOwnerTag())
				assert.Equal(t, []string{"10.0.0.1", "10.0.0.2"}, lb.GetIPAddresses(), "IPs should be sorted")
			},
		},
		{
			name:   "with resolver error (soft-fail)",
			client: elbV1ClientWithResources,
			resolver: func(t *testing.T) hostResolver {
				t.Helper()
				m := newMockHostResolver(t)
				m.EXPECT().LookupHost(mock.Anything, mock.Anything).Return(nil, errors.New("dns timeout"))
				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
			checkResult: func(t *testing.T, got []awslib.AwsResource) {
				t.Helper()
				lb, ok := got[0].(*ElasticLoadBalancerInfo)
				require.True(t, ok)
				assert.Equal(t, "team-infra", lb.GetOwnerTag())
				assert.Nil(t, lb.GetIPAddresses(), "IPs should be nil on DNS error (soft-fail)")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			var client Client
			for _, r := range tt.regions {
				clients[r] = tt.client()
				client = tt.client()
			}
			p := &Provider{
				log:      testhelper.NewLogger(t),
				clients:  clients,
				client:   client,
				resolver: tt.resolver(t),
			}
			got, err := p.DescribeAllLoadBalancers(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.expectedResults)
			if tt.checkResult != nil && len(got) > 0 {
				tt.checkResult(t, got)
			}
		})
	}
}
