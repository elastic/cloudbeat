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
	type encryptionProtectorResponse struct {
		assets []inventory.AzureAsset
		err    error
	}

	tcs := map[string]struct {
		input                        []inventory.AzureAsset
		expected                     []inventory.AzureAsset
		expectError                  bool
		encryptionProtectorResponses map[string]encryptionProtectorResponse
	}{
		"Some assets have encryption protection, others don't": {
			input: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
				mockOther("id4"),
				mockSQLServer("id2", "serverName2"),
			},
			expected: []inventory.AzureAsset{
				mockSQLServerWithEncryptionProtectorExtension("id1", "serverName1", []map[string]any{
					{
						"kind":                "azurekeyvault",
						"serverKeyType":       "AzureKeyVault",
						"autoRotationEnabled": true,
						"serverKeyName":       "serverKey1",
						"subregion":           "",
						"thumbprint":          "",
						"uri":                 "",
					},
				}),
				mockOther("id4"),
				mockSQLServer("id2", "serverName2"),
			},
			encryptionProtectorResponses: map[string]encryptionProtectorResponse{
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
		},
		"Multiple protectors": {
			input: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
			},
			expected: []inventory.AzureAsset{
				mockSQLServerWithEncryptionProtectorExtension("id1", "serverName1", []map[string]any{
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
				}),
			},
			encryptionProtectorResponses: map[string]encryptionProtectorResponse{
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
		},
		"Error in one protector": {
			expectError: true,
			input: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
				mockSQLServer("id2", "serverName2"),
			},
			expected: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
				mockSQLServerWithEncryptionProtectorExtension("id2", "serverName2", []map[string]any{
					{
						"kind":                "azurekeyvault",
						"serverKeyType":       "AzureKeyVault",
						"autoRotationEnabled": true,
						"serverKeyName":       "serverKey1",
						"subregion":           "",
						"thumbprint":          "",
						"uri":                 "",
					},
				}),
			},
			encryptionProtectorResponses: map[string]encryptionProtectorResponse{
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
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			cmd := cycle.Metadata{}

			provider := azurelib.NewMockProviderAPI(t)
			for serverName, r := range tc.encryptionProtectorResponses {
				provider.EXPECT().ListSQLEncryptionProtector(mock.Anything, "subId", "group", serverName).Return(r.assets, r.err)
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

func mockSQLServerWithEncryptionProtectorExtension(id, name string, p []map[string]any) inventory.AzureAsset {
	m := mockSQLServer(id, name)
	m.AddExtension(inventory.ExtensionEncryptionProtectors, p)
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

func mockOther(id string) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:   id,
		Type: "otherType",
	}
}
