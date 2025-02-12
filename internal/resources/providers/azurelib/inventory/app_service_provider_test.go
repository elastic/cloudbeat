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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type mockAzureAppServiceWrapper struct {
	mock.Mock
}

func (m *mockAzureAppServiceWrapper) AssetAuthSettings(_ context.Context, subscriptionID string, resourceGroupName string, appName string) (armappservice.WebAppsClientGetAuthSettingsResponse, error) {
	r := m.Called(subscriptionID, resourceGroupName, appName)
	return r.Get(0).(armappservice.WebAppsClientGetAuthSettingsResponse), r.Error(1)
}

func (m *mockAzureAppServiceWrapper) AssetSiteConfigs(_ context.Context, subscriptionID string, resourceGroupName string, appName string) (armappservice.WebAppsClientGetConfigurationResponse, error) {
	r := m.Called(subscriptionID, resourceGroupName, appName)
	return r.Get(0).(armappservice.WebAppsClientGetConfigurationResponse), r.Error(1)
}

func TestGetAppServiceAuthSettings(t *testing.T) {
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
				Kind: nil,
				ID:   to.Ptr(webAppId + "/config/authsettings"),
				Name: to.Ptr("authsettings"),
				Type: to.Ptr("Microsoft.Web/sites/config"),
				Properties: &armappservice.SiteAuthSettingsProperties{
					Enabled: authEnabled,
				},
			},
		}
		return response
	}

	buildExpectedAzureAsset := func(authEnabled *bool) AzureAsset {
		response := buildResponse(authEnabled)
		return AzureAsset{
			Id:             pointers.Deref(response.ID),
			Name:           pointers.Deref(response.Name),
			DisplayName:    "",
			Location:       "",
			Properties:     unwrapAuthSettingsResponseProperties(response.Properties),
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
			mockWrapperResponseError: errors.New("error fetching resource"),
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

			got, err := provider.GetAppServiceAuthSettings(context.Background(), tc.inputWebApp)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expected, got)
		})
	}
}

