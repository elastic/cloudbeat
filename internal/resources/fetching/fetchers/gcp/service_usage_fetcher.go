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

type GcpServiceUsageFetcher struct {
	log        *clog.Logger
	resourceCh chan fetching.ResourceInfo
	provider   inventory.ServiceAPI
}

type GcpServiceUsageAsset struct {
	Type    string
	subType string

	Asset *ServiceUsageAsset `json:"assets,omitempty"`
}

type ServiceUsageAsset struct {
	CloudAccount *fetching.CloudAccountMetadata
	Services     []*inventory.ExtendedGcpAsset `json:"services,omitempty"`
}

func NewGcpServiceUsageFetcher(_ context.Context, log *clog.Logger, ch chan fetching.ResourceInfo, provider inventory.ServiceAPI) *GcpServiceUsageFetcher {
	return &GcpServiceUsageFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

func (f *GcpServiceUsageFetcher) Fetch(ctx context.Context, cycleMetadata cycle.Metadata) error {
	f.log.Info("Starting GcpServiceUsageFetcher.Fetch")
	defer f.log.Info("GcpServiceUsageFetcher.Fetch done")
	defer f.provider.Clear()

	resultsCh := make(chan *inventory.ProjectAssets)
	go f.provider.ListProjectAssets(ctx, []string{inventory.ServiceUsageAssetType}, resultsCh)

	for asset := range resultsCh {
		select {
		case <-ctx.Done():
			f.log.Debugf("GcpServiceUsageFetcher.Fetch context done: %v", ctx.Err())
			return nil

		case f.resourceCh <- fetching.ResourceInfo{
			CycleMetadata: cycleMetadata,
			Resource: &GcpServiceUsageAsset{
				Type:    fetching.MonitoringIdentity,
				subType: fetching.GcpServiceUsage,
				Asset:   &ServiceUsageAsset{asset.CloudAccount, asset.Assets},
			},
		}:
		}
	}
	return nil
}

func (f *GcpServiceUsageFetcher) Stop() {
	f.provider.Close()
}

func (g *GcpServiceUsageAsset) GetMetadata() (fetching.ResourceMetadata, error) {
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

func (g *GcpServiceUsageAsset) buildId() string {
	return fmt.Sprintf("%s-%s", g.subType, g.Asset.CloudAccount.AccountId)
}

func (g *GcpServiceUsageAsset) GetData() any {
	return g.Asset
}

func (g *GcpServiceUsageAsset) GetIds() []string {
	return []string{g.buildId()}
}

func (g *GcpServiceUsageAsset) GetElasticCommonData() (map[string]any, error) {
	return nil, nil
}
