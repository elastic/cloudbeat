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

package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elastic/beats/v7/libbeat/beat"
	libevents "github.com/elastic/beats/v7/libbeat/beat/events"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

func TestAssetInventory_Run(t *testing.T) {
	now := func() time.Time { return time.Date(2024, 1, 1, 1, 1, 1, 0, time.Local) }
	expected := []beat.Event{
		{
			Meta:      mapstr.M{libevents.FieldMetaIndex: "logs-cloud_asset_inventory.asset_inventory-infrastructure_compute_virtual-machine_ec2-instance-default"},
			Timestamp: now(),
			Fields: mapstr.M{
				"asset": Asset{
					Id:   []string{"arn:aws:ec2:us-east::ec2/234567890"},
					Name: "test-server",
					AssetClassification: AssetClassification{
						Category:    CategoryInfrastructure,
						SubCategory: SubCategoryCompute,
						Type:        TypeVirtualMachine,
						SubType:     SubTypeEC2,
					},
					Tags: map[string]string{"Name": "test-server", "key": "value"},
				},
				"cloud": &AssetCloud{
					Provider: AwsCloudProvider,
					Region:   "us-east",
				},
				"host": &AssetHost{
					Architecture:    string(types.ArchitectureValuesX8664),
					ImageId:         pointers.Ref("image-id"),
					InstanceType:    "instance-type",
					Platform:        "linux",
					PlatformDetails: pointers.Ref("ubuntu"),
				},
				"network": &AssetNetwork{
					NetworkId:        pointers.Ref("vpc-id"),
					SubnetIds:        []string{"subnetId"},
					Ipv6Address:      pointers.Ref("ipv6"),
					PublicIpAddress:  pointers.Ref("public-ip-addr"),
					PrivateIpAddress: pointers.Ref("private-ip-addre"),
					PublicDnsName:    pointers.Ref("public-dns"),
					PrivateDnsName:   pointers.Ref("private-dns"),
				},
				"iam": &AssetIAM{
					Id:  pointers.Ref("a123123"),
					Arn: pointers.Ref("123123:123123:123123"),
				},
				"resource_policies": []AssetResourcePolicy{
					{
						Version:   pointers.Ref("2012-10-17"),
						Id:        pointers.Ref("Test 1"),
						Effect:    "Allow",
						Principal: map[string]any{"*": "*"},
						Action:    []string{"read"},
						Resource:  []string{"s3/bucket"},
					},
				},
				"related.entities": []string{"arn:aws:ec2:us-east::ec2/234567890"},
			},
		},
	}

	publishedCh := make(chan []beat.Event)
	publisher := NewMockAssetPublisher(t)
	publisher.EXPECT().PublishAll(mock.Anything).Run(func(e []beat.Event) {
		publishedCh <- e
	})

	fetcher := NewMockAssetFetcher(t)
	fetcher.EXPECT().Fetch(mock.Anything, mock.Anything).Run(func(_ context.Context, assetChannel chan<- AssetEvent) {
		assetChannel <- NewAssetEvent(
			AssetClassification{
				Category:    CategoryInfrastructure,
				SubCategory: SubCategoryCompute,
				Type:        TypeVirtualMachine,
				SubType:     SubTypeEC2,
			},
			[]string{"arn:aws:ec2:us-east::ec2/234567890"},
			"test-server",
			WithTags(map[string]string{"Name": "test-server", "key": "value"}),
			WithCloud(AssetCloud{
				Provider: AwsCloudProvider,
				Region:   "us-east",
			}),
			WithHost(AssetHost{
				Architecture:    string(types.ArchitectureValuesX8664),
				ImageId:         pointers.Ref("image-id"),
				InstanceType:    "instance-type",
				Platform:        "linux",
				PlatformDetails: pointers.Ref("ubuntu"),
			}),
			WithIAM(AssetIAM{
				Id:  pointers.Ref("a123123"),
				Arn: pointers.Ref("123123:123123:123123"),
			}),
			WithNetwork(AssetNetwork{
				NetworkId:        pointers.Ref("vpc-id"),
				SubnetIds:        []string{"subnetId"},
				Ipv6Address:      pointers.Ref("ipv6"),
				PublicIpAddress:  pointers.Ref("public-ip-addr"),
				PrivateIpAddress: pointers.Ref("private-ip-addre"),
				PublicDnsName:    pointers.Ref("public-dns"),
				PrivateDnsName:   pointers.Ref("private-dns"),
			}),
			WithResourcePolicies(AssetResourcePolicy{
				Version:   pointers.Ref("2012-10-17"),
				Id:        pointers.Ref("Test 1"),
				Effect:    "Allow",
				Principal: map[string]any{"*": "*"},
				Action:    []string{"read"},
				Resource:  []string{"s3/bucket"},
			}),
		)
	})

	logger := logp.NewLogger("test_run")
	inventory := AssetInventory{
		logger:              logger,
		fetchers:            []AssetFetcher{fetcher},
		publisher:           publisher,
		bufferFlushInterval: 10 * time.Millisecond,
		bufferMaxSize:       1,
		assetCh:             make(chan AssetEvent),
		now:                 now,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	go func() {
		inventory.Run(ctx)
	}()

	select {
	case <-ctx.Done():
		t.Errorf("Context done without receiving any events")
	case received := <-publishedCh:
		inventory.Stop()
		assert.ElementsMatch(t, received, expected)
	}
}
