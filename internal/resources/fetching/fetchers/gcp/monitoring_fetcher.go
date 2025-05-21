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

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/gcplib"
	"github.com/elastic/cloudbeat/internal/resources/providers/gcplib/inventory"
)

type GcpMonitoringFetcher struct {
	log        *clog.Logger
	resourceCh chan fetching.ResourceInfo
	provider   inventory.ServiceAPI
}

type GcpMonitoringAsset struct {
	Type    string
	subType string

	Asset *inventory.MonitoringAsset `json:"assets,omitempty"`
}

func NewGcpMonitoringFetcher(_ context.Context, log *clog.Logger, ch chan fetching.ResourceInfo, provider inventory.ServiceAPI) *GcpMonitoringFetcher {
	return &GcpMonitoringFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

func (f *GcpMonitoringFetcher) Fetch(ctx context.Context, cycleMetadata cycle.Metadata) error {
	f.log.Info("Starting GcpMonitoringFetcher.Fetch")
	defer f.log.Info("GcpMonitoringFetcher.Fetch done")
	defer f.provider.Clear()

	resultsCh := make(chan *inventory.MonitoringAsset)
	go f.provider.ListMonitoringAssets(ctx, resultsCh)

	for asset := range resultsCh {
		select {
		case <-ctx.Done():
			f.log.Debugf("GcpMonitoringFetcher.Fetch context done: %v", ctx.Err())
			return nil

		case f.resourceCh <- fetching.ResourceInfo{
			CycleMetadata: cycleMetadata,
			Resource: &GcpMonitoringAsset{
				Type:    fetching.MonitoringIdentity,
				subType: fetching.GcpMonitoringType,
				Asset:   asset,
			},
		}:
		}
	}
	return nil
}

func (f *GcpMonitoringFetcher) Stop() {
	f.provider.Close()
}

func (g *GcpMonitoringAsset) GetMetadata() (fetching.ResourceMetadata, error) {
	id := g.buildId()
	return fetching.ResourceMetadata{
		ID:                   id,
		Type:                 g.Type,
		SubType:              g.subType,
		Name:                 id,
		Region:               gcplib.GlobalRegion,
		CloudAccountMetadata: *g.Asset.CloudAccount,
	}, nil
}

func (g *GcpMonitoringAsset) buildId() string {
	return fmt.Sprintf("%s-%s", g.subType, g.Asset.CloudAccount.AccountId)
}

func (g *GcpMonitoringAsset) GetData() any {
	return g.Asset
}

func (g *GcpMonitoringAsset) GetIds() []string {
	return []string{g.buildId()}
}

func (g *GcpMonitoringAsset) GetElasticCommonData() (map[string]any, error) {
	return nil, nil
}
