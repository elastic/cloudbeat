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

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/rds"
	"github.com/elastic/elastic-agent-libs/logp"
)

type RdsFetcher struct {
	log        *logp.Logger
	cfg        RdsFetcherConfig
	resourceCh chan fetching.ResourceInfo
	provider   rds.Rds
}

type RdsFetcherConfig struct {
	fetching.AwsBaseFetcherConfig `config:",inline"`
}

type RdsResource struct {
	dbInstance awslib.AwsResource
}

func (f *RdsFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Info("Starting RdsFetcher.Fetch")
	dbInstances, err := f.provider.DescribeDBInstances(ctx)
	if err != nil {
		f.log.Errorf("failed to load some DB instances from rds: %v", err)
	}

	for _, dbInstance := range dbInstances {
		resource := RdsResource{dbInstance}
		f.log.Debugf("Fetched DB instance: %s", dbInstance.GetResourceName())
		f.resourceCh <- fetching.ResourceInfo{
			Resource:      resource,
			CycleMetadata: cMetadata,
		}
	}

	return nil
}

func (f *RdsFetcher) Stop() {}

func (r RdsResource) GetData() interface{} {
	return r.dbInstance
}

func (r RdsResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:      r.dbInstance.GetResourceArn(),
		Type:    fetching.CloudDatabase,
		SubType: r.dbInstance.GetResourceType(),
		Name:    r.dbInstance.GetResourceName(),
		Region:  r.dbInstance.GetRegion(),
	}, nil
}

func (r RdsResource) GetElasticCommonData() any { return nil }
