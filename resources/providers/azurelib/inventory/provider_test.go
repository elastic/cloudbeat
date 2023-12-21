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
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/elastic/cloudbeat/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/resources/utils/strings"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
)

type ProviderTestSuite struct {
	suite.Suite
	mockedClient *azureClientWrapper
}

var nonTruncatedResponse = armresourcegraph.QueryResponse{
	Count: to.Ptr(int64(1)),
	Data: []any{
		map[string]any{
			"id":             "3",
			"name":           "3",
			"location":       "3",
			"properties":     map[string]any{"test": "test"},
			"resourceGroup":  "3",
			"subscriptionId": "3",
			"tenantId":       "3",
			"type":           "3",
			"sku":            "3",
		},
	},
	ResultTruncated: to.Ptr(armresourcegraph.ResultTruncatedFalse),
}

var truncatedResponse = armresourcegraph.QueryResponse{
	Count: to.Ptr(int64(2)),
	Data: []any{
		map[string]any{
			"id":             "1",
			"name":           "1",
			"location":       "1",
			"properties":     map[string]any{"test": "test"},
			"resourceGroup":  "1",
			"subscriptionId": "1",
			"tenantId":       "1",
			"type":           "1",
			"sku":            "1",
		},
		map[string]any{
			"id":             "2",
			"name":           "2",
			"location":       "2",
			"properties":     map[string]any{"test": "test"},
			"resourceGroup":  "2",
			"subscriptionId": "2",
			"tenantId":       "2",
			"type":           "2",
			"sku":            "2",
		},
	},
	ResultTruncated: to.Ptr(armresourcegraph.ResultTruncatedTrue),
	SkipToken:       to.Ptr("token"),
}

func TestInventoryProviderTestSuite(t *testing.T) {
	s := new(ProviderTestSuite)

	suite.Run(t, s)
}

func (s *ProviderTestSuite) SetupTest() {
	s.mockedClient = &azureClientWrapper{
		AssetQuery: func(ctx context.Context, query armresourcegraph.QueryRequest, options *armresourcegraph.ClientResourcesOptions) (armresourcegraph.ClientResourcesResponse, error) {
			if query.Options.SkipToken != nil && *query.Options.SkipToken != "" {
				return armresourcegraph.ClientResourcesResponse{
					QueryResponse: nonTruncatedResponse,
				}, nil
			}
			return armresourcegraph.ClientResourcesResponse{
				QueryResponse: truncatedResponse,
			}, nil
		},
	}
}

func (s *ProviderTestSuite) TestGetString() {
	tests := []struct {
		name string
		data map[string]any
		key  string
		want string
	}{
		{
			name: "nil map",
			data: nil,
			key:  "key",
			want: "",
		},
		{
			name: "key does not exist",
			data: map[string]any{"key": "value"},
			key:  "other-key",
			want: "",
		},
		{
			name: "wrong type",
			data: map[string]any{"key": 1},
			key:  "key",
			want: "",
		},
		{
			name: "correct value",
			data: map[string]any{"key": "value", "other-key": 1},
			key:  "key",
			want: "value",
		},
	}
	for _, tt := range tests {
		s.Equal(tt.want, strings.FromMap(tt.data, tt.key), "getString(%v, %s) = %s", tt.data, tt.key, tt.want)
	}
}

func (s *ProviderTestSuite) TestListAllAssetTypesByName() {
	provider := &provider{
		log:    testhelper.NewLogger(s.T()),
		client: s.mockedClient,
	}

	values, err := provider.ListAllAssetTypesByName(context.Background(), "test", []string{"test"})
	s.Require().NoError(err)
	s.Len(values, int(*nonTruncatedResponse.Count+*truncatedResponse.Count))
	lo.ForEach(values, func(r AzureAsset, index int) {
		strIndex := fmt.Sprintf("%d", index+1)
		s.Equal(AzureAsset{
			Id:             strIndex,
			Name:           strIndex,
			Location:       strIndex,
			Properties:     map[string]any{"test": "test"},
			ResourceGroup:  strIndex,
			SubscriptionId: strIndex,
			TenantId:       strIndex,
			Type:           strIndex,
			Sku:            strIndex,
		}, r)
	})
}

