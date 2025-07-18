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
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type ResourceGraphAzureClientWrapper struct {
	AssetQuery func(ctx context.Context, query armresourcegraph.QueryRequest, options *armresourcegraph.ClientResourcesOptions) (armresourcegraph.ClientResourcesResponse, error)
}

type ResourceGraphProviderAPI interface {
	// ListAllAssetTypesByName List all content types of the given assets types
	ListAllAssetTypesByName(ctx context.Context, assetsGroup string, assets []string) ([]AzureAsset, error)
}

type ResourceGraphProvider struct {
	client *ResourceGraphAzureClientWrapper
	log    *clog.Logger
}

func NewResourceGraphProvider(log *clog.Logger, resourceGraphClient *armresourcegraph.Client) ResourceGraphProviderAPI {
	// We wrap the client, so we can mock it in tests
	wrapper := &ResourceGraphAzureClientWrapper{
		AssetQuery: func(ctx context.Context, query armresourcegraph.QueryRequest, options *armresourcegraph.ClientResourcesOptions) (armresourcegraph.ClientResourcesResponse, error) {
			return resourceGraphClient.Resources(ctx, query, options)
		},
	}

	return &ResourceGraphProvider{
		log:    log,
		client: wrapper,
	}
}

func (p *ResourceGraphProvider) ListAllAssetTypesByName(ctx context.Context, assetGroup string, assets []string) ([]AzureAsset, error) {
	p.log.Infof("Listing Azure assets: %v", assets)

	query := armresourcegraph.QueryRequest{
		Query: to.Ptr(generateQuery(assetGroup, assets)),
		Options: &armresourcegraph.QueryRequestOptions{
			ResultFormat: to.Ptr(armresourcegraph.ResultFormatObjectArray),
		},
	}

	return p.runPaginatedQuery(ctx, query)
}

func generateQuery(assetGroup string, assets []string) string {
	var query bytes.Buffer
	query.WriteString(assetGroup)
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

func (p *ResourceGraphProvider) runPaginatedQuery(ctx context.Context, query armresourcegraph.QueryRequest) ([]AzureAsset, error) {
	var resourceAssets []AzureAsset

	for {
		response, err := p.client.AssetQuery(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		data, ok := response.Data.([]any)
		if !ok {
			return nil, errors.New("expected []any from response.Data")
		}
		for _, asset := range data {
			assetMap, ok := asset.(map[string]any)
			if !ok {
				continue // Skip malformed assets
			}
			structuredAsset := getAssetFromData(assetMap)
			resourceAssets = append(resourceAssets, structuredAsset)
		}

		if *response.ResultTruncated == armresourcegraph.ResultTruncatedFalse ||
			pointers.Deref(response.SkipToken) == "" {
			break
		}
		query.Options.SkipToken = response.SkipToken
	}

	return resourceAssets, nil
}

func readPager[T any](ctx context.Context, pager *runtime.Pager[T]) ([]T, error) {
	var res []T
	for pager.More() {
		r, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		res = append(res, r)
	}

	return res, nil
}
