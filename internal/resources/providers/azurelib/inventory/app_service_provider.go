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

	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type azureAppServiceWrapper struct {
	AssetWebAppsAuthSettings func(ctx context.Context, subscriptionID, resourceGroupName, appName string) (armappservice.WebAppsClientGetAuthSettingsResponse, error)
	AssetWebAppsSiteConfig   func(ctx context.Context, subscriptionID, resourceGroupName, appName string) (armappservice.WebAppsClientGetConfigurationResponse, error)
}

func defaultAzureAppServiceWrapper(credentials azcore.TokenCredential) *azureAppServiceWrapper {
	return &azureAppServiceWrapper{
		AssetWebAppsAuthSettings: func(ctx context.Context, subscriptionID, resourceGroupName, appName string) (armappservice.WebAppsClientGetAuthSettingsResponse, error) {
			client, err := armappservice.NewWebAppsClient(subscriptionID, credentials, nil)
			if err != nil {
				return armappservice.WebAppsClientGetAuthSettingsResponse{}, err
			}
			return client.GetAuthSettings(ctx, resourceGroupName, appName, nil)
		},
		AssetWebAppsSiteConfig: func(ctx context.Context, subscriptionID, resourceGroupName, appName string) (armappservice.WebAppsClientGetConfigurationResponse, error) {
			client, err := armappservice.NewWebAppsClient(subscriptionID, credentials, nil)
			if err != nil {
				return armappservice.WebAppsClientGetConfigurationResponse{}, err
			}
			return client.GetConfiguration(ctx, resourceGroupName, appName, nil)
		},
	}
}

type AppServiceProviderAPI interface {
	GetAppServiceAuthSettings(ctx context.Context, webApp AzureAsset) ([]AzureAsset, error)
	GetAppServiceSiteConfig(ctx context.Context, webApp AzureAsset) ([]AzureAsset, error)
}

func NewAppServiceProvider(log *clog.Logger, credentials azcore.TokenCredential) AppServiceProviderAPI {
	return &azureAppServiceProvider{
		log:    log,
		client: defaultAzureAppServiceWrapper(credentials),
	}
}

type azureAppServiceProvider struct {
	log    *clog.Logger
	client *azureAppServiceWrapper
}

func (p *azureAppServiceProvider) GetAppServiceAuthSettings(ctx context.Context, webApp AzureAsset) ([]AzureAsset, error) {
	p.log.Info("Getting Azure AppService AuthSettings")

	response, err := p.client.AssetWebAppsAuthSettings(ctx, webApp.SubscriptionId, webApp.ResourceGroup, webApp.Name)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving AppService AuthSettings: %w", err)
	}

	if response.Properties == nil {
		return nil, fmt.Errorf("error: got empty AppService AuthSettings for %s", webApp.Name)
	}

	authSettings := AzureAsset{
		Id:             pointers.Deref(response.ID),
		Name:           pointers.Deref(response.Name),
		DisplayName:    webApp.DisplayName,
		Location:       webApp.Location,
		Properties:     unwrapAuthSettingsResponseProperties(response.Properties),
		Extension:      map[string]any{},
		ResourceGroup:  webApp.ResourceGroup,
		SubscriptionId: webApp.SubscriptionId,
		TenantId:       webApp.TenantId,
		Type:           pointers.Deref(response.Type),
	}

	return []AzureAsset{authSettings}, nil
}

func (p *azureAppServiceProvider) GetAppServiceSiteConfig(ctx context.Context, webApp AzureAsset) ([]AzureAsset, error) {
	p.log.Info("Getting Azure AppService SiteConfig")

	response, err := p.client.AssetWebAppsSiteConfig(ctx, webApp.SubscriptionId, webApp.ResourceGroup, webApp.Name)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving AppService SiteConfig: %w", err)
	}

	if response.Properties == nil {
		return nil, fmt.Errorf("error: got empty AppService SiteConfig for %s", webApp.Name)
	}

	authSettings := AzureAsset{
		Id:             pointers.Deref(response.ID),
		Name:           pointers.Deref(response.Name),
		DisplayName:    webApp.DisplayName,
		Location:       webApp.Location,
		Properties:     unwrapSiteConfigResponseProperties(response.Properties),
		Extension:      map[string]any{},
		ResourceGroup:  webApp.ResourceGroup,
		SubscriptionId: webApp.SubscriptionId,
		TenantId:       webApp.TenantId,
		Type:           pointers.Deref(response.Type),
	}

	return []AzureAsset{authSettings}, nil
}

