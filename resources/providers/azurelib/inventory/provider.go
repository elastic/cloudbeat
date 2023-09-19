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
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/providers/azurelib/auth"
)

type Provider struct {
	log    *logp.Logger
	client *AzureClientWrapper
	ctx    context.Context
	Config auth.AzureFactoryConfig
}

type ProviderInitializer struct{}

type AzureClientWrapper struct {
	AssetQuery func(ctx context.Context, query armresourcegraph.QueryRequest, options *armresourcegraph.ClientResourcesOptions) (armresourcegraph.ClientResourcesResponse, error)
}

type ServiceAPI interface {
	// ListAllAssetTypesByName List all content types of the given assets types
	ListAllAssetTypesByName(assets []string) ([]interface{}, error)
}

type ProviderInitializerAPI interface {
	// Init initializes the Azure asset client
	Init(ctx context.Context, log *logp.Logger, azureConfig auth.AzureFactoryConfig) (ServiceAPI, error)
}

func (p *ProviderInitializer) Init(ctx context.Context, log *logp.Logger, azureConfig auth.AzureFactoryConfig) (ServiceAPI, error) {
	clientFactory, err := armresourcegraph.NewClientFactory(azureConfig.Credentials, nil)
	if err != nil {
		return nil, err
	}

	client := clientFactory.NewClient()

	// We wrap the client so we can mock it in tests
	wrapper := &AzureClientWrapper{
		AssetQuery: func(ctx context.Context, query armresourcegraph.QueryRequest, options *armresourcegraph.ClientResourcesOptions) (armresourcegraph.ClientResourcesResponse, error) {
			return client.Resources(ctx, query, options)
		},
	}

	return &Provider{
		Config: azureConfig,
		client: wrapper,
		log:    log,
		ctx:    ctx,
	}, nil
}

func (p *Provider) ListAllAssetTypesByName(assets []string) ([]interface{}, error) {
	p.log.Infof("Listing Azure assets: %v", assets)
	var resourceAssets []interface{}

	query := armresourcegraph.QueryRequest{
		Query: to.Ptr(generateQuery(assets)),
		Options: &armresourcegraph.QueryRequestOptions{
			ResultFormat: to.Ptr(armresourcegraph.ResultFormatObjectArray),
		},
		Subscriptions: []*string{
			// TODO: Populate from config or query (not sensitive but still don't want to commit)
			to.Ptr(os.Getenv("AZURE_SUBSCRIPTION_ID"))},
	}

	resourceAssets, err := p.runPaginatedQuery(query)
	if err != nil {
		return nil, err
	}

	return resourceAssets, nil
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

func (p *Provider) runPaginatedQuery(query armresourcegraph.QueryRequest) ([]interface{}, error) {
	var resourceAssets []interface{}

	for {
		response, err := p.client.AssetQuery(p.ctx, query, nil)
		if err != nil {
			return nil, err
		}

		resourceAssets = append(resourceAssets, response.Data.([]interface{})...)

		if response.SkipToken != nil && *response.SkipToken != "" {
			query.Options = &armresourcegraph.QueryRequestOptions{
				SkipToken: response.SkipToken,
			}
		} else {
			break
		}
	}

	return resourceAssets, nil
}

// TODO: Handle this function
func (p *Provider) Close() error {
	return nil
}
