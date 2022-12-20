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
	awsSdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/cloudbeat/config"
	"github.com/stretchr/testify/mock"
	"testing"

	agentConfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
)

type S3FactoryTestSuite struct {
	suite.Suite

	log *logp.Logger
}

func TestS3FactoryTestSuite(t *testing.T) {
	s := new(S3FactoryTestSuite)
	s.log = logp.NewLogger("cloudbeat_s3_factory_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *S3FactoryTestSuite) SetupTest() {}

func (s *S3FactoryTestSuite) TestCreateFetcher() {
	var tests = []struct {
		config string
	}{
		{
			`
name: aws-s3
access_key_id: key
secret_access_key: secret
default_region: eu-west-2
`,
		},
	}

	for _, test := range tests {
		mockedConfigGetter := &config.MockAwsConfigProvider{}
		mockedConfigGetter.EXPECT().
			InitializeAWSConfig(mock.Anything, mock.Anything).
			Call.
			Return(func(ctx context.Context, config aws.ConfigAWS) awsSdk.Config {

				return CreateSdkConfig(config, "eu-west-2")
			},
				func(ctx context.Context, config aws.ConfigAWS) error {
					return nil
				},
			)

		factory := &S3Factory{
			AwsConfigProvider: mockedConfigGetter,
		}

		cfg, err := agentConfig.NewConfigFrom(test.config)
		s.NoError(err)

		fetcher, err := factory.Create(s.log, cfg, nil)
		s.NoError(err)
		s.NotNil(fetcher)

		s3Fetcher, ok := fetcher.(*S3Fetcher)
		s.True(ok)
		s.Equal("key", s3Fetcher.cfg.AwsConfig.AccessKeyID)
		s.Equal("secret", s3Fetcher.cfg.AwsConfig.SecretAccessKey)
		s.Equal("eu-west-2", s3Fetcher.cfg.AwsConfig.DefaultRegion)
	}
}
