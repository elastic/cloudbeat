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

	NetworkAsset *inventory.ExtendedGcpAsset `json:"asset,omitempty"`
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
	defer f.provider.Clear()

	resultsCh := make(chan *inventory.ExtendedGcpAsset)
	go f.provider.ListNetworkAssets(ctx, resultsCh)

	for asset := range resultsCh {
		select {
		case <-ctx.Done():
			f.log.Debugf("GcpNetworksFetcher.Fetch context done: %v", ctx.Err())
			return nil

		case f.resourceCh <- fetching.ResourceInfo{
			CycleMetadata: cycleMetadata,
			Resource: &GcpNetworksAsset{
				Type:         fetching.CloudCompute,
				subType:      "gcp-compute-network",
				NetworkAsset: asset,
			},
		}:
		}
	}
	return nil
}

func (f *GcpNetworksFetcher) Stop() {
	f.provider.Close()
}

func (g *GcpNetworksAsset) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:                   g.NetworkAsset.Name,
		Type:                 g.Type,
		SubType:              g.subType,
		Name:                 getAssetResourceName(g.NetworkAsset),
		Region:               gcplib.GlobalRegion,
		CloudAccountMetadata: *g.NetworkAsset.CloudAccount,
	}, nil
}

func (g *GcpNetworksAsset) GetData() any {
	return g.NetworkAsset
}

func (g *GcpNetworksAsset) GetIds() []string {
	return []string{g.NetworkAsset.Name}
}

func (g *GcpNetworksAsset) GetElasticCommonData() (map[string]any, error) {
	return nil, nil
}
