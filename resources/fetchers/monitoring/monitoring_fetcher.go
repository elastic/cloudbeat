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

package monitoring

import (
	"context"

	"github.com/elastic/cloudbeat/resources/providers/aws_cis/monitoring"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/securityhub"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
)

// Type fetcher
const Type = "aws-monitoring"

type MonitoringFetcher struct {
	log           *logp.Logger
	provider      monitoring.Client
	cfg           MonitoringFetcherConfig
	resourceCh    chan fetching.ResourceInfo
	cloudIdentity *awslib.Identity
	securityhub   securityhub.Service
}

type MonitoringFetcherConfig struct {
	fetching.AwsBaseFetcherConfig `config:",inline"`
}

func New(options ...Option) *MonitoringFetcher {
	f := &MonitoringFetcher{}
	for _, opt := range options {
		opt(f)
	}
	return f
}

func (m MonitoringFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	m.log.Debug("Starting MonitoringFetcher.Fetch")
	out, err := m.provider.AggregateResources(ctx)
	if err != nil {
		m.log.Errorf("failed to aggregate monitoring resources: %v", err)
	}
	if out != nil {
		m.resourceCh <- fetching.ResourceInfo{
			Resource:      MonitoringResource{*out, m.cloudIdentity},
			CycleMetadata: cMetadata,
		}
	}
	hubs, err := m.securityhub.Describe(ctx)
	if err != nil {
		m.log.Errorf("failed to describe security hub: %v", err)
		return nil
	}

	for _, hub := range hubs {
		m.resourceCh <- fetching.ResourceInfo{
			Resource: SecurityHubResource{
				SecurityHub: hub,
			},
			CycleMetadata: cMetadata,
		}
	}

	return nil
}

func (f MonitoringFetcher) Stop() {}
