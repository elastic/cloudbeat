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

package preset

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	fetchers "github.com/elastic/cloudbeat/internal/resources/fetching/fetchers/aws"
	"github.com/elastic/cloudbeat/internal/resources/fetching/registry"
	"github.com/elastic/cloudbeat/internal/resources/providers/aws_cis/logging"
	"github.com/elastic/cloudbeat/internal/resources/providers/aws_cis/monitoring"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/cloudtrail"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/cloudwatch"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/cloudwatch/logs"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/configservice"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/ec2"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/kms"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/rds"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/s3"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/securityhub"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/sns"
)

type ErrorPublisher interface {
	Publish(ctx context.Context, err error)
}

func NewCisAwsFetchers(ctx context.Context, log *clog.Logger, cfg aws.Config, ch chan fetching.ResourceInfo, identity *cloud.Identity, errorPublisher ErrorPublisher) registry.FetchersMap {
	log.Infof("Initializing AWS fetchers for account: '%s'", identity.Account)

	m := make(registry.FetchersMap)
	iamProvider := iam.NewIAMProvider(ctx, log, cfg, &awslib.MultiRegionClientFactory[iam.AccessAnalyzerClient]{})
	iamFetcher := fetchers.NewIAMFetcher(log, iamProvider, ch, identity, errorPublisher)
	m[fetching.IAMType] = registry.RegisteredFetcher{Fetcher: iamFetcher}

	kmsProvider := kms.NewKMSProvider(ctx, log, cfg, &awslib.MultiRegionClientFactory[kms.Client]{})
	kmsFetcher := fetchers.NewKMSFetcher(log, kmsProvider, ch, errorPublisher)
	m[fetching.KmsType] = registry.RegisteredFetcher{Fetcher: kmsFetcher}

	loggingProvider := logging.NewProvider(ctx, log, cfg, &awslib.MultiRegionClientFactory[cloudtrail.Client]{}, &awslib.MultiRegionClientFactory[s3.Client]{}, identity.Account)
	configserviceProvider := configservice.NewProvider(ctx, log, cfg, &awslib.MultiRegionClientFactory[configservice.Client]{}, identity.Account)
	loggingFetcher := fetchers.NewLoggingFetcher(log, loggingProvider, configserviceProvider, ch, identity, errorPublisher)
	m[fetching.TrailType] = registry.RegisteredFetcher{Fetcher: loggingFetcher}

	monitoringProvider := monitoring.NewProvider(
		ctx,
		log,
		cfg,
		&awslib.MultiRegionClientFactory[cloudtrail.Client]{},
		&awslib.MultiRegionClientFactory[cloudwatch.Client]{},
		&awslib.MultiRegionClientFactory[logs.Client]{},
		&awslib.MultiRegionClientFactory[sns.Client]{},
	)

	securityHubProvider := securityhub.NewProvider(ctx, log, cfg, &awslib.MultiRegionClientFactory[securityhub.Client]{}, identity.Account)
	monitoringFetcher := fetchers.NewMonitoringFetcher(log, monitoringProvider, securityHubProvider, ch, identity, errorPublisher)
	m[fetching.AwsMonitoringType] = registry.RegisteredFetcher{Fetcher: monitoringFetcher}

	ec2Provider := ec2.NewEC2Provider(ctx, log, identity.Account, cfg, &awslib.MultiRegionClientFactory[ec2.Client]{})
	networkFetcher := fetchers.NewNetworkFetcher(log, ec2Provider, ch, errorPublisher)
	m[fetching.EC2NetworkingType] = registry.RegisteredFetcher{Fetcher: networkFetcher}

	rdsProvider := rds.NewProvider(ctx, log, cfg, &awslib.MultiRegionClientFactory[rds.Client]{}, ec2Provider)
	rdsFetcher := fetchers.NewRdsFetcher(log, rdsProvider, ch, errorPublisher)
	m[fetching.RdsType] = registry.RegisteredFetcher{Fetcher: rdsFetcher}

	s3Provider := s3.NewProvider(ctx, log, cfg, &awslib.MultiRegionClientFactory[s3.Client]{}, identity.Account)
	s3Fetcher := fetchers.NewS3Fetcher(log, s3Provider, ch)
	m[fetching.S3Type] = registry.RegisteredFetcher{Fetcher: s3Fetcher}

	return m
}
