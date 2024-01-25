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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/utils/pointers"
)

type azureAppServiceWrapper struct {
	AssetWebAppsAuthSettings func(ctx context.Context, subscriptionID, resourceGroupName, webAppName string) (armappservice.WebAppsClientGetAuthSettingsResponse, error)
}

func defaultAzureAppServiceWrapper(credentials azcore.TokenCredential) *azureAppServiceWrapper {
	return &azureAppServiceWrapper{
		AssetWebAppsAuthSettings: func(ctx context.Context, subscriptionID, resourceGroupName, webAppName string) (armappservice.WebAppsClientGetAuthSettingsResponse, error) {
			client, err := armappservice.NewWebAppsClient(subscriptionID, credentials, nil)
			if err != nil {
				return armappservice.WebAppsClientGetAuthSettingsResponse{}, err
			}
			response, err := client.GetAuthSettings(ctx, resourceGroupName, webAppName, nil)
			if err != nil {
				return armappservice.WebAppsClientGetAuthSettingsResponse{}, err
			}
			return response, nil
		},
	}
}

type AppServiceProviderAPI interface {
	GetWebAppsAuthSettings(ctx context.Context, webApp AzureAsset) ([]AzureAsset, error)
}

func NewAppServiceProvider(log *logp.Logger, credentials azcore.TokenCredential) AppServiceProviderAPI {
	return &azureAppServiceProvider{
		log:    log,
		client: defaultAzureAppServiceWrapper(credentials),
	}
}

type azureAppServiceProvider struct {
	log    *logp.Logger
	client *azureAppServiceWrapper
}

func (p *azureAppServiceProvider) GetWebAppsAuthSettings(ctx context.Context, webApp AzureAsset) ([]AzureAsset, error) {
	p.log.Info("Getting Azure AppService WebApp Auth settings")

	response, err := p.client.AssetWebAppsAuthSettings(ctx, webApp.SubscriptionId, webApp.ResourceGroup, webApp.Name)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving WebApp Auth settings: %w", err)
	}

	if response.Properties == nil {
		return nil, fmt.Errorf("error: got empty WebApp Auth settings for %s", webApp.Name)
	}

	authSettings := AzureAsset{
		Id:             pointers.Deref(response.ID),
		Name:           pointers.Deref(response.Name),
		DisplayName:    webApp.DisplayName,
		Location:       webApp.Location,
		Properties:     unwrapResponseProperties(response.Properties),
		Extension:      map[string]any{},
		ResourceGroup:  webApp.ResourceGroup,
		SubscriptionId: webApp.SubscriptionId,
		TenantId:       webApp.TenantId,
		Type:           pointers.Deref(response.Type),
	}

	return []AzureAsset{authSettings}, nil
}

func unwrapResponseProperties(properties *armappservice.SiteAuthSettingsProperties) map[string]any {
	return map[string]any{
		"AADClaimsAuthorization":                  pointers.Deref(properties.AADClaimsAuthorization),
		"AdditionalLoginParams":                   properties.AdditionalLoginParams,
		"AllowedAudiences":                        properties.AllowedAudiences,
		"AllowedExternalRedirectUrls":             properties.AllowedExternalRedirectUrls,
		"AuthFilePath":                            pointers.Deref(properties.AuthFilePath),
		"ClientID":                                pointers.Deref(properties.ClientID),
		"ClientSecret":                            pointers.Deref(properties.ClientSecret),
		"ClientSecretCertificateThumbprint":       pointers.Deref(properties.ClientSecretCertificateThumbprint),
		"ClientSecretSettingName":                 pointers.Deref(properties.ClientSecretSettingName),
		"ConfigVersion":                           pointers.Deref(properties.ConfigVersion),
		"DefaultProvider":                         pointers.Deref(properties.DefaultProvider),
		"Enabled":                                 pointers.Deref(properties.Enabled),
		"FacebookAppID":                           pointers.Deref(properties.FacebookAppID),
		"FacebookAppSecret":                       pointers.Deref(properties.FacebookAppSecret),
		"FacebookAppSecretSettingName":            pointers.Deref(properties.FacebookAppSecretSettingName),
		"FacebookOAuthScopes":                     properties.FacebookOAuthScopes,
		"GitHubClientID":                          pointers.Deref(properties.GitHubClientID),
		"GitHubClientSecret":                      pointers.Deref(properties.GitHubClientSecret),
		"GitHubClientSecretSettingName":           pointers.Deref(properties.GitHubClientSecretSettingName),
		"GitHubOAuthScopes":                       properties.GitHubOAuthScopes,
		"GoogleClientID":                          pointers.Deref(properties.GoogleClientID),
		"GoogleClientSecret":                      pointers.Deref(properties.GoogleClientSecret),
		"GoogleClientSecretSettingName":           pointers.Deref(properties.GoogleClientSecretSettingName),
		"GoogleOAuthScopes":                       properties.GoogleOAuthScopes,
		"IsAuthFromFile":                          pointers.Deref(properties.IsAuthFromFile),
		"Issuer":                                  pointers.Deref(properties.Issuer),
		"MicrosoftAccountClientID":                pointers.Deref(properties.MicrosoftAccountClientID),
		"MicrosoftAccountClientSecret":            pointers.Deref(properties.MicrosoftAccountClientSecret),
		"MicrosoftAccountClientSecretSettingName": pointers.Deref(properties.MicrosoftAccountClientSecretSettingName),
		"MicrosoftAccountOAuthScopes":             properties.MicrosoftAccountOAuthScopes,
		"RuntimeVersion":                          pointers.Deref(properties.RuntimeVersion),
		"TokenRefreshExtensionHours":              pointers.Deref(properties.TokenRefreshExtensionHours),
		"TokenStoreEnabled":                       pointers.Deref(properties.TokenStoreEnabled),
		"TwitterConsumerKey":                      pointers.Deref(properties.TwitterConsumerKey),
		"TwitterConsumerSecret":                   pointers.Deref(properties.TwitterConsumerSecret),
		"TwitterConsumerSecretSettingName":        pointers.Deref(properties.TwitterConsumerSecretSettingName),
		"UnauthenticatedClientAction":             pointers.Deref(properties.UnauthenticatedClientAction),
		"ValidateIssuer":                          pointers.Deref(properties.ValidateIssuer),
	}
}
