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

	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/governance"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

type AzureLocationsNetworkWatcherAssetBatchFetcher struct {
	log        *clog.Logger
	resourceCh chan fetching.ResourceInfo
	provider   azurelib.ProviderAPI
}

func NewAzureLocationsNetworkWatcherAssetBatchFetcher(log *clog.Logger, ch chan fetching.ResourceInfo, provider azurelib.ProviderAPI) *AzureLocationsNetworkWatcherAssetBatchFetcher {
	return &AzureLocationsNetworkWatcherAssetBatchFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

func (f *AzureLocationsNetworkWatcherAssetBatchFetcher) Fetch(ctx context.Context, metadata cycle.Metadata) error {
	f.log.Info("Starting AzureLocationsNetworkWatcherAssetBatchFetcher.Fetch")
	subscriptions, err := f.provider.GetSubscriptions(ctx, metadata)
	if err != nil {
		return err
	}

	watchers, err := f.provider.ListAllAssetTypesByName(ctx, inventory.AssetGroupResources, []string{inventory.NetworkWatchersAssetType})
	if err != nil {
		return err
	}

	var errAgg error
	for _, subscription := range subscriptions {
		errAgg = errors.Join(errAgg, f.fetchNetworkWatchersPerLocation(ctx, metadata, watchers, subscription))
	}

	return errAgg
}

func (f *AzureLocationsNetworkWatcherAssetBatchFetcher) fetchNetworkWatchersPerLocation(ctx context.Context, metadata cycle.Metadata, watchers []inventory.AzureAsset, subscription governance.Subscription) error {
	subID := subscription.ShortID
	locations, err := f.provider.ListLocations(ctx, subID)
	if err != nil {
		return err
	}

	subscriptionWatchers := lo.Filter(watchers, func(watcher inventory.AzureAsset, _ int) bool {
		return watcher.SubscriptionId == subID
	})

	groupedWatchers := lo.GroupBy(subscriptionWatchers, func(watcher inventory.AzureAsset) string {
		return watcher.Location
	})

	var errAgg error
	for _, location := range locations {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			f.log.Infof("AzureLocationsNetworkWatcherAssetBatchFetcher.Fetch context err: %s", err.Error())
			errAgg = errors.Join(errAgg, err)
			return errAgg
		case f.resourceCh <- fetching.ResourceInfo{
			CycleMetadata: metadata,
			Resource: &NetworkWatchersBatchedByLocationResource{
				typePair:        newPair(fetching.AzureNetworkWatchersType, fetching.MonitoringIdentity),
				Subscription:    subscription,
				Location:        location,
				NetworkWatchers: groupedWatchers[location.Name],
			},
		}:
		}
	}

	return errAgg
}

func (f *AzureLocationsNetworkWatcherAssetBatchFetcher) Stop() {
}

type NetworkWatchersBatchedByLocationResource struct {
	typePair
	Subscription    governance.Subscription `json:"subscription"`
	Location        inventory.AzureAsset    `json:"location"`
	NetworkWatchers []inventory.AzureAsset  `json:"networkWatchers,omitempty"`
}

func (r *NetworkWatchersBatchedByLocationResource) GetMetadata() (fetching.ResourceMetadata, error) {
	id := r.buildId()
	return fetching.ResourceMetadata{
		ID:                   id,
		Name:                 id,
		Type:                 r.Type,
		SubType:              r.SubType,
		Region:               r.Location.Name,
		CloudAccountMetadata: r.Subscription.GetCloudAccountMetadata(),
	}, nil
}

func (r *NetworkWatchersBatchedByLocationResource) buildId() string {
	id := fmt.Sprintf("%s-%s-%s", r.SubType, r.Location.Name, r.Subscription.ShortID)
	return id
}

func (r *NetworkWatchersBatchedByLocationResource) GetData() any {
	return r
}

func (r *NetworkWatchersBatchedByLocationResource) GetIds() []string {
	return []string{r.buildId()}
}

func (r *NetworkWatchersBatchedByLocationResource) GetElasticCommonData() (map[string]any, error) {
	return nil, nil
}
