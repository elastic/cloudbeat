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
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func newBlobServicesClientListResponse(items ...*armstorage.BlobServiceProperties) armstorage.BlobServicesClientListResponse {
	return armstorage.BlobServicesClientListResponse{
		BlobServiceItems: armstorage.BlobServiceItems{
			Value: items,
		},
	}
}

func newAccountsClientListResponse(items ...*armstorage.Account) armstorage.AccountsClientListResponse {
	return armstorage.AccountsClientListResponse{
		AccountListResult: armstorage.AccountListResult{
			Value: items,
		},
	}
}

func TestTransformStorageAccounts(t *testing.T) {
	tests := map[string]struct {
		inputServicesPages  []armstorage.AccountsClientListResponse
		inputSubscriptionId string
		expected            []AzureAsset
		expectError         bool
	}{
		"noop": {},
		"transform response": {
			inputServicesPages: []armstorage.AccountsClientListResponse{
				newAccountsClientListResponse(
					&armstorage.Account{
						ID:   to.Ptr("id1"),
						Name: to.Ptr("name1"),
						Type: to.Ptr("Microsoft.Storage/storageaccounts"),
						Properties: &armstorage.AccountProperties{
							Encryption: &armstorage.Encryption{},
						},
					},
				),
			},
			inputSubscriptionId: "id1",
			expected: []AzureAsset{
				{
					Id:   "id1",
					Name: "name1",
					Properties: map[string]any{
						"encryption": map[string]any{},
					},
					Extension: map[string]any{
						ExtensionStorageAccountID:   "id1",
						ExtensionStorageAccountName: "name1",
					},
					Type:           StorageAccountAssetType,
					SubscriptionId: "id1",
				},
			},
			expectError: false,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			got, err := transformStorageAccounts(tc.inputServicesPages, tc.inputSubscriptionId)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expected, got)
		})
	}
}

func TestTransformBlobServices(t *testing.T) {
	tests := map[string]struct {
		inputServicesPages  []armstorage.BlobServicesClientListResponse
		inputStorageAccount AzureAsset
		expected            []AzureAsset
		expectError         bool
	}{
		"noop": {},
		"transform response": {
			inputServicesPages: []armstorage.BlobServicesClientListResponse{
				newBlobServicesClientListResponse(
					&armstorage.BlobServiceProperties{
						ID:                    to.Ptr("id1"),
						Name:                  to.Ptr("name1"),
						Type:                  to.Ptr("Microsoft.Storage/storageaccounts/blobservices"),
						BlobServiceProperties: &armstorage.BlobServicePropertiesProperties{IsVersioningEnabled: to.Ptr(true)},
					},
				),
			},
			inputStorageAccount: AzureAsset{
				Id:            "sa1",
				Name:          "sa name",
				ResourceGroup: "rg1",
				TenantId:      "t1",
			},
			expected: []AzureAsset{
				{
					Id:         "id1",
					Name:       "name1",
					Properties: map[string]any{"isVersioningEnabled": true},
					Extension: map[string]any{
						ExtensionStorageAccountID:   "sa1",
						ExtensionStorageAccountName: "sa name",
					},
					ResourceGroup: "rg1",
					TenantId:      "t1",
					Type:          BlobServiceAssetType,
				},
			},
			expectError: false,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			got, err := transformBlobServices(tc.inputServicesPages, tc.inputStorageAccount)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expected, got)
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
				"/subscriptions/sub1/": {
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
				},
			},
			expecterError: false,
		},
		"two elements one subscription": {
			subscriptions: map[string]string{"sub1": "subName1"},
			responsesPerSubscription: map[string][]armmonitor.DiagnosticSettingsClientListResponse{
				"/subscriptions/sub1/": {
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
				},
			},
			expecterError: false,
		},
		"two elements two subscriptions": {
			subscriptions: map[string]string{"sub1": "subName1", "sub2": "subName2"},
			responsesPerSubscription: map[string][]armmonitor.DiagnosticSettingsClientListResponse{
				"/subscriptions/sub1/": {
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
				"/subscriptions/sub2/": {
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
				},
			},
			expecterError: false,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			provider := &storageAccountProvider{
				log: log,
				client: &storageAccountAzureClientWrapper{
					AssetDiagnosticSettings: func(_ context.Context, resourceID string, _ *armmonitor.DiagnosticSettingsClientListOptions) ([]armmonitor.DiagnosticSettingsClientListResponse, error) {
						response := tc.responsesPerSubscription[resourceID]
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
