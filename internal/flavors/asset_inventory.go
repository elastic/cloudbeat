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
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/awsfetcher"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
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

	now := func() time.Time { return time.Now() } //nolint:gocritic
	newAssetInventory := inventory.NewAssetInventory(logger, awsFetchers, publisherClient, now)
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
