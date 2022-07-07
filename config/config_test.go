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

	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite

	log *logp.Logger
}

func TestConfigTestSuite(t *testing.T) {
	s := new(ConfigTestSuite)
	s.log = logp.NewLogger("cloudbeat_config_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *ConfigTestSuite) SetupTest() {
}

func (s *ConfigTestSuite) TestNew() {
	var tests = []struct {
		config                    string
		expectedActivatedK8sRules []string
		expectedAccessKey         string
		expectedSecret            string
		expectedSessionToken      string
	}{
		{
			`
   type : cloudbeat/vanilla
   streams:
    - data_yaml:
        activated_rules:
          cis_k8s:
            - a
            - b
            - c
            - d
            - e
      access_key_id: key
      secret_access_key: secret
      session_token: session
`,
			[]string{"a", "b", "c", "d", "e"},
			"key",
			"secret",
			"session",
		},
	}

	for _, test := range tests {
		cfg, err := config.NewConfigFrom(test.config)
		s.NoError(err)

		c, err := New(cfg)
		s.NoError(err)

		s.Equal(test.expectedActivatedK8sRules, c.Streams[0].DataYaml.ActivatedRules.CISK8S)
		s.Equal(test.expectedAccessKey, c.Streams[0].AWSConfig.AccessKeyID)
		s.Equal(test.expectedSecret, c.Streams[0].AWSConfig.SecretAccessKey)
		s.Equal(test.expectedSessionToken, c.Streams[0].AWSConfig.SessionToken)
	}
}

func (s *ConfigTestSuite) TestDataYamlExists() {
	var tests = []struct {
		config   string
		expected bool
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
			true,
		},
		{
			`
  streams:
    - not_data_yaml:
        something: true
`,
			false,
		},
	}

	for _, test := range tests {
		cfg, err := config.NewConfigFrom(test.config)
		s.NoError(err)

		c, err := New(cfg)
		s.NoError(err)

		s.Equal(test.expected, c.Streams[0].DataYaml != nil)
	}
}

func (s *ConfigTestSuite) TestConfigUpdate() {
	configYml := `
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

	cfg, err := config.NewConfigFrom(configYml)
	s.NoError(err)

	c, err := New(cfg)
	s.NoError(err)

	for _, test := range tests {
		cfg, err := config.NewConfigFrom(test.update)
		s.NoError(err)

		err = c.Update(s.log, cfg)
		s.NoError(err)

		s.Equal(test.expected, c.Streams[0].DataYaml.ActivatedRules.CISK8S)
	}
}

// TestConfigUpdateIsolated tests whether updates made to a config from
// are isolated; only those parts of the config specified in the incoming
// config should get updated.
func (s *ConfigTestSuite) TestConfigUpdateIsolated() {
	configYml := `
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

	cfg, err := config.NewConfigFrom(configYml)
	s.NoError(err)

	c, err := New(cfg)
	s.NoError(err)

	for _, test := range tests {
		cfg, err := config.NewConfigFrom(test.update)
		s.NoError(err)

		err = c.Update(s.log, cfg)
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
		cfg, err := config.NewConfigFrom(test.config)
		s.NoError(err)

		c, err := New(cfg)
		s.NoError(err)

		dy, err := c.DataYaml()
		s.NoError(err)

		s.Equal(strings.TrimSpace(test.expected), strings.TrimSpace(dy))
	}
}

func (s *ConfigTestSuite) TestConfigPeriod() {
	var tests = []struct {
		config         string
		expectedPeriod time.Duration
	}{
		{"", 4 * time.Hour},
		{"period: 50s", 50 * time.Second},
		{"period: 5m", 5 * time.Minute},
		{"period: 2h", 2 * time.Hour},
	}

	for _, test := range tests {
		cfg, err := config.NewConfigFrom(test.config)
		s.NoError(err)

		c, err := New(cfg)
		s.NoError(err)

		s.Equal(test.expectedPeriod, c.Period)
	}
}
