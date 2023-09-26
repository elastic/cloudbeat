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
	"bytes"
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/resources/providers/azurelib/auth"
	"github.com/elastic/cloudbeat/resources/utils/strings"
)

type Provider struct {
	log           *logp.Logger
	client        *AzureClientWrapper
	subscriptions []*string
	ctx           context.Context
	Config        auth.AzureFactoryConfig
}

type ProviderInitializer struct{}

type AzureClientWrapper struct {
	AssetQuery func(ctx context.Context, query armresourcegraph.QueryRequest, options *armresourcegraph.ClientResourcesOptions) (armresourcegraph.ClientResourcesResponse, error)
}

type AzureAsset struct {
	Id             string         `json:"id,omitempty"`
	Name           string         `json:"name,omitempty"`
	Location       string         `json:"location,omitempty"`
	Properties     map[string]any `json:"properties,omitempty"`
	ResourceGroup  string         `json:"resource_group,omitempty"`
	SubscriptionId string         `json:"subscription_id,omitempty"`
	TenantId       string         `json:"tenant_id,omitempty"`
	Type           string         `json:"type,omitempty"`
}

type ServiceAPI interface {
	// ListAllAssetTypesByName List all content types of the given assets types
	ListAllAssetTypesByName(assets []string) ([]AzureAsset, error)
}

type ProviderInitializerAPI interface {
	// Init initializes the Azure asset client
	Init(ctx context.Context, log *logp.Logger, azureConfig auth.AzureFactoryConfig) (ServiceAPI, error)
}

func (p *ProviderInitializer) Init(ctx context.Context, log *logp.Logger, azureConfig auth.AzureFactoryConfig) (ServiceAPI, error) {
	log = log.Named("azure")

	clientFactory, err := armresourcegraph.NewClientFactory(azureConfig.Credentials, nil)
	if err != nil {
		return nil, err
	}

	client := clientFactory.NewClient()

	// We wrap the client, so we can mock it in tests
	wrapper := &AzureClientWrapper{
		AssetQuery: func(ctx context.Context, query armresourcegraph.QueryRequest, options *armresourcegraph.ClientResourcesOptions) (armresourcegraph.ClientResourcesResponse, error) {
			return client.Resources(ctx, query, options)
		},
	}

	subscriptions, err := p.getSubscriptionIds(ctx, azureConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription ids: %w", err)
	}
	log.Info(
		lo.Reduce(subscriptions, func(agg string, item *string, _ int) string {
			return fmt.Sprintf("%s %s", agg, strings.Dereference(item))
		}, "subscriptions:"),
	)

	return &Provider{
		log:           log,
		client:        wrapper,
		subscriptions: subscriptions,
		ctx:           ctx,
		Config:        azureConfig,
	}, nil
}

func (p *ProviderInitializer) getSubscriptionIds(ctx context.Context, azureConfig auth.AzureFactoryConfig) ([]*string, error) {
	// TODO: mockable

	var result []*string

	clientFactory, err := armsubscriptions.NewClientFactory(azureConfig.Credentials, nil)
	if err != nil {
		return nil, err
	}
	pager := clientFactory.NewClient().NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, subscription := range page.Value {
			if subscription != nil {
				result = append(result, subscription.SubscriptionID)
			}
		}
	}
	return result, nil
}

func (p *Provider) ListAllAssetTypesByName(assets []string) ([]AzureAsset, error) {
	p.log.Infof("Listing Azure assets: %v", assets)
	var resourceAssets []AzureAsset

	query := armresourcegraph.QueryRequest{
		Query: to.Ptr(generateQuery(assets)),
		Options: &armresourcegraph.QueryRequestOptions{
			ResultFormat: to.Ptr(armresourcegraph.ResultFormatObjectArray),
		},
		Subscriptions: p.subscriptions,
	}

	resourceAssets, err := p.runPaginatedQuery(query)
	if err != nil {
		return nil, err
	}

	return resourceAssets, nil
}

func getAssetFromData(data map[string]any) AzureAsset {
	properties, _ := data["properties"].(map[string]any)

	return AzureAsset{
		Id:             getString(data, "id"),
		Name:           getString(data, "name"),
		Location:       getString(data, "location"),
		Properties:     properties,
		ResourceGroup:  getString(data, "resourceGroup"),
		SubscriptionId: getString(data, "subscriptionId"),
		TenantId:       getString(data, "tenantId"),
		Type:           getString(data, "type"),
	}
}

func getString(data map[string]any, key string) string {
	value, _ := data[key].(string)
	return value
}

func generateQuery(assets []string) string {
	var query bytes.Buffer
	query.WriteString("Resources")
	for index, asset := range assets {
		if index == 0 {
			query.WriteString(" | where type == '")
		} else {
			query.WriteString(" or type == '")
		}
		query.WriteString(asset)
		query.WriteString("'")
	}
	return query.String()
}

func (p *Provider) runPaginatedQuery(query armresourcegraph.QueryRequest) ([]AzureAsset, error) {
	var resourceAssets []AzureAsset

	for {
		response, err := p.client.AssetQuery(p.ctx, query, nil)
		if err != nil {
			return nil, err
		}

		for _, asset := range response.Data.([]interface{}) {
			structuredAsset := getAssetFromData(asset.(map[string]any))
			resourceAssets = append(resourceAssets, structuredAsset)
		}

		if *response.ResultTruncated == *to.Ptr(armresourcegraph.ResultTruncatedTrue) &&
			response.SkipToken != nil &&
			*response.SkipToken != "" {
			query.Options.SkipToken = response.SkipToken
		} else {
			break
		}
	}

	return resourceAssets, nil
}
