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

package logging

import (
	"context"

	"github.com/elastic/cloudbeat/resources/providers/aws_cis/logging"
	"github.com/elastic/cloudbeat/resources/providers/awslib/configservice"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
)

// Type fetcher
const Type = "aws-trail"

type LoggingFetcher struct {
	log                   *logp.Logger
	loggingProvider       logging.Client
	configserviceProvider configservice.ConfigService
	cfg                   fetching.AwsBaseFetcherConfig
	resourceCh            chan fetching.ResourceInfo
}

func New(options ...Option) *LoggingFetcher {
	f := &LoggingFetcher{}
	for _, opt := range options {
		opt(f)
	}
	return f
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
	}

	for _, resource := range configs {
		f.resourceCh <- fetching.ResourceInfo{
			Resource:      ConfigResource{AwsResource: resource},
			CycleMetadata: cMetadata,
		}
	}

	return nil
}

func (f LoggingFetcher) Stop() {}
