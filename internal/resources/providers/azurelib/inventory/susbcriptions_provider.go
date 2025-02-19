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
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type locationAzureClientWrapper struct {
	AssetLocations func(ctx context.Context, subID string, clientOptions *arm.ClientOptions, options *armsubscriptions.ClientListLocationsOptions) ([]armsubscriptions.ClientListLocationsResponse, error)
}

type tenantAzureClientWrapper struct {
	Tenants func(ctx context.Context, clientOptions *arm.ClientOptions, options *armsubscriptions.TenantsClientListOptions) ([]armsubscriptions.TenantsClientListResponse, error)
}

type subscriptionAzureClientWrapper struct {
	Subscriptions func(ctx context.Context, clientOptions *arm.ClientOptions, options *armsubscriptions.ClientListOptions) ([]armsubscriptions.ClientListResponse, error)
}

type SubscriptionProviderAPI interface {
	ListLocations(ctx context.Context, subID string) ([]AzureAsset, error)
	ListTenants(ctx context.Context) ([]AzureAsset, error)
	ListSubscriptions(ctx context.Context) ([]AzureAsset, error)
}

type subscriptionProvider struct {
	locationClient     locationAzureClientWrapper
	tenantClient       tenantAzureClientWrapper
	subscriptionClient subscriptionAzureClientWrapper
	log                *clog.Logger //nolint:unused
}

func NewSubscriptionProvider(log *clog.Logger, credentials azcore.TokenCredential) SubscriptionProviderAPI {
	locationClient := locationAzureClientWrapper{
		AssetLocations: func(ctx context.Context, subID string, clientOptions *arm.ClientOptions, options *armsubscriptions.ClientListLocationsOptions) ([]armsubscriptions.ClientListLocationsResponse, error) {
			cl, err := armsubscriptions.NewClient(credentials, clientOptions)
			if err != nil {
				return nil, err
			}

			return readPager(ctx, cl.NewListLocationsPager(subID, options))
		},
	}

	tenantClient := tenantAzureClientWrapper{
		Tenants: func(ctx context.Context, clientOptions *arm.ClientOptions, options *armsubscriptions.TenantsClientListOptions) ([]armsubscriptions.TenantsClientListResponse, error) {
			c, err := armsubscriptions.NewTenantsClient(credentials, clientOptions)
			if err != nil {
				return nil, err
			}
			return readPager(ctx, c.NewListPager(options))
		},
	}

	subscriptionClient := subscriptionAzureClientWrapper{
		Subscriptions: func(ctx context.Context, clientOptions *arm.ClientOptions, options *armsubscriptions.ClientListOptions) ([]armsubscriptions.ClientListResponse, error) {
			cl, err := armsubscriptions.NewClient(credentials, clientOptions)
			if err != nil {
				return nil, err
			}

			return readPager(ctx, cl.NewListPager(options))
		},
	}

	return &subscriptionProvider{
		locationClient:     locationClient,
		tenantClient:       tenantClient,
		subscriptionClient: subscriptionClient,
		log:                log,
	}
}

func (p *subscriptionProvider) ListLocations(ctx context.Context, subID string) ([]AzureAsset, error) {
	paged, err := p.locationClient.AssetLocations(ctx, subID, nil, nil)
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

func (p *subscriptionProvider) ListTenants(ctx context.Context) ([]AzureAsset, error) {
	paged, err := p.tenantClient.Tenants(ctx, nil, nil)
	if err != nil {
		return nil, err
	}

	tenants := lo.FlatMap(paged, func(item armsubscriptions.TenantsClientListResponse, _ int) []*armsubscriptions.TenantIDDescription {
		return item.TenantListResult.Value
	})

	assets := make([]AzureAsset, 0, len(tenants))
	for _, t := range tenants {
		if t == nil {
			continue
		}

		assets = append(assets, AzureAsset{
			Id:          pointers.Deref(t.ID),
			Name:        pointers.Deref(t.TenantID),
			DisplayName: pointers.Deref(t.DisplayName),
			TenantId:    pointers.Deref(t.TenantID),
			Type:        TenantAssetType,
		})
	}

	return assets, nil
}

func (p *subscriptionProvider) ListSubscriptions(ctx context.Context) ([]AzureAsset, error) {
	paged, err := p.subscriptionClient.Subscriptions(ctx, nil, nil)
	if err != nil {
		return nil, err
	}

	subscriptions := lo.FlatMap(paged, func(item armsubscriptions.ClientListResponse, _ int) []*armsubscriptions.Subscription {
		return item.SubscriptionListResult.Value
	})

	assets := make([]AzureAsset, 0, len(subscriptions))
	for _, s := range subscriptions {
		if s == nil {
			continue
		}

		tags := map[string]any{}
		for k, v := range s.Tags {
			if v == nil {
				continue
			}
			tags[k] = *v
		}

		assets = append(assets, AzureAsset{
			Id:             pointers.Deref(s.ID),
			Name:           pointers.Deref(s.SubscriptionID),
			DisplayName:    pointers.Deref(s.DisplayName),
			TenantId:       pointers.Deref(s.TenantID),
			SubscriptionId: pointers.Deref(s.SubscriptionID),
			Type:           SubscriptionAssetType,
			Tags:           tags,
		})
	}

	return assets, nil
}
