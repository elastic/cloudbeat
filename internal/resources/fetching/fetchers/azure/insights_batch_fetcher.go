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

package fetchers

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
)

type AzureInsightsBatchAssetFetcher struct {
	log        *clog.Logger
	resourceCh chan fetching.ResourceInfo
	provider   azurelib.ProviderAPI
}

func NewAzureInsightsBatchAssetFetcher(log *clog.Logger, ch chan fetching.ResourceInfo, provider azurelib.ProviderAPI) *AzureInsightsBatchAssetFetcher {
	return &AzureInsightsBatchAssetFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

func (f *AzureInsightsBatchAssetFetcher) Fetch(ctx context.Context, cycleMetadata cycle.Metadata) error {
	f.log.Info("Starting AzureInsightsBatchAssetFetcher.Fetch")
	subscriptions, err := f.provider.GetSubscriptions(ctx, cycleMetadata)
	if err != nil {
		return fmt.Errorf("azure insights fetcher: error while receiving subscriptions: %w", err)
	}

	assets, err := f.provider.ListDiagnosticSettingsAssetTypes(ctx, cycleMetadata, lo.Keys(subscriptions))
	if err != nil {
		return fmt.Errorf("azure insights fetcher: error while receiving diagnostic settings: %w", err)
	}

	// group and send by subscription id
	subscriptionGroups := lo.GroupBy(assets, func(item inventory.AzureAsset) string {
		return item.SubscriptionId
	})

	for subId, subscription := range subscriptions {
		batchAssets := subscriptionGroups[subId]
		if batchAssets == nil {
			batchAssets = []inventory.AzureAsset{} // Use empty array instead of nil
		}

		select {
		case <-ctx.Done():
			err := ctx.Err()
			f.log.Infof("AzureInsightsBatchAssetFetcher.Fetch context err: %s", err.Error())
			return err
		case f.resourceCh <- fetching.ResourceInfo{
			CycleMetadata: cycleMetadata,
			Resource: &AzureBatchResource{
				typePair: typePair{
					Type:    fetching.MonitoringIdentity,
					SubType: fetching.AzureDiagnosticSettingsType,
				},
				Subscription: subscription,
				Assets:       batchAssets,
			},
		}:
		}
	}

	return nil
}

func (f *AzureInsightsBatchAssetFetcher) Stop() {}
