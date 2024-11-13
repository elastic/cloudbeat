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
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type mockAzureKeyVaultWrapper struct {
	mock.Mock
}

func (m *mockAzureKeyVaultWrapper) AssetKeyVaultKeys(_ context.Context, subscriptionID string, resourceGroupName string, vaultName string) ([]armkeyvault.KeysClientListResponse, error) {
	r := m.Called(subscriptionID, resourceGroupName, vaultName)
	return r.Get(0).([]armkeyvault.KeysClientListResponse), r.Error(1)
}

func (m *mockAzureKeyVaultWrapper) AssetKeyVaultSecrets(_ context.Context, subscriptionID string, resourceGroupName string, vaultName string) ([]armkeyvault.SecretsClientListResponse, error) {
	r := m.Called(subscriptionID, resourceGroupName, vaultName)
	return r.Get(0).([]armkeyvault.SecretsClientListResponse), r.Error(1)
}

func (m *mockAzureKeyVaultWrapper) AssetDiagnosticSettings(_ context.Context, vaultId string, options *armmonitor.DiagnosticSettingsClientListOptions) ([]armmonitor.DiagnosticSettingsClientListResponse, error) {
	r := m.Called(vaultId, options)
	return r.Get(0).([]armmonitor.DiagnosticSettingsClientListResponse), r.Error(1)
}

func TestListKeyVaultDiagnosticSettings(t *testing.T) {
	log := testhelper.NewLogger(t)

	response := func(settings ...*armmonitor.DiagnosticSettingsResource) armmonitor.DiagnosticSettingsClientListResponse {
		return armmonitor.DiagnosticSettingsClientListResponse{
			DiagnosticSettingsResourceCollection: armmonitor.DiagnosticSettingsResourceCollection{
				Value: settings,
			},
		}
	}

	settings := func(id string) *armmonitor.DiagnosticSettingsResource {
		return &armmonitor.DiagnosticSettingsResource{
			ID:   to.Ptr(id),
			Name: to.Ptr("diagName"),
			Type: to.Ptr("Microsoft.KeyVault/vaults"),
			Properties: &armmonitor.DiagnosticSettings{
				StorageAccountID: to.Ptr("storage_account_id"),
				Logs: []*armmonitor.LogSettings{
					{
						Category: to.Ptr("AuditEvent"),
						Enabled:  to.Ptr(true),
					},
				},
			},
		}
	}

	vaultAsset := AzureAsset{
		Id:             "kv1",
		Name:           "diagName",
		ResourceGroup:  "rg1",
		SubscriptionId: "sub1",
		TenantId:       "ten1",
	}

	tests := map[string]struct {
		inputVault               AzureAsset
		mockWrapperResponse      []armmonitor.DiagnosticSettingsClientListResponse
		mockWrapperResponseError error
		expected                 []AzureAsset
		expectError              bool
	}{
		"test error": {
			inputVault:               vaultAsset,
			mockWrapperResponse:      nil,
			mockWrapperResponseError: errors.New("some error"),
			expected:                 []AzureAsset{},
			expectError:              true,
		},
		"test single": {
			inputVault: vaultAsset,
			mockWrapperResponse: []armmonitor.DiagnosticSettingsClientListResponse{
				response(nil, settings("diag1")),
			},
			mockWrapperResponseError: nil,
			expected: []AzureAsset{
				{
					Id:             "diag1",
					Name:           "diagName",
					DisplayName:    "",
					ResourceGroup:  "rg1",
					SubscriptionId: "sub1",
					TenantId:       "ten1",
					Type:           "Microsoft.KeyVault/vaults",
					Properties: map[string]any{
						"storageAccountId": settings("diag1").Properties.StorageAccountID,
						"logs":             settings("diag1").Properties.Logs,
					},
				},
			},
			expectError: false,
		},
		"test multiple": {
			inputVault: AzureAsset{
				Id:             "kv1",
				Name:           "name1",
				ResourceGroup:  "rg1",
				SubscriptionId: "sub1",
				TenantId:       "ten1",
			},
			mockWrapperResponse: []armmonitor.DiagnosticSettingsClientListResponse{
				response(nil, settings("diag1"), settings("diag2")),
			},
			mockWrapperResponseError: nil,
			expected: []AzureAsset{
				{
					Id:             "diag1",
					Name:           "diagName",
					DisplayName:    "",
					ResourceGroup:  "rg1",
					SubscriptionId: "sub1",
					TenantId:       "ten1",
					Type:           "Microsoft.KeyVault/vaults",
					Properties: map[string]any{
						"storageAccountId": settings("diag1").Properties.StorageAccountID,
						"logs":             settings("diag1").Properties.Logs,
					},
				},
				{
					Id:             "diag2",
					Name:           "diagName",
					DisplayName:    "",
					ResourceGroup:  "rg1",
					SubscriptionId: "sub1",
					TenantId:       "ten1",
					Type:           "Microsoft.KeyVault/vaults",
					Properties: map[string]any{
						"storageAccountId": settings("diag2").Properties.StorageAccountID,
						"logs":             settings("diag2").Properties.Logs,
					},
				},
			},
			expectError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockWrapper := &mockAzureKeyVaultWrapper{}
			mockWrapper.Test(t)
			mockWrapper.
				On("AssetDiagnosticSettings", tc.inputVault.Id, mock.Anything).
				Return(tc.mockWrapperResponse, tc.mockWrapperResponseError).
				Once()
			t.Cleanup(func() { mockWrapper.AssertExpectations(t) })

			provider := keyVaultProvider{
				log: log,
				client: &azureKeyVaultWrapper{
					AssetDiagnosticSettings: mockWrapper.AssetDiagnosticSettings,
				},
			}

			got, err := provider.ListKeyVaultDiagnosticSettings(context.Background(), tc.inputVault)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expected, got)
		})
	}
}

