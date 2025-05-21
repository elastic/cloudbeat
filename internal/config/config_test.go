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
		overwriteEnvs       func(*testing.T)
		expectedType        string
		expectedCloudConfig CloudConfig
	}{
		{
			config: `
config:
  v1:
    benchmark: cis_k8s
`,
			expectedType:        "cis_k8s",
			expectedCloudConfig: CloudConfig{Gcp: defaultGCPConfig()},
		},
		{
			config: `
config:
  v1:
    benchmark: cis_azure
`,
			expectedType:        "cis_azure",
			expectedCloudConfig: CloudConfig{Gcp: defaultGCPConfig()},
		},
		{
			config: `
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
			expectedType: "cis_eks",
			expectedCloudConfig: CloudConfig{
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
				Gcp: defaultGCPConfig(),
			},
		},
		{ // GCP config only, default values for page size and timeout.
			config: `
config:
  v1:
    type: cspm
    deployment: gcp
    benchmark: cis_gcp
    gcp:
      project_id: abc123
      organization_id: efg456
      account_type: organization-account
      credentials:
        credentials_file_path: /tmp/creds.json
`,
			expectedType: "cis_gcp",
			expectedCloudConfig: CloudConfig{
				Gcp: GcpConfig{
					ProjectId:      "abc123",
					OrganizationId: "efg456",
					AccountType:    "organization-account",
					GcpClientOpt: GcpClientOpt{
						CredentialsFilePath: "/tmp/creds.json",
					},
					GcpCallOpt: GcpCallOpt{
						ListAssetsTimeout:  4 * time.Minute,
						ListAssetsPageSize: 200,
					},
				},
			},
		},
		{ // GCP config with call_options.
			config: `
config:
  v1:
    type: cspm
    deployment: gcp
    benchmark: cis_gcp
    gcp:
      project_id: abc123
      organization_id: efg456
      account_type: organization-account
      credentials:
        credentials_file_path: /tmp/creds.json
      call_options:
        list_assets_timeout: 6m
        list_assets_page_size: 500
`,
			expectedType: "cis_gcp",
			expectedCloudConfig: CloudConfig{
				Gcp: GcpConfig{
					ProjectId:      "abc123",
					OrganizationId: "efg456",
					AccountType:    "organization-account",
					GcpClientOpt: GcpClientOpt{
						CredentialsFilePath: "/tmp/creds.json",
					},
					GcpCallOpt: GcpCallOpt{
						ListAssetsTimeout:  6 * time.Minute,
						ListAssetsPageSize: 500,
					},
				},
			},
		},
		{ // GCP with overwrite env vars
			overwriteEnvs: func(t *testing.T) {
				t.Helper()
				t.Setenv(CloudbeatGCPListAssetPageSizeEnvVar, "400")
				t.Setenv(CloudbeatGCPListAssetTimeoutEnvVar, "5m")
			},
			config: `
config:
  v1:
    type: cspm
    deployment: gcp
    benchmark: cis_gcp
    gcp:
      project_id: abc123
      organization_id: efg456
      account_type: organization-account
      credentials:
        credentials_file_path: /tmp/creds.json
`,
			expectedType: "cis_gcp",
			expectedCloudConfig: CloudConfig{
				Gcp: GcpConfig{
					ProjectId:      "abc123",
					OrganizationId: "efg456",
					AccountType:    "organization-account",
					GcpClientOpt: GcpClientOpt{
						CredentialsFilePath: "/tmp/creds.json",
					},
					GcpCallOpt: GcpCallOpt{
						ListAssetsTimeout:  5 * time.Minute,
						ListAssetsPageSize: 400,
					},
				},
			},
		},
	}

	for i, test := range tests {
		s.Run(fmt.Sprint(i), func() {
			if test.overwriteEnvs != nil {
				test.overwriteEnvs(s.T())
			}
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

func (s *ConfigTestSuite) TestCloudConnectorsConfig() {
	tests := map[string]struct {
		config              string
		overwriteEnv        func(t *testing.T)
		expectedType        string
		expectedCloudConfig CloudConfig
	}{
		"happy path cloud connectors enabled": {
			config: `
config:
  v1:
    benchmark: cis_aws
    aws:
        supports_cloud_connectors: true
        credentials:
            external_id: abc123
`,
			expectedType: "cis_aws",
			expectedCloudConfig: CloudConfig{
				Aws: AwsConfig{
					CloudConnectors: true,
					Cred: aws.ConfigAWS{
						ExternalID: "abc123",
					},
					CloudConnectorsConfig: CloudConnectorsConfig{},
				},
				Gcp: defaultGCPConfig(),
			},
		},
		"happy path cloud connectors enabled - attempt overwrite roles": {
			config: `
config:
  v1:
    benchmark: cis_aws
    aws:
        account_type: single-account
        supports_cloud_connectors: true
        credentials:
            external_id: abc123
        CloudConnectorsConfig:
            LocalRoleARN: "abc123"
            LocalRoleARN: "abc123"
`,
			expectedType: "cis_aws",
			expectedCloudConfig: CloudConfig{
				Aws: AwsConfig{
					AccountType:     SingleAccount,
					CloudConnectors: true,
					Cred: aws.ConfigAWS{
						ExternalID: "abc123",
					},
					CloudConnectorsConfig: CloudConnectorsConfig{},
				},
				Gcp: defaultGCPConfig(),
			},
		},
		"happy path cloud connectors enabled - env vars set": {
			config: `
config:
  v1:
    benchmark: cis_aws
    aws:
        account_type: single-account
        supports_cloud_connectors: true
        credentials:
            external_id: abc123
`,
			overwriteEnv: func(t *testing.T) {
				t.Helper()
				t.Setenv(CloudConnectorsLocalRoleEnvVar, "abc123")
				t.Setenv(CloudConnectorsGlobalRoleEnvVar, "abc456")
				t.Setenv(CloudResourceIDEnvVar, "abc789")
			},
			expectedType: "cis_aws",
			expectedCloudConfig: CloudConfig{
				Aws: AwsConfig{
					AccountType:     SingleAccount,
					CloudConnectors: true,
					Cred: aws.ConfigAWS{
						ExternalID: "abc123",
					},
					CloudConnectorsConfig: CloudConnectorsConfig{
						LocalRoleARN:  "abc123",
						GlobalRoleARN: "abc456",
						ResourceID:    "abc789",
					},
				},
				Gcp: defaultGCPConfig(),
			},
		},
	}

	for i, test := range tests {
		s.Run(fmt.Sprint(i), func() {
			if test.overwriteEnv != nil {
				test.overwriteEnv(s.T())
			}
			cfg, err := config.NewConfigFrom(test.config)
			s.Require().NoError(err)

			c, err := New(cfg)
			s.Require().NoError(err)

			s.Equal(test.expectedType, c.Benchmark)
			s.Equal(test.expectedCloudConfig, c.CloudConfig)
		})
	}
}
