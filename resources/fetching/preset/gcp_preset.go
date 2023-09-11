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
	"context"

	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
	fetchers "github.com/elastic/cloudbeat/resources/fetching/fetchers/gcp"
	"github.com/elastic/cloudbeat/resources/fetching/registry"
	"github.com/elastic/cloudbeat/resources/providers/gcplib/inventory"
)

func NewCisGcpFetchers(ctx context.Context, log *logp.Logger, ch chan fetching.ResourceInfo, inventory inventory.ServiceAPI) (registry.FetchersMap, error) {
	log.Infof("Initializing GCP fetchers")
	m := make(registry.FetchersMap)

	assetsFetcher := fetchers.NewGcpAssetsFetcher(ctx, log, ch, inventory)
	m["gcp_cloud_assets_fetcher"] = registry.RegisteredFetcher{Fetcher: assetsFetcher}

	monitoringFetcher := fetchers.NewGcpMonitoringFetcher(ctx, log, ch, inventory)
	m["gcp_monitoring_fetcher"] = registry.RegisteredFetcher{Fetcher: monitoringFetcher}

	return m, nil
}