func unwrapAuthSettingsResponseProperties(properties *armappservice.SiteAuthSettingsProperties) map[string]any {
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

func unwrapSiteConfigResponseProperties(properties *armappservice.SiteConfig) map[string]any {
	return map[string]any{
		"APIDefinition":                          pointers.Deref(properties.APIDefinition),
		"APIManagementConfig":                    pointers.Deref(properties.APIManagementConfig),
		"AcrUseManagedIdentityCreds":             pointers.Deref(properties.AcrUseManagedIdentityCreds),
		"AcrUserManagedIdentityID":               pointers.Deref(properties.AcrUserManagedIdentityID),
		"AlwaysOn":                               pointers.Deref(properties.AlwaysOn),
		"AppCommandLine":                         pointers.Deref(properties.AppCommandLine),
		"AppSettings":                            properties.AppSettings,
		"AutoHealEnabled":                        pointers.Deref(properties.AutoHealEnabled),
		"AutoHealRules":                          pointers.Deref(properties.AutoHealRules),
		"AutoSwapSlotName":                       pointers.Deref(properties.AutoSwapSlotName),
		"AzureStorageAccounts":                   properties.AzureStorageAccounts,
		"ConnectionStrings":                      properties.ConnectionStrings,
		"Cors":                                   pointers.Deref(properties.Cors),
		"DefaultDocuments":                       properties.DefaultDocuments,
		"DetailedErrorLoggingEnabled":            pointers.Deref(properties.DetailedErrorLoggingEnabled),
		"DocumentRoot":                           pointers.Deref(properties.DocumentRoot),
		"ElasticWebAppScaleLimit":                pointers.Deref(properties.ElasticWebAppScaleLimit),
		"Experiments":                            pointers.Deref(properties.Experiments),
		"FtpsState":                              pointers.Deref(properties.FtpsState),
		"FunctionAppScaleLimit":                  pointers.Deref(properties.FunctionAppScaleLimit),
		"FunctionsRuntimeScaleMonitoringEnabled": pointers.Deref(properties.FunctionsRuntimeScaleMonitoringEnabled),
		"HTTPLoggingEnabled":                     pointers.Deref(properties.HTTPLoggingEnabled),
		"HandlerMappings":                        properties.HandlerMappings,
		"HealthCheckPath":                        pointers.Deref(properties.HealthCheckPath),
		"Http20Enabled":                          pointers.Deref(properties.Http20Enabled),
		"IPSecurityRestrictions":                 properties.IPSecurityRestrictions,
		"IPSecurityRestrictionsDefaultAction":    pointers.Deref(properties.IPSecurityRestrictionsDefaultAction),
		"JavaContainer":                          pointers.Deref(properties.JavaContainer),
		"JavaContainerVersion":                   pointers.Deref(properties.JavaContainerVersion),
		"JavaVersion":                            pointers.Deref(properties.JavaVersion),
		"KeyVaultReferenceIdentity":              pointers.Deref(properties.KeyVaultReferenceIdentity),
		"Limits":                                 pointers.Deref(properties.Limits),
		"LinuxFxVersion":                         pointers.Deref(properties.LinuxFxVersion),
		"LoadBalancing":                          pointers.Deref(properties.LoadBalancing),
		"LocalMySQLEnabled":                      pointers.Deref(properties.LocalMySQLEnabled),
		"LogsDirectorySizeLimit":                 pointers.Deref(properties.LogsDirectorySizeLimit),
		"ManagedPipelineMode":                    pointers.Deref(properties.ManagedPipelineMode),
		"ManagedServiceIdentityID":               pointers.Deref(properties.ManagedServiceIdentityID),
		"Metadata":                               properties.Metadata,
		"MinTLSCipherSuite":                      pointers.Deref(properties.MinTLSCipherSuite),
		"MinTLSVersion":                          pointers.Deref(properties.MinTLSVersion),
		"MinimumElasticInstanceCount":            pointers.Deref(properties.MinimumElasticInstanceCount),
		"NetFrameworkVersion":                    pointers.Deref(properties.NetFrameworkVersion),
		"NodeVersion":                            pointers.Deref(properties.NodeVersion),
		"NumberOfWorkers":                        pointers.Deref(properties.NumberOfWorkers),
		"PhpVersion":                             pointers.Deref(properties.PhpVersion),
		"PowerShellVersion":                      pointers.Deref(properties.PowerShellVersion),
		"PreWarmedInstanceCount":                 pointers.Deref(properties.PreWarmedInstanceCount),
		"PublicNetworkAccess":                    pointers.Deref(properties.PublicNetworkAccess),
		"PublishingUsername":                     pointers.Deref(properties.PublishingUsername),
		"Push":                                   pointers.Deref(properties.Push),
		"PythonVersion":                          pointers.Deref(properties.PythonVersion),
		"RemoteDebuggingEnabled":                 pointers.Deref(properties.RemoteDebuggingEnabled),
		"RemoteDebuggingVersion":                 pointers.Deref(properties.RemoteDebuggingVersion),
		"RequestTracingEnabled":                  pointers.Deref(properties.RequestTracingEnabled),
		"RequestTracingExpirationTime":           pointers.Deref(properties.RequestTracingExpirationTime),
		"ScmIPSecurityRestrictions":              properties.ScmIPSecurityRestrictions,
		"ScmIPSecurityRestrictionsDefaultAction": pointers.Deref(properties.ScmIPSecurityRestrictionsDefaultAction),
		"ScmIPSecurityRestrictionsUseMain":       pointers.Deref(properties.ScmIPSecurityRestrictionsUseMain),
		"ScmMinTLSVersion":                       pointers.Deref(properties.ScmMinTLSVersion),
		"ScmType":                                pointers.Deref(properties.ScmType),
		"TracingOptions":                         pointers.Deref(properties.TracingOptions),
		"Use32BitWorkerProcess":                  pointers.Deref(properties.Use32BitWorkerProcess),
		"VirtualApplications":                    properties.VirtualApplications,
		"VnetName":                               pointers.Deref(properties.VnetName),
		"VnetPrivatePortsCount":                  pointers.Deref(properties.VnetPrivatePortsCount),
		"VnetRouteAllEnabled":                    pointers.Deref(properties.VnetRouteAllEnabled),
		"WebSocketsEnabled":                      pointers.Deref(properties.WebSocketsEnabled),
		"WebsiteTimeZone":                        pointers.Deref(properties.WebsiteTimeZone),
		"WindowsFxVersion":                       pointers.Deref(properties.WindowsFxVersion),
		"XManagedServiceIdentityID":              pointers.Deref(properties.XManagedServiceIdentityID),
		"MachineKey":                             pointers.Deref(properties.MachineKey),
	}
}
