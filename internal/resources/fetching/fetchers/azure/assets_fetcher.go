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

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/governance"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
)

type AzureAssetsFetcher struct {
	log        *clog.Logger
	resourceCh chan fetching.ResourceInfo
	provider   azurelib.ProviderAPI
	enrichers  []AssetsEnricher
}

type AzureResource struct {
	Type         string
	SubType      string
	Asset        inventory.AzureAsset `json:"asset,omitempty"`
	Subscription governance.Subscription
}

type typePair struct {
	SubType string
	Type    string
}

func newPair(subType string, tpe string) typePair {
	return typePair{
		SubType: subType,
		Type:    tpe,
	}
}

var AzureAssetTypeToTypePair = map[string]typePair{
	inventory.ClassicStorageAccountAssetType:     newPair(fetching.AzureClassicStorageAccountType, fetching.CloudStorage),
	inventory.DiskAssetType:                      newPair(fetching.AzureDiskType, fetching.CloudCompute),
	inventory.DocumentDBDatabaseAccountAssetType: newPair(fetching.AzureDocumentDBDatabaseAccountType, fetching.CloudDatabase),
	inventory.MySQLDBAssetType:                   newPair(fetching.AzureMySQLDBType, fetching.CloudDatabase),
	inventory.FlexibleMySQLDBAssetType:           newPair(fetching.AzureFlexibleMySQLDBType, fetching.CloudDatabase),
	inventory.NetworkWatchersFlowLogAssetType:    newPair(fetching.AzureNetworkWatchersFlowLogType, fetching.MonitoringIdentity),
	inventory.FlexiblePostgreSQLDBAssetType:      newPair(fetching.AzureFlexiblePostgreSQLDBType, fetching.CloudDatabase),
	inventory.PostgreSQLDBAssetType:              newPair(fetching.AzurePostgreSQLDBType, fetching.CloudDatabase),
	inventory.SQLServersAssetType:                newPair(fetching.AzureSQLServerType, fetching.CloudDatabase),
	inventory.StorageAccountAssetType:            newPair(fetching.AzureStorageAccountType, fetching.CloudStorage),
	inventory.VirtualMachineAssetType:            newPair(fetching.AzureVMType, fetching.CloudCompute),
	inventory.WebsitesAssetType:                  newPair(fetching.AzureWebSiteType, fetching.CloudCompute),
	inventory.VaultAssetType:                     newPair(fetching.AzureVaultType, fetching.KeyManagement),
	inventory.RoleDefinitionsType:                newPair(fetching.AzureRoleDefinitionType, fetching.CloudIdentity),

	// This asset type is used only for enrichment purposes, but is sent to OPA layer, producing no findings.
	inventory.NetworkSecurityGroupAssetType: newPair(fetching.AzureNetworkSecurityGroupType, fetching.MonitoringIdentity),
}

// In order to simplify the mappings, we are trying to query all AzureAssetTypeToTypePair on every asset group
// Because this is done with an "|"" this means that we won't get irrelevant data
var AzureAssetGroups = []string{inventory.AssetGroupResources, inventory.AssetGroupAuthorizationResources}

func NewAzureAssetsFetcher(log *clog.Logger, ch chan fetching.ResourceInfo, provider azurelib.ProviderAPI) *AzureAssetsFetcher {
	return &AzureAssetsFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
		enrichers:  initEnrichers(provider),
	}
}

func (f *AzureAssetsFetcher) Fetch(ctx context.Context, cycleMetadata cycle.Metadata) error {
	f.log.Info("Starting AzureAssetsFetcher.Fetch")
	var errAgg error
	// This might be relevant if we'd like to fetch assets in parallel in order to evaluate a rule that uses multiple resources
	var assets []inventory.AzureAsset
	for _, assetGroup := range AzureAssetGroups {
		// Fetching all types even if non-existent in asset group for simplicity
		r, err := f.provider.ListAllAssetTypesByName(ctx, assetGroup, slices.Collect(maps.Keys(AzureAssetTypeToTypePair)))
		if err != nil {
			f.log.Errorf("AzureAssetsFetcher.Fetch failed to fetch asset group %s: %s", assetGroup, err.Error())
			errAgg = errors.Join(errAgg, err)
			continue
		}
		assets = append(assets, r...)
	}

	subscriptions, err := f.provider.GetSubscriptions(ctx, cycleMetadata)
	if err != nil {
		f.log.Errorf("Error fetching subscription information: %v", err)
	}

	for _, e := range f.enrichers {
		if err := e.Enrich(ctx, cycleMetadata, assets); err != nil {
			errAgg = errors.Join(errAgg, fmt.Errorf("error while enriching assets: %w", err))
		}
	}

	for _, asset := range assets {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			f.log.Infof("AzureAssetsFetcher.Fetch context err: %s", err.Error())
			errAgg = errors.Join(errAgg, err)
			return errAgg
		case f.resourceCh <- resourceFromAsset(asset, cycleMetadata, subscriptions):
		}
	}

	return errAgg
}

func resourceFromAsset(asset inventory.AzureAsset, cycleMetadata cycle.Metadata, subscriptions map[string]governance.Subscription) fetching.ResourceInfo {
	pair := AzureAssetTypeToTypePair[asset.Type]
	subscription, ok := subscriptions[asset.SubscriptionId]
	if !ok {
		subscription = governance.Subscription{
			FullyQualifiedID: asset.SubscriptionId,
			ShortID:          "",
			DisplayName:      "",
			ManagementGroup: governance.ManagementGroup{
				FullyQualifiedID: "",
				DisplayName:      "",
			},
		}
	}
	return fetching.ResourceInfo{
		CycleMetadata: cycleMetadata,
		Resource: &AzureResource{
			Type:         pair.Type,
			SubType:      pair.SubType,
			Asset:        asset,
			Subscription: subscription,
		},
	}
}

func (f *AzureAssetsFetcher) Stop() {}

func (r *AzureResource) GetData() any {
	return r.Asset
}

func (r *AzureResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:                   r.Asset.Id,
		Type:                 r.Type,
		SubType:              r.SubType,
		Name:                 r.Asset.Name,
		Region:               r.Asset.Location,
		CloudAccountMetadata: r.Subscription.GetCloudAccountMetadata(),
	}, nil
}

func (r *AzureResource) GetIds() []string {
	return []string{r.Asset.Id}
}

func (r *AzureResource) GetElasticCommonData() (map[string]any, error) {
	m := map[string]any{}

	switch r.Asset.Type {
	case inventory.VirtualMachineAssetType:
		{
			m["host.name"] = r.Asset.Name

			// "host.hostname" = "properties.osProfile.computerName" if it exists
			osProfileRaw, ok := r.Asset.Properties["osProfile"]
			if !ok {
				break
			}
			osProfile, ok := osProfileRaw.(map[string]any)
			if !ok {
				break
			}
			computerNameRaw, ok := osProfile["computerName"]
			if !ok {
				break
			}
			computerName, ok := computerNameRaw.(string)
			if !ok {
				break
			}
			m["host.hostname"] = computerName
		}
	case inventory.RoleDefinitionsType:
		{
			m["user.effective.id"] = r.Asset.Id
			m["user.effective.name"] = r.Asset.Name
		}
	}

	return m, nil
}
