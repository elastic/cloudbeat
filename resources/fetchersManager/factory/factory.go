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

package factory

import (
	"context"
	"fmt"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/uniqueness"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
)

type RegisteredFetcher struct {
	Fetcher   fetching.Fetcher
	Condition []fetching.Condition
}

type FetchersMap map[string]RegisteredFetcher

// NewFactory Creates a new factory based on the benchmark name
func NewFactory(ctx context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo, le uniqueness.Manager) (FetchersMap, error) {
	switch cfg.Benchmark {
	case config.CIS_AWS:
		return NewCisAwsFactory(ctx, log, cfg, ch)
	case config.CIS_K8S:
		return NewCisK8sFactory(ctx, log, cfg, ch, le)
	case config.CIS_EKS:
		return NewCisEksFactory(ctx, log, cfg, ch, le)
	}

	return nil, fmt.Errorf("benchmark %s is not supported, no fetchers to return", cfg.Benchmark)
}
