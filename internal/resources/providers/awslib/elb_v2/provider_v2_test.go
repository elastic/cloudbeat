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
	"context"
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
				m.On("DescribeListeners", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeListenersOutput{
					Listeners: []types.Listener{},
				}, nil)
				m.On("DescribeLoadBalancers", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
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
				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
		},
		{
			name: "with resources",
			client: func() Client {
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
				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
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
				log:     testhelper.NewLogger(t),
				clients: clients,
			}
			got, err := p.DescribeLoadBalancers(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.expectedResults)
		})
	}
}
