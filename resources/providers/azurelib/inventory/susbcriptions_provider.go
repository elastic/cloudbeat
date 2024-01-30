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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/resources/utils/pointers"
)

type subscriptionAzureClientWrapper struct {
	AssetLocations func(ctx context.Context, subID string, clientOptions *arm.ClientOptions, options *armsubscriptions.ClientListLocationsOptions) ([]armsubscriptions.ClientListLocationsResponse, error)
}

type SubscriptionProviderAPI interface {
	ListLocations(ctx context.Context, subID string) ([]AzureAsset, error)
}

type subscriptionProvider struct {
	client subscriptionAzureClientWrapper
	log    *logp.Logger //nolint:unused
}

func NewSubscriptionProvider(log *logp.Logger, credentials azcore.TokenCredential) SubscriptionProviderAPI {
	client := subscriptionAzureClientWrapper{
		AssetLocations: func(ctx context.Context, subID string, clientOptions *arm.ClientOptions, options *armsubscriptions.ClientListLocationsOptions) ([]armsubscriptions.ClientListLocationsResponse, error) {
			cl, err := armsubscriptions.NewClient(credentials, clientOptions)
			if err != nil {
				return nil, err
			}

			return readPager(ctx, cl.NewListLocationsPager(subID, options))
		},
	}

	return &subscriptionProvider{
		client: client,
		log:    log,
	}
}

func (p *subscriptionProvider) ListLocations(ctx context.Context, subID string) ([]AzureAsset, error) {
	paged, err := p.client.AssetLocations(ctx, subID, nil, nil)
	if err != nil {
		return nil, err
	}

	locations := lo.FlatMap(paged, func(item armsubscriptions.ClientListLocationsResponse, _ int) []*armsubscriptions.Location {
		return item.LocationListResult.Value
	})

	assets := make([]AzureAsset, 0, len(locations))
	for _, loc := range locations {
		if loc == nil {
			continue
		}

		assets = append(assets, AzureAsset{
			Id:             pointers.Deref(loc.ID),
			Name:           pointers.Deref(loc.Name),
			DisplayName:    pointers.Deref(loc.DisplayName),
			Location:       pointers.Deref(loc.Name),
			SubscriptionId: pointers.Deref(loc.SubscriptionID),
			Type:           LocationAssetType,
		})
	}

	return assets, nil
}
