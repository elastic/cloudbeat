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

package beater

import (
	"context"
	"testing"
	"time"

	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
)

type BeaterTestSuite struct {
	suite.Suite

	log *logp.Logger
}

func TestBeaterTestSuite(t *testing.T) {
	s := new(BeaterTestSuite)
	s.log = logp.NewLogger("cloudbeat_beater_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *BeaterTestSuite) SetupTest() {
}

func (s *BeaterTestSuite) TestReconfigureWait() {
	ctx, cancel := context.WithCancel(context.Background())

	beat := &cloudbeat{
		ctx:    ctx,
		cancel: cancel,
		log:    s.log,
	}

	configNoStreams, err := config.NewConfigFrom(`
    not_streams:
      - data_yaml:
          activated_rules:
            cis_k8s:
              - a
              - b
              - c
              - d
              - e
`)
	s.NoError(err)

	configNoDataYaml, err := config.NewConfigFrom(`
    streams:
      - not_data_yaml:
          activated_rules:
            cis_k8s:
              - a
              - b
              - c
              - d
              - e
`)
	s.NoError(err)

	configWithDataYaml, err := config.NewConfigFrom(`
    streams:
      - data_yaml:
          activated_rules:
            cis_k8s:
              - a
              - b
              - c
              - d
              - e
`)
	s.NoError(err)

	type incomingConfigs []struct {
		after  time.Duration
		config *config.C
	}

	testcases := []struct {
		timeout  time.Duration
		configs  incomingConfigs
		expected *config.C
	}{
		{
			5 * time.Millisecond,
			incomingConfigs{},
			nil,
		},
		{
			40 * time.Millisecond,
			incomingConfigs{
				{1 * time.Millisecond, configWithDataYaml},
			},
			configWithDataYaml,
		},
		{
			40 * time.Millisecond,
			incomingConfigs{
				{1 * time.Millisecond, configNoStreams},
				{1 * time.Millisecond, configNoDataYaml},
				{1 * time.Millisecond, configNoStreams},
			},
			nil,
		},
		{
			40 * time.Millisecond,
			incomingConfigs{
				{1 * time.Millisecond, configNoStreams},
				{1 * time.Millisecond, configNoDataYaml},
				{1 * time.Millisecond, configNoStreams},
				{1 * time.Millisecond, configWithDataYaml},
			},
			configWithDataYaml,
		},
		{
			40 * time.Millisecond,
			incomingConfigs{
				{1 * time.Millisecond, configNoStreams},
				{1 * time.Millisecond, configNoDataYaml},
				{1 * time.Millisecond, configNoStreams},
				{40 * time.Millisecond, configWithDataYaml},
			},
			nil,
		},
		{
			40 * time.Millisecond,
			incomingConfigs{
				{1 * time.Millisecond, configNoStreams},
				{40 * time.Millisecond, configNoDataYaml},
				{1 * time.Millisecond, configNoStreams},
			},
			nil,
		},
		{
			40 * time.Millisecond,
			incomingConfigs{
				{1 * time.Millisecond, configNoDataYaml},
				{1 * time.Millisecond, configWithDataYaml},
				{1 * time.Millisecond, configNoStreams},
			},
			configWithDataYaml,
		},
		{
			40 * time.Millisecond,
			incomingConfigs{
				{1 * time.Millisecond, configWithDataYaml},
				{1 * time.Millisecond, configNoStreams},
			},
			configWithDataYaml,
		},
	}

	for _, tcase := range testcases {
		cu := make(chan *config.C)
		beat.configUpdates = cu

		go func(ic incomingConfigs) {
			defer close(cu)

			for _, c := range ic {
				time.Sleep(c.after)
				cu <- c.config
			}
		}(tcase.configs)

		u, err := beat.reconfigureWait(tcase.timeout)
		if tcase.expected == nil {
			s.Error(err)
		} else {
			s.Equal(tcase.expected, u)
		}
	}
}
