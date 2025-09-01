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

	"github.com/go-viper/mapstructure/v2"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/inventory"
	azurelib "github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

type resourceGraphFetcher struct {
	logger   *clog.Logger
	tenantID string //nolint:unused
	provider resourceGraphProvider
}

type (
	resourceGraphProvider interface {
		ListAllAssetTypesByName(ctx context.Context, assetGroup string, assets []string) ([]azurelib.AzureAsset, error)
	}
)

func newResourceGraphFetcher(logger *clog.Logger, tenantID string, provider resourceGraphProvider) inventory.AssetFetcher {
	return &resourceGraphFetcher{
		logger:   logger,
		tenantID: tenantID,
		provider: provider,
	}
}

func (f *resourceGraphFetcher) Fetch(ctx context.Context, assetChan chan<- inventory.AssetEvent) {
	resourcesToFetch := []struct {
		name           string
		serviceName    string
		azureGroup     string
		azureType      string
		classification inventory.AssetClassification
	}{
		{"App Services", "Azure App Services", azurelib.AssetGroupResources, azurelib.WebsitesAssetType, inventory.AssetClassificationAzureAppService},
		{"Container Registries", "Azure Container Registries", azurelib.AssetGroupResources, azurelib.ContainerRegistryAssetType, inventory.AssetClassificationAzureContainerRegistry},
		{"Cosmos DB Accounts", "Azure Cosmos DB", azurelib.AssetGroupResources, azurelib.DocumentDBDatabaseAccountAssetType, inventory.AssetClassificationAzureCosmosDBAccount},
		{"Cosmos DB SQL Databases", "Azure Cosmos DB", azurelib.AssetGroupResources, azurelib.CosmosDBForSQLDatabaseAssetType, inventory.AssetClassificationAzureCosmosDBSQLDatabase},
		{"Disks", "Azure Storage", azurelib.AssetGroupResources, azurelib.DiskAssetType, inventory.AssetClassificationAzureDisk},
		{"Elastic Pools", "Azure SQL Elastic Pools", azurelib.AssetGroupResources, azurelib.ElasticPoolAssetType, inventory.AssetClassificationAzureElasticPool},
		{"MySQL Flexible Servers", "Azure SQL Servers", azurelib.AssetGroupResources, azurelib.FlexibleMySQLDBServerAssetType, inventory.AssetClassificationAzureSQLServer},
		{"Resource Groups", "Azure Management", azurelib.AssetGroupResourceContainers, azurelib.ResouceGroupAssetType, inventory.AssetClassificationAzureResourceGroup},
		{"SQL Databases", "Azure SQL Databases", azurelib.AssetGroupResources, azurelib.MySQLDatabaseAssetType, inventory.AssetClassificationAzureSQLDatabase},
		{"Snapshots", "Azure Storage", azurelib.AssetGroupResources, azurelib.SnapshotAssetType, inventory.AssetClassificationAzureSnapshot},
		{"Storage Accounts", "Azure Storage", azurelib.AssetGroupResources, azurelib.StorageAccountAssetType, inventory.AssetClassificationAzureStorageAccount},
		{"Virtual Machines", "Azure Virtual Machines", azurelib.AssetGroupResources, azurelib.VirtualMachineAssetType, inventory.AssetClassificationAzureVirtualMachine},
	}
	for _, r := range resourcesToFetch {
		f.fetch(ctx, r.name, r.serviceName, r.azureGroup, r.azureType, r.classification, assetChan)
	}
}

func (f *resourceGraphFetcher) fetch(ctx context.Context, resourceName, serviceName, resourceGroup, resourceType string, classification inventory.AssetClassification, assetChan chan<- inventory.AssetEvent) {
	f.logger.Infof("Fetching %s", resourceName)
	defer f.logger.Infof("Fetching %s - Finished", resourceName)

	azureAssets, err := f.provider.ListAllAssetTypesByName(ctx, resourceGroup, []string{resourceType})
	if err != nil {
		f.logger.Errorf(ctx, "Could not fetch %s: %v", resourceName, err)
		return
	}

	for _, item := range azureAssets {
		asset := inventory.NewAssetEvent(
			classification,
			item.Id,
			pickName(item.Name, item.DisplayName, item.Id),
			inventory.WithRawAsset(item),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				Region:      item.Location,
				AccountID:   item.TenantId,
				ProjectID:   item.SubscriptionId,
				ServiceName: serviceName,
			}),
			inventory.WithLabelsFromAny(item.Tags),
		)

		if resourceType == azurelib.VirtualMachineAssetType {
			vmProperties := tryUnpackingVMProperties(item.Properties)
			if vmProperties != nil {
				asset.Host = &inventory.Host{
					ID:   item.Id,
					Name: vmProperties.Extended.InstanceView.ComputerName,
					Type: vmProperties.HardwareProfile.VmSize,
				}
				asset.Cloud.MachineType = vmProperties.HardwareProfile.VmSize
			}
			asset.Cloud.InstanceID = item.Id
			asset.Cloud.InstanceName = item.Name
		}

		assetChan <- asset
	}
}

//nolint:revive
type vmProperties struct {
	Extended struct {
		InstanceView struct {
			ComputerName string
		}
	}
	HardwareProfile struct {
		VmSize string
	}
}

func tryUnpackingVMProperties(m map[string]any) *vmProperties {
	o := &vmProperties{}
	err := mapstructure.Decode(m, o)
	if err != nil {
		return nil
	}
	return o
}
