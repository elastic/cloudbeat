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
	"testing"

	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/kms"
	"github.com/stretchr/testify/mock"

	agentConfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
)

type KmsFactoryTestSuite struct {
	suite.Suite

	log *logp.Logger
}

func TestKmsFactoryTestSuite(t *testing.T) {
	s := new(KmsFactoryTestSuite)
	s.log = logp.NewLogger("cloudbeat_kms_factory_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *KmsFactoryTestSuite) SetupTest() {}

func (s *KmsFactoryTestSuite) TestCreateFetcher() {
	var tests = []struct {
		config string
	}{
		{
			`
name: aws-kms
access_key_id: key
secret_access_key: secret
default_region: eu-west-2
`,
		},
	}

	for _, test := range tests {
		mockCrossRegionFetcher := &awslib.MockCrossRegionFetcher[kms.Client]{}
		mockCrossRegionFetcher.On("GetMultiRegionsClientMap").Return(nil)

		mockCrossRegionFactory := &awslib.MockCrossRegionFactory[kms.Client]{}
		mockCrossRegionFactory.On(
			"NewMultiRegionClients",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(mockCrossRegionFetcher)

		factory := &KmsFactory{
			CrossRegionFactory: mockCrossRegionFactory,
		}

		cfg, err := agentConfig.NewConfigFrom(test.config)
		s.NoError(err)

		fetcher, err := factory.Create(s.log, cfg, nil)
		s.NoError(err)
		s.NotNil(fetcher)

		KmsFetcher, ok := fetcher.(*KmsFetcher)
		s.True(ok)
		s.Equal("key", KmsFetcher.cfg.AwsConfig.AccessKeyID)
		s.Equal("secret", KmsFetcher.cfg.AwsConfig.SecretAccessKey)
		s.Equal("eu-west-2", KmsFetcher.cfg.AwsConfig.DefaultRegion)
	}
}
