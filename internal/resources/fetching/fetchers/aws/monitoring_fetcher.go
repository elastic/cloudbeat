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

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/aws_cis/monitoring"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/securityhub"
)

type MonitoringFetcher struct {
	log           *clog.Logger
	provider      monitoring.Client
	resourceCh    chan fetching.ResourceInfo
	cloudIdentity *cloud.Identity
	securityhub   securityhub.Service
}

type MonitoringResource struct {
	monitoring.Resource
	identity *cloud.Identity
}

type SecurityHubResource struct {
	securityhub.SecurityHub
}

func NewMonitoringFetcher(log *clog.Logger, provider monitoring.Client, securityHubProvider securityhub.Service, ch chan fetching.ResourceInfo, identity *cloud.Identity) *MonitoringFetcher {
	return &MonitoringFetcher{
		log:           log,
		provider:      provider,
		securityhub:   securityHubProvider,
		resourceCh:    ch,
		cloudIdentity: identity,
	}
}

func (m MonitoringFetcher) Fetch(ctx context.Context, cycleMetadata cycle.Metadata) error {
	m.log.Debug("Starting MonitoringFetcher.Fetch")
	out, err := m.provider.AggregateResources(ctx)
	if err != nil {
		m.log.Errorf(ctx, "failed to aggregate monitoring resources: %v", err)
	}
	if out != nil {
		m.resourceCh <- fetching.ResourceInfo{
			Resource:      MonitoringResource{*out, m.cloudIdentity},
			CycleMetadata: cycleMetadata,
		}
	}
	hubs, err := m.securityhub.Describe(ctx)
	if err != nil {
		m.log.Errorf(ctx, "failed to describe security hub: %v", err)
		return nil
	}

	for _, hub := range hubs {
		m.resourceCh <- fetching.ResourceInfo{
			Resource: SecurityHubResource{
				SecurityHub: hub,
			},
			CycleMetadata: cycleMetadata,
		}
	}

	return nil
}

func (m MonitoringFetcher) Stop() {}

func (r MonitoringResource) GetData() any {
	return r
}

func (r MonitoringResource) GetIds() []string {
	return []string{}
}

func (r MonitoringResource) GetMetadata() (fetching.ResourceMetadata, error) {
	id := fmt.Sprintf("cloudtrail-%s", r.identity.Account)
	return fetching.ResourceMetadata{
		ID:      id,
		Type:    fetching.MonitoringIdentity,
		SubType: fetching.MultiTrailsType,
		Name:    id,
		Region:  awslib.GlobalRegion,
	}, nil
}
func (r MonitoringResource) GetElasticCommonData() (map[string]any, error) { return nil, nil }

func (s SecurityHubResource) GetData() any {
	return s
}

func (s SecurityHubResource) GetIds() []string {
	return []string{s.GetResourceArn()}
}

func (s SecurityHubResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:      s.GetResourceArn(),
		Name:    s.GetResourceName(),
		Type:    fetching.MonitoringIdentity,
		SubType: fetching.SecurityHubType,
		Region:  s.GetRegion(),
	}, nil
}

func (s SecurityHubResource) GetElasticCommonData() (map[string]any, error) { return nil, nil }
