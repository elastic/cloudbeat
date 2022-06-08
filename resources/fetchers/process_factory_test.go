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

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/stretchr/testify/suite"
)

type ProcessFactoryTestSuite struct {
	suite.Suite
	factory fetching.Factory

	log *logp.Logger
}
type ProcessConfigTestValidator struct {
	processName string
	validate    func([]string)
}

func TestProcessFactoryTestSuite(t *testing.T) {
	s := new(ProcessFactoryTestSuite)
	s.log = logp.NewLogger("cloudbeat_process_factory_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *ProcessFactoryTestSuite) SetupTest() {
	s.factory = &ProcessFactory{}
}

func (s *ProcessFactoryTestSuite) TestCreateFetcher() {
	var tests = []struct {
		config            string
		expectedDirectory string
		processValidators []ProcessConfigTestValidator
	}{
		{
			`
name: process
directory: /hostfs
processes:
 etcd:
 kubelet:
  config-file-arguments:
  - config
`,
			"/hostfs",
			[]ProcessConfigTestValidator{
				{
					processName: "kubelet",
					validate: func(cmd []string) {
						s.Len(cmd, 1)
						s.Contains(cmd, "config")
					},
				},
				{
					processName: "etcd",
					validate: func(cmd []string) {
						s.Nil(cmd)
					},
				}},
		},
	}

	for _, test := range tests {
		cfg, err := common.NewConfigFrom(test.config)
		s.NoError(err)

		fetcher, err := s.factory.Create(s.log, cfg, nil)
		s.NoError(err)
		s.NotNil(fetcher)

		process, ok := fetcher.(*ProcessesFetcher)
		s.True(ok)
		s.Equal(test.expectedDirectory, process.cfg.Directory)
		s.NotNil(process.Fs)

		s.Equal(len(test.processValidators), len(process.cfg.RequiredProcesses))
		for _, validator := range test.processValidators {
			validator.validate(process.cfg.RequiredProcesses[validator.processName].ConfigFileArguments)
		}
	}
}
