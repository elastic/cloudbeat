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

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
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
		config string
	}{
		{
			`
name: aws-iam
`,
		},
	}

	for _, test := range tests {
		iamProvider := &awslib.MockIAMRolePermissionGetter{}
		factory := &IAMFactory{extraElements: func(log *logp.Logger) (IAMExtraElements, error) {
			return IAMExtraElements{
				iamProvider: iamProvider,
			}, nil
		}}

		cfg, err := common.NewConfigFrom(test.config)
		s.NoError(err)

		fetcher, err := factory.Create(s.log, cfg, nil)
		s.NoError(err)
		s.NotNil(fetcher)

		iamFetcher, ok := fetcher.(*IAMFetcher)
		s.True(ok)
		s.Equal(iamProvider, iamFetcher.iamProvider)
	}
}
