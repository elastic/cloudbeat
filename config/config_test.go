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

package config

import (
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
		config                 string
		expectedActivatedRules []string
		expectedType           string
		expectedAccessKey      string
		expectedSecret         string
		expectedSessionToken   string
	}{
		{
			`
   type : cloudbeat/cis_k8s
   streams:
    - runtime_cfg:
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
			"cloudbeat/cis_k8s",
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

		s.Equal(test.expectedType, c.Type)
		s.Equal(test.expectedActivatedRules, c.Streams[0].RuntimeCfg.ActivatedRules.CisK8s)
		s.Equal(test.expectedAccessKey, c.Streams[0].AWSConfig.AccessKeyID)
		s.Equal(test.expectedSecret, c.Streams[0].AWSConfig.SecretAccessKey)
		s.Equal(test.expectedSessionToken, c.Streams[0].AWSConfig.SessionToken)
	}
}

func (s *ConfigTestSuite) TestRuntimeCfgExists() {
	var tests = []struct {
		config   string
		expected bool
	}{
		{
			`
  streams:
    - runtime_cfg:
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
    - not_runtime_cfg:
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

		s.Equal(test.expected, c.Streams[0].RuntimeCfg != nil)
	}
}

func (s *ConfigTestSuite) TestRuntimeConfig() {
	var tests = []struct {
		config   string
		expected []string
	}{
		{
			`
  streams:
    - runtime_cfg:
        activated_rules:
          cis_k8s:
            - a
            - b
            - c
            - d
`, []string{"a", "b", "c", "d"}},
	}

	for _, test := range tests {
		cfg, err := config.NewConfigFrom(test.config)
		s.NoError(err)

		c, err := New(cfg)
		s.NoError(err)

		dy, err := c.GetActivatedRules()
		s.NoError(err)

		s.Equal(test.expected, dy.CisK8s)
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

func (s *ConfigTestSuite) TestActivatedRulesFrameWork() {
	var tests = []struct {
		config                    string
		expectedActivatedRules    []string
		expectedEksActivatedRules []string
		expectedType              string
	}{
		{
			`
type: cloudbeat/cis_k8s
streams:
  - runtime_cfg:
      activated_rules:
        cis_k8s:
          - a
          - b
`,
			[]string{"a", "b"},
			nil,
			"cloudbeat/cis_k8s",
		},
		{
			`
type: cloudbeat/cis_eks
streams:
  - runtime_cfg:
      activated_rules:
        cis_eks:
          - a
          - b
`,
			nil,
			[]string{"a", "b"},
			"cloudbeat/cis_eks",
		},
	}

	for _, test := range tests {
		cfg, err := config.NewConfigFrom(test.config)
		s.NoError(err)

		c, err := New(cfg)
		s.NoError(err)

		s.Equal(test.expectedType, c.Type)
		s.Equal(test.expectedActivatedRules, c.Streams[0].RuntimeCfg.ActivatedRules.CisK8s)
		s.Equal(test.expectedEksActivatedRules, c.Streams[0].RuntimeCfg.ActivatedRules.CisEks)
	}
}
