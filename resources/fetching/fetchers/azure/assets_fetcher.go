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

	"github.com/elastic/elastic-agent-libs/logp"
	"golang.org/x/exp/maps"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
)

type AzureAssetsFetcher struct {
	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
	provider   inventory.ServiceAPI
}

type AzureResource struct {
	Type    string
	SubType string
	Asset   inventory.AzureAsset `json:"asset,omitempty"`
}

var AzureResourceTypes = map[string]string{
	inventory.ActivityLogAlertAssetType:          fetching.AzureActivityLogAlertType,
	inventory.ClassicStorageAccountAssetType:     fetching.AzureClassicStorageAccountType,
	inventory.ClassicVirtualMachineAssetType:     fetching.AzureClassicVMType,
	inventory.DiskAssetType:                      fetching.AzureDiskType,
	inventory.DocumentDBDatabaseAccountAssetType: fetching.AzureDocumentDBDatabaseAccountType,
	inventory.MySQLDBAssetType:                   fetching.AzureMySQLDBType,
	inventory.NetworkWatchersAssetType:           fetching.AzureNetworkWatchersType,
	inventory.NetworkWatchersFlowLogAssetType:    fetching.AzureNetworkWatchersFlowLogType,
	inventory.PostgreSQLDBAssetType:              fetching.AzurePostgreSQLDBType,
	inventory.SQLServersAssetType:                fetching.AzureSQLServerType,
	inventory.StorageAccountAssetType:            fetching.AzureStorageAccountType,
	inventory.VirtualMachineAssetType:            fetching.AzureVMType,
	inventory.WebsitesAssetType:                  fetching.AzureWebSiteType,
}

func NewAzureAssetsFetcher(log *logp.Logger, ch chan fetching.ResourceInfo, provider inventory.ServiceAPI) *AzureAssetsFetcher {
	return &AzureAssetsFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

func (f *AzureAssetsFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Info("Starting AzureAssetsFetcher.Fetch")
	// This might be relevant if we'd like to fetch assets in parallel in order to evaluate a rule that uses multiple resources
	assets, err := f.provider.ListAllAssetTypesByName(maps.Keys(AzureResourceTypes))
	if err != nil {
		return err
	}

	for _, asset := range assets {
		select {
		case <-ctx.Done():
			f.log.Infof("AzureAssetsFetcher.Fetch context err: %s", ctx.Err().Error())
			return nil
		case f.resourceCh <- fetching.ResourceInfo{
			CycleMetadata: cMetadata,
			Resource: &AzureResource{
				Type:    AzureResourceTypes[asset.Type],
				SubType: getAzureSubType(asset.Type),
				Asset:   asset,
			},
		}:
		}
	}

	return nil
}

func getAzureSubType(assetType string) string {
	return ""
}

func (f *AzureAssetsFetcher) Stop() {}

func (r *AzureResource) GetData() any {
	return r.Asset
}

func (r *AzureResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:      r.Asset.Id,
		Type:    r.Type,
		SubType: r.SubType,
		Name:    r.Asset.Name,
		Region:  r.Asset.Location,
	}, nil
}

func (r *AzureResource) GetElasticCommonData() (map[string]any, error) { return nil, nil }
