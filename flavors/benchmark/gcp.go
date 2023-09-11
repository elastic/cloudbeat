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
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/preset"
	"github.com/elastic/cloudbeat/resources/fetching/registry"
	"github.com/elastic/cloudbeat/resources/providers/gcplib/auth"
	"github.com/elastic/cloudbeat/resources/providers/gcplib/inventory"
)

type GCP struct {
	CfgProvider          auth.ConfigProviderAPI
	inventoryInitializer inventory.ProviderInitializerAPI
}

func (g *GCP) Run(context.Context) error { return nil }

func (g *GCP) Initialize(ctx context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo) (registry.Registry, dataprovider.CommonDataProvider, dataprovider.IdProvider, error) {
	if err := g.checkDependencies(); err != nil {
		return nil, nil, nil, err
	}

	gcpConfig, err := g.CfgProvider.GetGcpClientConfig(ctx, cfg.CloudConfig.Gcp, log)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize gcp config: %w", err)
	}

	assetProvider, err := g.inventoryInitializer.Init(ctx, log, *gcpConfig)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize gcp asset inventory: %v", err)
	}

	fetchers, err := preset.NewCisGcpFactory(ctx, log, ch, assetProvider)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize gcp fetchers: %v", err)
	}

	return registry.NewRegistry(log, registry.WithFetchersMap(fetchers)),
		cloud.NewDataProvider(cloud.WithLogger(log)),
		cloud.NewIdProvider(),
		nil
}

func (g *GCP) Stop() {}

func (g *GCP) checkDependencies() error {
	if g.CfgProvider == nil {
		return errors.New("gcp config provider is uninitialized")
	}

	if g.inventoryInitializer == nil {
		return errors.New("gcp asset inventory is uninitialized")
	}
	return nil
}
