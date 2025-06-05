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
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elastic/beats/v7/libbeat/beat"
	libevents "github.com/elastic/beats/v7/libbeat/beat/events"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestAssetInventory_Run(t *testing.T) {
	var emptyRef *any
	now := func() time.Time { return time.Date(2024, 1, 1, 1, 1, 1, 0, time.Local) }
	expected := []beat.Event{
		{
			Meta:      mapstr.M{libevents.FieldMetaIndex: "logs-cloud_asset_inventory.asset_inventory-infrastructure_virtual_machine-default"},
			Timestamp: now(),
			Fields: mapstr.M{
				"entity": Entity{
					Id:   "arn:aws:ec2:us-east::ec2/234567890",
					Name: "test-server",
					AssetClassification: AssetClassification{
						Category: CategoryInfrastructure,
						Type:     "Virtual Machine",
					},
				},
				"event": Event{
					Kind: "asset",
				},
				"labels": map[string]string{"Name": "test-server", "key": "value"},
				"cloud": &Cloud{
					Provider: AwsCloudProvider,
					Region:   "us-east",
				},
				"host": &Host{
					Architecture: string(types.ArchitectureValuesX8664),
					Type:         "instance-type",
					ID:           "i-a2",
				},
				"network": &Network{
					Name: "vpc-id",
				},
				"user": &User{
					ID:   "a123123",
					Name: "name",
				},
				"related.entity": []string{"arn:aws:ec2:us-east::ec2/234567890"},
				"tags":           []string{"foo", "bar"},
				"orchestrator": &Orchestrator{
					Type: "kubernetes",
				},
				"container": &Container{},
				"organization": &Organization{
					ID: "org-id",
				},
				"fass": &Fass{
					Name: "fass",
				},
				"url": &URL{
					Full: "https://example.com",
				},
				"Attributes": emptyRef,
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
				Category: CategoryInfrastructure,
				Type:     "Virtual Machine",
			},
			"arn:aws:ec2:us-east::ec2/234567890",
			"test-server",
			WithLabels(map[string]string{"Name": "test-server", "key": "value"}),
			WithCloud(Cloud{
				Provider: AwsCloudProvider,
				Region:   "us-east",
			}),
			WithHost(Host{
				Architecture: string(types.ArchitectureValuesX8664),
				Type:         "instance-type",
				ID:           "i-a2",
			}),
			WithUser(User{
				ID:   "a123123",
				Name: "name",
			}),
			WithNetwork(Network{
				Name: "vpc-id",
			}),
			WithContainer(Container{}),
			WithOrchestrator(Orchestrator{
				Type: "kubernetes",
			}),
			WithOrganization(Organization{
				ID: "org-id",
			}),
			WithFass(Fass{
				Name: "fass",
			}),
			WithURL(URL{
				Full: "https://example.com",
			}),
			WithTags([]string{"foo", "bar"}),
		)
	})

	logger := clog.NewLogger("test_run")
	inventory := AssetInventory{
		logger:              logger,
		fetchers:            []AssetFetcher{fetcher},
		publisher:           publisher,
		bufferFlushInterval: 10 * time.Millisecond,
		bufferMaxSize:       1,
		period:              24 * time.Hour,
		assetCh:             make(chan AssetEvent),
		now:                 now,
	}

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
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

func TestAssetInventory_Period(t *testing.T) {
	testhelper.SkipLong(t)
	now := func() time.Time { return time.Date(2024, 1, 1, 1, 1, 1, 0, time.Local) }

	var cycleCounter int64

	publisher := NewMockAssetPublisher(t)
	publisher.EXPECT().PublishAll(mock.Anything).Maybe()

	fetcher := NewMockAssetFetcher(t)
	fetcher.EXPECT().Fetch(mock.Anything, mock.Anything).Run(func(_ context.Context, _ chan<- AssetEvent) {
		atomic.AddInt64(&cycleCounter, 1)
	})

	logger := clog.NewLogger("test_run")
	inventory := AssetInventory{
		logger:              logger,
		fetchers:            []AssetFetcher{fetcher},
		publisher:           publisher,
		bufferFlushInterval: 10 * time.Millisecond,
		bufferMaxSize:       1,
		period:              500 * time.Millisecond,
		assetCh:             make(chan AssetEvent),
		now:                 now,
	}

	// Run it enough for 2 cycles to finish; one starts immediately, the other after 500 milliseconds
	ctx, cancel := context.WithTimeout(t.Context(), 600*time.Millisecond)
	defer cancel()

	go func() {
		inventory.Run(ctx)
	}()

	<-ctx.Done()
	val := atomic.LoadInt64(&cycleCounter)
	assert.Equal(t, int64(2), val, "Expected to run 2 cycles, got %d", val)
}

func TestAssetInventory_RunAllFetchersOnce(t *testing.T) {
	now := func() time.Time { return time.Date(2024, 1, 1, 1, 1, 1, 0, time.Local) }
	publisher := NewMockAssetPublisher(t)
	publisher.EXPECT().PublishAll(mock.Anything).Maybe()

	fetchers := []AssetFetcher{}
	fetcherCounters := [](*int64){}
	for i := 0; i < 5; i++ {
		fetcher := NewMockAssetFetcher(t)
		counter := int64(0)
		fetcher.EXPECT().Fetch(mock.Anything, mock.Anything).Run(func(_ context.Context, _ chan<- AssetEvent) {
			atomic.AddInt64(&counter, 1)
		})
		fetchers = append(fetchers, fetcher)
		fetcherCounters = append(fetcherCounters, &counter)
	}

	logger := clog.NewLogger("test_run")
	inventory := AssetInventory{
		logger:              logger,
		fetchers:            fetchers,
		publisher:           publisher,
		bufferFlushInterval: 10 * time.Millisecond,
		bufferMaxSize:       1,
		period:              24 * time.Hour,
		assetCh:             make(chan AssetEvent),
		now:                 now,
	}

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	inventory.runAllFetchersOnce(ctx)
	<-ctx.Done()

	// Check that EVERY fetcher has been called EXACTLY ONCE
	for _, counter := range fetcherCounters {
		val := atomic.LoadInt64(counter)
		assert.Equal(t, int64(1), val, "Expected to run once, got %d", val)
	}
}