func TestListKeyVaultKeys(t *testing.T) {
	log := testhelper.NewLogger(t)

	response := func(keys ...*armkeyvault.Key) armkeyvault.KeysClientListResponse {
		return armkeyvault.KeysClientListResponse{
			KeyListResult: armkeyvault.KeyListResult{
				Value: keys,
			},
		}
	}

	key := func(id string) *armkeyvault.Key {
		return &armkeyvault.Key{
			ID:       to.Ptr(id),
			Name:     to.Ptr("keyName"),
			Location: to.Ptr("location"),
			Type:     to.Ptr("Microsoft.KeyVault/vaults/keys"),
			Properties: &armkeyvault.KeyProperties{
				KeyURI: to.Ptr("key_uri"),
				Attributes: &armkeyvault.KeyAttributes{
					Enabled: to.Ptr(true),
					Expires: to.Ptr(int64(1705356581)),
				},
			},
		}
	}

	assetKey := func(id string) AzureAsset {
		return AzureAsset{
			Id:             id,
			Name:           "keyName",
			DisplayName:    "",
			Location:       "location",
			ResourceGroup:  "rg1",
			SubscriptionId: "sub1",
			TenantId:       "ten1",
			Type:           "Microsoft.KeyVault/vaults/keys",
			Properties: map[string]any{
				"keyUri":     to.Ptr("key_uri"),
				"attributes": key("key1").Properties.Attributes,
			},
		}
	}

	vaultAsset := AzureAsset{
		Id:             "kv1",
		Name:           "name1",
		ResourceGroup:  "rg1",
		SubscriptionId: "sub1",
		TenantId:       "ten1",
	}

	tests := map[string]struct {
		inputVault               AzureAsset
		mockWrapperResponse      []armkeyvault.KeysClientListResponse
		mockWrapperResponseError error
		expected                 []AzureAsset
		expectError              bool
	}{
		"test with filter out": {
			inputVault: vaultAsset,
			mockWrapperResponse: []armkeyvault.KeysClientListResponse{
				response(nil, key("key1"), nil),
			},
			mockWrapperResponseError: nil,
			expected: []AzureAsset{
				assetKey("key1"),
			},
			expectError: false,
		},
		"test multiple": {
			inputVault: vaultAsset,
			mockWrapperResponse: []armkeyvault.KeysClientListResponse{
				response(key("key1"), key("key2")),
			},
			mockWrapperResponseError: nil,
			expected: []AzureAsset{
				assetKey("key1"),
				assetKey("key2"),
			},
			expectError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockWrapper := &mockAzureKeyVaultWrapper{}
			mockWrapper.Test(t)
			mockWrapper.
				On("AssetKeyVaultKeys", tc.inputVault.SubscriptionId, tc.inputVault.ResourceGroup, tc.inputVault.Name).
				Return(tc.mockWrapperResponse, tc.mockWrapperResponseError).
				Once()
			t.Cleanup(func() { mockWrapper.AssertExpectations(t) })

			provider := keyVaultProvider{
				log: log,
				client: &azureKeyVaultWrapper{
					AssetKeyVaultKeys: mockWrapper.AssetKeyVaultKeys,
				},
			}

			got, err := provider.ListKeyVaultKeys(context.Background(), tc.inputVault)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expected, got)
		})
	}
}

