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
		config            string
		expectedType      string
		expectedAWSConfig aws.ConfigAWS
		expectedFetchers  int
	}{
		{
			`
fetchers:
  - name: a
    directory: b
  - name: b
    directory: b
`,
			"cis_k8s",
			aws.ConfigAWS{},
			2,
		},
		{
			`
config:
  v1:
    benchmark: cis_eks
    aws:
      credentials:
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
			"cis_eks",
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

	for i, test := range tests {
		s.Run(fmt.Sprint(i), func() {
			cfg, err := config.NewConfigFrom(test.config)
			s.NoError(err)

			c, err := New(cfg)
			s.NoError(err)

			s.Equal(test.expectedType, c.BenchmarkConfig.ID)
			s.Equal(test.expectedAWSConfig, c.BenchmarkConfig.AWSConfig)
			s.Equal(test.expectedFetchers, len(c.Fetchers))
		})
	}
}

func (s *ConfigTestSuite) TestBenchmarkType() {
	tests := []struct {
		config    string
		expected  string
		wantError bool
	}{
		{
			`
config:
  v1:
    benchmark: cis_eks
`,
			"cis_eks",
			false,
		},
		{
			`
config:
  v1:
    benchmark: cis_gcp
`,
			"",
			true,
		},
	}

	for i, test := range tests {
		s.Run(fmt.Sprint(i), func() {
			cfg, err := config.NewConfigFrom(test.config)
			s.NoError(err)

			c, err := New(cfg)
			if test.wantError {
				s.Error(err)
				return
			}
			s.NoError(err)
			s.Equal(test.expected, c.BenchmarkConfig.ID)
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
