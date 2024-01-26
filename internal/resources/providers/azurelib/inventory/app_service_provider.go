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
	AssetWebAppsSiteConfig   func(ctx context.Context, subscriptionID, resourceGroupName, webAppName string) (*armappservice.SiteConfig, error)
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
		AssetWebAppsSiteConfig: func(ctx context.Context, subscriptionID, resourceGroupName, webAppName string) (*armappservice.SiteConfig, error) {
			client, err := armappservice.NewWebAppsClient(subscriptionID, credentials, nil)
			if err != nil {
				return nil, err
			}
			response, err := client.Get(ctx, resourceGroupName, webAppName, nil)
			if err != nil {
				return nil, err
			}
			if response.Properties == nil {
				return nil, nil
			}
			return response.Properties.SiteConfig, nil
		},
	}
}

type AppServiceProviderAPI interface {
	GetWebAppsAuthSettings(ctx context.Context, webApp AzureAsset) ([]AzureAsset, error)
	GetWebAppsSiteConfig(ctx context.Context, webApp AzureAsset) ([]AzureAsset, error)
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
		Properties:     unwrapResponseProperties(response.Properties),
		Extension:      map[string]any{},
		ResourceGroup:  webApp.ResourceGroup,
		SubscriptionId: webApp.SubscriptionId,
		TenantId:       webApp.TenantId,
		Type:           pointers.Deref(response.Type),
	}

	return []AzureAsset{authSettings}, nil
}

