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
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func TestConfigTestSuite(t *testing.T) {
	s := new(ConfigTestSuite)

	suite.Run(t, s)
}

func (s *ConfigTestSuite) SetupTest() {
}

func (s *ConfigTestSuite) TestNew() {
	tests := []struct {
		config              string
		expectedType        string
		expectedCloudConfig CloudConfig
	}{
		{
			`
config:
  v1:
    benchmark: cis_k8s
`,
			"cis_k8s",
			CloudConfig{},
		},
		{
			`
config:
  v1:
    benchmark: cis_azure
`,
			"cis_azure",
			CloudConfig{},
		},
		{
			`
config:
  v1:
    benchmark: cis_eks
    aws:
      account_type: organization-account
      credentials:
        access_key_id: key
        secret_access_key: secret
        session_token: session
        shared_credential_file: shared_credential_file
        credential_profile_name: credential_profile_name
        role_arn: role_arn
`,
			"cis_eks",
			CloudConfig{
				Aws: AwsConfig{
					Cred: aws.ConfigAWS{
						AccessKeyID:          "key",
						SecretAccessKey:      "secret",
						SessionToken:         "session",
						SharedCredentialFile: "shared_credential_file",
						ProfileName:          "credential_profile_name",
						RoleArn:              "role_arn",
					},
					AccountType: "organization-account",
				},
			},
		},
	}

	for i, test := range tests {
		s.Run(fmt.Sprint(i), func() {
			cfg, err := config.NewConfigFrom(test.config)
			s.Require().NoError(err)

			c, err := New(cfg)
			s.Require().NoError(err)

			s.Equal(test.expectedType, c.Benchmark)
			s.Equal(test.expectedCloudConfig, c.CloudConfig)
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
    benchmark: cis_azure
`,
			"cis_azure",
			false,
		},
	}

	for i, test := range tests {
		s.Run(fmt.Sprint(i), func() {
			cfg, err := config.NewConfigFrom(test.config)
			s.Require().NoError(err)

			c, err := New(cfg)
			if test.wantError {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			s.Equal(test.expected, c.Benchmark)
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
			s.Require().NoError(err)

			c, err := New(cfg)
			s.Require().NoError(err)

			s.Equal(test.expectedPeriod, c.Period)
		})
	}
}

func (s *ConfigTestSuite) TestPackagePolicyFields() {
	tests := []struct {
		config                        string
		expectedPackagePolicyID       string
		expectedPackagePolicyRevision int
	}{
		{
			config:                        "",
			expectedPackagePolicyID:       "",
			expectedPackagePolicyRevision: 0,
		},
		{
			config:                        `package_policy_id: 123`,
			expectedPackagePolicyID:       "123",
			expectedPackagePolicyRevision: 0,
		},
		{
			config:                        `revision: 123`,
			expectedPackagePolicyID:       "",
			expectedPackagePolicyRevision: 123,
		},
		{
			config: `package_policy_id: 123
revision: 1`,
			expectedPackagePolicyID:       "123",
			expectedPackagePolicyRevision: 1,
		},
	}

	for i, test := range tests {
		s.Run(fmt.Sprint(i), func() {
			cfg := config.MustNewConfigFrom(test.config)

			c, err := New(cfg)
			s.Require().NoError(err)

			s.Equal(test.expectedPackagePolicyID, c.PackagePolicyId)
			s.Equal(test.expectedPackagePolicyRevision, c.PackagePolicyRevision)
		})
	}
}
