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

	"github.com/elastic/cloudbeat/resources/providers/aws_cis/logging"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/configservice"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
)

type LoggingFetcher struct {
	log                   *logp.Logger
	loggingProvider       logging.Client
	configserviceProvider configservice.ConfigService
	cfg                   fetching.AwsBaseFetcherConfig
	resourceCh            chan fetching.ResourceInfo
	cloudIdentity         *awslib.Identity
}

type LoggingResource struct {
	awslib.AwsResource
}

type ConfigResource struct {
	configs  []awslib.AwsResource
	identity *awslib.Identity
}

func (f LoggingFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Debug("Starting LoggingFetcher.Fetch")
	trails, err := f.loggingProvider.DescribeTrails(ctx)
	if err != nil {
		f.log.Errorf("failed to describe trails: %v", err)
	}

	for _, resource := range trails {
		f.resourceCh <- fetching.ResourceInfo{
			Resource: LoggingResource{
				AwsResource: resource,
			},
			CycleMetadata: cMetadata,
		}
	}

	configs, err := f.configserviceProvider.DescribeConfigRecorders(ctx)
	if err != nil {
		f.log.Errorf("failed to describe config recorders: %v", err)
		return nil
	}

	f.resourceCh <- fetching.ResourceInfo{
		Resource:      ConfigResource{configs: configs, identity: f.cloudIdentity},
		CycleMetadata: cMetadata,
	}

	return nil
}

func (f LoggingFetcher) Stop() {}

func (r LoggingResource) GetData() any {
	return r.AwsResource
}

func (r LoggingResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:      r.GetResourceArn(),
		Type:    fetching.CloudAudit,
		SubType: r.GetResourceType(),
		Name:    r.GetResourceName(),
		Region:  r.GetRegion(),
	}, nil
}
func (r LoggingResource) GetElasticCommonData() any { return nil }

func (c ConfigResource) GetMetadata() (fetching.ResourceMetadata, error) {
	id := fmt.Sprintf("configservice-%s", *c.identity.Account)
	return fetching.ResourceMetadata{
		ID:      id,
		Type:    fetching.CloudConfig,
		SubType: fetching.ConfigServiceResourceType,
		Name:    id,
	}, nil
}

func (c ConfigResource) GetData() any {
	return c.configs
}

func (c ConfigResource) GetElasticCommonData() any { return nil }
