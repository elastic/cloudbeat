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
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/stretchr/testify/mock"
	"testing"

	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
)

type IamFactoryTestSuite struct {
	suite.Suite

	log *logp.Logger
}

func TestIamFactoryTestSuite(t *testing.T) {
	s := new(IamFactoryTestSuite)
	s.log = logp.NewLogger("cloudbeat_iam_factory_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *IamFactoryTestSuite) SetupTest() {

}

func (s *IamFactoryTestSuite) TestCreateFetcher() {
	var tests = []struct {
		config  string
		account string
	}{
		{
			`
name: aws-iam
access_key_id: key
secret_access_key: secret
session_token: session
default_region: us1-east
`,
			"my_account",
		},
	}

	for _, test := range tests {
		mockedConfigGetter := &awslib.MockConfigProviderAPI{}
		mockedConfigGetter.EXPECT().
			InitializeAWSConfig(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Call.
			Return(func(ctx context.Context, cfg aws.ConfigAWS, log *logp.Logger, useDefaultRegion bool) awssdk.Config {

				return CreateSdkConfig(cfg, "us1-east")
			},
				func(ctx context.Context, cfg aws.ConfigAWS, log *logp.Logger, useDefaultRegion bool) error {
					return nil
				},
			)
		identity := awslib.Identity{
			Account: &test.account,
		}
		mockedIdentityProvider := &awslib.MockIdentityProviderGetter{}
		mockedIdentityProvider.EXPECT().GetIdentity(mock.Anything).Return(&identity, nil)

		factory := &IAMFactory{
			AwsConfigProvider: mockedConfigGetter,
			IdentityProvider: func(cfg awssdk.Config) awslib.IdentityProviderGetter {
				return mockedIdentityProvider
			},
		}

		cfg, err := agentconfig.NewConfigFrom(test.config)
		s.NoError(err)

		fetcher, err := factory.Create(s.log, cfg, nil)
		s.NoError(err)
		s.NotNil(fetcher)

		iamFetcher, ok := fetcher.(*IAMFetcher)
		s.True(ok)
		s.Equal("key", iamFetcher.cfg.AwsConfig.AccessKeyID)
		s.Equal("secret", iamFetcher.cfg.AwsConfig.SecretAccessKey)
		s.Equal("session", iamFetcher.cfg.AwsConfig.SessionToken)
	}
}
