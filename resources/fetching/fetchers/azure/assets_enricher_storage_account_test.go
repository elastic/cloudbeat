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
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/governance"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
)

func TestStorageAccountEnricher(t *testing.T) {
	storageAccount := func(storageAccountID string) inventory.AzureAsset {
		return inventory.AzureAsset{Id: storageAccountID, Type: inventory.StorageAccountAssetType}
	}

	diagSettings := func(storageAccountID string) inventory.AzureAsset {
		return inventory.AzureAsset{Type: inventory.DiagnosticSettingsAssetType, Properties: map[string]any{"storageAccountId": storageAccountID}}
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

	tests := map[string]struct {
		inputAssets       []inventory.AzureAsset
		inputDiagSettings []inventory.AzureAsset
		inputBlobServices []inventory.AzureAsset
		expected          []inventory.AzureAsset
	}{
		"no storage account asset": {
			inputAssets:       []inventory.AzureAsset{{Id: "id_1", Type: inventory.DiskAssetType}},
			inputDiagSettings: []inventory.AzureAsset{},
			inputBlobServices: []inventory.AzureAsset{},
			expected:          []inventory.AzureAsset{{Id: "id_1", Type: inventory.DiskAssetType}},
		},
		"storage account asset not used for activity log": {
			inputAssets:       []inventory.AzureAsset{storageAccount("id_1")},
			inputDiagSettings: []inventory.AzureAsset{},
			inputBlobServices: []inventory.AzureAsset{},
			expected:          []inventory.AzureAsset{storageAccount("id_1")},
		},
		"storage account asset with blob service": {
			inputAssets:       []inventory.AzureAsset{storageAccount("id_1")},
			inputDiagSettings: []inventory.AzureAsset{{Properties: map[string]any{}}},
			inputBlobServices: []inventory.AzureAsset{blobService("id_1")},
			expected: []inventory.AzureAsset{
				{
					Id:        "id_1",
					Type:      inventory.StorageAccountAssetType,
					Extension: map[string]any{inventory.ExtensionBlobService: blobServiceAsMap("id_1")},
				},
			},
		},
		"storage account asset used for activity log": {
			inputAssets:       []inventory.AzureAsset{storageAccount("id_1")},
			inputDiagSettings: []inventory.AzureAsset{diagSettings("id_1")},
			inputBlobServices: []inventory.AzureAsset{},
			expected: []inventory.AzureAsset{
				{
					Id:        "id_1",
					Type:      inventory.StorageAccountAssetType,
					Extension: map[string]any{inventory.ExtensionUsedForActivityLogs: true},
				},
			},
		},
		"storage account asset with blob service and used for activity log": {
			inputAssets:       []inventory.AzureAsset{storageAccount("id_1")},
			inputDiagSettings: []inventory.AzureAsset{diagSettings("id_1")},
			inputBlobServices: []inventory.AzureAsset{blobService("id_1")},
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
		},
		"multiple storage account asset, one used for activity log": {
			inputAssets: []inventory.AzureAsset{
				storageAccount("id_1"),
				storageAccount("id_2"),
				storageAccount("id_3"),
			},
			inputDiagSettings: []inventory.AzureAsset{
				diagSettings("id_2"),
			},
			inputBlobServices: []inventory.AzureAsset{},
			expected: []inventory.AzureAsset{
				storageAccount("id_1"),
				{Id: "id_2", Type: inventory.StorageAccountAssetType, Extension: map[string]any{inventory.ExtensionUsedForActivityLogs: true}},
				storageAccount("id_3"),
			},
		},
		"multiple storage account asset, one with blob services": {
			inputAssets: []inventory.AzureAsset{
				storageAccount("id_1"),
				storageAccount("id_2"),
				storageAccount("id_3"),
			},
			inputDiagSettings: []inventory.AzureAsset{},
			inputBlobServices: []inventory.AzureAsset{
				blobService("id_2"),
			},
			expected: []inventory.AzureAsset{
				storageAccount("id_1"),
				{Id: "id_2", Type: inventory.StorageAccountAssetType, Extension: map[string]any{inventory.ExtensionBlobService: blobServiceAsMap("id_2")}},
				storageAccount("id_3"),
			},
		},
		"multiple storage account asset, two used for activity log": {
			inputAssets: []inventory.AzureAsset{
				storageAccount("id_1"),
				storageAccount("id_2"),
				storageAccount("id_3"),
			},
			inputDiagSettings: []inventory.AzureAsset{
				diagSettings("id_2"),
				diagSettings("id_3"),
			},
			inputBlobServices: []inventory.AzureAsset{},
			expected: []inventory.AzureAsset{
				storageAccount("id_1"),
				{Id: "id_2", Type: inventory.StorageAccountAssetType, Extension: map[string]any{inventory.ExtensionUsedForActivityLogs: true}},
				{Id: "id_3", Type: inventory.StorageAccountAssetType, Extension: map[string]any{inventory.ExtensionUsedForActivityLogs: true}},
			},
		},
		"multiple storage account asset, two with blob service": {
			inputAssets: []inventory.AzureAsset{
				storageAccount("id_1"),
				storageAccount("id_2"),
				storageAccount("id_3"),
			},
			inputDiagSettings: []inventory.AzureAsset{},
			inputBlobServices: []inventory.AzureAsset{
				blobService("id_2"),
				blobService("id_3"),
			},
			expected: []inventory.AzureAsset{
				storageAccount("id_1"),
				{Id: "id_2", Type: inventory.StorageAccountAssetType, Extension: map[string]any{inventory.ExtensionBlobService: blobServiceAsMap("id_2")}},
				{Id: "id_3", Type: inventory.StorageAccountAssetType, Extension: map[string]any{inventory.ExtensionBlobService: blobServiceAsMap("id_3")}},
			},
		},
		"multiple storage account asset, mixed": {
			inputAssets: []inventory.AzureAsset{
				storageAccount("id_1"),
				storageAccount("id_2"),
				otherAsset("oid_1", inventory.SQLServersAssetType),
				storageAccount("id_3"),
			},
			inputDiagSettings: []inventory.AzureAsset{
				diagSettings("id_1"),
				diagSettings("id_2"),
			},
			inputBlobServices: []inventory.AzureAsset{
				blobService("id_2"),
				blobService("id_3"),
			},
			expected: []inventory.AzureAsset{
				{Id: "id_1", Type: inventory.StorageAccountAssetType, Extension: map[string]any{inventory.ExtensionUsedForActivityLogs: true}},
				{Id: "id_2", Type: inventory.StorageAccountAssetType, Extension: map[string]any{inventory.ExtensionUsedForActivityLogs: true, inventory.ExtensionBlobService: blobServiceAsMap("id_2")}},
				otherAsset("oid_1", inventory.SQLServersAssetType),
				{Id: "id_3", Type: inventory.StorageAccountAssetType, Extension: map[string]any{inventory.ExtensionBlobService: blobServiceAsMap("id_3")}},
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			cmd := cycle.Metadata{}

			provider := azurelib.NewMockProviderAPI(t)
			provider.EXPECT().GetSubscriptions(mock.Anything, cmd).Return(map[string]governance.Subscription{"sub1": {}}, nil).Once() //nolint:exhaustruct
			provider.EXPECT().ListDiagnosticSettingsAssetTypes(mock.Anything, cmd, []string{"sub1"}).Return(tc.inputDiagSettings, nil)

			var storageAccounts []inventory.AzureAsset
			for _, a := range tc.inputAssets {
				if a.Type == inventory.StorageAccountAssetType {
					storageAccounts = append(storageAccounts, a)
				}
			}

			if len(storageAccounts) > 0 {
				provider.EXPECT().
					ListStorageAccountBlobServices(
						mock.Anything,
						mock.MatchedBy(func(sa []inventory.AzureAsset) bool {
							// loose comparing using only ids.
							expectedIDs := lo.Map(storageAccounts, func(item inventory.AzureAsset, _ int) string { return item.Id })
							gotIDs := lo.Map(sa, func(item inventory.AzureAsset, _ int) string { return item.Id })
							return assert.Equal(t, expectedIDs, gotIDs)
						})).
					Return(tc.inputBlobServices, nil)
			}

			e := storageAccountEnricher{provider: provider}
			err := e.Enrich(context.Background(), cmd, tc.inputAssets)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, tc.inputAssets)
		})
	}
}
