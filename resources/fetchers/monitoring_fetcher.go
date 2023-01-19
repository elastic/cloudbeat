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

	"github.com/elastic/cloudbeat/resources/providers/aws_cis"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
)

type MonitoringFetcher struct {
	log        *logp.Logger
	provider   aws_cis.Client
	cfg        MonitoringFetcherConfig
	resourceCh chan fetching.ResourceInfo
}

type MonitoringFetcherConfig struct {
	fetching.AwsBaseFetcherConfig `config:",inline"`
}

type MonitoringResource struct {
	aws_cis.Rule41Output
}

func (m MonitoringFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	m.log.Debug("Starting MonitoringFetcher.Fetch")
	out, err := m.provider.Rule41(ctx)
	if err != nil {
		return err
	}
	m.resourceCh <- fetching.ResourceInfo{
		Resource:      MonitoringResource{out},
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
	item := r.Items[0]
	return fetching.ResourceMetadata{
		ID:      *item.TrailInfo.Trail.TrailARN,
		Type:    fetching.MonitoringIdentity,
		SubType: fetching.TrailType,
		Name:    *item.TrailInfo.Trail.Name,
	}, nil
}
func (r MonitoringResource) GetElasticCommonData() any { return nil }
