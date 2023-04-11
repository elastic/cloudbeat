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

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/cloudbeat/resources/providers/aws_cis/logging"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudtrail"
	"github.com/elastic/cloudbeat/resources/providers/awslib/configservice"
	"github.com/elastic/cloudbeat/resources/providers/awslib/s3"

	"github.com/elastic/cloudbeat/resources/fetchersManager"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
)

func init() {
	fetchersManager.Factories.RegisterFactory(fetching.TrailType, &LoggingFactory{
		TrailCrossRegionFactory:  &awslib.MultiRegionClientFactory[cloudtrail.Client]{},
		S3CrossRegionFactory:     &awslib.MultiRegionClientFactory[s3.Client]{},
		ConfigCrossRegionFactory: &awslib.MultiRegionClientFactory[configservice.Client]{},
		IdentityProvider:         awslib.GetIdentityClient,
	})
}

type LoggingFactory struct {
	TrailCrossRegionFactory  awslib.CrossRegionFactory[cloudtrail.Client]
	S3CrossRegionFactory     awslib.CrossRegionFactory[s3.Client]
	ConfigCrossRegionFactory awslib.CrossRegionFactory[configservice.Client]
	IdentityProvider         func(cfg awssdk.Config) awslib.IdentityProviderGetter
}

func (f *LoggingFactory) Create(log *logp.Logger, c *agentconfig.C, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	log.Debug("Starting LoggingFactory.Create")

	cfg := fetching.AwsBaseFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(log, cfg, ch)
}

func (f *LoggingFactory) CreateFrom(log *logp.Logger, cfg fetching.AwsBaseFetcherConfig, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	awsConfig, err := aws.InitializeAWSConfig(cfg.AwsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	identityProvider := f.IdentityProvider(awsConfig)
	identity, err := identityProvider.GetIdentity(context.Background())
	if err != nil {
		return nil, fmt.Errorf("could not get cloud indentity: %w", err)
	}

	return &LoggingFetcher{
		log:                   log,
		loggingProvider:       logging.NewProvider(log, awsConfig, f.TrailCrossRegionFactory, f.S3CrossRegionFactory, *identity.Account),
		configserviceProvider: configservice.NewProvider(log, awsConfig, f.ConfigCrossRegionFactory, *identity.Account),
		cfg:                   cfg,
		resourceCh:            ch,
		cloudIdentity:         identity,
	}, nil
}
