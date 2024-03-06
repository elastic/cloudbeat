package inventory

import (
	"context"
	"github.com/elastic/elastic-agent-libs/logp"
	"time"
)

type AssetInventory struct {
	fetchers            []AssetFetcher
	publisher           assetInventoryPublisher
	bufferFlushInterval time.Duration
	bufferMaxSize       int
	logger              *logp.Logger
}

type AssetFetcher interface {
	Fetch(ctx context.Context, assetChannel chan<- Asset)
}

func NewAssetInventory(logger *logp.Logger, fetchers []AssetFetcher) AssetInventory {
	logger.Info("Initializing Asset Inventory POC")
	return AssetInventory{
		logger:              logger,
		fetchers:            fetchers,
		publisher:           newPublisher(logger),
		bufferFlushInterval: 15 * time.Second,
		bufferMaxSize:       50,
	}
}

func (a *AssetInventory) BuildInventory(ctx context.Context) {
	ch := make(chan Asset)
	for _, fetcher := range a.fetchers {
		go func(fetcher AssetFetcher) {
			fetcher.Fetch(ctx, ch)
		}(fetcher)
	}

	assetsBuffer := make([]Asset, 0, a.bufferMaxSize)
	flushTicker := time.NewTicker(a.bufferFlushInterval)
	for {
		select {
		case <-ctx.Done():
			a.logger.Warnf("Asset Inventory context is done: %v", ctx.Err())
			a.publisher.publish(assetsBuffer)
			return

		case <-flushTicker.C:
			if len(assetsBuffer) == 0 {
				a.logger.Infof("Interval reached without events")
				continue
			}

			a.logger.Infof("Asset Inventory buffer is being flushed (assets %d)", len(assetsBuffer))
			a.publisher.publish(assetsBuffer)
			assetsBuffer = assetsBuffer[:0] // clear keeping cap

		case assetToPublish := <-ch:
			assetsBuffer = append(assetsBuffer, assetToPublish)

			if len(assetsBuffer) == a.bufferMaxSize {
				a.logger.Infof("Asset Inventory buffer is being flushed (assets %d)", len(assetsBuffer))
				a.publisher.publish(assetsBuffer)
				assetsBuffer = assetsBuffer[:0] // clear keeping cap
			}
		}
	}
}
