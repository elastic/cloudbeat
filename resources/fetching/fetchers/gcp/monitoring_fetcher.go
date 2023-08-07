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

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/gcplib"
	"github.com/elastic/cloudbeat/resources/providers/gcplib/inventory"
)

type GcpMonitoringFetcher struct {
	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
	provider   inventory.ServiceAPI
}

type GcpMonitoringAsset struct {
	Type    string
	subType string

	Asset *inventory.MonitoringAsset `json:"assets,omitempty"`
}

func NewGcpMonitoringFetcher(_ context.Context, log *logp.Logger, ch chan fetching.ResourceInfo, provider inventory.ServiceAPI) *GcpMonitoringFetcher {
	return &GcpMonitoringFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

var monitoringAssetTypes = map[string][]string{
	"LogMetric":   {"logging.googleapis.com/LogMetric"},
	"AlertPolicy": {"monitoring.googleapis.com/AlertPolicy"},
}

func (f *GcpMonitoringFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Info("Starting GcpMonitoringFetcher.Fetch")

	monitoringAsset, err := f.provider.ListMonitoringAssets(monitoringAssetTypes)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		f.log.Info("GcpMonitoringFetcher.ListMonitoringAssets context canceled")
		return nil
	default:
		select {
		case <-ctx.Done():
			return nil
		case f.resourceCh <- fetching.ResourceInfo{
			CycleMetadata: cMetadata,
			Resource: &GcpMonitoringAsset{
				Type:    fetching.MonitoringIdentity,
				subType: fetching.GcpMonitoringType,
				Asset:   monitoringAsset,
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
	id := fmt.Sprintf("%s-%s", g.subType, g.Asset.ProjectId)
	return fetching.ResourceMetadata{
		ID:      id,
		Type:    g.Type,
		SubType: g.subType,
		Name:    id,
		Region:  gcplib.GlobalRegion,
	}, nil
}

func (g *GcpMonitoringAsset) GetData() any {
	return g.Asset
}

func (g *GcpMonitoringAsset) GetElasticCommonData() any { return nil }
