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

package assetinventory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/azurefetcher"
	"github.com/elastic/cloudbeat/internal/inventory/gcpfetcher"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	azure_auth "github.com/elastic/cloudbeat/internal/resources/providers/azurelib/auth"
	gcp_auth "github.com/elastic/cloudbeat/internal/resources/providers/gcplib/auth"
	gcp_inventory "github.com/elastic/cloudbeat/internal/resources/providers/gcplib/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/msgraph"
)

type Strategy interface {
	NewAssetInventory(ctx context.Context, client beat.Client) (inventory.AssetInventory, error)
}

type strategy struct {
	logger *clog.Logger
	cfg    *config.Config
}

func (s *strategy) NewAssetInventory(ctx context.Context, client beat.Client) (inventory.AssetInventory, error) {
	var fetchers []inventory.AssetFetcher
	var err error

	switch s.cfg.AssetInventoryProvider {
	case config.ProviderAWS:
		switch s.cfg.CloudConfig.Aws.AccountType {
		case config.SingleAccount, config.OrganizationAccount:
			fetchers, err = s.initAwsFetchers(ctx)
		default:
			err = fmt.Errorf("unsupported account_type: %q", s.cfg.CloudConfig.Aws.AccountType)
		}
	case config.ProviderAzure:
		fetchers, err = s.initAzureFetchers(ctx)
	case config.ProviderGCP:
		fetchers, err = s.initGcpFetchers(ctx)
	case "":
		err = errors.New("missing config.v1.asset_inventory_provider setting")
	default:
		err = fmt.Errorf("unsupported Asset Inventory provider %q", s.cfg.AssetInventoryProvider)
	}
	if err != nil {
		return inventory.AssetInventory{}, err
	}
	s.logger.Infof("Creating %s AssetInventory", strings.ToUpper(s.cfg.AssetInventoryProvider))

	now := func() time.Time { return time.Now() } //nolint:gocritic
	return inventory.NewAssetInventory(s.logger, fetchers, client, now, s.cfg.Period), nil
}

func (s *strategy) initAzureFetchers(_ context.Context) ([]inventory.AssetFetcher, error) {
	cfgProvider := &azure_auth.ConfigProvider{AuthProvider: &azure_auth.AzureAuthProvider{}}
	azureConfig, err := cfgProvider.GetAzureClientConfig(s.cfg.CloudConfig.Azure)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize azure config: %w", err)
	}
	initializer := &azurelib.ProviderInitializer{}
	provider, err := initializer.Init(s.logger, *azureConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize azure provider: %w", err)
	}

	msgraphProvider, err := msgraph.NewProvider(s.logger, *azureConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize azure msgraph provider: %w", err)
	}

	return azurefetcher.New(s.logger, s.cfg.CloudConfig.Azure.Credentials.TenantID, provider, msgraphProvider), nil
}

func (s *strategy) initGcpFetchers(ctx context.Context) ([]inventory.AssetFetcher, error) {
	cfgProvider := &gcp_auth.ConfigProvider{AuthProvider: &gcp_auth.GoogleAuthProvider{}}
	gcpConfig, err := cfgProvider.GetGcpClientConfig(ctx, s.cfg.CloudConfig.Gcp, s.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gcp config: %w", err)
	}
	inventoryInitializer := &gcp_inventory.ProviderInitializer{}
	provider, err := inventoryInitializer.Init(ctx, s.logger, *gcpConfig, s.cfg.CloudConfig.Gcp)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gcp asset inventory: %v", err)
	}
	return gcpfetcher.New(s.logger, provider), nil
}

func GetStrategy(logger *clog.Logger, cfg *config.Config) Strategy {
	return &strategy{
		logger: logger,
		cfg:    cfg,
	}
}
