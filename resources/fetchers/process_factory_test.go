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
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/stretchr/testify/suite"
)

type ProcessFactoryTestSuite struct {
	suite.Suite
	factory fetching.Factory
}

func TestProcessFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessFactoryTestSuite))
}

func (s *ProcessFactoryTestSuite) SetupTest() {
	s.factory = &ProcessFactory{}
}

func (s *ProcessFactoryTestSuite) TestCreateFetcher() {
	var tests = []struct {
		config            string
		expectedDirectory string
	}{
		{
			`
name: process
directory: /hostfs
`,
			"/hostfs",
		},
	}

	for _, test := range tests {
		cfg, err := common.NewConfigFrom(test.config)
		s.NoError(err)

		fetcher, err := s.factory.Create(cfg)
		s.NoError(err)
		s.NotNil(fetcher)

		process, ok := fetcher.(*ProcessesFetcher)
		s.True(ok)
		s.Equal(test.expectedDirectory, process.cfg.Directory)
	}
}
