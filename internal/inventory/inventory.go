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
	"strings"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	libevents "github.com/elastic/beats/v7/libbeat/beat/events"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/samber/lo"
)

const (
	indexTemplate = "logs-cloud_asset_inventory.asset_inventory-%s_%s-default"
	minimalPeriod = 30 * time.Second
)

type AssetInventory struct {
	fetchers            []AssetFetcher
	publisher           AssetPublisher
	bufferFlushInterval time.Duration
	bufferMaxSize       int
	period              time.Duration
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

func NewAssetInventory(logger *logp.Logger, fetchers []AssetFetcher, publisher AssetPublisher, now func() time.Time, period time.Duration) AssetInventory {
	if period < minimalPeriod {
		period = minimalPeriod
	}
	logger.Infof("Initializing Asset Inventory POC with period of %s", period)
	return AssetInventory{
		logger:    logger,
		fetchers:  fetchers,
		publisher: publisher,
		// move to a configuration parameter
		bufferFlushInterval: 10 * time.Second,
		bufferMaxSize:       1600,
		period:              period,
		assetCh:             make(chan AssetEvent),
		now:                 now,
	}
}

func (a *AssetInventory) Run(ctx context.Context) {
	a.runAllFetchersOnce(ctx)

	assetsBuffer := make([]AssetEvent, 0, a.bufferMaxSize)
	flushTicker := time.NewTicker(a.bufferFlushInterval)
	fetcherPeriod := time.NewTicker(a.period)
	for {
		select {
		case <-ctx.Done():
			a.logger.Warnf("Asset Inventory context is done: %v", ctx.Err())
			a.publish(assetsBuffer)
			return

		case <-fetcherPeriod.C:
			a.runAllFetchersOnce(ctx)

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

// runAllFetchersOnce runs every fetcher to collect assets to assetCh ONCE. It
// should be called every cycle, once every `a.period`.
func (a *AssetInventory) runAllFetchersOnce(ctx context.Context) {
	a.logger.Debug("Running all fetchers once")
	for _, fetcher := range a.fetchers {
		go func(fetcher AssetFetcher) {
			fetcher.Fetch(ctx, a.assetCh)
		}(fetcher)
	}
}

func (a *AssetInventory) publish(assets []AssetEvent) {
	events := lo.Map(assets, func(e AssetEvent, _ int) beat.Event {
		var relatedEntity []string
		relatedEntity = append(relatedEntity, e.Entity.Id)
		if len(e.Entity.relatedEntityId) > 0 {
			relatedEntity = append(relatedEntity, e.Entity.relatedEntityId...)
		}
		return beat.Event{
			Meta:      mapstr.M{libevents.FieldMetaIndex: generateIndex(e.Entity)},
			Timestamp: a.now(),
			Fields: mapstr.M{
				"entity":         e.Entity,
				"cloud":          e.Cloud,
				"host":           e.Host,
				"network":        e.Network,
				"user":           e.User,
				"Attributes":     e.RawAttributes,
				"labels":         e.Labels,
				"related.entity": relatedEntity,
			},
		}
	})

	a.publisher.PublishAll(events)
}

func generateIndex(a Entity) string {
	return fmt.Sprintf(indexTemplate, slugfy(string(a.Category)), slugfy(string(a.Type)))
}

func slugfy(s string) string {
	chunks := strings.Split(s, " ")
	clean := make([]string, len(chunks))
	for i, c := range chunks {
		clean[i] = strings.ToLower(c)
	}
	return strings.Join(clean, "_")
}

func (a *AssetInventory) Stop() {
	close(a.assetCh)
}

func removeEmpty(list []string) []string {
	return lo.Filter(list, func(item string, _ int) bool { return item != "" })
}
