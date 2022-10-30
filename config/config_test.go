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

	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
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
	tests := []struct {
		config                 string
		expectedActivatedRules *Benchmarks
		expectedType           string
		expectedAWSConfig      aws.ConfigAWS
		expectedFetchers       int
	}{
		{
			`
runtime_cfg:
  activated_rules:
    cis_k8s:
      - a
      - b
      - c
      - d
      - e
fetchers:
  - name: a
    directory: b
  - name: b
    directory: b
`,
			&Benchmarks{CisK8s: []string{"a", "b", "c", "d", "e"}},
			"cloudbeat/cis_k8s",
			aws.ConfigAWS{},
			2,
		},
		{
			`
runtime_cfg:
  activated_rules:
    cis_eks:
      - a
      - b
      - c
      - d
      - e
access_key_id: key
secret_access_key: secret
session_token: session
shared_credential_file: shared_credential_file
credential_profile_name: credential_profile_name
role_arn: role_arn
fetchers:
  - name: a
    directory: b
  - name: b
    directory: b
  - name: c
    directory: c
`,
			&Benchmarks{CisEks: []string{"a", "b", "c", "d", "e"}},
			"cloudbeat/cis_eks",
			aws.ConfigAWS{
				AccessKeyID:          "key",
				SecretAccessKey:      "secret",
				SessionToken:         "session",
				SharedCredentialFile: "shared_credential_file",
				ProfileName:          "credential_profile_name",
				RoleArn:              "role_arn",
			},
			3,
		},
	}

	for _, test := range tests {
		cfg, err := config.NewConfigFrom(test.config)
		s.NoError(err)

		c, err := New(cfg)
		s.NoError(err)

		s.Equal(test.expectedType, c.Type)
		s.EqualValues(test.expectedActivatedRules, c.RuntimeCfg.ActivatedRules)
		s.Equal(test.expectedAWSConfig, c.AWSConfig)
		s.Equal(test.expectedFetchers, len(c.Fetchers))
	}
}

func (s *ConfigTestSuite) TestRuntimeCfgExists() {
	tests := []struct {
		config   string
		expected bool
	}{
		{
			`
runtime_cfg:
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
not_runtime_cfg:
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

		s.Equal(test.expected, c.RuntimeCfg != nil)
	}
}

func (s *ConfigTestSuite) TestRuntimeConfig() {
	tests := []struct {
		config   string
		expected []string
	}{
		{
			`
runtime_cfg:
  activated_rules:
    cis_k8s:
      - a
      - b
      - c
      - d
`, []string{"a", "b", "c", "d"},
		},
	}

	for _, test := range tests {
		cfg, err := config.NewConfigFrom(test.config)
		s.NoError(err)

		c, err := New(cfg)
		s.NoError(err)

		rules := c.RuntimeCfg.ActivatedRules

		s.Equal(test.expected, rules.CisK8s)
	}
}

func (s *ConfigTestSuite) TestRuntimeEvaluatorConfig() {
	tests := []struct {
		config   string
		expected EvaluatorConfig
	}{
		{
			`
evaluator:
  decision_logs: true
`,
			EvaluatorConfig{
				DecisionLogs: true,
			},
		},
	}

	for _, test := range tests {
		cfg, err := config.NewConfigFrom(test.config)
		s.NoError(err)

		c, err := New(cfg)
		s.NoError(err)

		s.Equal(test.expected, c.Evaluator)
	}
}

func (s *ConfigTestSuite) TestConfigPeriod() {
	tests := []struct {
		config         string
		expectedPeriod time.Duration
	}{
		{"", 4 * time.Hour},
		{
			`
    period: 50s
`, 50 * time.Second,
		},
		{
			`
    period: 5m
`, 5 * time.Minute,
		},
		{
			`
    period: 2h
`, 2 * time.Hour,
		},
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
	tests := []struct {
		config                    string
		expectedActivatedRules    []string
		expectedEksActivatedRules []string
		expectedType              string
	}{
		{
			`
runtime_cfg:
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
runtime_cfg:
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
		s.Equal(test.expectedActivatedRules, c.RuntimeCfg.ActivatedRules.CisK8s)
		s.Equal(test.expectedEksActivatedRules, c.RuntimeCfg.ActivatedRules.CisEks)
	}
}
