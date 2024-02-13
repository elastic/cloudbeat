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

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

func TestKeyVaultEnricher(t *testing.T) {
	assetVault := func(id string) inventory.AzureAsset {
		return inventory.AzureAsset{
			Id:   id,
			Type: inventory.VaultAssetType,
		}
	}
	assetKey := func(id string) inventory.AzureAsset {
		return inventory.AzureAsset{
			Id:   id,
			Type: "Microsoft.KeyVault/vaults/keys",
		}
	}
	assetSecret := func(id string) inventory.AzureAsset {
		return inventory.AzureAsset{
			Id:   id,
			Type: "Microsoft.KeyVault/vaults/secrets",
		}
	}

	tests := map[string]struct {
		inputAssets                []inventory.AzureAsset
		mockKeysPerVaultID         map[string][]inventory.AzureAsset
		mockSecretsPerVaultID      map[string][]inventory.AzureAsset
		mockedDiagnosticPerVaultID map[string][]inventory.AzureAsset
		expected                   []inventory.AzureAsset
		expectError                bool
	}{
		"single": {
			inputAssets: []inventory.AzureAsset{
				assetVault("id1"),
			},
			mockKeysPerVaultID: map[string][]inventory.AzureAsset{
				"id1": {assetKey("key1")},
			},
			mockSecretsPerVaultID: map[string][]inventory.AzureAsset{
				"id1": {assetSecret("sec1")},
			},
			mockedDiagnosticPerVaultID: map[string][]inventory.AzureAsset{
				"id1": {assetVault("diag1")},
			},
			expected: func() []inventory.AzureAsset {
				a := assetVault("id1")
				a.Extension = map[string]any{}
				a.Extension[inventory.ExtensionKeyVaultKeys] = []inventory.AzureAsset{assetKey("key1")}
				a.Extension[inventory.ExtensionKeyVaultSecrets] = []inventory.AzureAsset{assetSecret("sec1")}
				a.Extension[inventory.ExtensionKeyVaultDiagnosticSettings] = []inventory.AzureAsset{assetVault("diag1")}
				return []inventory.AzureAsset{a}
			}(),
			expectError: false,
		},
		"multiple": {
			inputAssets: []inventory.AzureAsset{
				assetVault("id1"),
				assetVault("id2"),
				assetVault("id3"),
			},
			mockKeysPerVaultID: map[string][]inventory.AzureAsset{
				"id1": {assetKey("key1"), assetKey("key2")},
				"id2": {assetKey("key3")},
				"id3": {},
			},
			mockSecretsPerVaultID: map[string][]inventory.AzureAsset{
				"id1": {assetSecret("sec1")},
				"id2": {},
				"id3": {assetSecret("sec2"), assetSecret("sec3")},
			},
			mockedDiagnosticPerVaultID: map[string][]inventory.AzureAsset{
				"id1": {assetVault("diag1")},
				"id2": {},
				"id3": {assetVault("diag3")},
			},
			expected: func() []inventory.AzureAsset {
				a := assetVault("id1")
				a.Extension = map[string]any{}
				a.Extension[inventory.ExtensionKeyVaultKeys] = []inventory.AzureAsset{assetKey("key1"), assetKey("key2")}
				a.Extension[inventory.ExtensionKeyVaultSecrets] = []inventory.AzureAsset{assetSecret("sec1")}
				a.Extension[inventory.ExtensionKeyVaultDiagnosticSettings] = []inventory.AzureAsset{assetVault("diag1")}

				b := assetVault("id2")
				b.Extension = map[string]any{}
				b.Extension[inventory.ExtensionKeyVaultKeys] = []inventory.AzureAsset{assetKey("key3")}

				c := assetVault("id3")
				c.Extension = map[string]any{}
				c.Extension[inventory.ExtensionKeyVaultSecrets] = []inventory.AzureAsset{assetSecret("sec2"), assetSecret("sec3")}
				c.Extension[inventory.ExtensionKeyVaultDiagnosticSettings] = []inventory.AzureAsset{assetVault("diag3")}
				return []inventory.AzureAsset{a, b, c}
			}(),
			expectError: false,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			mockProvider := azurelib.NewMockProviderAPI(t)

			mockProvider.EXPECT().
				ListKeyVaultKeys(mock.Anything, mock.MatchedBy(func(a inventory.AzureAsset) bool {
					_, found := tc.mockKeysPerVaultID[a.Id]
					return found
				})).
				RunAndReturn(func(_ context.Context, a inventory.AzureAsset) ([]inventory.AzureAsset, error) {
					sl := tc.mockKeysPerVaultID[a.Id]
					return sl, nil
				}).
				Times(len(tc.inputAssets))

			mockProvider.EXPECT().
				ListKeyVaultSecrets(mock.Anything, mock.MatchedBy(func(a inventory.AzureAsset) bool {
					_, found := tc.mockSecretsPerVaultID[a.Id]
					return found
				})).
				RunAndReturn(func(_ context.Context, a inventory.AzureAsset) ([]inventory.AzureAsset, error) {
					sl := tc.mockSecretsPerVaultID[a.Id]
					return sl, nil
				}).
				Times(len(tc.inputAssets))
			mockProvider.EXPECT().
				ListKeyVaultDiagnosticSettings(mock.Anything, mock.MatchedBy(func(a inventory.AzureAsset) bool {
					_, found := tc.mockedDiagnosticPerVaultID[a.Id]
					return found
				})).
				RunAndReturn(func(_ context.Context, a inventory.AzureAsset) ([]inventory.AzureAsset, error) {
					sl := tc.mockedDiagnosticPerVaultID[a.Id]
					return sl, nil
				}).
				Times(len(tc.inputAssets))

			enricher := keyVaultEnricher{provider: mockProvider}

			err := enricher.Enrich(context.Background(), cycle.Metadata{}, tc.inputAssets)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expected, tc.inputAssets)
		})
	}
}
