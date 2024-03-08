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

	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type mockAssetPublisher struct {
	mockF func(events []beat.Event)
}

func (m mockAssetPublisher) PublishAll(events []beat.Event) {
	m.mockF(events)
}

type mockAssetFetcher struct {
	eventsToPublish []AssetEvent
}

func (m mockAssetFetcher) Fetch(_ context.Context, assetChannel chan<- AssetEvent) {
	for _, e := range m.eventsToPublish {
		assetChannel <- e
	}
}

func TestAssetInventory_Run(t *testing.T) {
	now := func() time.Time { return time.Date(2024, 1, 1, 1, 1, 1, 0, time.Local) }
	expected := []beat.Event{
		{
			Meta:      mapstr.M{libevents.FieldMetaIndex: "asset_inventory_infrastructure_compute_virtual-machine_ec2"},
			Timestamp: now(),
			Fields: mapstr.M{
				"asset": Asset{
					UUID: "WH25UKB5ExLpkCAwRJDStPN3U+VGaYg9bF5qKu4L7Ro=",
					Id:   "arn:aws:ec2:us-east::ec2/234567890",
					Name: "test-server",
					AssetClassification: AssetClassification{
						Category:    CategoryInfrastructure,
						SubCategory: SubCategoryCompute,
						Type:        TypeVirtualMachine,
						SubStype:    SubTypeEC2,
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
					SubnetId:         pointers.Ref("subnetId"),
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
			},
		},
	}

	received := make([]beat.Event, 0, 1)
	done := make(chan bool)
	publisher := mockAssetPublisher{
		mockF: func(e []beat.Event) {
			received = append(received, e...)
			done <- true
		},
	}

	fetchers := []AssetFetcher{
		mockAssetFetcher{
			eventsToPublish: []AssetEvent{
				NewAssetEvent(
					AssetClassification{
						Category:    CategoryInfrastructure,
						SubCategory: SubCategoryCompute,
						Type:        TypeVirtualMachine,
						SubStype:    SubTypeEC2,
					},
					"arn:aws:ec2:us-east::ec2/234567890",
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
						SubnetId:         pointers.Ref("subnetId"),
						Ipv6Address:      pointers.Ref("ipv6"),
						PublicIpAddress:  pointers.Ref("public-ip-addr"),
						PrivateIpAddress: pointers.Ref("private-ip-addre"),
						PublicDnsName:    pointers.Ref("public-dns"),
						PrivateDnsName:   pointers.Ref("private-dns"),
					}),
				),
			},
		},
	}

	logger := logp.NewLogger("test_run")
	inventory := AssetInventory{
		logger:              logger,
		fetchers:            fetchers,
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
	case <-done:
		inventory.Stop()
		assert.ElementsMatch(t, received, expected)
	}
}
