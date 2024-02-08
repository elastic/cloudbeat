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

	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

type AssetsEnricher interface {
	Enrich(ctx context.Context, cycleMetadata cycle.Metadata, assets []inventory.AzureAsset) error
}

func initEnrichers(provider azurelib.ProviderAPI) []AssetsEnricher {
	var enrichers []AssetsEnricher

	enrichers = append(enrichers, storageAccountEnricher{provider: provider})
	enrichers = append(enrichers, vmNetworkSecurityGroupEnricher{})
	enrichers = append(enrichers, sqlServerEnricher{provider: provider})
	enrichers = append(enrichers, postgresqlEnricher{provider: provider})
	enrichers = append(enrichers, mysqlAssetEnricher{provider: provider})
	enrichers = append(enrichers, keyVaultEnricher{provider: provider})
	enrichers = append(enrichers, appServiceEnricher{provider: provider})

	return enrichers
}
