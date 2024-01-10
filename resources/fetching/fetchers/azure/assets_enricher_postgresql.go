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

	"github.com/elastic/cloudbeat/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
)

type postgresqlEnricher struct {
	provider azurelib.ProviderAPI
}

func (p postgresqlEnricher) Enrich(ctx context.Context, _ cycle.Metadata, assets []inventory.AzureAsset) error {
	var errs error

	for i, a := range assets {
		if a.Type != inventory.PostgreSQLDBAssetType &&
			a.Type != inventory.FlexiblePostgreSQLDBAssetType {
			continue
		}

		if err := p.enrichConfigurations(ctx, &a); err != nil {
			errs = errors.Join(errs, err)
		}

		assets[i] = a
	}

	return errs
}

func (p postgresqlEnricher) enrichConfigurations(ctx context.Context, a *inventory.AzureAsset) error {
	configs, err := p.listConfigurations(ctx, a)
	if err != nil {
		return err
	}

	enrichExtension(a, inventory.ExtensionPostgresqlConfigurations, configs)
	return nil
}

func (p postgresqlEnricher) listConfigurations(ctx context.Context, a *inventory.AzureAsset) ([]inventory.AzureAsset, error) {
	if a.Type == inventory.PostgreSQLDBAssetType {
		return p.provider.ListPostgresConfigurations(ctx, a.SubscriptionId, a.ResourceGroup, a.Name)
	}

	return p.provider.ListFlexiblePostgresConfigurations(ctx, a.SubscriptionId, a.ResourceGroup, a.Name)
}
