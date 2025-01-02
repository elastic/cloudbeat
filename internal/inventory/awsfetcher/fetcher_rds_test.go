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

	"github.com/elastic/beats/v7/libbeat/ecs"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/testutil"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/rds"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

func TestRDSInstanceFetcher_Fetch(t *testing.T) {
	instance1 := rds.DBInstance{
		Identifier:              "db1",
		Arn:                     "arn:aws:rds:eu-west-1:123:db:db1",
		StorageEncrypted:        true,
		AutoMinorVersionUpgrade: true,
		PubliclyAccessible:      false,
		Subnets: []rds.Subnet{
			{
				ID: "subnet-aabbccdd",
				RouteTable: &rds.RouteTable{
					ID: "rtb-aabbccddee",
					Routes: []rds.Route{
						{
							GatewayId:            pointers.Ref("local"),
							DestinationCidrBlock: pointers.Ref("172.31.0.0/16"),
						},
					},
				},
			},
		},
	}
	instance2 := rds.DBInstance{
		Identifier:              "db2",
		Arn:                     "arn:aws:rds:eu-west-1:123:db:db2",
		StorageEncrypted:        true,
		AutoMinorVersionUpgrade: true,
		PubliclyAccessible:      true,
		Subnets: []rds.Subnet{
			{
				ID: "subnet-aabbccdd",
				RouteTable: &rds.RouteTable{
					ID: "rtb-aabbccddee",
					Routes: []rds.Route{
						{
							GatewayId:            pointers.Ref("local"),
							DestinationCidrBlock: pointers.Ref("172.31.0.0/16"),
						},
						{
							GatewayId:            pointers.Ref("igw-aabbccdd"),
							DestinationCidrBlock: pointers.Ref("0.0.0.0/0"),
						},
					},
				},
			},
		},
	}

	in := []awslib.AwsResource{instance1, instance2}

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsRds,
			"arn:aws:rds:eu-west-1:123:db:db1",
			"db1",
			inventory.WithRawAsset(instance1),
			inventory.WithCloud(ecs.Cloud{
				Provider:    inventory.AwsCloudProvider,
				AccountID:   "123",
				AccountName: "alias",
				ServiceName: "AWS RDS",
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsRds,
			"arn:aws:rds:eu-west-1:123:db:db2",
			"db2",
			inventory.WithRawAsset(instance2),
			inventory.WithCloud(ecs.Cloud{
				Provider:    inventory.AwsCloudProvider,
				AccountID:   "123",
				AccountName: "alias",
				ServiceName: "AWS RDS",
			}),
		),
	}

	logger := logp.NewLogger("test_fetcher_rds_instance")
	provider := newMockRdsProvider(t)
	provider.EXPECT().DescribeDBInstances(mock.Anything).Return(in, nil)

	identity := &cloud.Identity{Account: "123", AccountAlias: "alias"}
	fetcher := newRDSFetcher(logger, identity, provider)

	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}
