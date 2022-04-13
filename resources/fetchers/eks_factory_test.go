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
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/stretchr/testify/suite"
	"testing"
)

type EksFactoryTestSuite struct {
	suite.Suite
	factory fetching.Factory
}

func TestEksFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(EksFactoryTestSuite))
}

func (s *EksFactoryTestSuite) SetupTest() {

}

func (s *EksFactoryTestSuite) TestCreateFetcher() {
	var tests = []struct {
		config string
	}{
		{
			`
name: aws-eks
`,
		},
	}

	for _, test := range tests {
		eksProvider := &awslib.MockedEksClusterDescriber{}
		factory := &EKSFactory{extraElements: func() (eksExtraElements, error) {
			return eksExtraElements{eksProvider: eksProvider}, nil
		}}

		cfg, err := common.NewConfigFrom(test.config)
		s.NoError(err)

		fetcher, err := factory.Create(cfg)
		s.NoError(err)
		s.NotNil(fetcher)

		eksFetcher, ok := fetcher.(*EKSFetcher)
		s.True(ok)
		s.Equal(eksProvider, eksFetcher.eksProvider)
	}
}
