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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudtrail"
	"github.com/elastic/cloudbeat/resources/providers/awslib/configservice"
	"github.com/elastic/cloudbeat/resources/providers/awslib/s3"
	"testing"

	"github.com/elastic/cloudbeat/resources/providers/awslib"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoggingFactory_Create(t *testing.T) {
	var config = `
name: aws-trail
access_key_id: key
secret_access_key: secret
session_token: session
`
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

	mockCrossRegionS3Fetcher := &awslib.MockCrossRegionFetcher[s3.Client]{}
	mockCrossRegionS3Fetcher.On("GetMultiRegionsClientMap").Return(nil)

	mockCrossRegionS3Factory := &awslib.MockCrossRegionFactory[s3.Client]{}
	mockCrossRegionS3Factory.On(
		"NewMultiRegionClients",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(mockCrossRegionS3Fetcher)

	mockCrossRegionConfigFetcher := &awslib.MockCrossRegionFetcher[configservice.Client]{}
	mockCrossRegionConfigFetcher.On("GetMultiRegionsClientMap").Return(nil)

	mockCrossRegionConfigFactory := &awslib.MockCrossRegionFactory[configservice.Client]{}
	mockCrossRegionConfigFactory.On(
		"NewMultiRegionClients",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(mockCrossRegionConfigFetcher)

	identity := awslib.Identity{
		Account: aws.String("123456789012"),
	}
	identityProvider := &awslib.MockIdentityProviderGetter{}
	identityProvider.EXPECT().GetIdentity(mock.Anything).Return(&identity, nil)

	f := &LoggingFactory{
		TrailCrossRegionFactory:  mockCrossRegionTrailFactory,
		S3CrossRegionFactory:     mockCrossRegionS3Factory,
		ConfigCrossRegionFactory: mockCrossRegionConfigFactory,
		IdentityProvider: func(cfg aws.Config) awslib.IdentityProviderGetter {
			return identityProvider
		},
	}

	cfg, err := agentconfig.NewConfigFrom(config)
	assert.NoError(t, err)
	fetcher, err := f.Create(logp.NewLogger("logging-factory-test"), cfg, nil)
	assert.NoError(t, err)
	assert.NotNil(t, fetcher)

	nacl, ok := fetcher.(*LoggingFetcher)
	assert.True(t, ok)
	assert.Equal(t, nacl.cfg.AwsConfig.AccessKeyID, "key")
	assert.Equal(t, nacl.cfg.AwsConfig.SecretAccessKey, "secret")
}
