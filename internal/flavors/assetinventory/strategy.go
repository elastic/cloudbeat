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
	"fmt"
	"strings"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/awsfetcher"
	"github.com/elastic/cloudbeat/internal/inventory/azurefetcher"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	azure_auth "github.com/elastic/cloudbeat/internal/resources/providers/azurelib/auth"
)

type Strategy interface {
	NewAssetInventory(ctx context.Context, client beat.Client) (inventory.AssetInventory, error)
}

type strategy struct {
	logger *logp.Logger
	cfg    *config.Config
}

func (s *strategy) NewAssetInventory(ctx context.Context, client beat.Client) (inventory.AssetInventory, error) {
	var fetchers []inventory.AssetFetcher
	var err error

	switch s.cfg.AssetInventoryProvider {
	case config.ProviderAWS:
		fetchers, err = s.initAwsFetchers(ctx)
	case config.ProviderAzure:
		fetchers, err = s.initAzureFetchers(ctx)
	case config.ProviderGCP:
		err = fmt.Errorf("GCP branch not implemented")
	case "":
		err = fmt.Errorf("missing config.v1.asset_inventory_provider setting")
	default:
		err = fmt.Errorf("unsupported Asset Inventory provider %q", s.cfg.AssetInventoryProvider)
	}
	if err != nil {
		return inventory.AssetInventory{}, err
	}
	s.logger.Infof("Creating %s AssetInventory", strings.ToUpper(s.cfg.AssetInventoryProvider))

	now := func() time.Time { return time.Now() } //nolint:gocritic
	return inventory.NewAssetInventory(s.logger, fetchers, client, now), nil
}

func (s *strategy) initAwsFetchers(ctx context.Context) ([]inventory.AssetFetcher, error) {
	awsConfig, err := awslib.InitializeAWSConfig(s.cfg.CloudConfig.Aws.Cred)
	if err != nil {
		return nil, err
	}

	idProvider := awslib.IdentityProvider{Logger: s.logger}
	awsIdentity, err := idProvider.GetIdentity(ctx, *awsConfig)
	if err != nil {
		return nil, err
	}

	return awsfetcher.New(s.logger, awsIdentity, *awsConfig), nil
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
		return nil, fmt.Errorf("failed to initialize azure config: %w", err)
	}

	return azurefetcher.New(s.logger, provider, azureConfig), nil
}

func GetStrategy(logger *logp.Logger, cfg *config.Config) Strategy {
	return &strategy{
		logger: logger,
		cfg:    cfg,
	}
}
