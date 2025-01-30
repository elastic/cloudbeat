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
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/infra/clog"
)

type (
	locationsFn     func() ([]armsubscriptions.ClientListLocationsResponse, error)
	tenantsFn       func() ([]armsubscriptions.TenantsClientListResponse, error)
	subscriptionsFn func() ([]armsubscriptions.ClientListResponse, error)
)

func mockLocationsAsset(fn locationsFn) SubscriptionProviderAPI {
	wrapper := locationAzureClientWrapper{
		AssetLocations: func(_ context.Context, _ string, _ *arm.ClientOptions, _ *armsubscriptions.ClientListLocationsOptions) ([]armsubscriptions.ClientListLocationsResponse, error) {
			return fn()
		},
	}

	return &subscriptionProvider{locationClient: wrapper, log: clog.NewLogger("mock_subscriptions_locations_asset_provider")}
}

func mockTenantAsset(fn tenantsFn) SubscriptionProviderAPI {
	tenantClientWrapper := tenantAzureClientWrapper{
		Tenants: func(_ context.Context, _ *arm.ClientOptions, _ *armsubscriptions.TenantsClientListOptions) ([]armsubscriptions.TenantsClientListResponse, error) {
			return fn()
		},
	}

	return &subscriptionProvider{tenantClient: tenantClientWrapper, log: clog.NewLogger("mock_subscriptions_tenants_asset_provider")}
}

func mockSubscriptionAsset(fn subscriptionsFn) SubscriptionProviderAPI {
	subscriptionClientWrapper := subscriptionAzureClientWrapper{
		Subscriptions: func(_ context.Context, _ *arm.ClientOptions, _ *armsubscriptions.ClientListOptions) ([]armsubscriptions.ClientListResponse, error) {
			return fn()
		},
	}

	return &subscriptionProvider{subscriptionClient: subscriptionClientWrapper, log: clog.NewLogger("mock_subscriptions_subscription_asset_provider")}
}

func TestSubscriptionProvider_ListLocations(t *testing.T) {
	cases := map[string]struct {
		expectError bool
		configMock  locationsFn
		expected    []AzureAsset
	}{
		"Returns error": {
			expectError: true,
			configMock: func() ([]armsubscriptions.ClientListLocationsResponse, error) {
				return nil, errors.New("error")
			},
			expected: nil,
		},
		"Returns multiple locations": {
			expectError: false,
			configMock: func() ([]armsubscriptions.ClientListLocationsResponse, error) {
				return []armsubscriptions.ClientListLocationsResponse{
					{
						LocationListResult: armsubscriptions.LocationListResult{
							Value: []*armsubscriptions.Location{
								{
									ID:             to.Ptr("/subscriptions/sub/locations/eastus2euap"),
									Name:           to.Ptr("eastus2euap"),
									DisplayName:    to.Ptr("East US 2 EUAP"),
									SubscriptionID: to.Ptr("sub"),
								},
								{
									ID:             to.Ptr("/subscriptions/sub/locations/westcentralus"),
									Name:           to.Ptr("westcentralus"),
									DisplayName:    to.Ptr("West Central US"),
									SubscriptionID: to.Ptr("sub"),
								},
							},
						},
					},
					{
						LocationListResult: armsubscriptions.LocationListResult{
							Value: []*armsubscriptions.Location{
								{
									ID:             to.Ptr("/subscriptions/sub/locations/southafricawest"),
									Name:           to.Ptr("southafricawest"),
									DisplayName:    to.Ptr("South Africa West"),
									SubscriptionID: to.Ptr("sub"),
								},
								{
									ID:             to.Ptr("/subscriptions/sub/locations/australiacentral"),
									Name:           to.Ptr("australiacentral"),
									DisplayName:    to.Ptr("Australia Central"),
									SubscriptionID: to.Ptr("sub"),
								},
								{
									ID:             to.Ptr("/subscriptions/sub/locations/australiacentral2"),
									Name:           to.Ptr("australiacentral2"),
									DisplayName:    to.Ptr("Australia Central 2"),
									SubscriptionID: to.Ptr("sub"),
								},
								{
									ID:             to.Ptr("/subscriptions/sub/locations/australiasoutheast"),
									Name:           to.Ptr("australiasoutheast"),
									DisplayName:    to.Ptr("Australia Southeast"),
									SubscriptionID: to.Ptr("sub"),
								},
							},
						},
					},
				}, nil
			},

			expected: []AzureAsset{
				{
					Id:             "/subscriptions/sub/locations/eastus2euap",
					Name:           "eastus2euap",
					DisplayName:    "East US 2 EUAP",
					Location:       "eastus2euap",
					SubscriptionId: "sub",
					Type:           LocationAssetType,
				},
				{
					Id:             "/subscriptions/sub/locations/westcentralus",
					Name:           "westcentralus",
					DisplayName:    "West Central US",
					Location:       "westcentralus",
					SubscriptionId: "sub",
					Type:           LocationAssetType,
				},
				{
					Id:             "/subscriptions/sub/locations/southafricawest",
					Name:           "southafricawest",
					DisplayName:    "South Africa West",
					Location:       "southafricawest",
					SubscriptionId: "sub",
					Type:           LocationAssetType,
				},
				{
					Id:             "/subscriptions/sub/locations/australiacentral",
					Name:           "australiacentral",
					DisplayName:    "Australia Central",
					Location:       "australiacentral",
					SubscriptionId: "sub",
					Type:           LocationAssetType,
				},
				{
					Id:             "/subscriptions/sub/locations/australiacentral2",
					Name:           "australiacentral2",
					DisplayName:    "Australia Central 2",
					Location:       "australiacentral2",
					SubscriptionId: "sub",
					Type:           LocationAssetType,
				},
				{
					Id:             "/subscriptions/sub/locations/australiasoutheast",
					Name:           "australiasoutheast",
					DisplayName:    "Australia Southeast",
					Location:       "australiasoutheast",
					SubscriptionId: "sub",
					Type:           LocationAssetType,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			provider := mockLocationsAsset(tc.configMock)

			assets, err := provider.ListLocations(context.Background(), "subscription")

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.expected, assets)
		})
	}
}

