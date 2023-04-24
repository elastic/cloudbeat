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
	"github.com/elastic/cloudbeat/resources/providers/awslib/ec2"
	"testing"

	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/rds"
	"github.com/stretchr/testify/mock"

	agentConfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
)

type RdsFactoryTestSuite struct {
	suite.Suite

	log *logp.Logger
}

func TestRdsFactoryTestSuite(t *testing.T) {
	s := new(RdsFactoryTestSuite)
	s.log = logp.NewLogger("cloudbeat_rds_factory_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *RdsFactoryTestSuite) SetupTest() {}

func (s *RdsFactoryTestSuite) TestCreateFetcher() {
	tests := []struct {
		config string
	}{
		{
			`
name: aws-rds
access_key_id: key
secret_access_key: secret
`,
		},
	}

	for _, test := range tests {
		mockCrossRegionFetcher := &awslib.MockCrossRegionFetcher[rds.Client]{}
		mockCrossRegionFetcher.On("GetMultiRegionsClientMap").Return(nil)

		mockCrossRegionFactory := &awslib.MockCrossRegionFactory[rds.Client]{}
		mockCrossRegionFactory.On(
			"NewMultiRegionClients",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(mockCrossRegionFetcher)

		mockEc2CrossRegionFetcher := &awslib.MockCrossRegionFetcher[ec2.Client]{}
		mockEc2CrossRegionFetcher.On("GetMultiRegionsClientMap").Return(nil)

		mockEc2CrossRegionFactory := &awslib.MockCrossRegionFactory[ec2.Client]{}
		mockEc2CrossRegionFactory.On(
			"NewMultiRegionClients",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(mockEc2CrossRegionFetcher)

		identity := awslib.Identity{
			Account: aws.String("123456789012"),
		}
		identityProvider := &awslib.MockIdentityProviderGetter{}
		identityProvider.EXPECT().GetIdentity(mock.Anything).Return(&identity, nil)

		factory := &RdsFactory{
			CrossRegionFactory:    mockCrossRegionFactory,
			Ec2CrossRegionFactory: mockEc2CrossRegionFactory,
			IdentityProvider: func(cfg aws.Config) awslib.IdentityProviderGetter {
				return identityProvider
			},
		}

		cfg, err := agentConfig.NewConfigFrom(test.config)
		s.NoError(err)

		fetcher, err := factory.Create(s.log, cfg, nil)
		s.NoError(err)
		s.NotNil(fetcher)

		rdsFetcher, ok := fetcher.(*RdsFetcher)
		s.True(ok)
		s.Equal("key", rdsFetcher.cfg.AwsConfig.AccessKeyID)
		s.Equal("secret", rdsFetcher.cfg.AwsConfig.SecretAccessKey)
	}
}
