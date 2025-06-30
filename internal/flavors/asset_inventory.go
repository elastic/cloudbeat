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

	"github.com/elastic/beats/v7/libbeat/beat"
	agentconfig "github.com/elastic/elastic-agent-libs/config"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/flavors/assetinventory"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/inventory"
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
	logger := clog.NewLogger("asset_inventory")
	ctx, cancel := context.WithCancel(context.Background())

	beatClient, err := NewClient(b.Publisher, cfg.Processors)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to init client: %w", err)
	}

	strategy := assetinventory.GetStrategy(logger, cfg)
	newAssetInventory, err := strategy.NewAssetInventory(ctx, beatClient)
	if err != nil {
		cancel()
		return nil, err
	}

	publisher := NewPublisher(ctx, logger, flushInterval, eventsThreshold, beatClient)
	return &assetInventory{
		flavorBase: flavorBase{
			ctx:       ctx,
			client:    beatClient,
			cancel:    cancel,
			publisher: publisher,
			config:    cfg,
			log:       logger,
		},
		assetInventory: newAssetInventory,
	}, nil
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
