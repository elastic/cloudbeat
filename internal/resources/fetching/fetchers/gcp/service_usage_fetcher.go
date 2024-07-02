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

	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/gcplib"
	"github.com/elastic/cloudbeat/internal/resources/providers/gcplib/inventory"
)

type GcpServiceUsageFetcher struct {
	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
	provider   inventory.ServiceAPI
}

type GcpServiceUsageAsset struct {
	Type    string
	subType string

	Asset *inventory.ServiceUsageAsset `json:"assets,omitempty"`
}

func NewGcpServiceUsageFetcher(_ context.Context, log *logp.Logger, ch chan fetching.ResourceInfo, provider inventory.ServiceAPI) *GcpServiceUsageFetcher {
	return &GcpServiceUsageFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

func (f *GcpServiceUsageFetcher) Fetch(ctx context.Context, cycleMetadata cycle.Metadata) error {
	f.log.Info("Starting GcpServiceUsageFetcher.Fetch")

	serviceUsageAssets, err := f.provider.ListServiceUsageAssets(ctx)
	if err != nil {
		return err
	}

	for _, serviceUsageAsset := range serviceUsageAssets {
		select {
		case <-ctx.Done():
			f.log.Infof("GcpServiceUsageFetcher.ListMonitoringAssets context err: %s", ctx.Err().Error())
			return nil
		case f.resourceCh <- fetching.ResourceInfo{
			CycleMetadata: cycleMetadata,
			Resource: &GcpServiceUsageAsset{
				Type:    fetching.MonitoringIdentity,
				subType: fetching.GcpServiceUsage,
				Asset:   serviceUsageAsset,
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
	id := fmt.Sprintf("%s-%s", g.subType, g.Asset.CloudAccount.AccountId)
	return fetching.ResourceMetadata{
		ID:                   id,
		Type:                 g.Type,
		SubType:              g.subType,
		Name:                 id,
		Region:               gcplib.GlobalRegion,
		CloudAccountMetadata: *g.Asset.CloudAccount,
	}, nil
}

func (g *GcpServiceUsageAsset) GetData() any {
	return g.Asset
}

func (g *GcpServiceUsageAsset) GetElasticCommonData() (map[string]any, error) {
	return nil, nil
}
