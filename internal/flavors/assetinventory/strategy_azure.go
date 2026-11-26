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
	"github.com/elastic/cloudbeat/internal/inventory/azurefetcher"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	azure_auth "github.com/elastic/cloudbeat/internal/resources/providers/azurelib/auth"
	"github.com/elastic/cloudbeat/internal/resources/providers/msgraph"
)

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
