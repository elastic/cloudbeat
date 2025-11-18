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

package assetinventory

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestStrategyPicks(t *testing.T) {
	testhelper.SkipLong(t)

	testCases := []struct {
		name        string
		cfg         *config.Config
		env         map[string]string
		expectedErr string
	}{
		{
			name:        "expected error: asset_inventory_provider not set",
			cfg:         &config.Config{},
			env:         nil,
			expectedErr: "missing config.v1.asset_inventory_provider",
		},
		{
			name: "expected error: unsupported provider",
			cfg: &config.Config{
				AssetInventoryProvider: "NOPE",
			},
			env:         nil,
			expectedErr: "unsupported Asset Inventory provider \"NOPE\"",
		},
		{
			name: "expected success: Azure",
			cfg: &config.Config{
				AssetInventoryProvider: config.ProviderAzure,
			},
			env:         nil,
			expectedErr: "",
		},
		{
			name: "expected error: GCP missing account type",
			cfg: &config.Config{
				AssetInventoryProvider: config.ProviderGCP,
			},
			env:         nil,
			expectedErr: "invalid gcp account type",
		},
		{
			name: "expected success: GCP",
			cfg: &config.Config{
				AssetInventoryProvider: config.ProviderGCP,
				CloudConfig: config.CloudConfig{
					Gcp: config.GcpConfig{
						AccountType:    config.SingleAccount,
						ProjectId:      "nonexistent",
						OrganizationId: "nonexistent",
						GcpClientOpt: config.GcpClientOpt{
							CredentialsJSON: "{\"type\": \"service_account\"}",
						},
					},
				},
			},
			env:         nil,
			expectedErr: "could not parse key",
		},
		{
			name: "expected error: AWS unsupported account type",
			cfg: &config.Config{
				AssetInventoryProvider: config.ProviderAWS,
				CloudConfig: config.CloudConfig{
					Aws: config.AwsConfig{
						AccountType: "NOPE",
					},
				},
			},
			env:         nil,
			expectedErr: "unsupported account_type: \"NOPE\"",
		},
		{
			name: "expected success: AWS",
			cfg: &config.Config{
				AssetInventoryProvider: config.ProviderAWS,
				CloudConfig: config.CloudConfig{
					Aws: config.AwsConfig{
						AccountType: config.SingleAccount,
						Cred: aws.ConfigAWS{
							AccessKeyID:     "key",
							SecretAccessKey: "key",
						},
					},
				},
			},
			env:         nil,
			expectedErr: "STS: GetCallerIdentity",
		},
		{
			name: "expected success: AWS with cloud connectors",
			cfg: &config.Config{
				AssetInventoryProvider: config.ProviderAWS,
				CloudConfig: config.CloudConfig{
					Aws: config.AwsConfig{
						AccountType: config.SingleAccount,
						Cred: aws.ConfigAWS{
							AccessKeyID:     "key",
							SecretAccessKey: "key",
						},
						CloudConnectors: true,
						CloudConnectorsConfig: config.CloudConnectorsConfig{
							LocalRoleARN:  "abc",
							GlobalRoleARN: "xyz",
							ResourceID:    "123",
						},
					},
				},
			},
			env: map[string]string{
				"AWS_WEB_IDENTITY_TOKEN_FILE": "/tmp/fake-token-file",
				"AWS_ROLE_ARN":                "arn:aws:iam::123456789012:role/test-local-role",
			},
			expectedErr: "STS: GetCallerIdentity",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			s := strategy{
				logger: testhelper.NewLogger(t),
				cfg:    tc.cfg,
			}
			ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
			defer cancel()
			obj, err := s.NewAssetInventory(ctx, nil)
			if tc.expectedErr != "" {
				assert.Equal(t, inventory.AssetInventory{}, obj)
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetOrgIAMRoleNamesProvider(t *testing.T) {
	tests := []struct {
		cloudConnectors    bool
		expectedRootRole   string
		expectedMemberRole string
	}{
		{
			cloudConnectors:    false,
			expectedRootRole:   awslib.AssetDiscoveryOrgIAMRoleNamesProvider{}.RootRoleName(),
			expectedMemberRole: awslib.AssetDiscoveryOrgIAMRoleNamesProvider{}.MemberRoleName(),
		},
		{
			cloudConnectors:    true,
			expectedRootRole:   awslib.BenchmarkOrgIAMRoleNamesProvider{}.RootRoleName(),
			expectedMemberRole: awslib.BenchmarkOrgIAMRoleNamesProvider{}.MemberRoleName(),
		},
	}

	for i, tc := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := getOrgIAMRoleNamesProvider(config.AwsConfig{CloudConnectors: tc.cloudConnectors})
			assert.Equal(t, tc.expectedRootRole, got.RootRoleName())
			assert.Equal(t, tc.expectedMemberRole, got.MemberRoleName())
		})
	}
}
