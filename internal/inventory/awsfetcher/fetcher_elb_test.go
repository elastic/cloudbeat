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

package awsfetcher

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	typesv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/elb"
	elbv2 "github.com/elastic/cloudbeat/internal/resources/providers/awslib/elb_v2"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

func TestELBv1Fetcher_Fetch(t *testing.T) {
	asset := elb.ElasticLoadBalancerInfo{
		LoadBalancer: types.LoadBalancerDescription{
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
	}
	in := []awslib.AwsResource{asset}

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			newElbClassification(inventory.SubTypeELBv1),
			[]string{"arn:aws:elasticloadbalancing:::loadbalancer/my-elb-v1"},
			"my-elb-v1",
			inventory.WithRawAsset(asset),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AwsCloudProvider,
				Account: inventory.AssetCloudAccount{
					Id:   "123",
					Name: "alias",
				},
				Service: &inventory.AssetCloudService{
					Name: "AWS Networking",
				},
			}),
		),
	}

	logger := logp.NewLogger("test_fetcher_elb_v1")
	providerv1 := newMockV1Provider(t)
	providerv1.EXPECT().DescribeAllLoadBalancers(mock.Anything).Return(in, nil)
	providerv2 := newMockV2Provider(t)
	providerv2.EXPECT().DescribeLoadBalancers(mock.Anything).Return(nil, nil)

	identity := &cloud.Identity{Account: "123", AccountAlias: "alias"}
	fetcher := newElbFetcher(logger, identity, providerv1, providerv2)

	collectResourcesAndMatch(t, fetcher, expected)
}

func TestELBv2Fetcher_Fetch(t *testing.T) {
	asset := elbv2.ElasticLoadBalancerInfo{
		LoadBalancer: typesv2.LoadBalancer{
			AvailabilityZones:     []typesv2.AvailabilityZone{},
			CanonicalHostedZoneId: pointers.Ref("HZ-ID"),
			CreatedTime:           pointers.Ref(time.Now()),
			CustomerOwnedIpv4Pool: pointers.Ref("10.0.0.0/24"),
			DNSName:               pointers.Ref("internal-my-elb-v2.us-east-1.elb.amazonaws.com"),
			EnforceSecurityGroupInboundRulesOnPrivateLinkTraffic: pointers.Ref(""),
			LoadBalancerArn:  pointers.Ref("arn:aws:elasticloadbalancing:::loadbalancer/my-elb-v2"),
			LoadBalancerName: pointers.Ref("my-elb-v2"),
			Scheme:           typesv2.LoadBalancerSchemeEnumInternal,
			SecurityGroups:   []string{},
			Type:             typesv2.LoadBalancerTypeEnumApplication,
			VpcId:            pointers.Ref(""),
		},
	}
	in := []awslib.AwsResource{asset}

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			newElbClassification(inventory.SubTypeELBv2),
			[]string{"arn:aws:elasticloadbalancing:::loadbalancer/my-elb-v2"},
			"my-elb-v2",
			inventory.WithRawAsset(asset),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AwsCloudProvider,
				Account: inventory.AssetCloudAccount{
					Id:   "123",
					Name: "alias",
				},
				Service: &inventory.AssetCloudService{
					Name: "AWS Networking",
				},
			}),
		),
	}

	logger := logp.NewLogger("test_fetcher_elb_v2")
	providerv1 := newMockV1Provider(t)
	providerv1.EXPECT().DescribeAllLoadBalancers(mock.Anything).Return(nil, nil)
	providerv2 := newMockV2Provider(t)
	providerv2.EXPECT().DescribeLoadBalancers(mock.Anything).Return(in, nil)

	identity := &cloud.Identity{Account: "123", AccountAlias: "alias"}
	fetcher := newElbFetcher(logger, identity, providerv1, providerv2)

	collectResourcesAndMatch(t, fetcher, expected)
}
