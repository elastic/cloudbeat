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

func TestProvider_DescribeAllLoadBalancers(t *testing.T) {
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
				m.On("DescribeLoadBalancers", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
		{
			name: "with resources",
			client: func() Client {
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
			got, err := p.DescribeAllLoadBalancers(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.expectedResults)
		})
	}
}
