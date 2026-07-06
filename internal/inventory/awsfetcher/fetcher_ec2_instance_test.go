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
	"errors"
	"testing"
	"time"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/testutil"
	ec2beat "github.com/elastic/cloudbeat/internal/resources/providers/awslib/ec2"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
	"github.com/elastic/cloudbeat/internal/statushandler"
)

const (
	testProfileArn = "arn:aws:iam::123456789012:instance-profile/test-profile"
	testRoleArn    = "arn:aws:iam::123456789012:role/test-role"
)

func makeTestInstance(launchTime *time.Time) *ec2beat.Ec2Instance {
	return &ec2beat.Ec2Instance{
		Instance: ec2types.Instance{
			IamInstanceProfile: &ec2types.IamInstanceProfile{
				Id:  pointers.Ref("a123123"),
				Arn: pointers.Ref(testProfileArn),
			},
			Tags: []ec2types.Tag{
				{Key: pointers.Ref("Name"), Value: pointers.Ref("test-server")},
				{Key: pointers.Ref("key"), Value: pointers.Ref("value")},
				{Key: pointers.Ref("Owner"), Value: pointers.Ref("team-infra")},
				{Key: pointers.Ref("CostCenter"), Value: pointers.Ref("cc-1234")},
			},
			InstanceId:       pointers.Ref("234567890"),
			Architecture:     ec2types.ArchitectureValuesX8664,
			ImageId:          pointers.Ref("image-id"),
			InstanceType:     "instance-type",
			Platform:         "linux",
			PlatformDetails:  pointers.Ref("ubuntu"),
			VpcId:            pointers.Ref("vpc-id"),
			SubnetId:         pointers.Ref("subnetId"),
			State:            &ec2types.InstanceState{Name: ec2types.InstanceStateNameRunning},
			Ipv6Address:      pointers.Ref("ipv6"),
			PublicIpAddress:  pointers.Ref("public-ip-addr"),
			PrivateIpAddress: pointers.Ref("private-ip-addre"),
			LaunchTime:       launchTime,
			PublicDnsName:    pointers.Ref("public-dns"),
			PrivateDnsName:   pointers.Ref("private-dns"),
			Placement:        &ec2types.Placement{AvailabilityZone: pointers.Ref("1a")},
			NetworkInterfaces: []ec2types.InstanceNetworkInterface{
				{MacAddress: pointers.Ref("mac1")},
				{MacAddress: pointers.Ref("mac2")},
			},
		},
		Region: "us-east",
	}
}

func TestEC2InstanceFetcher_Fetch(t *testing.T) {
	launchTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	instance1 := makeTestInstance(&launchTime)
	instance2 := &ec2beat.Ec2Instance{
		Instance: ec2types.Instance{},
		Region:   "us-east",
	}

	in := []*ec2beat.Ec2Instance{instance1, nil, instance2}

	resolvedProfile := &iamtypes.InstanceProfile{
		InstanceProfileName: pointers.Ref("test-profile"),
		Arn:                 pointers.Ref(testProfileArn),
		Roles: []iamtypes.Role{
			{Arn: pointers.Ref(testRoleArn)},
		},
	}

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsEc2Instance,
			"arn:aws:ec2:us-east::ec2/234567890",
			"private-dns",
			inventory.WithRelatedAssetIds([]string{"234567890"}),
			inventory.WithRawAsset(instance1),
			inventory.WithLabels(map[string]string{"Name": "test-server", "key": "value", "Owner": "team-infra", "CostCenter": "cc-1234"}),
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
				Architecture: string(ec2types.ArchitectureValuesX8664),
				Type:         "instance-type",
				IP:           []string{"public-ip-addr", "private-ip-addre"},
				MacAddress:   []string{"mac1", "mac2"},
			}),
			inventory.WithEntityAttributes(map[string]any{
				"ImageId":            "image-id",
				"Platform":           "linux",
				"VpcId":              "vpc-id",
				"SubnetId":           "subnetId",
				"State":              "running",
				"InstanceProfileArn": testProfileArn,
				"RoleArn":            testRoleArn,
				"Owner":              "team-infra",
				"CostCenter":         "cc-1234",
			}),
			inventory.WithCreatedAt(&launchTime),
			inventory.WithUser(inventory.User{
				ID: testRoleArn,
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

	resolver := newMockInstanceProfileResolver(t)
	resolver.EXPECT().GetInstanceProfile(mock.Anything, "test-profile").Return(resolvedProfile, nil)

	msh := statushandler.NewMockStatusHandlerAPI(t)

	identity := &cloud.Identity{Account: "123", AccountAlias: "alias"}
	fetcher := newEc2InstancesFetcher(logger, identity, provider, resolver, msh)

	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}

func TestEC2InstanceFetcher_Fetch_ResolverError(t *testing.T) {
	launchTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	instance := makeTestInstance(&launchTime)
	in := []*ec2beat.Ec2Instance{instance}

	// When the resolver fails, InstanceProfileArn is emitted but RoleArn is not.
	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsEc2Instance,
			"arn:aws:ec2:us-east::ec2/234567890",
			"private-dns",
			inventory.WithRelatedAssetIds([]string{"234567890"}),
			inventory.WithRawAsset(instance),
			inventory.WithLabels(map[string]string{"Name": "test-server", "key": "value", "Owner": "team-infra", "CostCenter": "cc-1234"}),
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
				Architecture: string(ec2types.ArchitectureValuesX8664),
				Type:         "instance-type",
				IP:           []string{"public-ip-addr", "private-ip-addre"},
				MacAddress:   []string{"mac1", "mac2"},
			}),
			inventory.WithEntityAttributes(map[string]any{
				"ImageId":            "image-id",
				"Platform":           "linux",
				"VpcId":              "vpc-id",
				"SubnetId":           "subnetId",
				"State":              "running",
				"InstanceProfileArn": testProfileArn,
				// RoleArn is absent because the resolver failed.
				"Owner":      "team-infra",
				"CostCenter": "cc-1234",
			}),
			inventory.WithCreatedAt(&launchTime),
			// WithUser falls back to the profile ARN when role resolution fails.
			inventory.WithUser(inventory.User{
				ID: testProfileArn,
			}),
		),
	}

	logger := testhelper.NewLogger(t)
	provider := newMockEc2InstancesProvider(t)
	provider.EXPECT().DescribeInstances(mock.Anything).Return(in, nil)

	resolver := newMockInstanceProfileResolver(t)
	resolver.EXPECT().GetInstanceProfile(mock.Anything, "test-profile").Return(nil, errors.New("AccessDenied"))

	msh := statushandler.NewMockStatusHandlerAPI(t)

	identity := &cloud.Identity{Account: "123", AccountAlias: "alias"}
	fetcher := newEc2InstancesFetcher(logger, identity, provider, resolver, msh)

	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}
