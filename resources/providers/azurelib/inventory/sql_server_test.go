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

package inventory

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/require"
)

func mockAssetSQLEncryptionProtector(f func() ([]armsql.EncryptionProtectorsClientListByServerResponse, error)) ProviderAPI {
	wrapper := &azureClientWrapper{
		AssetSQLEncryptionProtector: func(_ context.Context, _, _, _ string, _ *arm.ClientOptions, _ *armsql.EncryptionProtectorsClientListByServerOptions) ([]armsql.EncryptionProtectorsClientListByServerResponse, error) {
			return f()
		},
	}

	return &provider{
		log:    logp.NewLogger("mock_asset_sql_encryption_protector"),
		client: wrapper,
	}
}

func mockAssetSQLBlobAuditingPolicies(f func() (armsql.ServerBlobAuditingPoliciesClientGetResponse, error)) ProviderAPI {
	wrapper := &azureClientWrapper{
		AssetSQLBlobAuditingPolicies: func(ctx context.Context, subID, resourceGroup, sqlServerName string, clientOptions *arm.ClientOptions, options *armsql.ServerBlobAuditingPoliciesClientGetOptions) (armsql.ServerBlobAuditingPoliciesClientGetResponse, error) {
			return f()
		},
	}

	return &provider{
		log:    logp.NewLogger("mock_asset_sql_encryption_protector"),
		client: wrapper,
	}
}

func TestListSQLEncryptionProtector(t *testing.T) {
	tcs := map[string]struct {
		apiMockCall    func() ([]armsql.EncryptionProtectorsClientListByServerResponse, error)
		expectError    bool
		expectedAssets []AzureAsset
	}{
		"Error on calling api": {
			apiMockCall: func() ([]armsql.EncryptionProtectorsClientListByServerResponse, error) {
				return nil, errors.New("error")
			},
			expectError:    true,
			expectedAssets: nil,
		},
		"No Encryption Protector Response": {
			apiMockCall: func() ([]armsql.EncryptionProtectorsClientListByServerResponse, error) {
				return nil, nil
			},
			expectError:    false,
			expectedAssets: nil,
		},
		"Response with encryption protectors in different pages": {
			apiMockCall: func() ([]armsql.EncryptionProtectorsClientListByServerResponse, error) {
				return []armsql.EncryptionProtectorsClientListByServerResponse{
					{
						EncryptionProtectorListResult: armsql.EncryptionProtectorListResult{
							Value: []*armsql.EncryptionProtector{
								{
									Name:     ref("Encryption Protector"),
									ID:       ref("id1"),
									Kind:     ref("azurekeyvault"),
									Location: ref("eu-west"),
									Type:     ref("encryptionProtector"),
									Properties: &armsql.EncryptionProtectorProperties{
										ServerKeyType:       ref(armsql.ServerKeyTypeAzureKeyVault),
										AutoRotationEnabled: ref(true),
										ServerKeyName:       ref("serverKeyName1"),
										Subregion:           ref("eu-west-1"),
									},
								},
								{
									Name:     ref("Encryption Protector"),
									ID:       ref("id2"),
									Kind:     ref("azurekeyvault"),
									Location: ref("eu-west"),
									Type:     ref("encryptionProtector"),
									Properties: &armsql.EncryptionProtectorProperties{
										ServerKeyType:       ref(armsql.ServerKeyTypeAzureKeyVault),
										AutoRotationEnabled: ref(true),
										ServerKeyName:       ref("serverKeyName2"),
										Subregion:           ref("eu-west-1"),
									},
								},
							},
						},
					},
					{
						EncryptionProtectorListResult: armsql.EncryptionProtectorListResult{
							Value: []*armsql.EncryptionProtector{
								{
									Name:     ref("Encryption Protector"),
									ID:       ref("id3"),
									Kind:     ref("azurekeyvault"),
									Location: ref("eu-west"),
									Type:     ref("encryptionProtector"),
									Properties: &armsql.EncryptionProtectorProperties{
										ServerKeyType:       ref(armsql.ServerKeyTypeServiceManaged),
										AutoRotationEnabled: ref(true),
										ServerKeyName:       ref("serverKeyName3"),
										Subregion:           ref("eu-west-1"),
									},
								},
							},
						},
					},
				}, nil
			},
			expectError: false,
			expectedAssets: []AzureAsset{
				{
					Id:             "id1",
					Name:           "Encryption Protector",
					DisplayName:    "",
					Location:       "eu-west",
					ResourceGroup:  "resourceGroup",
					SubscriptionId: "subId",
					Type:           "encryptionProtector",
					TenantId:       "",
					Sku:            nil,
					Identity:       nil,
					Properties: map[string]any{
						"kind":                "azurekeyvault",
						"serverKeyType":       "AzureKeyVault",
						"autoRotationEnabled": true,
						"serverKeyName":       "serverKeyName1",
						"subregion":           "eu-west-1",
						"thumbprint":          "",
						"uri":                 "",
					},
					Extension: nil,
				},
				{
					Id:             "id2",
					Name:           "Encryption Protector",
					DisplayName:    "",
					Location:       "eu-west",
					ResourceGroup:  "resourceGroup",
					SubscriptionId: "subId",
					Type:           "encryptionProtector",
					TenantId:       "",
					Sku:            nil,
					Identity:       nil,
					Properties: map[string]any{
						"kind":                "azurekeyvault",
						"serverKeyType":       "AzureKeyVault",
						"autoRotationEnabled": true,
						"serverKeyName":       "serverKeyName2",
						"subregion":           "eu-west-1",
						"thumbprint":          "",
						"uri":                 "",
					},
					Extension: nil,
				},
				{
					Id:             "id3",
					Name:           "Encryption Protector",
					DisplayName:    "",
					Location:       "eu-west",
					ResourceGroup:  "resourceGroup",
					SubscriptionId: "subId",
					Type:           "encryptionProtector",
					TenantId:       "",
					Sku:            nil,
					Identity:       nil,
					Properties: map[string]any{
						"kind":                "azurekeyvault",
						"serverKeyType":       "ServiceManaged",
						"autoRotationEnabled": true,
						"serverKeyName":       "serverKeyName3",
						"subregion":           "eu-west-1",
						"thumbprint":          "",
						"uri":                 "",
					},
					Extension: nil,
				},
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			p := mockAssetSQLEncryptionProtector(tc.apiMockCall)
			got, err := p.ListSQLEncryptionProtector(context.Background(), "subId", "resourceGroup", "sqlServerInstanceName")

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expectedAssets, got)
		})
	}
}