func (p *azureAppServiceProvider) GetWebAppsSiteConfig(ctx context.Context, webApp AzureAsset) ([]AzureAsset, error) {
	p.log.Info("Getting Azure AppService SiteConfig")

	siteConfig, err := p.client.AssetWebAppsSiteConfig(ctx, webApp.SubscriptionId, webApp.ResourceGroup, webApp.Name)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving AppService SiteConfig: %w", err)
	}

	if siteConfig == nil {
		return nil, fmt.Errorf("error: got empty AppService SiteConfig for %s", webApp.Name)
	}

	authSettings := AzureAsset{
		Id:             webApp.Id,
		Name:           webApp.Name,
		DisplayName:    webApp.DisplayName,
		Location:       webApp.Location,
		Properties:     unwrapSiteConfig(siteConfig),
		Extension:      map[string]any{},
		ResourceGroup:  webApp.ResourceGroup,
		SubscriptionId: webApp.SubscriptionId,
		TenantId:       webApp.TenantId,
		Type:           webApp.Type,
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

func unwrapSiteConfig(siteConfig *armappservice.SiteConfig) map[string]any {
	return map[string]any{
		"APIDefinition":                          pointers.Deref(siteConfig.APIDefinition),
		"APIManagementConfig":                    pointers.Deref(siteConfig.APIManagementConfig),
		"AcrUseManagedIdentityCreds":             pointers.Deref(siteConfig.AcrUseManagedIdentityCreds),
		"AcrUserManagedIdentityID":               pointers.Deref(siteConfig.AcrUserManagedIdentityID),
		"AlwaysOn":                               pointers.Deref(siteConfig.AlwaysOn),
		"AppCommandLine":                         pointers.Deref(siteConfig.AppCommandLine),
		"AppSettings":                            siteConfig.AppSettings,
		"AutoHealEnabled":                        pointers.Deref(siteConfig.AutoHealEnabled),
		"AutoHealRules":                          pointers.Deref(siteConfig.AutoHealRules),
		"AutoSwapSlotName":                       pointers.Deref(siteConfig.AutoSwapSlotName),
		"AzureStorageAccounts":                   siteConfig.AzureStorageAccounts,
		"ConnectionStrings":                      siteConfig.ConnectionStrings,
		"Cors":                                   pointers.Deref(siteConfig.Cors),
		"DefaultDocuments":                       siteConfig.DefaultDocuments,
		"DetailedErrorLoggingEnabled":            pointers.Deref(siteConfig.DetailedErrorLoggingEnabled),
		"DocumentRoot":                           pointers.Deref(siteConfig.DocumentRoot),
		"ElasticWebAppScaleLimit":                pointers.Deref(siteConfig.ElasticWebAppScaleLimit),
		"Experiments":                            pointers.Deref(siteConfig.Experiments),
		"FtpsState":                              pointers.Deref(siteConfig.FtpsState),
		"FunctionAppScaleLimit":                  pointers.Deref(siteConfig.FunctionAppScaleLimit),
		"FunctionsRuntimeScaleMonitoringEnabled": pointers.Deref(siteConfig.FunctionsRuntimeScaleMonitoringEnabled),
		"HTTPLoggingEnabled":                     pointers.Deref(siteConfig.HTTPLoggingEnabled),
		"HandlerMappings":                        siteConfig.HandlerMappings,
		"HealthCheckPath":                        pointers.Deref(siteConfig.HealthCheckPath),
		"Http20Enabled":                          pointers.Deref(siteConfig.Http20Enabled),
		"IPSecurityRestrictions":                 siteConfig.IPSecurityRestrictions,
		"IPSecurityRestrictionsDefaultAction":    pointers.Deref(siteConfig.IPSecurityRestrictionsDefaultAction),
		"JavaContainer":                          pointers.Deref(siteConfig.JavaContainer),
		"JavaContainerVersion":                   pointers.Deref(siteConfig.JavaContainerVersion),
		"JavaVersion":                            pointers.Deref(siteConfig.JavaVersion),
		"KeyVaultReferenceIdentity":              pointers.Deref(siteConfig.KeyVaultReferenceIdentity),
		"Limits":                                 pointers.Deref(siteConfig.Limits),
		"LinuxFxVersion":                         pointers.Deref(siteConfig.LinuxFxVersion),
		"LoadBalancing":                          pointers.Deref(siteConfig.LoadBalancing),
		"LocalMySQLEnabled":                      pointers.Deref(siteConfig.LocalMySQLEnabled),
		"LogsDirectorySizeLimit":                 pointers.Deref(siteConfig.LogsDirectorySizeLimit),
		"ManagedPipelineMode":                    pointers.Deref(siteConfig.ManagedPipelineMode),
		"ManagedServiceIdentityID":               pointers.Deref(siteConfig.ManagedServiceIdentityID),
		"Metadata":                               siteConfig.Metadata,
		"MinTLSCipherSuite":                      pointers.Deref(siteConfig.MinTLSCipherSuite),
		"MinTLSVersion":                          pointers.Deref(siteConfig.MinTLSVersion),
		"MinimumElasticInstanceCount":            pointers.Deref(siteConfig.MinimumElasticInstanceCount),
		"NetFrameworkVersion":                    pointers.Deref(siteConfig.NetFrameworkVersion),
		"NodeVersion":                            pointers.Deref(siteConfig.NodeVersion),
		"NumberOfWorkers":                        pointers.Deref(siteConfig.NumberOfWorkers),
		"PhpVersion":                             pointers.Deref(siteConfig.PhpVersion),
		"PowerShellVersion":                      pointers.Deref(siteConfig.PowerShellVersion),
		"PreWarmedInstanceCount":                 pointers.Deref(siteConfig.PreWarmedInstanceCount),
		"PublicNetworkAccess":                    pointers.Deref(siteConfig.PublicNetworkAccess),
		"PublishingUsername":                     pointers.Deref(siteConfig.PublishingUsername),
		"Push":                                   pointers.Deref(siteConfig.Push),
		"PythonVersion":                          pointers.Deref(siteConfig.PythonVersion),
		"RemoteDebuggingEnabled":                 pointers.Deref(siteConfig.RemoteDebuggingEnabled),
		"RemoteDebuggingVersion":                 pointers.Deref(siteConfig.RemoteDebuggingVersion),
		"RequestTracingEnabled":                  pointers.Deref(siteConfig.RequestTracingEnabled),
		"RequestTracingExpirationTime":           pointers.Deref(siteConfig.RequestTracingExpirationTime),
		"ScmIPSecurityRestrictions":              siteConfig.ScmIPSecurityRestrictions,
		"ScmIPSecurityRestrictionsDefaultAction": pointers.Deref(siteConfig.ScmIPSecurityRestrictionsDefaultAction),
		"ScmIPSecurityRestrictionsUseMain":       pointers.Deref(siteConfig.ScmIPSecurityRestrictionsUseMain),
		"ScmMinTLSVersion":                       pointers.Deref(siteConfig.ScmMinTLSVersion),
		"ScmType":                                pointers.Deref(siteConfig.ScmType),
		"TracingOptions":                         pointers.Deref(siteConfig.TracingOptions),
		"Use32BitWorkerProcess":                  pointers.Deref(siteConfig.Use32BitWorkerProcess),
		"VirtualApplications":                    siteConfig.VirtualApplications,
		"VnetName":                               pointers.Deref(siteConfig.VnetName),
		"VnetPrivatePortsCount":                  pointers.Deref(siteConfig.VnetPrivatePortsCount),
		"VnetRouteAllEnabled":                    pointers.Deref(siteConfig.VnetRouteAllEnabled),
		"WebSocketsEnabled":                      pointers.Deref(siteConfig.WebSocketsEnabled),
		"WebsiteTimeZone":                        pointers.Deref(siteConfig.WebsiteTimeZone),
		"WindowsFxVersion":                       pointers.Deref(siteConfig.WindowsFxVersion),
		"XManagedServiceIdentityID":              pointers.Deref(siteConfig.XManagedServiceIdentityID),
		"MachineKey":                             pointers.Deref(siteConfig.MachineKey),
	}
}
