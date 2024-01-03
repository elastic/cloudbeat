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
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/resources/utils/strings"
)

type azureClientWrapper struct {
	AssetQuery              func(ctx context.Context, query armresourcegraph.QueryRequest, options *armresourcegraph.ClientResourcesOptions) (armresourcegraph.ClientResourcesResponse, error)
	AssetDiagnosticSettings func(ctx context.Context, subID string, options *armmonitor.DiagnosticSettingsClientListOptions) ([]armmonitor.DiagnosticSettingsClientListResponse, error)
}

type ProviderAPI interface {
	// ListAllAssetTypesByName List all content types of the given assets types
	ListAllAssetTypesByName(ctx context.Context, assetsGroup string, assets []string) ([]AzureAsset, error)
	ListDiagnosticSettingsAssetTypes(ctx context.Context, cycleMetadata cycle.Metadata, subscriptionIDs []string) ([]AzureAsset, error)
}

type provider struct {
	client                  *azureClientWrapper
	log                     *logp.Logger
	diagnosticSettingsCache *cycle.Cache[[]AzureAsset]
}

func NewProvider(log *logp.Logger, resourceGraphClient *armresourcegraph.Client, diagnosticSettingsClient *armmonitor.DiagnosticSettingsClient) ProviderAPI {
	// We wrap the client, so we can mock it in tests
	wrapper := &azureClientWrapper{
		AssetQuery: func(ctx context.Context, query armresourcegraph.QueryRequest, options *armresourcegraph.ClientResourcesOptions) (armresourcegraph.ClientResourcesResponse, error) {
			return resourceGraphClient.Resources(ctx, query, options)
		},
		AssetDiagnosticSettings: func(ctx context.Context, subID string, options *armmonitor.DiagnosticSettingsClientListOptions) ([]armmonitor.DiagnosticSettingsClientListResponse, error) {
			pager := diagnosticSettingsClient.NewListPager(fmt.Sprintf("/subscriptions/%s/", subID), options)
			return readPager(ctx, pager)
		},
	}

	return &provider{log: log, client: wrapper, diagnosticSettingsCache: cycle.NewCache[[]AzureAsset](log)}
}

func (p *provider) ListAllAssetTypesByName(ctx context.Context, assetGroup string, assets []string) ([]AzureAsset, error) {
	p.log.Infof("Listing Azure assets: %v", assets)

	query := armresourcegraph.QueryRequest{
		Query: to.Ptr(generateQuery(assetGroup, assets)),
		Options: &armresourcegraph.QueryRequestOptions{
			ResultFormat: to.Ptr(armresourcegraph.ResultFormatObjectArray),
		},
	}

	return p.runPaginatedQuery(ctx, query)
}

func (p *provider) ListDiagnosticSettingsAssetTypes(ctx context.Context, cycleMetadata cycle.Metadata, subscriptionIDs []string) ([]AzureAsset, error) {
	p.log.Info("Listing Azure Diagnostic Monitor Settings")

	return p.diagnosticSettingsCache.GetValue(ctx, cycleMetadata, func(ctx context.Context) ([]AzureAsset, error) {
		return p.getDiagnosticSettings(ctx, subscriptionIDs)
	})
}

func (p *provider) getDiagnosticSettings(ctx context.Context, subscriptionIDs []string) ([]AzureAsset, error) {
	var assets []AzureAsset

	for _, subID := range subscriptionIDs {
		responses, err := p.client.AssetDiagnosticSettings(ctx, subID, nil)
		if err != nil {
			return nil, err
		}
		a, err := transformDiagnosticSettingsClientListResponses(responses, subID)
		if err != nil {
			return nil, err
		}
		assets = append(assets, a...)
	}

	return assets, nil
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

func (p *provider) runPaginatedQuery(ctx context.Context, query armresourcegraph.QueryRequest) ([]AzureAsset, error) {
	var resourceAssets []AzureAsset

	for {
		response, err := p.client.AssetQuery(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		for _, asset := range response.Data.([]any) {
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

func transformDiagnosticSettingsClientListResponses(response []armmonitor.DiagnosticSettingsClientListResponse, subID string) ([]AzureAsset, error) {
	var assets []AzureAsset

	for _, settingsCollection := range response {
		for _, v := range settingsCollection.Value {
			if v == nil {
				continue
			}
			a, err := transformDiagnosticSettingsResource(v, subID)
			if err != nil {
				return nil, fmt.Errorf("error parsing azure asset model: %w", err)
			}
			assets = append(assets, a)
		}
	}

	return assets, nil
}

func transformDiagnosticSettingsResource(v *armmonitor.DiagnosticSettingsResource, subID string) (AzureAsset, error) {
	properties, err := transformDiagnosticSettings(v.Properties)
	if err != nil {
		return AzureAsset{}, err
	}

	return AzureAsset{
		Id:             strings.Dereference(v.ID),
		Name:           strings.Dereference(v.Name),
		Location:       "global",
		Properties:     properties,
		ResourceGroup:  "",
		SubscriptionId: subID,
		TenantId:       "",
		Type:           strings.Dereference(v.Type),
	}, nil
}

func transformDiagnosticSettings(d *armmonitor.DiagnosticSettings) (map[string]any, error) {
	if d == nil {
		return nil, nil
	}

	js, err := d.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json: %w", err)
	}

	m := map[string]any{}
	err = json.Unmarshal(js, &m)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	return m, nil
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