func TestGetSQLBlobAuditingPolicies(t *testing.T) {
	tcs := map[string]struct {
		apiMockCall    func() (armsql.ServerBlobAuditingPoliciesClientGetResponse, error)
		expectError    bool
		expectedAssets []AzureAsset
	}{
		"Error on calling api": {
			apiMockCall: func() (armsql.ServerBlobAuditingPoliciesClientGetResponse, error) {
				return armsql.ServerBlobAuditingPoliciesClientGetResponse{}, errors.New("error")
			},
			expectError:    true,
			expectedAssets: nil,
		},
		"Response with encryption protectors in different pages": {
			apiMockCall: func() (armsql.ServerBlobAuditingPoliciesClientGetResponse, error) {
				return armsql.ServerBlobAuditingPoliciesClientGetResponse{
					ServerBlobAuditingPolicy: armsql.ServerBlobAuditingPolicy{
						ID:   ref("id1"),
						Name: ref("policy"),
						Type: ref("audit-policy"),
						Properties: &armsql.ServerBlobAuditingPolicyProperties{
							State:                        ref(armsql.BlobAuditingPolicyStateEnabled),
							IsAzureMonitorTargetEnabled:  ref(true),
							IsDevopsAuditEnabled:         ref(false),
							IsManagedIdentityInUse:       ref(true),
							IsStorageSecondaryKeyInUse:   ref(true),
							QueueDelayMs:                 ref(int32(100)),
							RetentionDays:                ref(int32(90)),
							StorageAccountAccessKey:      ref("access-key"),
							StorageAccountSubscriptionID: ref("sub-id"),
							StorageEndpoint:              nil,
							AuditActionsAndGroups:        []*string{ref("a"), ref("b")},
						},
					},
				}, nil
			},
			expectError: false,
			expectedAssets: []AzureAsset{
				{
					Id:             "id1",
					Name:           "policy",
					DisplayName:    "",
					Location:       "global",
					ResourceGroup:  "resourceGroup",
					SubscriptionId: "subId",
					Type:           "audit-policy",
					TenantId:       "",
					Sku:            nil,
					Identity:       nil,
					Properties: map[string]any{
						"state":                        "Enabled",
						"isAzureMonitorTargetEnabled":  true,
						"isDevopsAuditEnabled":         false,
						"isManagedIdentityInUse":       true,
						"isStorageSecondaryKeyInUse":   true,
						"queueDelayMs":                 int32(100),
						"retentionDays":                int32(90),
						"storageAccountAccessKey":      "access-key",
						"storageAccountSubscriptionID": "sub-id",
						"storageEndpoint":              "",
						"auditActionsAndGroups":        []string{"a", "b"},
					},
					Extension: nil,
				},
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			p := mockAssetSQLBlobAuditingPolicies(tc.apiMockCall)
			got, err := p.GetSQLBlobAuditingPolicies(context.Background(), "subId", "resourceGroup", "sqlServerInstanceName")

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expectedAssets, got)
		})
	}
}

func ref[T any](v T) *T {
	return &v
}
