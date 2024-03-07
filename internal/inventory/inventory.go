package inventory

import (
	"context"
	"fmt"
	"github.com/elastic/beats/v7/libbeat/beat"
	libevents "github.com/elastic/beats/v7/libbeat/beat/events"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/samber/lo"
	"time"
)

type AssetInventory struct {
	fetchers            []AssetFetcher
	publisher           AssetPublisher
	bufferFlushInterval time.Duration
	bufferMaxSize       int
	logger              *logp.Logger
	assetCh             chan Asset
}

type AssetFetcher interface {
	Fetch(ctx context.Context, assetChannel chan<- Asset)
}

type AssetPublisher interface {
	PublishAll([]beat.Event)
}

func NewAssetInventory(logger *logp.Logger, fetchers []AssetFetcher, publisher AssetPublisher) AssetInventory {
	logger.Info("Initializing Asset Inventory POC")
	return AssetInventory{
		logger:    logger,
		fetchers:  fetchers,
		publisher: publisher,
		// move to a configuration parameter
		bufferFlushInterval: 15 * time.Second,
		bufferMaxSize:       50,
		assetCh:             make(chan Asset),
	}
}

func (a *AssetInventory) Run(ctx context.Context) {
	for _, fetcher := range a.fetchers {
		go func(fetcher AssetFetcher) {
			fetcher.Fetch(ctx, a.assetCh)
		}(fetcher)
	}

	assetsBuffer := make([]Asset, 0, a.bufferMaxSize)
	flushTicker := time.NewTicker(a.bufferFlushInterval)
	for {
		select {
		case <-ctx.Done():
			a.logger.Warnf("Asset Inventory context is done: %v", ctx.Err())
			a.publish(assetsBuffer)
			return

		case <-flushTicker.C:
			if len(assetsBuffer) == 0 {
				a.logger.Infof("Interval reached without events")
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

func (a *AssetInventory) publish(assets []Asset) {
	events := lo.Map(assets, func(a Asset, _ int) beat.Event {
		return beat.Event{
			Meta:      mapstr.M{libevents.FieldMetaIndex: generateIndex(a)},
			Timestamp: time.Now(),
			Fields: mapstr.M{
				"asset": a,
			},
		}
	})

	a.publisher.PublishAll(events)
}

func generateIndex(a Asset) string {
	return fmt.Sprintf("asset_inventory_%s_%s_%s_%s", a.Category, a.SubCategory, a.Type, a.SubStype)
}

func (a *AssetInventory) Stop() {
	close(a.assetCh)
}
