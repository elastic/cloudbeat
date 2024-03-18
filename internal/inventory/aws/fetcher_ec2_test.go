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

package aws

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/inventory"
	ec2beat "github.com/elastic/cloudbeat/internal/resources/providers/awslib/ec2"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

func TestFetch(t *testing.T) {
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
			ec2Classification,
			"arn:aws:ec2:us-east::ec2/234567890",
			"test-server",
			inventory.WithRawAsset(instance1),
			inventory.WithTags(map[string]string{"Name": "test-server", "key": "value"}),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AwsCloudProvider,
				Region:   "us-east",
			}),
			inventory.WithHost(inventory.AssetHost{
				Architecture:    string(types.ArchitectureValuesX8664),
				ImageId:         pointers.Ref("image-id"),
				InstanceType:    "instance-type",
				Platform:        "linux",
				PlatformDetails: pointers.Ref("ubuntu"),
			}),
			inventory.WithIAM(inventory.AssetIAM{
				Id:  pointers.Ref("a123123"),
				Arn: pointers.Ref("123123:123123:123123"),
			}),
			inventory.WithNetwork(inventory.AssetNetwork{
				NetworkId:        pointers.Ref("vpc-id"),
				SubnetId:         pointers.Ref("subnetId"),
				Ipv6Address:      pointers.Ref("ipv6"),
				PublicIpAddress:  pointers.Ref("public-ip-addr"),
				PrivateIpAddress: pointers.Ref("private-ip-addre"),
				PublicDnsName:    pointers.Ref("public-dns"),
				PrivateDnsName:   pointers.Ref("private-dns"),
			}),
		),

		inventory.NewAssetEvent(
			ec2Classification,
			"",
			"",
			inventory.WithRawAsset(instance2),
			inventory.WithTags(map[string]string{}),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AwsCloudProvider,
				Region:   "us-east",
			}),
			inventory.WithHost(inventory.AssetHost{}),
			inventory.WithNetwork(inventory.AssetNetwork{}),
		),
	}

	logger := logp.NewLogger("test_fetcher_ec2")
	provider := newMockInstancesProvider(t)
	provider.EXPECT().DescribeInstances(mock.Anything).Return(in, nil)

	fetcher := Ec2Fetcher{
		logger:   logger,
		provider: provider,
	}

	ch := make(chan inventory.AssetEvent)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	go func() {
		fetcher.Fetch(ctx, ch)
	}()

	received := make([]inventory.AssetEvent, 0, len(expected))
	for len(expected) != len(received) {
		select {
		case <-ctx.Done():
			assert.ElementsMatch(t, expected, received)
			return
		case event := <-ch:
			received = append(received, event)
		}
	}

	assert.ElementsMatch(t, expected, received)
}
