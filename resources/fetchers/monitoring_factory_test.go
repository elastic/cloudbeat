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
	"testing"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudtrail"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudwatch"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudwatch/logs"
	"github.com/elastic/cloudbeat/resources/providers/awslib/securityhub"
	"github.com/elastic/cloudbeat/resources/providers/awslib/sns"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMonitoringFactory_Create(t *testing.T) {
	awsconfig := &awslib.MockConfigProviderAPI{}
	awsconfig.EXPECT().InitializeAWSConfig(mock.Anything, mock.Anything).
		Call.
		Return(func(ctx context.Context, config aws.ConfigAWS) *awssdk.Config {
			return CreateSdkConfig(config, "us1-east")
		},
			func(ctx context.Context, config aws.ConfigAWS) error {
				return nil
			},
		)

	mockCrossRegionTrailFetcher := &awslib.MockCrossRegionFetcher[cloudtrail.Client]{}
	mockCrossRegionTrailFetcher.On("GetMultiRegionsClientMap").Return(nil)
	mockCrossRegionTrailFactory := &awslib.MockCrossRegionFactory[cloudtrail.Client]{}
	mockCrossRegionTrailFactory.On(
		"NewMultiRegionClients",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(mockCrossRegionTrailFetcher)

	mockCrossRegionCloudwatchFetcher := &awslib.MockCrossRegionFetcher[cloudwatch.Client]{}
	mockCrossRegionCloudwatchFetcher.On("GetMultiRegionsClientMap").Return(nil)
	mockCrossRegionCloudwatchFactory := &awslib.MockCrossRegionFactory[cloudwatch.Client]{}
	mockCrossRegionCloudwatchFactory.On(
		"NewMultiRegionClients",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(mockCrossRegionCloudwatchFetcher)

	mockCrossRegionCloudwatchlogsFetcher := &awslib.MockCrossRegionFetcher[logs.Client]{}
	mockCrossRegionCloudwatchlogsFetcher.On("GetMultiRegionsClientMap").Return(nil)
	mockCrossRegionCloudwatchlogsFactory := &awslib.MockCrossRegionFactory[logs.Client]{}
	mockCrossRegionCloudwatchlogsFactory.On(
		"NewMultiRegionClients",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(mockCrossRegionCloudwatchlogsFetcher)

	mockCrossRegionSNSFetcher := &awslib.MockCrossRegionFetcher[sns.Client]{}
	mockCrossRegionSNSFetcher.On("GetMultiRegionsClientMap").Return(nil)
	mockCrossRegionSNSFactory := &awslib.MockCrossRegionFactory[sns.Client]{}
	mockCrossRegionSNSFactory.On(
		"NewMultiRegionClients",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(mockCrossRegionSNSFetcher)

	mockSecurityhubService := &securityhub.MockService{}
	mockSecurityhubService.On("Describe").Return(securityhub.SecurityHub{}, nil)
	mockCrossRegionSecurityHubFetcher := &awslib.MockCrossRegionFetcher[securityhub.Client]{}
	mockCrossRegionSecurityHubFetcher.On("GetMultiRegionsClientMap").Return(nil)
	mockCrossRegionSecurityHubFactory := &awslib.MockCrossRegionFactory[securityhub.Client]{}
	mockCrossRegionSecurityHubFactory.On(
		"NewMultiRegionClients",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(mockCrossRegionSecurityHubFetcher)

	identity := &awslib.MockIdentityProviderGetter{}
	identity.EXPECT().GetIdentity(mock.Anything).Return(&awslib.Identity{
		Account: awssdk.String("test-account"),
	}, nil)

	f := &MonitoringFactory{
		AwsConfigProvider:                awsconfig,
		TrailCrossRegionFactory:          mockCrossRegionTrailFactory,
		CloudwatchCrossRegionFactory:     mockCrossRegionCloudwatchFactory,
		CloudwatchlogsCrossRegionFactory: mockCrossRegionCloudwatchlogsFactory,
		SNSCrossRegionFactory:            mockCrossRegionSNSFactory,
		SecurityhubRegionFactory:         mockCrossRegionSecurityHubFactory,
		IdentityProvider: func(cfg awssdk.Config) awslib.IdentityProviderGetter {
			return identity
		},
	}
	cfg, err := agentconfig.NewConfigFrom(awsConfig)
	assert.NoError(t, err)
	fetcher, err := f.Create(logp.NewLogger("test"), cfg, nil)
	assert.NoError(t, err)
	assert.NotNil(t, fetcher)
	monitoringFetcher, ok := fetcher.(*MonitoringFetcher)
	assert.True(t, ok)
	assert.Equal(t, monitoringFetcher.cfg.AwsConfig.AccessKeyID, "key")
	assert.Equal(t, monitoringFetcher.cfg.AwsConfig.SecretAccessKey, "secret")
}
