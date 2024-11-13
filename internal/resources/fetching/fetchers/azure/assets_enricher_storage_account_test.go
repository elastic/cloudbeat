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

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/governance"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

func TestStorageAccountEnricher(t *testing.T) {
	storageAccount := func(storageAccountID string) inventory.AzureAsset {
		return inventory.AzureAsset{Id: storageAccountID, Type: inventory.StorageAccountAssetType}
	}

	diagSettings := func(storageAccountID string) inventory.AzureAsset {
		return inventory.AzureAsset{Type: inventory.DiagnosticSettingsAssetType, Properties: map[string]any{"storageAccountId": storageAccountID}}
	}

	serviceDiagSettings := func(storageAccountID string) inventory.AzureAsset {
		return inventory.AzureAsset{Type: inventory.DiagnosticSettingsAssetType, Extension: map[string]any{inventory.ExtensionStorageAccountID: storageAccountID}}
	}
	serviceDiagSettingsAsMap := func(storageAccountID string) map[string]any {
		return map[string]any{"type": inventory.DiagnosticSettingsAssetType, "extension": map[string]any{inventory.ExtensionStorageAccountID: storageAccountID}}
	}

	blobService := func(storageAccountID string) inventory.AzureAsset {
		return inventory.AzureAsset{
			Type:      inventory.BlobServiceAssetType,
			Extension: map[string]any{"storageAccountId": storageAccountID},
		}
	}

	blobServiceAsMap := func(storageAccountID string) map[string]any {
		return map[string]any{"type": inventory.BlobServiceAssetType, "extension": map[string]any{"storageAccountId": storageAccountID}}
	}

	otherAsset := func(id string, assetType string) inventory.AzureAsset {
		return inventory.AzureAsset{Id: id, Type: assetType}
	}

	type mockProviderResponse struct {
		err    error
		assets []inventory.AzureAsset
	}
	mockEmpty := func() mockProviderResponse { return mockProviderResponse{assets: nil, err: nil} }
	mockSuccess := func(a []inventory.AzureAsset) mockProviderResponse { return mockProviderResponse{assets: a, err: nil} }
	mockFail := func(e error) mockProviderResponse { return mockProviderResponse{assets: nil, err: e} }

	tests := map[string]struct {
		inputAssets                              []inventory.AzureAsset
		inputMockDiagSettings                    mockProviderResponse
		inputMockBlobServices                    mockProviderResponse
		inputMockBlobServicesDiagnosticSettings  mockProviderResponse
		inputMockTableServicesDiagnosticSettings mockProviderResponse
		inputMockQueueServicesDiagnosticSettings mockProviderResponse
		inputMockStorageAccounts                 mockProviderResponse
		expected                                 []inventory.AzureAsset
		expectError                              bool
	}{
		"no storage account asset": {
			inputAssets:                              []inventory.AzureAsset{{Id: "id_1", Type: inventory.DiskAssetType}},
			inputMockDiagSettings:                    mockEmpty(),
			inputMockBlobServices:                    mockEmpty(),
			inputMockBlobServicesDiagnosticSettings:  mockEmpty(),
			inputMockTableServicesDiagnosticSettings: mockEmpty(),
			inputMockQueueServicesDiagnosticSettings: mockEmpty(),
			inputMockStorageAccounts:                 mockEmpty(),
			expected:                                 []inventory.AzureAsset{{Id: "id_1", Type: inventory.DiskAssetType}},
			expectError:                              false,
		},
		"storage account asset not used for activity log": {
			inputAssets:                              []inventory.AzureAsset{storageAccount("id_1")},
			inputMockDiagSettings:                    mockEmpty(),
			inputMockBlobServices:                    mockEmpty(),
			inputMockBlobServicesDiagnosticSettings:  mockEmpty(),
			inputMockTableServicesDiagnosticSettings: mockEmpty(),
			inputMockQueueServicesDiagnosticSettings: mockEmpty(),
			inputMockStorageAccounts:                 mockEmpty(),
			expected:                                 []inventory.AzureAsset{storageAccount("id_1")},
			expectError:                              false,
		},
		"storage account asset with blob service": {
			inputAssets:                              []inventory.AzureAsset{storageAccount("id_1")},
			inputMockDiagSettings:                    mockSuccess([]inventory.AzureAsset{{Properties: map[string]any{}}}),
			inputMockBlobServices:                    mockSuccess([]inventory.AzureAsset{blobService("id_1")}),
			inputMockBlobServicesDiagnosticSettings:  mockEmpty(),
			inputMockTableServicesDiagnosticSettings: mockEmpty(),
			inputMockQueueServicesDiagnosticSettings: mockEmpty(),
			inputMockStorageAccounts:                 mockEmpty(),
			expected: []inventory.AzureAsset{
				{
					Id:        "id_1",
					Type:      inventory.StorageAccountAssetType,
					Extension: map[string]any{inventory.ExtensionBlobService: blobServiceAsMap("id_1")},
				},
			},
			expectError: false,
		},
		"storage account asset used for activity log": {
			inputAssets:                              []inventory.AzureAsset{storageAccount("id_1")},
			inputMockDiagSettings:                    mockSuccess([]inventory.AzureAsset{diagSettings("id_1")}),
			inputMockBlobServices:                    mockSuccess([]inventory.AzureAsset{}),
			inputMockBlobServicesDiagnosticSettings:  mockEmpty(),
			inputMockTableServicesDiagnosticSettings: mockEmpty(),
			inputMockQueueServicesDiagnosticSettings: mockEmpty(),
			inputMockStorageAccounts:                 mockEmpty(),
			expected: []inventory.AzureAsset{
				{
					Id:        "id_1",
					Type:      inventory.StorageAccountAssetType,
					Extension: map[string]any{inventory.ExtensionUsedForActivityLogs: true},
				},
			},
			expectError: false,
		},
		"storage account asset with blob service and used for activity log": {
			inputAssets:                              []inventory.AzureAsset{storageAccount("id_1")},
			inputMockDiagSettings:                    mockSuccess([]inventory.AzureAsset{diagSettings("id_1")}),
			inputMockBlobServices:                    mockSuccess([]inventory.AzureAsset{blobService("id_1")}),
			inputMockBlobServicesDiagnosticSettings:  mockEmpty(),
			inputMockTableServicesDiagnosticSettings: mockEmpty(),
			inputMockQueueServicesDiagnosticSettings: mockEmpty(),
			inputMockStorageAccounts:                 mockEmpty(),
			expected: []inventory.AzureAsset{
				{
					Id:   "id_1",
					Type: inventory.StorageAccountAssetType,
					Extension: map[string]any{
						inventory.ExtensionUsedForActivityLogs: true,
						inventory.ExtensionBlobService:         blobServiceAsMap("id_1"),
					},
				},
			},
			expectError: false,
		},
		"multiple storage account asset, one used for activity log": {
			inputAssets: []inventory.AzureAsset{
				storageAccount("id_1"),
				storageAccount("id_2"),
				storageAccount("id_3"),
			},
			inputMockDiagSettings:                    mockSuccess([]inventory.AzureAsset{diagSettings("id_2")}),
			inputMockBlobServices:                    mockEmpty(),
			inputMockBlobServicesDiagnosticSettings:  mockEmpty(),
			inputMockTableServicesDiagnosticSettings: mockEmpty(),
			inputMockQueueServicesDiagnosticSettings: mockEmpty(),
			inputMockStorageAccounts:                 mockEmpty(),
			expected: []inventory.AzureAsset{
				storageAccount("id_1"),
				{Id: "id_2", Type: inventory.StorageAccountAssetType, Extension: map[string]any{inventory.ExtensionUsedForActivityLogs: true}},
				storageAccount("id_3"),
			},
			expectError: false,
		},
		"multiple storage account asset, one with blob services": {
			inputAssets: []inventory.AzureAsset{
				storageAccount("id_1"),
				storageAccount("id_2"),
				storageAccount("id_3"),
			},
			inputMockDiagSettings:                    mockEmpty(),
			inputMockBlobServices:                    mockSuccess([]inventory.AzureAsset{blobService("id_2")}),
			inputMockBlobServicesDiagnosticSettings:  mockEmpty(),
			inputMockTableServicesDiagnosticSettings: mockEmpty(),
			inputMockQueueServicesDiagnosticSettings: mockEmpty(),
			inputMockStorageAccounts:                 mockEmpty(),
			expected: []inventory.AzureAsset{
				storageAccount("id_1"),
				{Id: "id_2", Type: inventory.StorageAccountAssetType, Extension: map[string]any{inventory.ExtensionBlobService: blobServiceAsMap("id_2")}},
				storageAccount("id_3"),
			},
			expectError: false,
		},
		"multiple storage account asset, two used for activity log": {
			inputAssets: []inventory.AzureAsset{
				storageAccount("id_1"),
				storageAccount("id_2"),
				storageAccount("id_3"),
			},
			inputMockDiagSettings: mockSuccess([]inventory.AzureAsset{
				diagSettings("id_2"),
				diagSettings("id_3"),
			}),
			inputMockBlobServices:                    mockEmpty(),
			inputMockBlobServicesDiagnosticSettings:  mockEmpty(),
			inputMockTableServicesDiagnosticSettings: mockEmpty(),
			inputMockQueueServicesDiagnosticSettings: mockEmpty(),
			inputMockStorageAccounts:                 mockEmpty(),
			expected: []inventory.AzureAsset{
				storageAccount("id_1"),
				{Id: "id_2", Type: inventory.StorageAccountAssetType, Extension: map[string]any{inventory.ExtensionUsedForActivityLogs: true}},
				{Id: "id_3", Type: inventory.StorageAccountAssetType, Extension: map[string]any{inventory.ExtensionUsedForActivityLogs: true}},
			},
			expectError: false,
		},
		"multiple storage account asset, two with blob service": {
			inputAssets: []inventory.AzureAsset{
				storageAccount("id_1"),
				storageAccount("id_2"),
				storageAccount("id_3"),
			},
			inputMockDiagSettings:                    mockEmpty(),
			inputMockBlobServices:                    mockSuccess([]inventory.AzureAsset{blobService("id_2"), blobService("id_3")}),
			inputMockBlobServicesDiagnosticSettings:  mockEmpty(),
			inputMockTableServicesDiagnosticSettings: mockEmpty(),
			inputMockQueueServicesDiagnosticSettings: mockEmpty(),
			inputMockStorageAccounts:                 mockEmpty(),
			expected: []inventory.AzureAsset{
				storageAccount("id_1"),
				{Id: "id_2", Type: inventory.StorageAccountAssetType, Extension: map[string]any{inventory.ExtensionBlobService: blobServiceAsMap("id_2")}},
				{Id: "id_3", Type: inventory.StorageAccountAssetType, Extension: map[string]any{inventory.ExtensionBlobService: blobServiceAsMap("id_3")}},
			},
			expectError: false,
		},
		"multiple storage account asset, mixed": {
			inputAssets: []inventory.AzureAsset{
				storageAccount("id_1"),
				storageAccount("id_2"),
				otherAsset("oid_1", inventory.SQLServersAssetType),
				storageAccount("id_3"),
			},
			inputMockDiagSettings:                    mockSuccess([]inventory.AzureAsset{diagSettings("id_1"), diagSettings("id_2")}),
			inputMockBlobServices:                    mockSuccess([]inventory.AzureAsset{blobService("id_2"), blobService("id_3")}),
			inputMockBlobServicesDiagnosticSettings:  mockEmpty(),
			inputMockTableServicesDiagnosticSettings: mockEmpty(),
			inputMockQueueServicesDiagnosticSettings: mockEmpty(),
			inputMockStorageAccounts:                 mockEmpty(),
			expected: []inventory.AzureAsset{
				{Id: "id_1", Type: inventory.StorageAccountAssetType, Extension: map[string]any{inventory.ExtensionUsedForActivityLogs: true}},
				{Id: "id_2", Type: inventory.StorageAccountAssetType, Extension: map[string]any{inventory.ExtensionUsedForActivityLogs: true, inventory.ExtensionBlobService: blobServiceAsMap("id_2")}},
				otherAsset("oid_1", inventory.SQLServersAssetType),
				{Id: "id_3", Type: inventory.StorageAccountAssetType, Extension: map[string]any{inventory.ExtensionBlobService: blobServiceAsMap("id_3")}},
			},
			expectError: false,
		},
		"storage account asset with blob diagnostic settings": {
			inputAssets:                              []inventory.AzureAsset{storageAccount("id_1")},
			inputMockDiagSettings:                    mockEmpty(),
			inputMockBlobServices:                    mockEmpty(),
			inputMockBlobServicesDiagnosticSettings:  mockSuccess([]inventory.AzureAsset{serviceDiagSettings("id_1")}),
			inputMockTableServicesDiagnosticSettings: mockEmpty(),
			inputMockQueueServicesDiagnosticSettings: mockEmpty(),
			inputMockStorageAccounts:                 mockEmpty(),
			expected: []inventory.AzureAsset{
				{
					Id:        "id_1",
					Type:      inventory.StorageAccountAssetType,
					Extension: map[string]any{inventory.ExtensionBlobDiagnosticSettings: serviceDiagSettingsAsMap("id_1")},
				},
			},
			expectError: false,
		},
		"storage account asset with table diagnostic settings": {
			inputAssets:                              []inventory.AzureAsset{storageAccount("id_1")},
			inputMockDiagSettings:                    mockEmpty(),
			inputMockBlobServices:                    mockEmpty(),
			inputMockBlobServicesDiagnosticSettings:  mockEmpty(),
			inputMockTableServicesDiagnosticSettings: mockSuccess([]inventory.AzureAsset{serviceDiagSettings("id_1")}),
			inputMockQueueServicesDiagnosticSettings: mockEmpty(),
			inputMockStorageAccounts:                 mockEmpty(),
			expected: []inventory.AzureAsset{
				{
					Id:        "id_1",
					Type:      inventory.StorageAccountAssetType,
					Extension: map[string]any{inventory.ExtensionTableDiagnosticSettings: serviceDiagSettingsAsMap("id_1")},
				},
			},
			expectError: false,
		},
		"storage account asset with queue diagnostic settings": {
			inputAssets:                              []inventory.AzureAsset{storageAccount("id_1")},
			inputMockDiagSettings:                    mockEmpty(),
			inputMockBlobServices:                    mockEmpty(),
			inputMockBlobServicesDiagnosticSettings:  mockEmpty(),
			inputMockTableServicesDiagnosticSettings: mockEmpty(),
			inputMockQueueServicesDiagnosticSettings: mockSuccess([]inventory.AzureAsset{serviceDiagSettings("id_1")}),
			inputMockStorageAccounts:                 mockEmpty(),
			expected: []inventory.AzureAsset{
				{
					Id:        "id_1",
					Type:      inventory.StorageAccountAssetType,
					Extension: map[string]any{inventory.ExtensionQueueDiagnosticSettings: serviceDiagSettingsAsMap("id_1")},
				},
			},
			expectError: false,
		},
		"enriched storage account asset": {
			inputAssets: []inventory.AzureAsset{
				storageAccount("id_1"),
			},
			inputMockDiagSettings:                    mockEmpty(),
			inputMockBlobServices:                    mockEmpty(),
			inputMockBlobServicesDiagnosticSettings:  mockEmpty(),
			inputMockTableServicesDiagnosticSettings: mockEmpty(),
			inputMockQueueServicesDiagnosticSettings: mockEmpty(),
			inputMockStorageAccounts: mockSuccess([]inventory.AzureAsset{
				{
					Id:   "id_1",
					Type: inventory.StorageAccountAssetType,
					Extension: map[string]any{
						inventory.ExtensionStorageAccountID: "id_1",
					},
				},
			}),
			expected: []inventory.AzureAsset{
				{
					Id:   "id_1",
					Type: inventory.StorageAccountAssetType,
					Extension: map[string]any{
						inventory.ExtensionStorageAccount: map[string]any{
							"id":   "id_1",
							"type": inventory.StorageAccountAssetType,
							"extension": map[string]any{
								inventory.ExtensionStorageAccountID: "id_1",
							},
						},
					},
				},
			},
			expectError: false,
		},
		"multiple storage account asset, mixed with errors": {
			inputAssets: []inventory.AzureAsset{
				storageAccount("id_1"),
				storageAccount("id_2"),
				otherAsset("oid_1", inventory.SQLServersAssetType),
				storageAccount("id_3"),
			},
			inputMockDiagSettings:                    mockSuccess([]inventory.AzureAsset{diagSettings("id_1"), diagSettings("id_2")}),
			inputMockBlobServices:                    mockSuccess([]inventory.AzureAsset{blobService("id_2"), blobService("id_3")}),
			inputMockBlobServicesDiagnosticSettings:  mockFail(errors.New("mock error")),
			inputMockTableServicesDiagnosticSettings: mockEmpty(),
			inputMockQueueServicesDiagnosticSettings: mockSuccess([]inventory.AzureAsset{serviceDiagSettings("id_1")}),
			inputMockStorageAccounts:                 mockEmpty(),
			expected: []inventory.AzureAsset{
				{Id: "id_1", Type: inventory.StorageAccountAssetType, Extension: map[string]any{
					inventory.ExtensionUsedForActivityLogs:     true,
					inventory.ExtensionQueueDiagnosticSettings: serviceDiagSettingsAsMap("id_1"),
				}},
				{Id: "id_2", Type: inventory.StorageAccountAssetType, Extension: map[string]any{
					inventory.ExtensionUsedForActivityLogs: true,
					inventory.ExtensionBlobService:         blobServiceAsMap("id_2"),
				}},
				otherAsset("oid_1", inventory.SQLServersAssetType),
				{Id: "id_3", Type: inventory.StorageAccountAssetType, Extension: map[string]any{
					inventory.ExtensionBlobService: blobServiceAsMap("id_3"),
				}},
			},
			expectError: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := cycle.Metadata{}

			provider := azurelib.NewMockProviderAPI(t)
			provider.EXPECT().GetSubscriptions(mock.Anything, cmd).Return(map[string]governance.Subscription{"sub1": {}}, nil).Once() //nolint:exhaustruct
			provider.EXPECT().
				ListDiagnosticSettingsAssetTypes(mock.Anything, cmd, []string{"sub1"}).
				Return(tc.inputMockDiagSettings.assets, tc.inputMockDiagSettings.err)

			var storageAccounts []inventory.AzureAsset
			for _, a := range tc.inputAssets {
				if a.Type == inventory.StorageAccountAssetType {
					storageAccounts = append(storageAccounts, a)
				}
			}

			if len(storageAccounts) > 0 {
				storageAccountsCompare := func(sa []inventory.AzureAsset) bool {
					// loose comparing using only ids because of enricher mutation.
					expectedIDs := lo.Map(storageAccounts, func(item inventory.AzureAsset, _ int) string { return item.Id })
					gotIDs := lo.Map(sa, func(item inventory.AzureAsset, _ int) string { return item.Id })
					return assert.ElementsMatch(t, expectedIDs, gotIDs)
				}
				provider.EXPECT().
					ListStorageAccountBlobServices(mock.Anything, mock.MatchedBy(storageAccountsCompare)).
					Return(tc.inputMockBlobServices.assets, tc.inputMockBlobServices.err)
				provider.EXPECT().
					ListStorageAccountsBlobDiagnosticSettings(mock.Anything, mock.MatchedBy(storageAccountsCompare)).
					Return(tc.inputMockBlobServicesDiagnosticSettings.assets, tc.inputMockBlobServicesDiagnosticSettings.err)
				provider.EXPECT().
					ListStorageAccountsTableDiagnosticSettings(mock.Anything, mock.MatchedBy(storageAccountsCompare)).
					Return(tc.inputMockTableServicesDiagnosticSettings.assets, tc.inputMockTableServicesDiagnosticSettings.err)
				provider.EXPECT().
					ListStorageAccountsQueueDiagnosticSettings(mock.Anything, mock.MatchedBy(storageAccountsCompare)).
					Return(tc.inputMockQueueServicesDiagnosticSettings.assets, tc.inputMockQueueServicesDiagnosticSettings.err)
				provider.EXPECT().
					ListStorageAccounts(mock.Anything, mock.AnythingOfType("[]string")).
					Return(tc.inputMockStorageAccounts.assets, tc.inputMockStorageAccounts.err)
			}

			e := storageAccountEnricher{provider: provider}
			err := e.Enrich(context.Background(), cmd, tc.inputAssets)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expected, tc.inputAssets)
		})
	}
}