func TestGetAppServiceSiteConfig(t *testing.T) {
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

	buildResponse := func(minTlsVersion armappservice.SupportedTLSVersions, ftpsState armappservice.FtpsState) armappservice.WebAppsClientGetConfigurationResponse {
		response := armappservice.WebAppsClientGetConfigurationResponse{
			SiteConfigResource: armappservice.SiteConfigResource{
				Kind: nil,
				ID:   to.Ptr(webAppId + "/config/authsettings"),
				Name: to.Ptr("authsettings"),
				Type: to.Ptr("Microsoft.Web/sites/config"),
				Properties: &armappservice.SiteConfig{
					MinTLSVersion: to.Ptr(minTlsVersion),
					FtpsState:     to.Ptr(ftpsState),
				},
			},
		}
		return response
	}

	buildExpectedAzureAsset := func(minTlsVersion armappservice.SupportedTLSVersions, ftpsState armappservice.FtpsState) AzureAsset {
		response := buildResponse(minTlsVersion, ftpsState)
		return AzureAsset{
			Id:             pointers.Deref(response.ID),
			Name:           pointers.Deref(response.Name),
			DisplayName:    "",
			Location:       "",
			Properties:     unwrapSiteConfigResponseProperties(response.Properties),
			Extension:      map[string]any{},
			ResourceGroup:  "galaxy",
			SubscriptionId: "space",
			TenantId:       "???",
			Type:           pointers.Deref(response.Type),
		}
	}

	tests := map[string]struct {
		inputWebApp              AzureAsset
		mockWrapperResponse      armappservice.WebAppsClientGetConfigurationResponse
		mockWrapperResponseError error
		expected                 []AzureAsset
		expectError              bool
	}{
		"expected error: could not fetch SiteConfig from Azure": {
			inputWebApp:              webAppAsset,
			mockWrapperResponse:      armappservice.WebAppsClientGetConfigurationResponse{},
			mockWrapperResponseError: errors.New("error fetching resource"),
			expected:                 []AzureAsset{},
			expectError:              true,
		},
		"expected error: SiteConfigResource.Properties are <nil>": {
			inputWebApp: webAppAsset,
			mockWrapperResponse: armappservice.WebAppsClientGetConfigurationResponse{
				SiteConfigResource: armappservice.SiteConfigResource{
					Properties: nil,
				},
			},
			mockWrapperResponseError: nil,
			expected:                 []AzureAsset{},
			expectError:              true,
		},
		"fetch AuthSettings with MinTLSVersion == 1.0": {
			inputWebApp:              webAppAsset,
			mockWrapperResponse:      buildResponse(armappservice.SupportedTLSVersionsOne0, armappservice.FtpsStateFtpsOnly),
			mockWrapperResponseError: nil,
			expected:                 []AzureAsset{buildExpectedAzureAsset(armappservice.SupportedTLSVersionsOne0, armappservice.FtpsStateFtpsOnly)},
			expectError:              false,
		},
		"fetch AuthSettings with MinTLSVersion == 1.1": {
			inputWebApp:              webAppAsset,
			mockWrapperResponse:      buildResponse(armappservice.SupportedTLSVersionsOne1, armappservice.FtpsStateFtpsOnly),
			mockWrapperResponseError: nil,
			expected:                 []AzureAsset{buildExpectedAzureAsset(armappservice.SupportedTLSVersionsOne1, armappservice.FtpsStateFtpsOnly)},
			expectError:              false,
		},
		"fetch AuthSettings with MinTLSVersion == 1.2": {
			inputWebApp:              webAppAsset,
			mockWrapperResponse:      buildResponse(armappservice.SupportedTLSVersionsOne2, armappservice.FtpsStateFtpsOnly),
			mockWrapperResponseError: nil,
			expected:                 []AzureAsset{buildExpectedAzureAsset(armappservice.SupportedTLSVersionsOne2, armappservice.FtpsStateFtpsOnly)},
			expectError:              false,
		},
		"fetch AuthSettings with FtpsState == FtpsOnly": {
			inputWebApp:              webAppAsset,
			mockWrapperResponse:      buildResponse(armappservice.SupportedTLSVersionsOne2, armappservice.FtpsStateFtpsOnly),
			mockWrapperResponseError: nil,
			expected:                 []AzureAsset{buildExpectedAzureAsset(armappservice.SupportedTLSVersionsOne2, armappservice.FtpsStateFtpsOnly)},
			expectError:              false,
		},
		"fetch AuthSettings with FtpsState == AllAllowed": {
			inputWebApp:              webAppAsset,
			mockWrapperResponse:      buildResponse(armappservice.SupportedTLSVersionsOne2, armappservice.FtpsStateAllAllowed),
			mockWrapperResponseError: nil,
			expected:                 []AzureAsset{buildExpectedAzureAsset(armappservice.SupportedTLSVersionsOne2, armappservice.FtpsStateAllAllowed)},
			expectError:              false,
		},
		"fetch AuthSettings with FtpsState == Disabled": {
			inputWebApp:              webAppAsset,
			mockWrapperResponse:      buildResponse(armappservice.SupportedTLSVersionsOne2, armappservice.FtpsStateDisabled),
			mockWrapperResponseError: nil,
			expected:                 []AzureAsset{buildExpectedAzureAsset(armappservice.SupportedTLSVersionsOne2, armappservice.FtpsStateDisabled)},
			expectError:              false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockWrapper := &mockAzureAppServiceWrapper{}
			mockWrapper.Test(t)
			mockWrapper.
				On("AssetSiteConfigs", tc.inputWebApp.SubscriptionId, tc.inputWebApp.ResourceGroup, tc.inputWebApp.Name).
				Return(tc.mockWrapperResponse, tc.mockWrapperResponseError).
				Once()
			t.Cleanup(func() { mockWrapper.AssertExpectations(t) })

			provider := azureAppServiceProvider{
				log: log,
				client: &azureAppServiceWrapper{
					AssetWebAppsSiteConfig: mockWrapper.AssetSiteConfigs,
				},
			}

			got, err := provider.GetAppServiceSiteConfig(context.Background(), tc.inputWebApp)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expected, got)
		})
	}
}
