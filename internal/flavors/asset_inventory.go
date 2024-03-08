package flavors

import (
	"context"
	"fmt"
	"github.com/elastic/beats/v7/libbeat/beat"
	awsbeat "github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/inventory"
	awsinventory "github.com/elastic/cloudbeat/internal/inventory/aws"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"time"
)

type assetInventory struct {
	flavorBase
	assetInventory inventory.AssetInventory
}

func NewAssetInventory(b *beat.Beat, agentConfig *agentconfig.C) (beat.Beater, error) {
	cfg, err := config.New(agentConfig)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	return newAssetInventoryFromCfg(b, cfg)
}

func newAssetInventoryFromCfg(b *beat.Beat, cfg *config.Config) (*assetInventory, error) {
	logger := logp.NewLogger("asset_inventory")
	ctx, cancel := context.WithCancel(context.Background())

	logger.Info("Creating AWS AssetInventory")

	awsFetchers, err := initAwsFetchers(ctx, cfg, logger)
	if err != nil {
		cancel()
		return nil, err
	}

	publisherClient, err := NewClient(b.Publisher, cfg.Processors)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to init client: %w", err)
	}

	now := func() time.Time { return time.Now() }
	newAssetInventory := inventory.NewAssetInventory(logger, awsFetchers, publisherClient, now)
	if err != nil {
		cancel()
		return nil, err
	}

	publisher := NewPublisher(logger, flushInterval, eventsThreshold, publisherClient)

	return &assetInventory{
		flavorBase: flavorBase{
			ctx:       ctx,
			cancel:    cancel,
			publisher: publisher,
			config:    cfg,
			log:       logger,
		},
		assetInventory: newAssetInventory,
	}, nil
}

func initAwsFetchers(ctx context.Context, cfg *config.Config, logger *logp.Logger) ([]inventory.AssetFetcher, error) {
	awsConfig, err := awsbeat.InitializeAWSConfig(cfg.CloudConfig.Aws.Cred)
	if err != nil {
		return nil, err
	}

	idProvider := awslib.IdentityProvider{}
	awsIdentity, err := idProvider.GetIdentity(ctx, awsConfig)
	if err != nil {
		return nil, err
	}

	return awsinventory.Fetchers(logger, awsIdentity, awsConfig), nil
}

func (bt *assetInventory) Run(*beat.Beat) error {
	bt.log.Info("Asset Inventory is running! Hit CTRL-C to stop it")
	bt.assetInventory.Run(bt.ctx)
	bt.log.Warn("Asset Inventory has finished running")
	return nil
}

func (bt *assetInventory) Stop() {
	bt.assetInventory.Stop()

	if err := bt.client.Close(); err != nil {
		bt.log.Fatal("Cannot close client", err)
	}

	bt.cancel()
}
