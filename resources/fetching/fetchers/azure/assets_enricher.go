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

	"github.com/elastic/cloudbeat/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
	"github.com/elastic/cloudbeat/resources/utils/strings"
)

type AssetsEnricherAPI interface {
	Enrich(ctx context.Context, cycleMetadata cycle.Metadata, assets []inventory.AzureAsset) error
}

func initEnrichers(provider azurelib.ProviderAPI) []AssetsEnricherAPI {
	var enrichers []AssetsEnricherAPI

	enrichers = append(enrichers, storageAccountEnricher{provider: provider})

	return enrichers
}

type storageAccountEnricher struct {
	provider azurelib.ProviderAPI
}

func (e storageAccountEnricher) Enrich(ctx context.Context, cycleMetadata cycle.Metadata, assets []inventory.AzureAsset) error {
	subscriptions, err := e.provider.GetSubscriptions(ctx, cycleMetadata)
	if err != nil {
		return fmt.Errorf("storageAccountEnricher: error while getting subscription: %w", err)
	}

	diagSettings, err := e.provider.ListDiagnosticSettingsAssetTypes(ctx, cycleMetadata, lo.Keys(subscriptions))
	if err != nil {
		return fmt.Errorf("storageAccountEnricher: error while getting diagnostic settings: %w", err)
	}
	e.addUsedForActivityLogsFlag(assets, diagSettings)

	return nil
}

func (*storageAccountEnricher) addUsedForActivityLogsFlag(assets []inventory.AzureAsset, diagSettings []inventory.AzureAsset) {
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

		if a.Extension == nil {
			a.Extension = make(map[string]any)
		}
		a.Extension["usedForActivityLogs"] = true
		assets[i] = a
	}
}
