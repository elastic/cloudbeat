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

package azurefetcher

import (
	"context"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/inventory"
	azurelib "github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

type accountFetcher struct {
	logger   *clog.Logger
	tenantID string //nolint:unused
	provider accountProvider
}

type (
	accountProviderFunc func(context.Context) ([]azurelib.AzureAsset, error)
	accountProvider     interface {
		ListTenants(ctx context.Context) ([]azurelib.AzureAsset, error)
		ListSubscriptions(ctx context.Context) ([]azurelib.AzureAsset, error)
	}
)

func newAccountFetcher(logger *clog.Logger, tenantID string, provider accountProvider) inventory.AssetFetcher {
	return &accountFetcher{
		logger:   logger,
		tenantID: tenantID,
		provider: provider,
	}
}

func (f *accountFetcher) Fetch(ctx context.Context, assetChan chan<- inventory.AssetEvent) {
	resourcesToFetch := []struct {
		name           string
		function       accountProviderFunc
		classification inventory.AssetClassification
	}{
		{"Tenants", f.provider.ListTenants, inventory.AssetClassificationAzureTenant},
		{"Subscriptions", f.provider.ListSubscriptions, inventory.AssetClassificationAzureSubscription},
	}
	for _, r := range resourcesToFetch {
		f.fetch(ctx, r.name, r.function, r.classification, assetChan)
	}
}

func (f *accountFetcher) fetch(ctx context.Context, resourceName string, function accountProviderFunc, classification inventory.AssetClassification, assetChan chan<- inventory.AssetEvent) {
	f.logger.Infof("Fetching %s", resourceName)
	defer f.logger.Infof("Fetching %s - Finished", resourceName)

	azureAssets, err := function(ctx)
	if err != nil {
		f.logger.Errorf("Could not fetch %s: %v", resourceName, err)
		return
	}

	for _, item := range azureAssets {
		assetChan <- inventory.NewAssetEvent(
			classification,
			item.Id,
			pickName(item.DisplayName, item.Name, item.Id),
			inventory.WithRawAsset(item),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				AccountID:   item.TenantId,
				ServiceName: "Azure",
			}),
			inventory.WithLabelsFromAny(item.Tags),
			inventory.WithOrganization(inventory.Organization{
				ID: item.TenantId,
			}),
		)
	}
}
