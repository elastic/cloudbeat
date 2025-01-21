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
	"fmt"

	"github.com/elastic/cloudbeat/internal/inventory"
	azurelib "github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
)

type storageFetcher struct {
	logger   *clog.Logger
	provider storageProvider
}

type (
	storageProviderFunc func(context.Context, []azurelib.AzureAsset) ([]azurelib.AzureAsset, error)
	storageProvider     interface {
		ListSubscriptions(ctx context.Context) ([]azurelib.AzureAsset, error)
		ListStorageAccountBlobServices(ctx context.Context, storageAccounts []azurelib.AzureAsset) ([]azurelib.AzureAsset, error)
		ListStorageAccountQueues(ctx context.Context, storageAccounts []azurelib.AzureAsset) ([]azurelib.AzureAsset, error)
		ListStorageAccountQueueServices(ctx context.Context, storageAccounts []azurelib.AzureAsset) ([]azurelib.AzureAsset, error)
		ListStorageAccounts(ctx context.Context, storageAccountsSubscriptionsIds []string) ([]azurelib.AzureAsset, error)
	}
)

func newStorageFetcher(logger *clog.Logger, provider storageProvider) inventory.AssetFetcher {
	return &storageFetcher{
		logger:   logger,
		provider: provider,
	}
}

func (f *storageFetcher) Fetch(ctx context.Context, assetChan chan<- inventory.AssetEvent) {
	resourcesToFetch := []struct {
		name           string
		function       storageProviderFunc
		classification inventory.AssetClassification
	}{
		{"Storage Blob Services", f.provider.ListStorageAccountBlobServices, inventory.AssetClassificationAzureStorageBlobService},
		{"Storage Queue Services", f.provider.ListStorageAccountQueueServices, inventory.AssetClassificationAzureStorageQueueService},
		{"Storage Queues", f.provider.ListStorageAccountQueues, inventory.AssetClassificationAzureStorageQueue},
	}

	storageAccounts, err := f.listStorageAccounts(ctx)
	if err != nil {
		f.logger.Errorf("Could not fetch anything: %v", err)
		return
	}

	for _, r := range resourcesToFetch {
		f.fetch(ctx, storageAccounts, r.name, r.function, r.classification, assetChan)
	}
}

func (f *storageFetcher) listStorageAccounts(ctx context.Context) ([]azurelib.AzureAsset, error) {
	subscriptions, err := f.provider.ListSubscriptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing subscriptions: %v", err)
	}

	subscriptionIds := make([]string, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		subscriptionIds = append(subscriptionIds, subscription.Name)
	}

	storageAccounts, err := f.provider.ListStorageAccounts(ctx, subscriptionIds)
	if err != nil {
		return nil, fmt.Errorf("error listing storage accounts: %v", err)
	}

	return storageAccounts, nil
}

func (f *storageFetcher) fetch(ctx context.Context, storageAccounts []azurelib.AzureAsset, resourceName string, function storageProviderFunc, classification inventory.AssetClassification, assetChan chan<- inventory.AssetEvent) {
	f.logger.Infof("Fetching %s", resourceName)
	defer f.logger.Infof("Fetching %s - Finished", resourceName)

	azureAssets, err := function(ctx, storageAccounts)
	if err != nil {
		f.logger.Errorf("Could not fetch %s: %v", resourceName, err)
		return
	}

	for _, item := range azureAssets {
		assetChan <- inventory.NewAssetEvent(
			classification,
			item.Id,
			item.DisplayName,
			inventory.WithRawAsset(item),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				AccountID:   item.TenantId,
				ServiceName: "Azure",
			}),
		)
	}
}
