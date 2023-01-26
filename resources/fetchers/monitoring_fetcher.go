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

	"github.com/elastic/cloudbeat/resources/providers/aws_cis/monitoring"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
)

type MonitoringFetcher struct {
	log           *logp.Logger
	provider      monitoring.Client
	cfg           MonitoringFetcherConfig
	resourceCh    chan fetching.ResourceInfo
	cloudIdentity *awslib.Identity
}

type MonitoringFetcherConfig struct {
	fetching.AwsBaseFetcherConfig `config:",inline"`
}

type MonitoringResource struct {
	monitoring.Resource
	identity *awslib.Identity
}

func (m MonitoringFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	m.log.Debug("Starting MonitoringFetcher.Fetch")
	out, err := m.provider.AggregateResources(ctx)
	if err != nil {
		return err
	}
	m.resourceCh <- fetching.ResourceInfo{
		Resource:      MonitoringResource{*out, m.cloudIdentity},
		CycleMetadata: cMetadata,
	}
	return nil
}

func (f MonitoringFetcher) Stop() {}

func (r MonitoringResource) GetData() any {
	return r
}

func (r MonitoringResource) GetMetadata() (fetching.ResourceMetadata, error) {
	if len(r.Items) == 0 {
		return fetching.ResourceMetadata{}, nil
	}
	id := fmt.Sprintf("cloudtrail-%d", r.identity.Account)
	return fetching.ResourceMetadata{
		ID:      id,
		Type:    fetching.MonitoringIdentity,
		SubType: fetching.TrailType,
		Name:    id,
	}, nil
}
func (r MonitoringResource) GetElasticCommonData() any { return nil }
