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
	"fmt"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	libevents "github.com/elastic/beats/v7/libbeat/beat/events"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/samber/lo"
)

const indexTemplate = "logs-cloud_asset_inventory.asset_inventory-%s_%s_%s_%s-default"

type AssetInventory struct {
	fetchers            []AssetFetcher
	publisher           AssetPublisher
	bufferFlushInterval time.Duration
	bufferMaxSize       int
	logger              *logp.Logger
	assetCh             chan AssetEvent
	now                 func() time.Time
}

type AssetFetcher interface {
	Fetch(ctx context.Context, assetChannel chan<- AssetEvent)
}

type AssetPublisher interface {
	PublishAll([]beat.Event)
}

func NewAssetInventory(logger *logp.Logger, fetchers []AssetFetcher, publisher AssetPublisher, now func() time.Time) AssetInventory {
	logger.Info("Initializing Asset Inventory POC")
	return AssetInventory{
		logger:    logger,
		fetchers:  fetchers,
		publisher: publisher,
		// move to a configuration parameter
		bufferFlushInterval: 10 * time.Second,
		bufferMaxSize:       1600,
		assetCh:             make(chan AssetEvent),
		now:                 now,
	}
}

func (a *AssetInventory) Run(ctx context.Context) {
	for _, fetcher := range a.fetchers {
		go func(fetcher AssetFetcher) {
			fetcher.Fetch(ctx, a.assetCh)
		}(fetcher)
	}

	assetsBuffer := make([]AssetEvent, 0, a.bufferMaxSize)
	flushTicker := time.NewTicker(a.bufferFlushInterval)
	for {
		select {
		case <-ctx.Done():
			a.logger.Warnf("Asset Inventory context is done: %v", ctx.Err())
			a.publish(assetsBuffer)
			return

		case <-flushTicker.C:
			if len(assetsBuffer) == 0 {
				a.logger.Debugf("Interval reached without events")
				continue
			}

			a.logger.Infof("Asset Inventory buffer is being flushed (assets %d)", len(assetsBuffer))
			a.publish(assetsBuffer)
			assetsBuffer = assetsBuffer[:0] // clear keeping cap

		case assetToPublish := <-a.assetCh:
			assetsBuffer = append(assetsBuffer, assetToPublish)

			if len(assetsBuffer) == a.bufferMaxSize {
				a.logger.Infof("Asset Inventory buffer is being flushed (assets %d)", len(assetsBuffer))
				a.publish(assetsBuffer)
				assetsBuffer = assetsBuffer[:0] // clear keeping cap
			}
		}
	}
}

func (a *AssetInventory) publish(assets []AssetEvent) {
	events := lo.Map(assets, func(e AssetEvent, _ int) beat.Event {
		return beat.Event{
			Meta:      mapstr.M{libevents.FieldMetaIndex: generateIndex(e.Asset)},
			Timestamp: a.now(),
			Fields: mapstr.M{
				"asset":             e.Asset,
				"cloud":             e.Cloud,
				"host":              e.Host,
				"network":           e.Network,
				"iam":               e.IAM,
				"resource_policies": e.ResourcePolicies,
				"related.entities":  e.Asset.Id,
			},
		}
	})

	a.publisher.PublishAll(events)
}

func generateIndex(a Asset) string {
	return fmt.Sprintf(indexTemplate, a.Category, a.SubCategory, a.Type, a.SubType)
}

func (a *AssetInventory) Stop() {
	close(a.assetCh)
}

func removeEmpty(list []string) []string {
	return lo.Filter(list, func(item string, _ int) bool { return item != "" })
}
