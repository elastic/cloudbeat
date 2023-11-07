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

package benchmark

import (
	"context"
	"errors"
	"fmt"

	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider"
	"github.com/elastic/cloudbeat/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/flavors/benchmark/builder"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/preset"
	"github.com/elastic/cloudbeat/resources/fetching/registry"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/auth"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
)

type Azure struct {
	CfgProvider          auth.ConfigProviderAPI
	inventoryInitializer inventory.ProviderInitializerAPI
}

func (a *Azure) Run(context.Context) error { return nil }

func (a *Azure) NewBenchmark(ctx context.Context, log *logp.Logger, cfg *config.Config) (builder.Benchmark, error) {
	resourceCh := make(chan fetching.ResourceInfo, resourceChBufferSize)
	reg, bdp, _, err := a.initialize(ctx, log, cfg, resourceCh)
	if err != nil {
		return nil, err
	}

	return builder.New(
		builder.WithBenchmarkDataProvider(bdp),
	).Build(ctx, log, cfg, resourceCh, reg)
}

func (a *Azure) initialize(ctx context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo) (registry.Registry, dataprovider.CommonDataProvider, dataprovider.IdProvider, error) {
	if err := a.checkDependencies(); err != nil {
		return nil, nil, nil, err
	}

	azureConfig, err := a.CfgProvider.GetAzureClientConfig(cfg.CloudConfig.Azure)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize azure config: %w", err)
	}

	assetProvider, err := a.inventoryInitializer.Init(ctx, log, *azureConfig)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize azure asset inventory: %v", err)
	}

	fetchers, err := preset.NewCisAzureFactory(log, ch, assetProvider)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize azure fetchers: %v", err)
	}

	return registry.NewRegistry(log, registry.WithFetchersMap(fetchers)),
		cloud.NewDataProvider(cloud.WithLogger(log)),
		nil,
		nil
}

func (a *Azure) checkDependencies() error {
	if a.CfgProvider == nil {
		return errors.New("azure config provider is uninitialized")
	}

	if a.inventoryInitializer == nil {
		return errors.New("azure asset inventory is uninitialized")
	}
	return nil
}