func Test_generateQuery(t *testing.T) {
	tests := []struct {
		assetsGroup string
		assets      []string
		want        string
	}{
		{
			assetsGroup: "empty assets",
			want:        "empty assets",
		},
		{
			assetsGroup: "resources",
			assets:      []string{"one"},
			want:        "resources | where type == 'one'",
		},
		{
			assetsGroup: "resources",
			assets:      []string{"one", "two", "three four five"},
			want:        "resources | where type == 'one' or type == 'two' or type == 'three four five'",
		},
		{
			assetsGroup: "authorizationresources",
			assets:      []string{"one"},
			want:        "authorizationresources | where type == 'one'",
		},
		{
			assetsGroup: "authorizationresources",
			assets:      []string{"one", "two", "three four five"},
			want:        "authorizationresources | where type == 'one' or type == 'two' or type == 'three four five'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, generateQuery(tt.assetsGroup, tt.assets))
		})
	}
}

func TestListDiagnosticSettingsAssetTypes(t *testing.T) {
	log := testhelper.NewLogger(t)

	response := func(v []*armmonitor.DiagnosticSettingsResource) armmonitor.DiagnosticSettingsClientListResponse {
		return armmonitor.DiagnosticSettingsClientListResponse{
			DiagnosticSettingsResourceCollection: armmonitor.DiagnosticSettingsResourceCollection{Value: v},
		}
	}

	tests := map[string]struct {
		subscriptions            map[string]string
		responsesPerSubscription map[string][]armmonitor.DiagnosticSettingsClientListResponse
		expected                 []AzureAsset
		expecterError            bool
	}{
		"one element one subscription": {
			subscriptions: map[string]string{"sub1": "subName1"},
			responsesPerSubscription: map[string][]armmonitor.DiagnosticSettingsClientListResponse{
				"sub1": {
					response([]*armmonitor.DiagnosticSettingsResource{
						{
							ID:   to.Ptr("id1"),
							Name: to.Ptr("name1"),
							Type: to.Ptr("Microsoft.Insights/diagnosticSettings"),
							Properties: &armmonitor.DiagnosticSettings{
								EventHubAuthorizationRuleID: nil,
								EventHubName:                nil,
								LogAnalyticsDestinationType: nil,
								MarketplacePartnerID:        nil,
								ServiceBusRuleID:            nil,
								StorageAccountID:            nil,
								WorkspaceID:                 to.Ptr("/workspace1"),
								Logs: []*armmonitor.LogSettings{
									{
										Category:        to.Ptr("Administrative"),
										Enabled:         to.Ptr(true),
										CategoryGroup:   nil,
										RetentionPolicy: nil,
									},
									{
										Category:        to.Ptr("Security"),
										Enabled:         to.Ptr(false),
										CategoryGroup:   nil,
										RetentionPolicy: nil,
									},
								},
								Metrics: nil,
							},
						},
					}),
				},
			},
			expected: []AzureAsset{
				{
					Id:       "id1",
					Name:     "name1",
					Location: "global",
					Properties: map[string]any{
						"logs": []any{
							map[string]any{
								"category": "Administrative",
								"enabled":  true,
							},
							map[string]any{
								"category": "Security",
								"enabled":  false,
							},
						},
						"workspaceId": "/workspace1",
					},
					ResourceGroup:  "",
					SubscriptionId: "sub1",
					TenantId:       "",
					Type:           "Microsoft.Insights/diagnosticSettings",
					Sku:            "",
				},
			},
			expecterError: false,
		},
		"two elements one subscription": {
			subscriptions: map[string]string{"sub1": "subName1"},
			responsesPerSubscription: map[string][]armmonitor.DiagnosticSettingsClientListResponse{
				"sub1": {
					response([]*armmonitor.DiagnosticSettingsResource{
						{
							ID:   to.Ptr("id2"),
							Name: to.Ptr("name2"),
							Type: to.Ptr("Microsoft.Insights/diagnosticSettings"),
							Properties: &armmonitor.DiagnosticSettings{
								EventHubAuthorizationRuleID: nil,
								EventHubName:                nil,
								LogAnalyticsDestinationType: nil,
								MarketplacePartnerID:        nil,
								ServiceBusRuleID:            nil,
								StorageAccountID:            nil,
								WorkspaceID:                 to.Ptr("/workspace2"),
								Logs: []*armmonitor.LogSettings{
									{
										Category:        to.Ptr("Administrative"),
										Enabled:         to.Ptr(false),
										CategoryGroup:   nil,
										RetentionPolicy: nil,
									},
									{
										Category:        to.Ptr("Security"),
										Enabled:         to.Ptr(true),
										CategoryGroup:   nil,
										RetentionPolicy: nil,
									},
								},
								Metrics: nil,
							},
						},
						{
							ID:   to.Ptr("id3"),
							Name: to.Ptr("name3"),
							Type: to.Ptr("Microsoft.Insights/diagnosticSettings"),
							Properties: &armmonitor.DiagnosticSettings{
								EventHubAuthorizationRuleID: nil,
								EventHubName:                nil,
								LogAnalyticsDestinationType: nil,
								MarketplacePartnerID:        nil,
								ServiceBusRuleID:            nil,
								StorageAccountID:            nil,
								WorkspaceID:                 to.Ptr("/workspace3"),
								Logs: []*armmonitor.LogSettings{
									{
										Category:        to.Ptr("Administrative"),
										Enabled:         to.Ptr(true),
										CategoryGroup:   nil,
										RetentionPolicy: nil,
									},
									{
										Category:        to.Ptr("Security"),
										Enabled:         to.Ptr(true),
										CategoryGroup:   nil,
										RetentionPolicy: nil,
									},
								},
								Metrics: nil,
							},
						},
					}),
				},
			},
			expected: []AzureAsset{
				{
					Id:       "id2",
					Name:     "name2",
					Location: "global",
					Properties: map[string]any{
						"logs": []any{
							map[string]any{
								"category": "Administrative",
								"enabled":  false,
							},
							map[string]any{
								"category": "Security",
								"enabled":  true,
							},
						},
						"workspaceId": "/workspace2",
					},
					ResourceGroup:  "",
					SubscriptionId: "sub1",
					TenantId:       "",
					Type:           "Microsoft.Insights/diagnosticSettings",
					Sku:            "",
				},
				{
					Id:       "id3",
					Name:     "name3",
					Location: "global",
					Properties: map[string]any{
						"logs": []any{
							map[string]any{
								"category": "Administrative",
								"enabled":  true,
							},
							map[string]any{
								"category": "Security",
								"enabled":  true,
							},
						},
						"workspaceId": "/workspace3",
					},
					ResourceGroup:  "",
					SubscriptionId: "sub1",
					TenantId:       "",
					Type:           "Microsoft.Insights/diagnosticSettings",
					Sku:            "",
				},
			},
			expecterError: false,
		},
		"two elements two subscriptions": {
			subscriptions: map[string]string{"sub1": "subName1", "sub2": "subName2"},
			responsesPerSubscription: map[string][]armmonitor.DiagnosticSettingsClientListResponse{
				"sub1": {
					response([]*armmonitor.DiagnosticSettingsResource{
						{
							ID:   to.Ptr("id2"),
							Name: to.Ptr("name2"),
							Type: to.Ptr("Microsoft.Insights/diagnosticSettings"),
							Properties: &armmonitor.DiagnosticSettings{
								EventHubAuthorizationRuleID: nil,
								EventHubName:                nil,
								LogAnalyticsDestinationType: nil,
								MarketplacePartnerID:        nil,
								ServiceBusRuleID:            nil,
								StorageAccountID:            nil,
								WorkspaceID:                 to.Ptr("/workspace2"),
								Logs: []*armmonitor.LogSettings{
									{
										Category:        to.Ptr("Administrative"),
										Enabled:         to.Ptr(false),
										CategoryGroup:   nil,
										RetentionPolicy: nil,
									},
									{
										Category:        to.Ptr("Security"),
										Enabled:         to.Ptr(true),
										CategoryGroup:   nil,
										RetentionPolicy: nil,
									},
								},
								Metrics: nil,
							},
						},
					}),
				},
				"sub2": {
					response([]*armmonitor.DiagnosticSettingsResource{
						{
							ID:   to.Ptr("id3"),
							Name: to.Ptr("name3"),
							Type: to.Ptr("Microsoft.Insights/diagnosticSettings"),
							Properties: &armmonitor.DiagnosticSettings{
								EventHubAuthorizationRuleID: nil,
								EventHubName:                nil,
								LogAnalyticsDestinationType: nil,
								MarketplacePartnerID:        nil,
								ServiceBusRuleID:            nil,
								StorageAccountID:            nil,
								WorkspaceID:                 to.Ptr("/workspace3"),
								Logs: []*armmonitor.LogSettings{
									{
										Category:        to.Ptr("Administrative"),
										Enabled:         to.Ptr(true),
										CategoryGroup:   nil,
										RetentionPolicy: nil,
									},
									{
										Category:        to.Ptr("Security"),
										Enabled:         to.Ptr(true),
										CategoryGroup:   nil,
										RetentionPolicy: nil,
									},
								},
								Metrics: nil,
							},
						},
					}),
				},
			},
			expected: []AzureAsset{
				{
					Id:       "id2",
					Name:     "name2",
					Location: "global",
					Properties: map[string]any{
						"logs": []any{
							map[string]any{
								"category": "Administrative",
								"enabled":  false,
							},
							map[string]any{
								"category": "Security",
								"enabled":  true,
							},
						},
						"workspaceId": "/workspace2",
					},
					ResourceGroup:  "",
					SubscriptionId: "sub1",
					TenantId:       "",
					Type:           "Microsoft.Insights/diagnosticSettings",
					Sku:            "",
				},
				{
					Id:       "id3",
					Name:     "name3",
					Location: "global",
					Properties: map[string]any{
						"logs": []any{
							map[string]any{
								"category": "Administrative",
								"enabled":  true,
							},
							map[string]any{
								"category": "Security",
								"enabled":  true,
							},
						},
						"workspaceId": "/workspace3",
					},
					ResourceGroup:  "",
					SubscriptionId: "sub2",
					TenantId:       "",
					Type:           "Microsoft.Insights/diagnosticSettings",
					Sku:            "",
				},
			},
			expecterError: false,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			provider := &provider{
				log: log,
				client: &azureClientWrapper{
					AssetDiagnosticSettings: func(_ context.Context, subID string, _ *armmonitor.DiagnosticSettingsClientListOptions) ([]armmonitor.DiagnosticSettingsClientListResponse, error) {
						response := tc.responsesPerSubscription[subID]
						return response, nil
					},
				},
				diagnosticSettingsCache: cycle.NewCache[[]AzureAsset](log),
			}

			got, err := provider.ListDiagnosticSettingsAssetTypes(context.Background(), cycle.Metadata{}, lo.Keys[string](tc.subscriptions))
			if tc.expecterError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.ElementsMatch(t, tc.expected, got)
		})
	}
}

func TestReadPager(t *testing.T) {
	tests := map[string]struct {
		moreFn    func(i int) bool
		fetchFn   func(context.Context, *int) (int, error)
		expectErr bool
		expected  []int
	}{
		"happy path, four pages": {
			moreFn: func(i int) bool {
				return i < 4
			},
			fetchFn: func(_ context.Context, i *int) (int, error) {
				if i == nil {
					i = to.Ptr(0)
				}
				*i++
				return *i, nil
			},
			expectErr: false,
			expected:  []int{1, 2, 3, 4},
		},

		"error at third of four pages": {
			moreFn: func(i int) bool {
				return i < 4
			},
			fetchFn: func(_ context.Context, i *int) (int, error) {
				if i == nil {
					i = to.Ptr(0)
				}
				*i++
				if *i == 3 {
					return *i, errors.New("mock error")
				}
				return *i, nil
			},
			expectErr: true,
			expected:  nil,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			pagerHandlerMock := runtime.PagingHandler[int]{
				More:    tc.moreFn,
				Fetcher: tc.fetchFn,
			}
			pager := runtime.NewPager[int](pagerHandlerMock)
			intSlice, err := readPager[int](context.Background(), pager)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.expected, intSlice)
		})
	}
}
