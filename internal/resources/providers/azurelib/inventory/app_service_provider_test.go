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
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/resources/utils/testhelper"
)

type mockAzureAppServiceWrapper struct {
	mock.Mock
}

func (m *mockAzureAppServiceWrapper) AssetAuthSettings(_ context.Context, subscriptionID string, resourceGroupName string, webAppName string) (armappservice.WebAppsClientGetAuthSettingsResponse, error) {
	r := m.Called(subscriptionID, resourceGroupName, webAppName)
	return r.Get(0).(armappservice.WebAppsClientGetAuthSettingsResponse), r.Error(1)
}

func TestGetAuthSettings(t *testing.T) {
	log := testhelper.NewLogger(t)

	//	response := func(keys ...*armkeyvault.Key) armkeyvault.KeysClientListResponse {
	//		return armkeyvault.KeysClientListResponse{
	//			KeyListResult: armkeyvault.KeyListResult{
	//				Value: keys,
	//			},
	//		}
	//	}
	//
	//	key := func(id string) *armkeyvault.Key {
	//		return &armkeyvault.Key{
	//			ID:       to.Ptr(id),
	//			Name:     to.Ptr("keyName"),
	//			Location: to.Ptr("location"),
	//			Type:     to.Ptr("Microsoft.KeyVault/vaults/keys"),
	//			Properties: &armkeyvault.KeyProperties{
	//				KeyURI: to.Ptr("key_uri"),
	//				Attributes: &armkeyvault.KeyAttributes{
	//					Enabled: to.Ptr(true),
	//					Expires: to.Ptr(int64(1705356581)),
	//				},
	//			},
	//		}
	//	}
	//
	//	assetKey := func(id string) AzureAsset {
	//		return AzureAsset{
	//			Id:             id,
	//			Name:           "keyName",
	//			DisplayName:    "",
	//			Location:       "location",
	//			ResourceGroup:  "rg1",
	//			SubscriptionId: "sub1",
	//			TenantId:       "ten1",
	//			Type:           "Microsoft.KeyVault/vaults/keys",
	//			Properties: map[string]any{
	//				"keyUri":     to.Ptr("key_uri"),
	//				"attributes": key("key1").Properties.Attributes,
	//			},
	//		}
	//	}

	vaultAsset := AzureAsset{
		Id:             "kv1",
		Name:           "name1",
		ResourceGroup:  "rg1",
		SubscriptionId: "sub1",
		TenantId:       "ten1",
	}

	webAppAsset := AzureAsset{
		Id:             "/subscriptions/space/resourceGroups/galaxy/providers/Microsoft.Web/sites/cats-in-space",
		Name:           "cats-in-space",
		ResourceGroup:  "galaxy",
		SubscriptionId: "space",
		TenantId:       "???",
		Type:           "Microsoft.Web/sites",
	}

	authSettings := func(authEnabled *bool) AzureAsset {
		return AzureAsset{}
	}

	response := func(authEnabled *bool) armappservice.WebAppsClientGetAuthSettingsResponse {
		return armappservice.WebAppsClientGetAuthSettingsResponse{}
	}

	tests := map[string]struct {
		inputWebApp              AzureAsset
		mockWrapperResponse      armappservice.WebAppsClientGetAuthSettingsResponse
		mockWrapperResponseError error
		expected                 []AzureAsset
		expectError              bool
	}{
		"test WebApp with no AuthSettings (should not happen)": {
			inputWebApp:              webAppAsset,
			mockWrapperResponse:      armappservice.WebAppsClientGetAuthSettingsResponse{},
			mockWrapperResponseError: nil,
			expected:                 []AzureAsset{},
			expectError:              true,
		},

		/*
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
		*/
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			mockWrapper := &mockAzureAppServiceWrapper{}
			mockWrapper.Test(t)
			mockWrapper.
				On("AssetAuthSettings", tc.inputWebApp.SubscriptionId, tc.inputWebApp.ResourceGroup, tc.inputWebApp.Name).
				Return(tc.mockWrapperResponse, tc.mockWrapperResponseError).
				Once()
			t.Cleanup(func() { mockWrapper.AssertExpectations(t) })

			provider := azureAppServiceProvider{
				log: log,
				client: &azureAppServiceWrapper{
					AssetWebAppsAuthSettings: mockWrapper.AssetAuthSettings,
				},
			}

			got, err := provider.GetWebAppsAuthSettings(context.Background(), tc.inputWebApp)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expected, got)
		})
	}
}
