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

package factory

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/aws_cis/logging"
	"github.com/elastic/cloudbeat/resources/providers/aws_cis/monitoring"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudtrail"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudwatch"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudwatch/logs"
	"github.com/elastic/cloudbeat/resources/providers/awslib/configservice"
	"github.com/elastic/cloudbeat/resources/providers/awslib/ec2"
	"github.com/elastic/cloudbeat/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/resources/providers/awslib/kms"
	"github.com/elastic/cloudbeat/resources/providers/awslib/rds"
	"github.com/elastic/cloudbeat/resources/providers/awslib/s3"
	"github.com/elastic/cloudbeat/resources/providers/awslib/securityhub"
	"github.com/elastic/cloudbeat/resources/providers/awslib/sns"
	"github.com/elastic/elastic-agent-libs/logp"
)

func NewCisAwsFactory(log *logp.Logger, cfg aws.Config, ch chan fetching.ResourceInfo, identity *awslib.Identity) (FetchersMap, error) {
	log.Infof("Initializing AWS fetchers")
	if identity == nil {
		return nil, fmt.Errorf("NewCisAwsFactory was not provided with an aws identity")
	}

	m := make(FetchersMap)
	iamProvider := iam.NewIAMProvider(log, cfg, &awslib.MultiRegionClientFactory[iam.AccessAnalyzerClient]{})
	iamFetcher := fetchers.NewIAMFetcher(log, iamProvider, ch, identity)
	m[fetching.IAMType] = RegisteredFetcher{Fetcher: iamFetcher}

	kmsProvider := kms.NewKMSProvider(log, cfg, &awslib.MultiRegionClientFactory[kms.Client]{})
	kmsFetcher := fetchers.NewKMSFetcher(log, kmsProvider, ch)
	m[fetching.KmsType] = RegisteredFetcher{Fetcher: kmsFetcher}

	loggingProvider := logging.NewProvider(log, cfg, &awslib.MultiRegionClientFactory[cloudtrail.Client]{}, &awslib.MultiRegionClientFactory[s3.Client]{}, *identity.Account)
	configserviceProvider := configservice.NewProvider(log, cfg, &awslib.MultiRegionClientFactory[configservice.Client]{}, *identity.Account)
	loggingFetcher := fetchers.NewLoggingFetcher(log, loggingProvider, configserviceProvider, ch, identity)
	m[fetching.TrailType] = RegisteredFetcher{Fetcher: loggingFetcher}

	monitoringProvider := monitoring.NewProvider(
		log,
		cfg,
		&awslib.MultiRegionClientFactory[cloudtrail.Client]{},
		&awslib.MultiRegionClientFactory[cloudwatch.Client]{},
		&awslib.MultiRegionClientFactory[logs.Client]{},
		&awslib.MultiRegionClientFactory[sns.Client]{},
	)

	securityHubProvider := securityhub.NewProvider(log, cfg, &awslib.MultiRegionClientFactory[securityhub.Client]{}, *identity.Account)
	monitoringFetcher := fetchers.NewMonitoringFetcher(log, monitoringProvider, securityHubProvider, ch, identity)
	m[fetching.MonitoringType] = RegisteredFetcher{Fetcher: monitoringFetcher}

	ec2Provider := ec2.NewEC2Provider(log, *identity.Account, cfg, &awslib.MultiRegionClientFactory[ec2.Client]{})
	networkFetcher := fetchers.NewNetworkFetcher(log, ec2Provider, ch, identity)
	m[fetching.EC2NetworkingType] = RegisteredFetcher{Fetcher: networkFetcher}

	rdsProvider := rds.NewProvider(log, cfg, &awslib.MultiRegionClientFactory[rds.Client]{}, ec2Provider)
	rdsFetcher := fetchers.NewRdsFetcher(log, rdsProvider, ch)
	m[fetching.RdsType] = RegisteredFetcher{Fetcher: rdsFetcher}

	s3Provider := s3.NewProvider(log, cfg, &awslib.MultiRegionClientFactory[s3.Client]{}, *identity.Account)
	s3Fetcher := fetchers.NewS3Fetcher(log, s3Provider, ch)
	m[fetching.S3Type] = RegisteredFetcher{Fetcher: s3Fetcher}

	return m, nil
}
