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

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/testutil"
	ec2beat "github.com/elastic/cloudbeat/internal/resources/providers/awslib/ec2"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestEC2InstanceFetcher_Fetch(t *testing.T) {
	instance1 := &ec2beat.Ec2Instance{
		Instance: types.Instance{
			IamInstanceProfile: &types.IamInstanceProfile{
				Id:  pointers.Ref("a123123"),
				Arn: pointers.Ref("123123:123123:123123"),
			},
			Tags: []types.Tag{
				{
					Key:   pointers.Ref("Name"),
					Value: pointers.Ref("test-server"),
				},
				{
					Key:   pointers.Ref("key"),
					Value: pointers.Ref("value"),
				},
			},
			InstanceId:       pointers.Ref("234567890"),
			Architecture:     types.ArchitectureValuesX8664,
			ImageId:          pointers.Ref("image-id"),
			InstanceType:     "instance-type",
			Platform:         "linux",
			PlatformDetails:  pointers.Ref("ubuntu"),
			VpcId:            pointers.Ref("vpc-id"),
			SubnetId:         pointers.Ref("subnetId"),
			Ipv6Address:      pointers.Ref("ipv6"),
			PublicIpAddress:  pointers.Ref("public-ip-addr"),
			PrivateIpAddress: pointers.Ref("private-ip-addre"),
			PublicDnsName:    pointers.Ref("public-dns"),
			PrivateDnsName:   pointers.Ref("private-dns"),
			Placement: &types.Placement{
				AvailabilityZone: pointers.Ref("1a"),
			},
			NetworkInterfaces: []types.InstanceNetworkInterface{
				{
					MacAddress: pointers.Ref("mac1"),
				},
				{
					MacAddress: pointers.Ref("mac2"),
				},
			},
		},
		Region: "us-east",
	}

	instance2 := &ec2beat.Ec2Instance{
		Instance: types.Instance{},
		Region:   "us-east",
	}

	in := []*ec2beat.Ec2Instance{instance1, nil, instance2}

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsEc2Instance,
			"arn:aws:ec2:us-east::ec2/234567890",
			"private-dns",
			inventory.WithRelatedAssetIds([]string{"234567890"}),
			inventory.WithRawAsset(instance1),
			inventory.WithLabels(map[string]string{"Name": "test-server", "key": "value"}),
			inventory.WithCloud(inventory.Cloud{
				Provider:         inventory.AwsCloudProvider,
				Region:           "us-east",
				AvailabilityZone: "1a",
				AccountID:        "123",
				AccountName:      "alias",
				InstanceID:       "234567890",
				InstanceName:     "test-server",
				MachineType:      "instance-type",
				ServiceName:      "AWS EC2",
			}),
			inventory.WithHost(inventory.Host{
				ID:           "234567890",
				Name:         "private-dns",
				Architecture: string(types.ArchitectureValuesX8664),
				Type:         "instance-type",
				IP:           "public-ip-addr",
				MacAddress:   []string{"mac1", "mac2"},
			}),
			inventory.WithUser(inventory.User{
				ID: "123123:123123:123123",
			}),
		),

		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsEc2Instance,
			"",
			"",
			inventory.WithRawAsset(instance2),
			inventory.WithLabels(map[string]string{}),
			inventory.WithCloud(inventory.Cloud{
				Provider:         inventory.AwsCloudProvider,
				Region:           "us-east",
				AvailabilityZone: "",
				AccountID:        "123",
				AccountName:      "alias",
				InstanceID:       "",
				InstanceName:     "",
				MachineType:      "",
				ServiceName:      "AWS EC2",
			}),
			inventory.WithHost(inventory.Host{
				MacAddress: []string{},
			}),
		),
	}

	logger := testhelper.NewLogger(t)
	provider := newMockEc2InstancesProvider(t)
	provider.EXPECT().DescribeInstances(mock.Anything).Return(in, nil)

	identity := &cloud.Identity{Account: "123", AccountAlias: "alias"}
	fetcher := newEc2InstancesFetcher(logger, identity, provider)

	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}
