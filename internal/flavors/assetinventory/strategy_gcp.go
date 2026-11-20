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
//go:build !requirefips

package assetinventory

import (
	"context"
	"fmt"

	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/gcpfetcher"
	gcp_auth "github.com/elastic/cloudbeat/internal/resources/providers/gcplib/auth"
	gcp_inventory "github.com/elastic/cloudbeat/internal/resources/providers/gcplib/inventory"
)

func (s *strategy) initGcpFetchers(ctx context.Context) ([]inventory.AssetFetcher, error) {
	cfgProvider := gcp_auth.NewConfigProvider()
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
