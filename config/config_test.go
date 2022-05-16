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

//go:build !integration
// +build !integration

package config

import (
	"strings"
	"testing"
	"time"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) SetupTest() {
}

func (s *ConfigTestSuite) TestNew() {
	var tests = []struct {
		config   string
		expected []string
	}{
		{
			`
  streams:
    - data_yaml:
        activated_rules:
          cis_k8s:
            - a
            - b
            - c
            - d
            - e
`,
			[]string{"a", "b", "c", "d", "e"},
		},
	}

	for _, test := range tests {
		cfg, err := common.NewConfigFrom(test.config)
		s.NoError(err)

		c, err := New(cfg)
		s.NoError(err)

		s.Equal(test.expected, c.Streams[0].DataYaml.ActivatedRules.CISK8S)
	}
}

func (s *ConfigTestSuite) TestConfigUpdate() {
	config := `
    streams:
      - data_yaml:
          activated_rules:
            cis_k8s:
              - a
              - b
              - c
              - d
              - e`

	var tests = []struct {
		update   string
		expected []string
	}{
		{
			`
          streams:
            - data_yaml:
                activated_rules:
                  cis_k8s:
                    - a
                    - b
                    - c
                    - d
        `,
			[]string{"a", "b", "c", "d"},
		},
		{
			`
          streams:
            - data_yaml:
                activated_rules:
                  cis_k8s:
                    - b
                    - c
                    - d
                    - e
                    - f
        `,
			[]string{"b", "c", "d", "e", "f"},
		},
		{
			`
          streams:
            - data_yaml:
                activated_rules:
                  cis_k8s: []
        `,
			[]string{},
		},
		{
			`
          streams:
            - data_yaml:
                activated_rules:
                  cis_k8s:
                    - a
                    - b
                    - c
                    - d
                    - e
        `,
			[]string{"a", "b", "c", "d", "e"},
		},
		// We're not currently aware of a scenario where we receive
		// multiple input streams, but still to make sure that it doesn't
		// break for this scenario.
		{
			`
          streams:
            - data_yaml:
                activated_rules:
                  cis_k8s:
                    - x
                    - "y" # Just YAML 1.1 things
                    - z
            - data_yaml:
                activated_rules:
                    - f
                    - g
                    - h
                    - i
                    - j
                    - k
        `,
			[]string{"x", "y", "z"},
		},
	}

	cfg, err := common.NewConfigFrom(config)
	s.NoError(err)

	c, err := New(cfg)
	s.NoError(err)

	for _, test := range tests {
		cfg, err := common.NewConfigFrom(test.update)
		s.NoError(err)

		err = c.Update(cfg)
		s.NoError(err)

		s.Equal(test.expected, c.Streams[0].DataYaml.ActivatedRules.CISK8S)
	}
}

// TestConfigUpdateIsolated tests whether updates made to a config from
// are isolated; only those parts of the config specified in the incoming
// config should get updated.
func (s *ConfigTestSuite) TestConfigUpdateIsolated() {
	config := `
    period: 10s
    kube_config: some_path
    streams:
      - data_yaml:
          activated_rules:
            cis_k8s:
              - a
              - b
              - c
              - d
              - e`

	var tests = []struct {
		update             string
		expectedPeriod     time.Duration
		expectedKubeConfig string
		expectedCISK8S     []string
	}{
		{
			update: `
            streams:
              - data_yaml:
                  activated_rules:
                    cis_k8s:
                      - a
                      - b
                      - c
                      - d`,
			expectedPeriod:     10 * time.Second,
			expectedKubeConfig: "some_path",
			expectedCISK8S:     []string{"a", "b", "c", "d"},
		},
		{
			update:             `period: 4h`,
			expectedPeriod:     4 * time.Hour,
			expectedKubeConfig: "some_path",
			expectedCISK8S:     []string{"a", "b", "c", "d"},
		},
		{
			update: `
            kube_config: some_other_path
            streams:
              - data_yaml:
                  activated_rules:
                    cis_k8s:
                      - a
                      - b
                      - c
                      - d
                      - e`,
			expectedPeriod:     4 * time.Hour,
			expectedKubeConfig: "some_other_path",
			expectedCISK8S:     []string{"a", "b", "c", "d", "e"},
		},
	}

	cfg, err := common.NewConfigFrom(config)
	s.NoError(err)

	c, err := New(cfg)
	s.NoError(err)

	for _, test := range tests {
		cfg, err := common.NewConfigFrom(test.update)
		s.NoError(err)

		err = c.Update(cfg)
		s.NoError(err)

		s.Equal(test.expectedPeriod, c.Period)
		s.Equal(test.expectedKubeConfig, c.KubeConfig)
		s.Equal(test.expectedCISK8S, c.Streams[0].DataYaml.ActivatedRules.CISK8S)
	}
}

func (s *ConfigTestSuite) TestConfigDataYaml() {
	var tests = []struct {
		config   string
		expected string
	}{
		{
			`
  streams:
    - data_yaml:
        activated_rules:
          cis_k8s:
            - a
            - b
            - c
            - d
`,
			`
activated_rules:
    cis_k8s:
        - a
        - b
        - c
        - d
`,
		},
	}

	for _, test := range tests {
		cfg, err := common.NewConfigFrom(test.config)
		s.NoError(err)

		c, err := New(cfg)
		s.NoError(err)

		dy, err := c.DataYaml()
		s.NoError(err)

		s.Equal(strings.TrimSpace(test.expected), strings.TrimSpace(dy))
	}
}
