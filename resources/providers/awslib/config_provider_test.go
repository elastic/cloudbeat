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

package awslib

import (
	"context"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"testing"

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ConfigProviderTestSuite struct {
	suite.Suite

	log *logp.Logger
}

func TestEcrFactoryTestSuite(t *testing.T) {
	s := new(ConfigProviderTestSuite)
	s.log = logp.NewLogger("cloudbeat_config_provider_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *ConfigProviderTestSuite) TestInitializeAWSConfig() {
	var tests = []struct {
		accessKey string
		secret    string
		session   string
		region    string
		mock      func() MetadataProvider
	}{
		{
			accessKey: "key",
			secret:    "secret",
			session:   "session",
			region:    "us-east-1",
			mock: func() MetadataProvider {
				m := &MockMetadataProvider{}
				m.EXPECT().
					GetMetadata(mock.Anything, mock.Anything).
					Return(Ec2Metadata{
						Region: "us-east-1",
					}, nil)
				return m
			},
		},
		{
			accessKey: "key-1",
			secret:    "secret-1",
			session:   "session-1",
			region:    "us-east-2",
			mock: func() MetadataProvider {
				m := &MockMetadataProvider{}
				m.EXPECT().
					GetMetadata(mock.Anything, mock.Anything).
					Return(Ec2Metadata{
						Region: "us-east-2",
					}, nil)
				return m
			},
		},
		{
			accessKey: "key-1",
			secret:    "secret-1",
			session:   "session-1",
			region:    "us-east-1",
			mock:      func() MetadataProvider { return nil },
		},
	}

	for _, test := range tests {

		configProvider := ConfigProvider{
			MetadataProvider: test.mock(),
		}

		agentAwsConfig := aws.ConfigAWS{
			AccessKeyID:     test.accessKey,
			SecretAccessKey: test.secret,
			SessionToken:    test.session,
		}
		awsConfig, err := configProvider.InitializeAWSConfig(context.Background(), agentAwsConfig)
		s.NoError(err)

		cred, err := awsConfig.Credentials.Retrieve(context.Background())
		s.NoError(err)
		s.Equal(test.accessKey, cred.AccessKeyID)
		s.Equal(test.secret, cred.SecretAccessKey)
		s.Equal(test.session, cred.SessionToken)
		s.Equal(test.region, awsConfig.Region)

	}
}
