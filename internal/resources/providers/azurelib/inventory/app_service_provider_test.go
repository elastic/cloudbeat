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
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/resources/utils/pointers"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
)

type mockAzureAppServiceWrapper struct {
	mock.Mock
}

func (m *mockAzureAppServiceWrapper) AssetAuthSettings(_ context.Context, subscriptionID string, resourceGroupName string, webAppName string) (armappservice.WebAppsClientGetAuthSettingsResponse, error) {
	r := m.Called(subscriptionID, resourceGroupName, webAppName)
	return r.Get(0).(armappservice.WebAppsClientGetAuthSettingsResponse), r.Error(1)
}

func (m *mockAzureAppServiceWrapper) AssetSiteConfigs(_ context.Context, subscriptionID string, resourceGroupName string, webAppName string) (armappservice.SiteConfig, error) {
	r := m.Called(subscriptionID, resourceGroupName, webAppName)
	return r.Get(0).(armappservice.SiteConfig), r.Error(1)
}

func TestGetAuthSettings(t *testing.T) {
	log := testhelper.NewLogger(t)

	webAppId := "/subscriptions/space/resourceGroups/galaxy/providers/Microsoft.Web/sites/cats-in-space"

	webAppAsset := AzureAsset{
		Id:             webAppId,
		Name:           "cats-in-space",
		ResourceGroup:  "galaxy",
		SubscriptionId: "space",
		TenantId:       "???",
		Type:           "Microsoft.Web/sites",
	}

	buildResponse := func(authEnabled *bool) armappservice.WebAppsClientGetAuthSettingsResponse {
		response := armappservice.WebAppsClientGetAuthSettingsResponse{
			SiteAuthSettings: armappservice.SiteAuthSettings{
				Kind:       nil,
				ID:         to.Ptr(webAppId + "/config/authsettings"),
				Name:       to.Ptr("authsettings"),
				Type:       to.Ptr("Microsoft.Web/sites/config"),
				Properties: &armappservice.SiteAuthSettingsProperties{},
			},
		}
		response.Properties.Enabled = authEnabled
		return response
	}

	buildExpectedAzureAsset := func(authEnabled *bool) AzureAsset {
		response := buildResponse(authEnabled)
		return AzureAsset{
			Id:             pointers.Deref(response.ID),
			Name:           pointers.Deref(response.Name),
			DisplayName:    "",
			Location:       "",
			Properties:     unwrapResponseProperties(response.Properties),
			Extension:      map[string]any{},
			ResourceGroup:  "galaxy",
			SubscriptionId: "space",
			TenantId:       "???",
			Type:           pointers.Deref(response.Type),
		}
	}

	tests := map[string]struct {
		inputWebApp              AzureAsset
		mockWrapperResponse      armappservice.WebAppsClientGetAuthSettingsResponse
		mockWrapperResponseError error
		expected                 []AzureAsset
		expectError              bool
	}{
		"expected error: could not fetch AuthSettings from Azure": {
			inputWebApp:              webAppAsset,
			mockWrapperResponse:      armappservice.WebAppsClientGetAuthSettingsResponse{},
			mockWrapperResponseError: fmt.Errorf("error fetching resource"),
			expected:                 []AzureAsset{},
			expectError:              true,
		},
		"expected error: AuthSettings.Properties are <nil>": {
			inputWebApp: webAppAsset,
			mockWrapperResponse: armappservice.WebAppsClientGetAuthSettingsResponse{
				SiteAuthSettings: armappservice.SiteAuthSettings{
					Properties: nil,
				},
			},
			mockWrapperResponseError: nil,
			expected:                 []AzureAsset{},
			expectError:              true,
		},
		"fetch AuthSettings with authorization set to null": {
			inputWebApp:              webAppAsset,
			mockWrapperResponse:      buildResponse(nil),
			mockWrapperResponseError: nil,
			expected:                 []AzureAsset{buildExpectedAzureAsset(nil)},
			expectError:              false,
		},
		"fetch AuthSettings with authorization set to false": {
			inputWebApp:              webAppAsset,
			mockWrapperResponse:      buildResponse(to.Ptr(false)),
			mockWrapperResponseError: nil,
			expected:                 []AzureAsset{buildExpectedAzureAsset(to.Ptr(false))},
			expectError:              false,
		},
		"fetch AuthSettings with authorization set to true": {
			inputWebApp:              webAppAsset,
			mockWrapperResponse:      buildResponse(to.Ptr(true)),
			mockWrapperResponseError: nil,
			expected:                 []AzureAsset{buildExpectedAzureAsset(to.Ptr(true))},
			expectError:              false,
		},
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