func TestListKeyVaultSecrets(t *testing.T) {
	log := testhelper.NewLogger(t)

	response := func(secrets ...*armkeyvault.Secret) armkeyvault.SecretsClientListResponse {
		return armkeyvault.SecretsClientListResponse{
			SecretListResult: armkeyvault.SecretListResult{
				Value: secrets,
			},
		}
	}

	secret := func(id string) *armkeyvault.Secret {
		return &armkeyvault.Secret{
			ID:       to.Ptr(id),
			Name:     to.Ptr("keyName"),
			Location: to.Ptr("location"),
			Type:     to.Ptr("Microsoft.KeyVault/vaults/secrets"),
			Properties: &armkeyvault.SecretProperties{
				SecretURI: to.Ptr("secret_uri"),
				Attributes: &armkeyvault.SecretAttributes{
					Enabled: to.Ptr(true),
					Expires: to.Ptr(time.Unix(int64(1705356581), 0)),
				},
			},
		}
	}

	assetSecret := func(id string) AzureAsset {
		return AzureAsset{
			Id:             id,
			Name:           "keyName",
			DisplayName:    "",
			Location:       "location",
			ResourceGroup:  "rg1",
			SubscriptionId: "sub1",
			TenantId:       "ten1",
			Type:           "Microsoft.KeyVault/vaults/secrets",
			Properties: map[string]any{
				"secretUri":  to.Ptr("secret_uri"),
				"attributes": secret("").Properties.Attributes,
			},
		}
	}

	vaultAsset := AzureAsset{
		Id:             "kv1",
		Name:           "name1",
		ResourceGroup:  "rg1",
		SubscriptionId: "sub1",
		TenantId:       "ten1",
	}

	tests := map[string]struct {
		inputVault               AzureAsset
		mockWrapperResponse      []armkeyvault.SecretsClientListResponse
		mockWrapperResponseError error
		expected                 []AzureAsset
		expectError              bool
	}{
		"test with filter out": {
			inputVault: vaultAsset,
			mockWrapperResponse: []armkeyvault.SecretsClientListResponse{
				response(nil, secret("secret1"), nil),
			},
			mockWrapperResponseError: nil,
			expected: []AzureAsset{
				assetSecret("secret1"),
			},
			expectError: false,
		},
		"test multiple": {
			inputVault: AzureAsset{
				Id:             "kv1",
				Name:           "name1",
				ResourceGroup:  "rg1",
				SubscriptionId: "sub1",
				TenantId:       "ten1",
			},
			mockWrapperResponse: []armkeyvault.SecretsClientListResponse{
				response(secret("secret1"), secret("secret2")),
			},
			mockWrapperResponseError: nil,
			expected: []AzureAsset{
				assetSecret("secret1"),
				assetSecret("secret2"),
			},
			expectError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockWrapper := &mockAzureKeyVaultWrapper{}
			mockWrapper.Test(t)
			mockWrapper.
				On("AssetKeyVaultSecrets", tc.inputVault.SubscriptionId, tc.inputVault.ResourceGroup, tc.inputVault.Name).
				Return(tc.mockWrapperResponse, tc.mockWrapperResponseError).
				Once()
			t.Cleanup(func() { mockWrapper.AssertExpectations(t) })

			provider := keyVaultProvider{
				log: log,
				client: &azureKeyVaultWrapper{
					AssetKeyVaultSecrets: mockWrapper.AssetKeyVaultSecrets,
				},
			}

			got, err := provider.ListKeyVaultSecrets(context.Background(), tc.inputVault)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expected, got)
		})
	}
}
