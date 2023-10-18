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

type GcpPoliciesFetcher struct {
	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
	provider   inventory.ServiceAPI
}

type GcpPoliciesAsset struct {
	Type    string
	subType string

	Asset *inventory.ProjectPoliciesAsset `json:"assets,omitempty"`
}

func NewGcpPoliciesFetcher(_ context.Context, log *logp.Logger, ch chan fetching.ResourceInfo, provider inventory.ServiceAPI) *GcpPoliciesFetcher {
	return &GcpPoliciesFetcher{
		log:        log,
		resourceCh: ch,
		provider:   provider,
	}
}

func (f *GcpPoliciesFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Info("Starting GcpPoliciesFetcher.Fetch")

	projectsAssets, err := f.provider.ListProjectsAncestorsPolicies(ctx)
	if err != nil {
		return err
	}

	for _, projectPolicies := range projectsAssets {
		select {
		case <-ctx.Done():
			f.log.Infof("GcpPoliciesFetcher context err: %s", ctx.Err().Error())
			return nil
		case f.resourceCh <- fetching.ResourceInfo{
			CycleMetadata: cMetadata,
			Resource: &GcpPoliciesAsset{
				Type:    fetching.ProjectManagement,
				subType: fetching.GcpPolicies,
				Asset:   projectPolicies,
			},
		}:
		}
	}
	return nil
}

func (f *GcpPoliciesFetcher) Stop() {
	f.provider.Close()
}

func (g *GcpPoliciesAsset) GetMetadata() (fetching.ResourceMetadata, error) {
	id := fmt.Sprintf("%s-%s", g.subType, g.Asset.Ecs.ProjectId)
	return fetching.ResourceMetadata{
		ID:      id,
		Type:    g.Type,
		SubType: g.subType,
		Name:    id,
		Region:  gcplib.GlobalRegion,
	}, nil
}

func (g *GcpPoliciesAsset) GetData() any {
	return g.Asset.Policies
}

func (g *GcpPoliciesAsset) GetElasticCommonData() (map[string]any, error) {
	return map[string]any{
		"cloud": map[string]any{
			"provider": "gcp",
			"account": map[string]any{
				"id":   g.Asset.Ecs.ProjectId,
				"name": g.Asset.Ecs.ProjectName,
			},
			"Organization": map[string]any{
				"id":   g.Asset.Ecs.OrganizationId,
				"name": g.Asset.Ecs.OrganizationName,
			},
		},
	}, nil
}
