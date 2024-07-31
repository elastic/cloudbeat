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

package flavors

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/awsfetcher"
	"github.com/elastic/cloudbeat/internal/inventory/azurefetcher"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	azure_auth "github.com/elastic/cloudbeat/internal/resources/providers/azurelib/auth"
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

	var fetchers []inventory.AssetFetcher
	var err error

	switch cfg.AssetInventoryProvider {
	case config.ProviderAWS:
		fetchers, err = initAwsFetchers(ctx, cfg, logger)
	case config.ProviderAzure:
		fetchers, err = initAzureFetchers(ctx, cfg, logger)
	case config.ProviderGCP:
		err = fmt.Errorf("GCP branch not implemented")
	default:
		err = fmt.Errorf("unsupported Asset Inventory provider %q", cfg.AssetInventoryProvider)
	}
	if err != nil {
		cancel()
		return nil, err
	}
	logger.Infof("Creating %s AssetInventory", strings.ToUpper(cfg.AssetInventoryProvider))

	publisherClient, err := NewClient(b.Publisher, cfg.Processors)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to init client: %w", err)
	}

	now := func() time.Time { return time.Now() } //nolint:gocritic
	newAssetInventory := inventory.NewAssetInventory(logger, fetchers, publisherClient, now)
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
	awsConfig, err := awslib.InitializeAWSConfig(cfg.CloudConfig.Aws.Cred)
	if err != nil {
		return nil, err
	}

	idProvider := awslib.IdentityProvider{Logger: logger}
	awsIdentity, err := idProvider.GetIdentity(ctx, *awsConfig)
	if err != nil {
		return nil, err
	}

	return awsfetcher.New(logger, awsIdentity, *awsConfig), nil
}

func initAzureFetchers(_ context.Context, cfg *config.Config, logger *logp.Logger) ([]inventory.AssetFetcher, error) {
	cfgProvider := &azure_auth.ConfigProvider{AuthProvider: &azure_auth.AzureAuthProvider{}}
	azureConfig, err := cfgProvider.GetAzureClientConfig(cfg.CloudConfig.Azure)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize azure config: %w", err)
	}
	initializer := &azurelib.ProviderInitializer{}
	provider, err := initializer.Init(logger, *azureConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize azure config: %w", err)
	}

	return azurefetcher.New(logger, provider, azureConfig), nil
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
