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
	"github.com/elastic/cloudbeat/resources/fetchersManager"
	"github.com/elastic/cloudbeat/resources/providers/aws_cis/monitoring"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudtrail"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudwatch"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudwatch/logs"
	"github.com/elastic/cloudbeat/resources/providers/awslib/securityhub"
	"github.com/elastic/cloudbeat/resources/providers/awslib/sns"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
)

func init() {
	fetchersManager.Factories.RegisterFactory(fetching.MonitoringType, &MonitoringFactory{
		AwsConfigProvider:                awslib.ConfigProvider{MetadataProvider: awslib.Ec2MetadataProvider{}},
		TrailCrossRegionFactory:          &awslib.MultiRegionClientFactory[cloudtrail.Client]{},
		CloudwatchCrossRegionFactory:     &awslib.MultiRegionClientFactory[cloudwatch.Client]{},
		CloudwatchlogsCrossRegionFactory: &awslib.MultiRegionClientFactory[logs.Client]{},
		SNSCrossRegionFactory:            &awslib.MultiRegionClientFactory[sns.Client]{},
		SecurityhubRegionFactory:         &awslib.MultiRegionClientFactory[securityhub.Client]{},
		IdentityProvider:                 awslib.GetIdentityClient,
	})
}

type MonitoringFactory struct {
	AwsConfigProvider                awslib.ConfigProviderAPI
	TrailCrossRegionFactory          awslib.CrossRegionFactory[cloudtrail.Client]
	CloudwatchCrossRegionFactory     awslib.CrossRegionFactory[cloudwatch.Client]
	CloudwatchlogsCrossRegionFactory awslib.CrossRegionFactory[logs.Client]
	SNSCrossRegionFactory            awslib.CrossRegionFactory[sns.Client]
	SecurityhubRegionFactory         awslib.CrossRegionFactory[securityhub.Client]
	IdentityProvider                 func(cfg awssdk.Config) awslib.IdentityProviderGetter
}

func (f *MonitoringFactory) Create(log *logp.Logger, c *agentconfig.C, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	log.Debug("Starting MonitoringFactory.Create")

	cfg := MonitoringFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(log, cfg, ch)
}

func (f *MonitoringFactory) CreateFrom(log *logp.Logger, cfg MonitoringFetcherConfig, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	awsConfig, err := aws.InitializeAWSConfig(cfg.AwsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	provider := monitoring.Provider{
		Cloudtrail:     cloudtrail.NewProvider(awsConfig, log, f.TrailCrossRegionFactory),
		Cloudwatch:     cloudwatch.NewProvider(log, awsConfig, f.CloudwatchCrossRegionFactory),
		Cloudwatchlogs: logs.NewCloudwatchLogsProvider(log, awsConfig, f.CloudwatchlogsCrossRegionFactory),
		Sns:            sns.NewSNSProvider(log, awsConfig, f.SNSCrossRegionFactory),
		Log:            log,
	}

	identityProvider := f.IdentityProvider(awsConfig)
	identity, err := identityProvider.GetIdentity(context.Background())
	if err != nil {
		return nil, fmt.Errorf("could not get cloud indentity: %w", err)
	}

	return &MonitoringFetcher{
		log:           log,
		cfg:           cfg,
		provider:      &provider,
		resourceCh:    ch,
		cloudIdentity: identity,
		securityhub:   securityhub.NewProvider(awsConfig, log, f.SecurityhubRegionFactory),
	}, nil
}
