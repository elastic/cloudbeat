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

type GcpLogSinkFetcher struct {
	log        *clog.Logger
	resourceCh chan fetching.ResourceInfo
	provider   inventory.ServiceAPI
}

type GcpLoggingAsset struct {
	Type    string
	subType string

	Asset *LoggingAsset `json:"asset,omitempty"`
}

type LoggingAsset struct {
	CloudAccount *fetching.CloudAccountMetadata
	LogSinks     []*inventory.ExtendedGcpAsset `json:"log_sinks,omitempty"`
}

func NewGcpLogSinkFetcher(_ context.Context, log *clog.Logger, ch chan fetching.ResourceInfo, provider inventory.ServiceAPI) *GcpLogSinkFetcher {
	return &GcpLogSinkFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

func (f *GcpLogSinkFetcher) Fetch(ctx context.Context, cycleMetadata cycle.Metadata) error {
	f.log.Info("Starting GcpLogSinkFetcher.Fetch")
	defer f.log.Info("GcpLogSinkFetcher.Fetch done")

	resultsCh := make(chan *inventory.ProjectAssets)
	go f.provider.ListProjectAssets(ctx, []string{inventory.LogSinkAssetType}, resultsCh)

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
				Resource: &GcpLoggingAsset{
					Type:    fetching.LoggingIdentity,
					subType: fetching.GcpLoggingType,
					Asset:   &LoggingAsset{asset.CloudAccount, asset.Assets},
				},
			}
		}
	}
}

func (f *GcpLogSinkFetcher) Stop() {
	f.provider.Close()
}

func (g *GcpLoggingAsset) GetMetadata() (fetching.ResourceMetadata, error) {
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

func (g *GcpLoggingAsset) buildId() string {
	return fmt.Sprintf("%s-%s", g.subType, g.Asset.CloudAccount.AccountId)
}

func (g *GcpLoggingAsset) GetData() any {
	return g.Asset
}

func (g *GcpLoggingAsset) GetIds() []string {
	return []string{g.buildId()}
}

func (g *GcpLoggingAsset) GetElasticCommonData() (map[string]any, error) {
	return nil, nil
}
