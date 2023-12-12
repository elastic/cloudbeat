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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/governance"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
)

func TestAddUsedForActivityLogsFlag(t *testing.T) {
	tests := map[string]struct {
		inputAssets       []inventory.AzureAsset
		inputDiagSettings []inventory.AzureAsset
		expected          []inventory.AzureAsset
	}{
		"no storage account asset": {
			inputAssets:       []inventory.AzureAsset{{Id: "id_1", Type: inventory.DiskAssetType}},
			inputDiagSettings: []inventory.AzureAsset{{Properties: map[string]any{}}},
			expected:          []inventory.AzureAsset{{Id: "id_1", Type: inventory.DiskAssetType}},
		},
		"storage account asset not used for activity log": {
			inputAssets:       []inventory.AzureAsset{{Id: "id_1", Type: inventory.StorageAccountAssetType}},
			inputDiagSettings: []inventory.AzureAsset{{Properties: map[string]any{}}},
			expected:          []inventory.AzureAsset{{Id: "id_1", Type: inventory.StorageAccountAssetType}},
		},
		"storage account asset used for activity log": {
			inputAssets:       []inventory.AzureAsset{{Id: "id_1", Type: inventory.StorageAccountAssetType}},
			inputDiagSettings: []inventory.AzureAsset{{Properties: map[string]any{"storageAccountId": "id_1"}}},
			expected:          []inventory.AzureAsset{{Id: "id_1", Type: inventory.StorageAccountAssetType, Extension: map[string]any{"usedForActivityLogs": true}}},
		},
		"multiple storage account asset, one used for activity log": {
			inputAssets: []inventory.AzureAsset{
				{Id: "id_1", Type: inventory.StorageAccountAssetType},
				{Id: "id_2", Type: inventory.StorageAccountAssetType},
				{Id: "id_3", Type: inventory.StorageAccountAssetType},
			},
			inputDiagSettings: []inventory.AzureAsset{{Properties: map[string]any{"storageAccountId": "id_2"}}},
			expected: []inventory.AzureAsset{
				{Id: "id_1", Type: inventory.StorageAccountAssetType},
				{Id: "id_2", Type: inventory.StorageAccountAssetType, Extension: map[string]any{"usedForActivityLogs": true}},
				{Id: "id_3", Type: inventory.StorageAccountAssetType},
			},
		},
		"multiple storage account asset, two used for activity log": {
			inputAssets: []inventory.AzureAsset{
				{Id: "id_1", Type: inventory.StorageAccountAssetType},
				{Id: "id_2", Type: inventory.StorageAccountAssetType},
				{Id: "id_3", Type: inventory.StorageAccountAssetType},
			},
			inputDiagSettings: []inventory.AzureAsset{
				{Properties: map[string]any{"storageAccountId": "id_2"}},
				{Properties: map[string]any{"storageAccountId": "id_3"}},
			},
			expected: []inventory.AzureAsset{
				{Id: "id_1", Type: inventory.StorageAccountAssetType},
				{Id: "id_2", Type: inventory.StorageAccountAssetType, Extension: map[string]any{"usedForActivityLogs": true}},
				{Id: "id_3", Type: inventory.StorageAccountAssetType, Extension: map[string]any{"usedForActivityLogs": true}},
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			cmd := cycle.Metadata{}

			provider := azurelib.NewMockProviderAPI(t)
			provider.EXPECT().GetSubscriptions(mock.Anything, cmd).Return(map[string]governance.Subscription{"sub1": {}}, nil).Once()
			provider.EXPECT().ListDiagnosticSettingsAssetTypes(mock.Anything, cmd, []string{"sub1"}).Return(tc.inputDiagSettings, nil)

			e := storageAccountEnricher{provider: provider}
			err := e.Enrich(context.Background(), cmd, tc.inputAssets)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, tc.inputAssets)
		})
	}
}