func TestSubscriptionProvider_ListTenants(t *testing.T) {
	cases := map[string]struct {
		expectError bool
		configMock  tenantsFn
		expected    []AzureAsset
	}{
		"Returns error": {
			expectError: true,
			configMock: func() ([]armsubscriptions.TenantsClientListResponse, error) {
				return nil, errors.New("error")
			},
			expected: nil,
		},
		"Returns tenants": {
			expectError: false,
			configMock: func() ([]armsubscriptions.TenantsClientListResponse, error) {
				return []armsubscriptions.TenantsClientListResponse{
					{
						TenantListResult: armsubscriptions.TenantListResult{
							Value: []*armsubscriptions.TenantIDDescription{
								{
									ID:          to.Ptr("/tenants/uuid"),
									TenantID:    to.Ptr("uuid"),
									DisplayName: to.Ptr("main account"),
								},
							},
						},
					},
				}, nil
			},

			expected: []AzureAsset{
				{
					Id:          "/tenants/uuid",
					Name:        "uuid",
					DisplayName: "main account",
					TenantId:    "uuid",
					Type:        TenantAssetType,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			provider := mockTenantAsset(tc.configMock)

			assets, err := provider.ListTenants(context.Background())

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.expected, assets)
		})
	}
}

func TestSubscriptionProvider_ListSubscriptions(t *testing.T) {
	cases := map[string]struct {
		expectError bool
		configMock  subscriptionsFn
		expected    []AzureAsset
	}{
		"Returns error": {
			expectError: true,
			configMock: func() ([]armsubscriptions.ClientListResponse, error) {
				return nil, errors.New("error")
			},
			expected: nil,
		},
		"Returns subscriptions": {
			expectError: false,
			configMock: func() ([]armsubscriptions.ClientListResponse, error) {
				return []armsubscriptions.ClientListResponse{
					{
						SubscriptionListResult: armsubscriptions.SubscriptionListResult{
							Value: []*armsubscriptions.Subscription{
								{
									ID:             to.Ptr("/subscriptions/uuid"),
									SubscriptionID: to.Ptr("uuid"),
									DisplayName:    to.Ptr("main sub"),
									TenantID:       to.Ptr("uuid2"),
								},
							},
						},
					},
				}, nil
			},

			expected: []AzureAsset{
				{
					Id:             "/subscriptions/uuid",
					Name:           "uuid",
					DisplayName:    "main sub",
					SubscriptionId: "uuid",
					TenantId:       "uuid2",
					Type:           SubscriptionAssetType,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			provider := mockSubscriptionAsset(tc.configMock)

			assets, err := provider.ListSubscriptions(context.Background())

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.expected, assets)
		})
	}
}
