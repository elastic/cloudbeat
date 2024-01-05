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

package fetchers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
)

func TestSQLServerEnricher_Enrich(t *testing.T) {
	type enricherResponse struct {
		assets []inventory.AzureAsset
		err    error
	}

	tcs := map[string]struct {
		input                       []inventory.AzureAsset
		expected                    []inventory.AzureAsset
		expectError                 bool
		encryptionProtectorResponse map[string]enricherResponse
		blobAuditPolicyResponse     map[string]enricherResponse
	}{
		"Some assets have encryption protection, others don't": {
			input: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
				mockOther("id4"),
				mockSQLServer("id2", "serverName2"),
			},
			expected: []inventory.AzureAsset{
				mockSQLServerWithEncryptionProtectorExtension("id1", "serverName1", map[string]any{
					inventory.ExtensionSQLEncryptionProtectors: []map[string]any{
						{
							"kind":                "azurekeyvault",
							"serverKeyType":       "AzureKeyVault",
							"autoRotationEnabled": true,
							"serverKeyName":       "serverKey1",
							"subregion":           "",
							"thumbprint":          "",
							"uri":                 "",
						},
					},
				}),
				mockOther("id4"),
				mockSQLServer("id2", "serverName2"),
			},
			encryptionProtectorResponse: map[string]enricherResponse{
				"serverName1": { //nolint:exhaustruct
					assets: []inventory.AzureAsset{
						mockEncryptionProtector("ep1", map[string]any{
							"kind":                "azurekeyvault",
							"serverKeyType":       "AzureKeyVault",
							"autoRotationEnabled": true,
							"serverKeyName":       "serverKey1",
							"subregion":           "",
							"thumbprint":          "",
							"uri":                 "",
						}),
					},
				},
				"serverName2": {}, //nolint:exhaustruct
			},
			blobAuditPolicyResponse: map[string]enricherResponse{
				"serverName1": {}, //nolint:exhaustruct
				"serverName2": {}, //nolint:exhaustruct
			},
		},
		"Multiple policy protectors": {
			input: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
			},
			expected: []inventory.AzureAsset{
				mockSQLServerWithEncryptionProtectorExtension("id1", "serverName1", map[string]any{
					inventory.ExtensionSQLEncryptionProtectors: []map[string]any{
						{
							"kind":                "azurekeyvault",
							"serverKeyType":       "AzureKeyVault",
							"autoRotationEnabled": true,
							"serverKeyName":       "serverKey1",
							"subregion":           "",
							"thumbprint":          "",
							"uri":                 "",
						},
						{
							"serverKeyType":       "ServiceManaged",
							"autoRotationEnabled": false,
							"serverKeyName":       "serverKey2",
							"subregion":           "",
							"thumbprint":          "",
							"uri":                 "",
						},
					},
				}),
			},
			encryptionProtectorResponse: map[string]enricherResponse{
				"serverName1": { //nolint:exhaustruct
					assets: []inventory.AzureAsset{
						mockEncryptionProtector("ep1", map[string]any{
							"kind":                "azurekeyvault",
							"serverKeyType":       "AzureKeyVault",
							"autoRotationEnabled": true,
							"serverKeyName":       "serverKey1",
							"subregion":           "",
							"thumbprint":          "",
							"uri":                 "",
						}),
						mockEncryptionProtector("ep2", map[string]any{
							"serverKeyType":       "ServiceManaged",
							"autoRotationEnabled": false,
							"serverKeyName":       "serverKey2",
							"subregion":           "",
							"thumbprint":          "",
							"uri":                 "",
						}),
					},
				},
			},
			blobAuditPolicyResponse: map[string]enricherResponse{
				"serverName1": {}, //nolint:exhaustruct
			},
		},
		"Error in one policy protector": {
			expectError: true,
			input: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
				mockSQLServer("id2", "serverName2"),
			},
			expected: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
				mockSQLServerWithEncryptionProtectorExtension("id2", "serverName2", map[string]any{
					inventory.ExtensionSQLEncryptionProtectors: []map[string]any{
						{
							"kind":                "azurekeyvault",
							"serverKeyType":       "AzureKeyVault",
							"autoRotationEnabled": true,
							"serverKeyName":       "serverKey1",
							"subregion":           "",
							"thumbprint":          "",
							"uri":                 "",
						},
					},
				}),
			},
			encryptionProtectorResponse: map[string]enricherResponse{
				"serverName1": { //nolint:exhaustruct
					err: errors.New("error"),
				},
				"serverName2": { //nolint:exhaustruct
					assets: []inventory.AzureAsset{
						mockEncryptionProtector("ep1", map[string]any{
							"kind":                "azurekeyvault",
							"serverKeyType":       "AzureKeyVault",
							"autoRotationEnabled": true,
							"serverKeyName":       "serverKey1",
							"subregion":           "",
							"thumbprint":          "",
							"uri":                 "",
						}),
					},
				},
			},
			blobAuditPolicyResponse: map[string]enricherResponse{
				"serverName1": {}, //nolint:exhaustruct
				"serverName2": {}, //nolint:exhaustruct
			},
		},
		"Error in one blob audit policy": {
			expectError: true,
			input: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
				mockSQLServer("id2", "serverName2"),
			},
			expected: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
				mockSQLServerWithEncryptionProtectorExtension("id2", "serverName2", map[string]any{
					inventory.ExtensionSQLEncryptionProtectors: []map[string]any{
						{
							"kind":                "azurekeyvault",
							"serverKeyType":       "AzureKeyVault",
							"autoRotationEnabled": true,
							"serverKeyName":       "serverKey1",
							"subregion":           "",
							"thumbprint":          "",
							"uri":                 "",
						},
					},
				}),
			},
			encryptionProtectorResponse: map[string]enricherResponse{
				"serverName1": {}, //nolint:exhaustruct
				"serverName2": { //nolint:exhaustruct
					assets: []inventory.AzureAsset{
						mockEncryptionProtector("ep1", map[string]any{
							"kind":                "azurekeyvault",
							"serverKeyType":       "AzureKeyVault",
							"autoRotationEnabled": true,
							"serverKeyName":       "serverKey1",
							"subregion":           "",
							"thumbprint":          "",
							"uri":                 "",
						}),
					},
				},
			},
			blobAuditPolicyResponse: map[string]enricherResponse{
				"serverName1": {},                         //nolint:exhaustruct
				"serverName2": {err: errors.New("error")}, //nolint:exhaustruct
			},
		},
		"Policy protector and blob audit policies": {
			input: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
			},
			expected: []inventory.AzureAsset{
				mockSQLServerWithEncryptionProtectorExtension("id1", "serverName1", map[string]any{
					inventory.ExtensionSQLEncryptionProtectors: []map[string]any{
						{
							"kind":                "azurekeyvault",
							"serverKeyType":       "AzureKeyVault",
							"autoRotationEnabled": true,
							"serverKeyName":       "serverKey1",
							"subregion":           "",
							"thumbprint":          "",
							"uri":                 "",
						},
					},
					inventory.ExtensionSQLBlobAuditPolicy: map[string]any{
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
				}),
			},
			encryptionProtectorResponse: map[string]enricherResponse{
				"serverName1": { //nolint:exhaustruct
					assets: []inventory.AzureAsset{
						mockEncryptionProtector("ep1", map[string]any{
							"kind":                "azurekeyvault",
							"serverKeyType":       "AzureKeyVault",
							"autoRotationEnabled": true,
							"serverKeyName":       "serverKey1",
							"subregion":           "",
							"thumbprint":          "",
							"uri":                 "",
						}),
					},
				},
			},
			blobAuditPolicyResponse: map[string]enricherResponse{
				"serverName1": { //nolint:exhaustruct
					assets: []inventory.AzureAsset{
						mockBlobAuditingPolicies("ep1", map[string]any{
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
						}),
					},
				},
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			cmd := cycle.Metadata{}

			provider := azurelib.NewMockProviderAPI(t)
			for serverName, r := range tc.encryptionProtectorResponse {
				provider.EXPECT().ListSQLEncryptionProtector(mock.Anything, "subId", "group", serverName).Return(r.assets, r.err)
			}

			for serverName, r := range tc.blobAuditPolicyResponse {
				provider.EXPECT().GetSQLBlobAuditingPolicies(mock.Anything, "subId", "group", serverName).Return(r.assets, r.err)
			}

			e := sqlServerEnricher{provider: provider}

			err := e.Enrich(context.Background(), cmd, tc.input)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.expected, tc.input)
		})
	}
}

func mockSQLServerWithEncryptionProtectorExtension(id, name string, ext map[string]any) inventory.AzureAsset {
	m := mockSQLServer(id, name)
	m.Extension = ext
	return m
}

func mockSQLServer(id, name string) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:             id,
		SubscriptionId: "subId",
		ResourceGroup:  "group",
		Name:           name,
		Type:           inventory.SQLServersAssetType,
	}
}

func mockEncryptionProtector(id string, props map[string]any) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:         id,
		Type:       inventory.SQLServersAssetType + "/encryptionProtector",
		Properties: props,
	}
}

func mockBlobAuditingPolicies(id string, props map[string]any) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:         id,
		Type:       inventory.SQLServersAssetType + "/blobAuditPolicy",
		Properties: props,
	}
}

func mockOther(id string) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:   id,
		Type: "otherType",
	}
}
