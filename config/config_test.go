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
	"fmt"
	"testing"
	"time"

	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
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
		config                     string
		expectedActivatedRules     *Benchmarks
		expectedType               string
		expectedAWSConfig          aws.ConfigAWS
		expectedFetchers           int
		expectedCompatibleVersions *CompatibleVersions
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
compatible_versions:
  cloudbeat: some_range
`,
			&Benchmarks{CisK8s: []string{"a", "b", "c", "d", "e"}},
			"cloudbeat/cis_k8s",
			aws.ConfigAWS{},
			2,
			&CompatibleVersions{Cloudbeat: awssdk.String("some_range")},
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
			nil,
		},
	}

	for i, test := range tests {
		s.Run(fmt.Sprint(i), func() {
			cfg, err := config.NewConfigFrom(test.config)
			s.NoError(err)

			c, err := New(cfg)
			s.NoError(err)

			s.Equal(test.expectedType, c.Type)
			s.EqualValues(test.expectedActivatedRules, c.RuntimeConfig.ActivatedRules)
			s.Equal(test.expectedAWSConfig, c.AWSConfig)
			s.Equal(test.expectedFetchers, len(c.Fetchers))
			s.Equal(test.expectedCompatibleVersions, c.CompatibleVersions)
		})
	}
}

func (s *ConfigTestSuite) TestRuntimeConfigExists() {
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

	for i, test := range tests {
		s.Run(fmt.Sprint(i), func() {
			cfg, err := config.NewConfigFrom(test.config)
			s.NoError(err)

			c, err := New(cfg)
			s.NoError(err)

			s.Equal(test.expected, c.RuntimeConfig != nil)
		})
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

	for i, test := range tests {
		s.Run(fmt.Sprint(i), func() {
			cfg, err := config.NewConfigFrom(test.config)
			s.NoError(err)

			c, err := New(cfg)
			s.NoError(err)

			rules := c.RuntimeConfig.ActivatedRules

			s.Equal(test.expected, rules.CisK8s)
		})
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

	for i, test := range tests {
		s.Run(fmt.Sprint(i), func() {
			cfg, err := config.NewConfigFrom(test.config)
			s.NoError(err)

			c, err := New(cfg)
			s.NoError(err)

			s.Equal(test.expectedPeriod, c.Period)
		})
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

	for i, test := range tests {
		s.Run(fmt.Sprint(i), func() {
			cfg, err := config.NewConfigFrom(test.config)
			s.NoError(err)

			c, err := New(cfg)
			s.NoError(err)

			s.Equal(test.expectedType, c.Type)
			s.Equal(test.expectedActivatedRules, c.RuntimeConfig.ActivatedRules.CisK8s)
			s.Equal(test.expectedEksActivatedRules, c.RuntimeConfig.ActivatedRules.CisEks)
		})
	}
}

func (s *ConfigTestSuite) TestConfigValidatefVersionCompatibility() {
	tests := []struct {
		config  string
		version string
		valid   bool
	}{
		{
			`
compatible_versions:
  cloudbeat: ">= 8.4.0 <= 8.5.0"
`,
			"8.4.0",
			true,
		},
		{
			`
compatible_versions:
  cloudbeat: "8.5.0"
`,
			"v8.5.0-SNAPSHOT",
			true,
		},
		{
			`
compatible_versions:
  cloudbeat: ">= v8.5.0 <= v8.7.0"
`,
			"8.3.0",
			false,
		},
		{
			`
compatible_versions:
  cloudbeat: ">= 8.5.0 <= 8.7.0"
`,
			"8.3.0-SNAPSHOT",
			false,
		},
		{
			`
compatible_versions:
  cloudbeat: ">= 8.5.0 <= 8.7.0"
`,
			"8.6.0-SNAPSHOT",
			true,
		},
		{
			`
compatible_versions:
  cloudbeat: ">= 8.5.0 <= 8.7.0"
`,
			"v8.8.0",
			false,
		},
		{
			`
compatible_versions:
  cloudbeat: ">= v8.5.0 <= v8.7.0"
`,
			"v8.5.0",
			true,
		},
		{
			`
compatible_versions:
  cloudbeat: ">= v8.5.0 < v8.7.0"
`,
			"8.7.0",
			false,
		},
		{
			`
compatible_versions:
  cloudbeat: ">= 8.5.0"
`,
			"v8.3.0-SNAPSHOT",
			false,
		},
		{
			`
compatible_versions:
  cloudbeat: ">= v8.5.0"
`,
			"v8.8.0-SNAPSHOT",
			true,
		},
		{
			`
compatible_versions:
  cloudbeat: "< 8.7.0"
`,
			"8.3.0",
			true,
		},
		{
			`
compatible_versions:
  cloudbeat: "<= 8.7.0"
`,
			"v8.7.0-SNAPSHOT",
			true,
		},
		{
			`
compatible_versions:
  cloudbeat: "< v8.7.0"
`,
			"8.3.0",
			true,
		},
	}

	versionfuncer := func(s string) func() string {
		return func() string {
			return s
		}
	}

	for i, test := range tests {
		s.Run(fmt.Sprint(i), func() {
			cfg, err := config.NewConfigFrom(test.config)
			s.NoError(err)

			c, err := New(cfg)
			s.NoError(err)

			valid, err := c.validateVersionCompatibility(versionfuncer(test.version))
			s.NoError(err)

			s.Equal(test.valid, valid)
		})
	}
}
