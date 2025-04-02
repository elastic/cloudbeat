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

type GcpNetworksFetcher struct {
	log        *clog.Logger
	resourceCh chan fetching.ResourceInfo
	provider   inventory.ServiceAPI
}

type GcpNetworksAsset struct {
	Type    string
	subType string

	Asset *inventory.ExtendedGcpAsset `json:"assets,omitempty"`
}

func NewGcpNetworksFetcher(_ context.Context, log *clog.Logger, ch chan fetching.ResourceInfo, provider inventory.ServiceAPI) *GcpNetworksFetcher {
	return &GcpNetworksFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

func (f *GcpNetworksFetcher) Fetch(ctx context.Context, cycleMetadata cycle.Metadata) error {
	f.log.Info("Starting GcpNetworksFetcher.Fetch")
	defer f.log.Info("GcpNetworksFetcher.Fetch done")

	resultsCh := make(chan *inventory.ExtendedGcpAsset)
	go f.provider.ListNetworkAssets(ctx, resultsCh)

	for {
		select {
		case <-ctx.Done():
			return nil
		case asset, ok := <-resultsCh:
			if !ok {
				return nil
			}
			f.resourceCh <- fetching.ResourceInfo{
				CycleMetadata: cycleMetadata,
				Resource: &GcpNetworksAsset{
					Type:    fetching.ProjectManagement,
					subType: fetching.GcpPolicies,
					Asset:   asset,
				},
			}
		}
	}
}

func (f *GcpNetworksFetcher) Stop() {
	f.provider.Close()
}

func (g *GcpNetworksAsset) GetMetadata() (fetching.ResourceMetadata, error) {
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

func (g *GcpNetworksAsset) buildId() string {
	id := fmt.Sprintf("%s-%s", g.subType, g.Asset.CloudAccount.AccountId)
	return id
}

func (g *GcpNetworksAsset) GetData() any {
	return g.Asset
}

func (g *GcpNetworksAsset) GetIds() []string {
	return []string{g.buildId()}
}

func (g *GcpNetworksAsset) GetElasticCommonData() (map[string]any, error) {
	return nil, nil
}
