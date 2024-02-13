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

	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
	"github.com/elastic/cloudbeat/internal/resources/utils/maps"
	"github.com/elastic/cloudbeat/internal/resources/utils/strings"
)

type storageAccountEnricher struct {
	provider azurelib.ProviderAPI
}

func (e storageAccountEnricher) Enrich(ctx context.Context, cycleMetadata cycle.Metadata, assets []inventory.AzureAsset) error {
	var errs []error

	subscriptions, err := e.provider.GetSubscriptions(ctx, cycleMetadata)
	if err != nil {
		return fmt.Errorf("storageAccountEnricher: error while getting subscription: %w", err)
	}

	if err := e.addUsedForActivityLogsFlag(ctx, cycleMetadata, assets, lo.Keys(subscriptions)); err != nil {
		errs = append(errs, err)
	}

	storageAccounts := lo.Filter(assets, func(item inventory.AzureAsset, _ int) bool {
		return item.Type == inventory.StorageAccountAssetType
	})

	if err := e.addStorageAccountBlobServices(ctx, storageAccounts, assets); err != nil {
		errs = append(errs, fmt.Errorf("storageAccountEnricher: error while getting data protection settings: %w", err))
	}

	if err := e.addStorageAccountServicesDiagnosticSettings(ctx, storageAccounts, assets); err != nil {
		errs = append(errs, fmt.Errorf("storageAccountEnricher: error while getting services diagnostic settings: %w", err))
	}

	storageAccountsSubscriptionsIds := lo.Uniq(lo.Map(storageAccounts, func(item inventory.AzureAsset, _ int) string {
		return item.SubscriptionId
	}))

	// Assets of the storage account type that are returned from Azure Resource Graph API need to
	// be enriched with additional data from the Azure Go SDK Account client
	if err := e.addStorageAccounts(ctx, storageAccountsSubscriptionsIds, assets); err != nil {
		errs = append(errs, fmt.Errorf("storageAccountEnricher: error while getting storage accounts: %w", err))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (e storageAccountEnricher) addStorageAccounts(ctx context.Context, storageAccountsSubscriptionsIds []string, assets []inventory.AzureAsset) error {
	if len(storageAccountsSubscriptionsIds) == 0 {
		return nil
	}

	sa, err := e.provider.ListStorageAccounts(ctx, storageAccountsSubscriptionsIds)
	if err != nil {
		return fmt.Errorf("storageAccountEnricher: error while getting storage accounts: %w", err)
	}
	return e.appendExtensionAssets(inventory.ExtensionStorageAccount, sa, assets)
}

func (e storageAccountEnricher) addUsedForActivityLogsFlag(ctx context.Context, cycleMetadata cycle.Metadata, assets []inventory.AzureAsset, subscriptionIDs []string) error {
	diagSettings, err := e.provider.ListDiagnosticSettingsAssetTypes(ctx, cycleMetadata, subscriptionIDs)
	if err != nil {
		return fmt.Errorf("storageAccountEnricher: error while getting diagnostic settings: %w", err)
	}

	usedStorageAccountIDs := map[string]struct{}{}
	for _, d := range diagSettings {
		storageAccountID := strings.FromMap(d.Properties, "storageAccountId")
		if storageAccountID == "" {
			continue
		}
		usedStorageAccountIDs[storageAccountID] = struct{}{}
	}

	for i, a := range assets {
		if a.Type != inventory.StorageAccountAssetType {
			continue
		}

		if _, exists := usedStorageAccountIDs[a.Id]; !exists {
			continue
		}

		a.AddExtension(inventory.ExtensionUsedForActivityLogs, true)

		assets[i] = a
	}

	return nil
}

func (e storageAccountEnricher) addStorageAccountBlobServices(ctx context.Context, storageAccounts []inventory.AzureAsset, assets []inventory.AzureAsset) error {
	if len(storageAccounts) == 0 {
		return nil
	}

	dataProtection, err := e.provider.ListStorageAccountBlobServices(ctx, storageAccounts)
	if err != nil {
		return err
	}

	return e.appendExtensionAssets(inventory.ExtensionBlobService, dataProtection, assets)
}

func (e storageAccountEnricher) addStorageAccountServicesDiagnosticSettings(ctx context.Context, storageAccounts []inventory.AzureAsset, assets []inventory.AzureAsset) error {
	if len(storageAccounts) == 0 {
		return nil
	}

	var errs []error

	if err := e.addBlobDiagnosticSettings(ctx, storageAccounts, assets); err != nil {
		errs = append(errs, err)
	}
	if err := e.addTableDiagnosticSettings(ctx, storageAccounts, assets); err != nil {
		errs = append(errs, err)
	}
	if err := e.addQueueDiagnosticSettings(ctx, storageAccounts, assets); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (e storageAccountEnricher) addBlobDiagnosticSettings(ctx context.Context, storageAccounts []inventory.AzureAsset, assets []inventory.AzureAsset) error {
	diagSettings, err := e.provider.ListStorageAccountsBlobDiagnosticSettings(ctx, storageAccounts)
	if err != nil {
		return err
	}
	return e.appendExtensionAssets(inventory.ExtensionBlobDiagnosticSettings, diagSettings, assets)
}

func (e storageAccountEnricher) addTableDiagnosticSettings(ctx context.Context, storageAccounts []inventory.AzureAsset, assets []inventory.AzureAsset) error {
	diagSettings, err := e.provider.ListStorageAccountsTableDiagnosticSettings(ctx, storageAccounts)
	if err != nil {
		return err
	}
	return e.appendExtensionAssets(inventory.ExtensionTableDiagnosticSettings, diagSettings, assets)
}

func (e storageAccountEnricher) addQueueDiagnosticSettings(ctx context.Context, storageAccounts []inventory.AzureAsset, assets []inventory.AzureAsset) error {
	diagSettings, err := e.provider.ListStorageAccountsQueueDiagnosticSettings(ctx, storageAccounts)
	if err != nil {
		return err
	}
	return e.appendExtensionAssets(inventory.ExtensionQueueDiagnosticSettings, diagSettings, assets)
}

func (e storageAccountEnricher) appendExtensionAssets(extensionField string, extensions []inventory.AzureAsset, assets []inventory.AzureAsset) error {
	if len(extensions) == 0 {
		return nil
	}

	// map per storage account id
	perStorageAccountID := map[string]inventory.AzureAsset{}
	for _, d := range extensions {
		perStorageAccountID[strings.FromMap(d.Extension, inventory.ExtensionStorageAccountID)] = d
	}

	var errs []error
	for i, a := range assets {
		if a.Type != inventory.StorageAccountAssetType {
			continue
		}

		ext, exist := perStorageAccountID[a.Id]
		if !exist {
			continue
		}

		extMap, err := maps.AsMapStringAny(ext)
		if err != nil {
			errs = append(errs, err)
		}

		a.AddExtension(extensionField, extMap)

		assets[i] = a
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
