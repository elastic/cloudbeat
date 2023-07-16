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

package benchmark

import (
	"context"
	"fmt"

	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider"
	gcpdataprovider "github.com/elastic/cloudbeat/dataprovider/providers/gcp"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/factory"
	"github.com/elastic/cloudbeat/resources/fetching/registry"

	gcplib "github.com/elastic/cloudbeat/resources/providers/gcplib/auth"
)

type GCP struct{}

func (G *GCP) Run(context.Context) error { return nil }

func (G *GCP) Initialize(ctx context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo, dependencies *Dependencies) (registry.Registry, dataprovider.CommonDataProvider, error) {
	gcpClientConfig, err := gcplib.GetGcpClientConfig(cfg, log)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize gcp config: %w", err)
	}

	gcpFactoryConfig := &gcplib.GcpFactoryConfig{
		ProjectId:  cfg.CloudConfig.Gcp.ProjectId,
		ClientOpts: gcpClientConfig,
	}

	fetchers, err := factory.NewCisGcpFactory(ctx, log, ch, *gcpFactoryConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize gcp fetchers: %w", err)
	}

	return registry.NewRegistry(log, fetchers), gcpdataprovider.New(
		gcpdataprovider.WithLogger(log),
	), nil
}

func (G *GCP) Stop() {}
