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
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/governance"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

type AzureBatchAssetFetcher struct {
	log        *clog.Logger
	resourceCh chan fetching.ResourceInfo
	provider   azurelib.ProviderAPI
}

var AzureBatchAssets = map[string]typePair{
	inventory.ActivityLogAlertAssetType: newPair(fetching.AzureActivityLogAlertType, fetching.MonitoringIdentity),
	inventory.ApplicationInsights:       newPair(fetching.AzureInsightsComponentType, fetching.MonitoringIdentity),
	inventory.BastionAssetType:          newPair(fetching.AzureBastionType, fetching.CloudDns),
}

// In order to simplify the mappings, we are trying to query all AzureBatchAssets on every asset group
// Because this is done with an "|"" this means that we won't get irrelevant data
var AzureBatchAssetGroups = []string{inventory.AssetGroupResources}

func NewAzureBatchAssetFetcher(log *clog.Logger, ch chan fetching.ResourceInfo, provider azurelib.ProviderAPI) *AzureBatchAssetFetcher {
	return &AzureBatchAssetFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

func (f *AzureBatchAssetFetcher) Fetch(ctx context.Context, cycleMetadata cycle.Metadata) error {
	f.log.Info("Starting AzureBatchAssetFetcher.Fetch")
	subscriptions, err := f.provider.GetSubscriptions(ctx, cycleMetadata)
	if err != nil {
		return fmt.Errorf("failed to fetch governance info: %w", err)
	}

	var errAgg error
	assets := []inventory.AzureAsset{}
	for _, assetGroup := range AzureBatchAssetGroups {
		r, err := f.provider.ListAllAssetTypesByName(ctx, assetGroup, slices.Collect(maps.Keys(AzureBatchAssets)))
		if err != nil {
			f.log.Errorf("AzureBatchAssetFetcher.Fetch failed to fetch asset group %s: %s", assetGroup, err.Error())
			errAgg = errors.Join(errAgg, err)
			continue
		}
		assets = append(assets, r...)
	}

	subscriptionGroups := lo.GroupBy(assets, func(item inventory.AzureAsset) string {
		return item.SubscriptionId
	})

	for _, sub := range subscriptions {
		assetGroups := lo.GroupBy(subscriptionGroups[sub.ShortID], func(item inventory.AzureAsset) string {
			return item.Type
		})
		for assetType, pair := range AzureBatchAssets {
			batchAssets := assetGroups[assetType]
			if batchAssets == nil {
				batchAssets = []inventory.AzureAsset{} // Use empty array instead of nil
			}

			select {
			case <-ctx.Done():
				err := ctx.Err()
				f.log.Infof("AzureBatchAssetFetcher.Fetch context err: %s", err.Error())
				errAgg = errors.Join(errAgg, err)
				return errAgg
			case f.resourceCh <- fetching.ResourceInfo{
				CycleMetadata: cycleMetadata,
				Resource: &AzureBatchResource{
					// Every asset in the list has the same type and subtype
					typePair:     pair,
					Subscription: sub,
					Assets:       batchAssets,
				},
			}:
			}
		}
	}

	return errAgg
}

func (f *AzureBatchAssetFetcher) Stop() {}

type AzureBatchResource struct {
	typePair
	Subscription governance.Subscription
	Assets       []inventory.AzureAsset `json:"assets,omitempty"`
}

func (r *AzureBatchResource) GetData() any {
	return r.Assets
}

func (r *AzureBatchResource) GetIds() []string {
	return lo.Map(r.Assets, func(item inventory.AzureAsset, _ int) string {
		return item.Id
	})
}

func (r *AzureBatchResource) GetMetadata() (fetching.ResourceMetadata, error) {
	// Assuming all batch in not empty includes assets of the same subscription
	id := fmt.Sprintf("%s-%s", r.SubType, r.Subscription.ShortID)
	return fetching.ResourceMetadata{
		ID:      id,
		Type:    r.Type,
		SubType: r.SubType,
		Name:    id,
		// TODO: Make sure ActivityLogAlerts are not location scoped (benchmarks do not check location)
		Region:               azurelib.GlobalRegion,
		CloudAccountMetadata: r.Subscription.GetCloudAccountMetadata(),
	}, nil
}

func (r *AzureBatchResource) GetElasticCommonData() (map[string]any, error) {
	return nil, nil
}
