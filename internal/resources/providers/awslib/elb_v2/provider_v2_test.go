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

package elb_v2

import (
	"errors"
	"testing"
	"time"

	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

var onlyDefaultRegion = []string{awslib.DefaultRegion}

func TestProvider_DescribeLoadBalancers(t *testing.T) {
	elbWithDNSNoIPs := func() *MockClient {
		m := &MockClient{}
		m.On("DescribeListeners", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeListenersOutput{
			Listeners: []types.Listener{},
		}, nil)
		m.On("DescribeLoadBalancers", mock.Anything, mock.Anything).
			Return(&elb.DescribeLoadBalancersOutput{
				LoadBalancers: []types.LoadBalancer{
					{
						AvailabilityZones:     []types.AvailabilityZone{},
						CanonicalHostedZoneId: pointers.Ref("HZ-ID"),
						CreatedTime:           pointers.Ref(time.Now()),
						DNSName:               pointers.Ref("my-nlb.us-east-1.elb.amazonaws.com"),
						LoadBalancerArn:       pointers.Ref("arn:aws:elasticloadbalancing:::loadbalancer/my-elb-v2"),
						LoadBalancerName:      pointers.Ref("my-elb-v2"),
						Scheme:                types.LoadBalancerSchemeEnumInternetFacing,
						Type:                  types.LoadBalancerTypeEnumNetwork,
					},
				},
			}, nil)
		m.On("DescribeTags", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeTagsOutput{}, nil)
		return m
	}

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
				m.On("DescribeListeners", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeListenersOutput{
					Listeners: []types.Listener{},
				}, nil)
				m.On("DescribeLoadBalancers", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			resolver: func(t *testing.T) hostResolver {
				// DescribeLoadBalancers fails before any lookup is attempted.
				return newMockHostResolver(t)
			},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
		{
			name: "listeners error does not cause global error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeListeners", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				m.On("DescribeLoadBalancers", mock.Anything, mock.Anything).Return(&elb.DescribeLoadBalancersOutput{
					LoadBalancers: []types.LoadBalancer{
						{
							LoadBalancerArn:  pointers.Ref("arn:aws:elasticloadbalancing:::loadbalancer/my-elb-v2"),
							LoadBalancerName: pointers.Ref("my-elb-v2"),
						},
					},
				}, nil)
				m.On("DescribeTags", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeTagsOutput{}, nil)
				return m
			},
			resolver: func(t *testing.T) hostResolver {
				// No DNSName on the LB, so LookupHost is never reached.
				return newMockHostResolver(t)
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
		},
		{
			name: "API returns IPs — resolver not called",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeListeners", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeListenersOutput{
					Listeners: []types.Listener{},
				}, nil)
				m.On("DescribeLoadBalancers", mock.Anything, mock.Anything).
					Return(&elb.DescribeLoadBalancersOutput{
						LoadBalancers: []types.LoadBalancer{
							{
								AvailabilityZones: []types.AvailabilityZone{
									{
										LoadBalancerAddresses: []types.LoadBalancerAddress{
											{IpAddress: pointers.Ref("203.0.113.10")},
											{PrivateIPv4Address: pointers.Ref("10.0.1.5")},
										},
									},
								},
								CanonicalHostedZoneId: pointers.Ref("HZ-ID"),
								CreatedTime:           pointers.Ref(time.Now()),
								CustomerOwnedIpv4Pool: pointers.Ref("10.0.0.0/24"),
								DNSName:               pointers.Ref("internal-my-elb-v2.us-east-1.elb.amazonaws.com"),
								EnforceSecurityGroupInboundRulesOnPrivateLinkTraffic: pointers.Ref(""),
								LoadBalancerArn:  pointers.Ref("arn:aws:elasticloadbalancing:::loadbalancer/my-elb-v2"),
								LoadBalancerName: pointers.Ref("my-elb-v2"),
								Scheme:           types.LoadBalancerSchemeEnumInternal,
								SecurityGroups:   []string{},
								State:            &types.LoadBalancerState{Code: types.LoadBalancerStateEnumActive},
								Type:             types.LoadBalancerTypeEnumNetwork,
								VpcId:            pointers.Ref(""),
							},
						},
					}, nil)
				m.On("DescribeTags", mock.Anything, mock.Anything, mock.Anything).
					Return(&elb.DescribeTagsOutput{
						TagDescriptions: []types.TagDescription{
							{
								ResourceArn: pointers.Ref("arn:aws:elasticloadbalancing:::loadbalancer/my-elb-v2"),
								Tags: []types.Tag{
									{Key: pointers.Ref("Owner"), Value: pointers.Ref("team-infra")},
								},
							},
						},
					}, nil)
				return m
			},
			resolver: func(t *testing.T) hostResolver {
				// API returned IPs, so the DNS fallback is not reached.
				return newMockHostResolver(t)
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
			checkResult: func(t *testing.T, got []awslib.AwsResource) {
				t.Helper()
				lb, ok := got[0].(*ElasticLoadBalancerInfo)
				require.True(t, ok)
				assert.Equal(t, "team-infra", lb.GetOwnerTag())
				assert.Equal(t, "network", lb.GetLoadBalancerType())
				assert.Equal(t, "active", lb.GetState())
				assert.Equal(t, []string{"203.0.113.10", "10.0.1.5"}, lb.GetIPAddresses())
			},
		},
		{
			name:   "DNS fallback used when API returns no IPs",
			client: func() Client { return elbWithDNSNoIPs() },
			resolver: func(t *testing.T) hostResolver {
				m := newMockHostResolver(t)
				// Intentionally unsorted to verify the provider sorts the result.
				m.EXPECT().LookupHost(mock.Anything, "my-nlb.us-east-1.elb.amazonaws.com").
					Return([]string{"203.0.113.2", "203.0.113.1"}, nil)
				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
			checkResult: func(t *testing.T, got []awslib.AwsResource) {
				t.Helper()
				lb, ok := got[0].(*ElasticLoadBalancerInfo)
				require.True(t, ok)
				assert.Equal(t, []string{"203.0.113.1", "203.0.113.2"}, lb.GetIPAddresses())
			},
		},
		{
			name:   "resolver error — soft-fail, LB still emitted with no IPs",
			client: func() Client { return elbWithDNSNoIPs() },
			resolver: func(t *testing.T) hostResolver {
				m := newMockHostResolver(t)
				m.EXPECT().LookupHost(mock.Anything, "my-nlb.us-east-1.elb.amazonaws.com").
					Return(nil, errors.New("dns timeout"))
				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
			checkResult: func(t *testing.T, got []awslib.AwsResource) {
				t.Helper()
				lb, ok := got[0].(*ElasticLoadBalancerInfo)
				require.True(t, ok)
				assert.Empty(t, lb.GetIPAddresses())
			},
		},
		{
			name: "with resources + listeners",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeListeners", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeListenersOutput{
					Listeners: []types.Listener{
						{
							ListenerArn: pointers.Ref("arn"),
							Port:        pointers.Ref(int32(8080)),
						},
					},
				}, nil)
				m.On("DescribeLoadBalancers", mock.Anything, mock.Anything).
					Return(&elb.DescribeLoadBalancersOutput{
						LoadBalancers: []types.LoadBalancer{
							{
								AvailabilityZones:     []types.AvailabilityZone{},
								CanonicalHostedZoneId: pointers.Ref("HZ-ID"),
								CreatedTime:           pointers.Ref(time.Now()),
								CustomerOwnedIpv4Pool: pointers.Ref("10.0.0.0/24"),
								DNSName:               pointers.Ref("internal-my-elb-v2.us-east-1.elb.amazonaws.com"),
								EnforceSecurityGroupInboundRulesOnPrivateLinkTraffic: pointers.Ref(""),
								LoadBalancerArn:  pointers.Ref("arn:aws:elasticloadbalancing:::loadbalancer/my-elb-v2"),
								LoadBalancerName: pointers.Ref("my-elb-v2"),
								Scheme:           types.LoadBalancerSchemeEnumInternal,
								SecurityGroups:   []string{},
								Type:             types.LoadBalancerTypeEnumApplication,
								VpcId:            pointers.Ref(""),
							},
						},
					}, nil)
				m.On("DescribeTags", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeTagsOutput{}, nil)
				return m
			},
			resolver: func(t *testing.T) hostResolver {
				m := newMockHostResolver(t)
				m.EXPECT().LookupHost(mock.Anything, "internal-my-elb-v2.us-east-1.elb.amazonaws.com").
					Return([]string{"10.0.0.5"}, nil)
				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = tt.client()
			}
			p := &Provider{
				log:      testhelper.NewLogger(t),
				clients:  clients,
				resolver: tt.resolver(t),
			}
			got, err := p.DescribeLoadBalancers(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.expectedResults)
			if tt.checkResult != nil {
				tt.checkResult(t, got)
			}
		})
	}
}
