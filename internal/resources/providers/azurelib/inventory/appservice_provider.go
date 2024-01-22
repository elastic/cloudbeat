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
	ListWebAppsAuthSettings(ctx context.Context, webApp AzureAsset) ([]AzureAsset, error)
}

func NewAppServiceProvider(log *logp.Logger, credentials azcore.TokenCredential) AppServiceProviderAPI {
	return &appServiceProvider{
		log:    log,
		client: defaultAzureAppServiceWrapper(credentials),
	}
}

type appServiceProvider struct {
	log    *logp.Logger
	client *azureAppServiceWrapper
}

func (p *appServiceProvider) ListWebAppsAuthSettings(ctx context.Context, webApp AzureAsset) ([]AzureAsset, error) {
	p.log.Info("Getting Azure AppService WebApp Auth settings")

	response, err := p.client.AssetWebAppsAuthSettings(ctx, webApp.SubscriptionId, webApp.ResourceGroup, webApp.Name)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving WebApp Auth settings: %w", err)
	}

	webApp.Extension["authSettings"] = response.Properties

	return []AzureAsset{webApp}, nil
}
