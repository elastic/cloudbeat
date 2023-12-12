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

package preset

import (
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
	fetchers "github.com/elastic/cloudbeat/resources/fetching/fetchers/azure"
	"github.com/elastic/cloudbeat/resources/fetching/registry"
	"github.com/elastic/cloudbeat/resources/providers/azurelib"
)

func NewCisAzureFactory(log *logp.Logger, ch chan fetching.ResourceInfo, provider azurelib.ProviderAPI) (registry.FetchersMap, error) {
	log.Infof("Initializing Azure fetchers")
	m := make(registry.FetchersMap)

	assetsFetcher := fetchers.NewAzureAssetsFetcher(log, ch, provider)
	m["azure_cloud_assets_fetcher"] = registry.RegisteredFetcher{Fetcher: assetsFetcher}

	batchFetcher := fetchers.NewAzureBatchAssetFetcher(log, ch, provider)
	m["azure_cloud_batch_asset_fetcher"] = registry.RegisteredFetcher{Fetcher: batchFetcher}

	insightsBatchFetcher := fetchers.NewAzureInsightsBatchAssetFetcher(log, ch, provider)
	m["azure_cloud_insights_batch_asset_fetcher"] = registry.RegisteredFetcher{Fetcher: insightsBatchFetcher}

	return m, nil
}
