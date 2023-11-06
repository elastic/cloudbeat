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
	inventory.ClassicVirtualMachineAssetType:     newPair(fetching.AzureClassicVMType, fetching.CloudCompute),
	inventory.DiskAssetType:                      newPair(fetching.AzureDiskType, fetching.CloudCompute),
	inventory.DocumentDBDatabaseAccountAssetType: newPair(fetching.AzureDocumentDBDatabaseAccountType, fetching.CloudDatabase),
	inventory.MySQLDBAssetType:                   newPair(fetching.AzureMySQLDBType, fetching.CloudDatabase),
	inventory.NetworkWatchersAssetType:           newPair(fetching.AzureNetworkWatchersType, fetching.MonitoringIdentity),
	inventory.NetworkWatchersFlowLogAssetType:    newPair(fetching.AzureNetworkWatchersFlowLogType, fetching.MonitoringIdentity),
	inventory.PostgreSQLDBAssetType:              newPair(fetching.AzurePostgreSQLDBType, fetching.CloudDatabase),
	inventory.SQLServersAssetType:                newPair(fetching.AzureSQLServerType, fetching.CloudDatabase),
	inventory.StorageAccountAssetType:            newPair(fetching.AzureStorageAccountType, fetching.CloudStorage),
	inventory.VirtualMachineAssetType:            newPair(fetching.AzureVMType, fetching.CloudCompute),
	inventory.WebsitesAssetType:                  newPair(fetching.AzureWebSiteType, fetching.CloudCompute),
	inventory.VaultAssetType:                     newPair(fetching.AzureVaultType, fetching.KeyManagement),
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
	assets, err := f.provider.ListAllAssetTypesByName(ctx, maps.Keys(AzureAssetTypeToTypePair))
	if err != nil {
		return err
	}

	for _, asset := range assets {
		select {
		case <-ctx.Done():
			f.log.Infof("AzureAssetsFetcher.Fetch context err: %s", ctx.Err().Error())
			return nil
		case f.resourceCh <- resourceFromAsset(asset, cMetadata):
		}
	}

	return nil
}

func resourceFromAsset(asset inventory.AzureAsset, cMetadata fetching.CycleMetadata) fetching.ResourceInfo {
	pair := AzureAssetTypeToTypePair[asset.Type]
	return fetching.ResourceInfo{
		CycleMetadata: cMetadata,
		Resource: &AzureResource{
			Type:    pair.Type,
			SubType: pair.SubType,
			Asset:   asset,
		},
	}
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

func (r *AzureResource) GetElasticCommonData() (map[string]any, error) {
	return map[string]any{
		"cloud": map[string]any{
			"provider": "azure",
			"account": map[string]any{
				"id":   r.Asset.SubscriptionId,
				"name": r.Asset.SubscriptionName,
			},
			// TODO: Organization fields
		},
	}, nil
}
