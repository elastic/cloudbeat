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

	"github.com/elastic/cloudbeat/resources/fetchersManager"
	"github.com/elastic/cloudbeat/resources/providers/aws_cis"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudtrail"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudwatch"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudwatch/logs"
	"github.com/elastic/cloudbeat/resources/providers/awslib/sns"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
)

func init() {
	fetchersManager.Factories.RegisterFactory(fetching.MonitoringType, &MonitoringFactory{
		AwsConfigProvider: awslib.ConfigProvider{MetadataProvider: awslib.Ec2MetadataProvider{}},
	})
}

type MonitoringFactory struct {
	AwsConfigProvider awslib.ConfigProviderAPI
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
	ctx := context.Background()
	awsConfig, err := f.AwsConfigProvider.InitializeAWSConfig(ctx, cfg.AwsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	provider := aws_cis.Provider{
		Cloudtrail:     cloudtrail.NewCloudtrailProvider(log, *awsConfig),
		Cloudwatch:     cloudwatch.NewCloudwatchProvider(log, *awsConfig),
		Cloudwatchlogs: logs.NewCloudwatchLogsProvider(log, *awsConfig),
		Sns:            sns.NewSNSProvider(log, *awsConfig),
		Log:            log,
	}

	return &MonitoringFetcher{
		log:        log,
		cfg:        cfg,
		provider:   &provider,
		resourceCh: ch,
	}, nil
}
