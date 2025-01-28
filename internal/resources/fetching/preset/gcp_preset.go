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

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	fetchers "github.com/elastic/cloudbeat/internal/resources/fetching/fetchers/gcp"
	"github.com/elastic/cloudbeat/internal/resources/fetching/registry"
	"github.com/elastic/cloudbeat/internal/resources/providers/gcplib/inventory"
)

func NewCisGcpFetchers(ctx context.Context, log *clog.Logger, ch chan fetching.ResourceInfo, inventory inventory.ServiceAPI, cfg config.GcpConfig) (registry.FetchersMap, error) {
	log.Infof("Initializing GCP fetchers")
	m := make(registry.FetchersMap)

	assetsFetcher := fetchers.NewGcpAssetsFetcher(ctx, log, ch, inventory)
	m["gcp_cloud_assets_fetcher"] = registry.RegisteredFetcher{Fetcher: assetsFetcher}

	monitoringFetcher := fetchers.NewGcpMonitoringFetcher(ctx, log, ch, inventory)
	m["gcp_monitoring_fetcher"] = registry.RegisteredFetcher{Fetcher: monitoringFetcher}

	serviceUsageFetcher := fetchers.NewGcpServiceUsageFetcher(ctx, log, ch, inventory)
	m["gcp_service_usage_fetcher"] = registry.RegisteredFetcher{Fetcher: serviceUsageFetcher}

	// The logging fetcher is only available for the organization scope as it requires the Cloud Asset Inventory API
	// to be enabled for the organization/folders level.
	if cfg.AccountType == config.OrganizationAccount {
		loggingFetcher := fetchers.NewGcpLogSinkFetcher(ctx, log, ch, inventory)
		m["gcp_logging_fetcher"] = registry.RegisteredFetcher{Fetcher: loggingFetcher}

		policiesFetcher := fetchers.NewGcpPoliciesFetcher(ctx, log, ch, inventory)
		m["gcp_policies_fetcher"] = registry.RegisteredFetcher{Fetcher: policiesFetcher}
	}

	return m, nil
}
