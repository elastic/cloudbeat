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

package azurelib

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/auth"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/governance"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
)

type ProviderAPI interface {
	inventory.AssetsARGProviderAPI
	inventory.SQLProviderAPI
	inventory.StorageAccountProviderAPI
	governance.ProviderAPI
}

type ProviderInitializer struct{}

type ProviderInitializerAPI interface {
	// Init initializes the Azure clients
	Init(log *logp.Logger, azureConfig auth.AzureFactoryConfig) (ProviderAPI, error)
}

func (p *ProviderInitializer) Init(log *logp.Logger, azureConfig auth.AzureFactoryConfig) (ProviderAPI, error) {
	log = log.Named("azure")

	factory, err := armresourcegraph.NewClientFactory(azureConfig.Credentials, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize resource graph factory: %w", err)
	}
	resourceGraphClientFactory := factory.NewClient()

	diagnosticSettingsClient, err := armmonitor.NewDiagnosticSettingsClient(azureConfig.Credentials, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to init monitor client: %w", err)
	}

	inventoryProvider := inventory.NewAssetsARGProvider(log, resourceGraphClientFactory)
	return &provider{
		inventory:        inventoryProvider,
		sqlInventory:     inventory.NewSQLProvider(log, azureConfig.Credentials),
		storageInventory: inventory.NewStorageAccountProvider(log, diagnosticSettingsClient, azureConfig.Credentials),
		governance:       governance.NewProvider(log, inventoryProvider),
	}, nil
}

type provider struct {
	inventory        inventory.AssetsARGProviderAPI
	sqlInventory     inventory.SQLProviderAPI
	storageInventory inventory.StorageAccountProviderAPI
	governance       governance.ProviderAPI
}

func (p provider) ListAllAssetTypesByName(ctx context.Context, assetGroup string, assets []string) ([]inventory.AzureAsset, error) {
	return p.inventory.ListAllAssetTypesByName(ctx, assetGroup, assets)
}

func (p provider) ListDiagnosticSettingsAssetTypes(ctx context.Context, cycleMetadata cycle.Metadata, subscriptionIDs []string) ([]inventory.AzureAsset, error) {
	return p.storageInventory.ListDiagnosticSettingsAssetTypes(ctx, cycleMetadata, subscriptionIDs)
}

func (p provider) ListStorageAccountBlobServices(ctx context.Context, storageAccounts []inventory.AzureAsset) ([]inventory.AzureAsset, error) {
	return p.storageInventory.ListStorageAccountBlobServices(ctx, storageAccounts)
}

func (p provider) ListSQLEncryptionProtector(ctx context.Context, subID, resourceGroup, sqlServerName string) ([]inventory.AzureAsset, error) {
	return p.sqlInventory.ListSQLEncryptionProtector(ctx, subID, resourceGroup, sqlServerName)
}

func (p provider) ListSQLTransparentDataEncryptions(ctx context.Context, subID, resourceGroup, sqlServerName string) ([]inventory.AzureAsset, error) {
	return p.sqlInventory.ListSQLTransparentDataEncryptions(ctx, subID, resourceGroup, sqlServerName)
}

func (p provider) GetSQLBlobAuditingPolicies(ctx context.Context, subID, resourceGroup, sqlServerName string) ([]inventory.AzureAsset, error) {
	return p.sqlInventory.GetSQLBlobAuditingPolicies(ctx, subID, resourceGroup, sqlServerName)
}

func (p provider) ListStorageAccountsBlobDiagnosticSettings(ctx context.Context, storageAccounts []inventory.AzureAsset) ([]inventory.AzureAsset, error) {
	return p.storageInventory.ListStorageAccountsBlobDiagnosticSettings(ctx, storageAccounts)
}

func (p provider) ListStorageAccountsTableDiagnosticSettings(ctx context.Context, storageAccounts []inventory.AzureAsset) ([]inventory.AzureAsset, error) {
	return p.storageInventory.ListStorageAccountsTableDiagnosticSettings(ctx, storageAccounts)
}

func (p provider) ListStorageAccountsQueueDiagnosticSettings(ctx context.Context, storageAccounts []inventory.AzureAsset) ([]inventory.AzureAsset, error) {
	return p.storageInventory.ListStorageAccountsQueueDiagnosticSettings(ctx, storageAccounts)
}

func (p provider) GetSubscriptions(ctx context.Context, cycleMetadata cycle.Metadata) (map[string]governance.Subscription, error) {
	return p.governance.GetSubscriptions(ctx, cycleMetadata)
}
