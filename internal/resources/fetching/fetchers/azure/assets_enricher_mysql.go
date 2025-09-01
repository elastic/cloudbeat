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

	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

type mysqlAssetEnricher struct {
	provider azurelib.ProviderAPI
}

func (e mysqlAssetEnricher) Enrich(ctx context.Context, _ cycle.Metadata, assets []inventory.AzureAsset) error {
	var errAgg error
	for i, a := range assets {
		if a.Type != inventory.FlexibleMySQLDBServerAssetType {
			continue
		}

		if err := e.enrichTLSVersion(ctx, &a); err != nil {
			errAgg = errors.Join(errAgg, err)
		}

		assets[i] = a
	}

	return errAgg
}

func (e mysqlAssetEnricher) enrichTLSVersion(ctx context.Context, asset *inventory.AzureAsset) error {
	configs, err := e.provider.GetFlexibleTLSVersionConfiguration(ctx, asset.SubscriptionId, asset.ResourceGroup, asset.Name)
	if err != nil {
		return err
	}

	if len(configs) == 0 {
		return nil
	}

	asset.AddExtension(inventory.ExtensionMysqlConfigurations, configs)
	return nil
}
