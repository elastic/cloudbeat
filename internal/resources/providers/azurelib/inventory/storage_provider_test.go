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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
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
	id := "/subscriptions/<subid>/resourceGroups/<name>/providers/Microsoft.Storage/storageAccounts/id1"
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
						ID:   to.Ptr(id),
						Name: to.Ptr("id1"),
						Type: to.Ptr("Microsoft.Storage/storageaccounts"),
						Properties: &armstorage.AccountProperties{
							Encryption: &armstorage.Encryption{},
						},
					},
				),
			},
			inputSubscriptionId: "subid",
			expected: []AzureAsset{
				{
					Id:   id,
					Name: "id1",
					Properties: map[string]any{
						"encryption": map[string]any{},
					},
					Extension: map[string]any{
						ExtensionStorageAccountID:   id,
						ExtensionStorageAccountName: "id1",
					},
					Type:           StorageAccountAssetType,
					SubscriptionId: "subid",
					ResourceGroup:  "<name>",
				},
			},
			expectError: false,
		},
	}

	for name, tc := range tests {
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

func TestListStorageAccountQueues(t *testing.T) {
	response := func(v []*armstorage.ListQueue) armstorage.QueueClientListResponse {
		return armstorage.QueueClientListResponse{
			ListQueueResource: armstorage.ListQueueResource{Value: v},
		}
	}

	tests := map[string]struct {
		input         []*armstorage.ListQueue
		expected      []AzureAsset
		expectedError bool
	}{
		"list queues": {
			input: []*armstorage.ListQueue{
				{
					ID:              pointers.Ref("/subscriptions/<subid>/resourceGroups/<rgname>/providers/Microsoft.Storage/storageAccounts/<storageid>/queues/queue1"),
					Name:            pointers.Ref("queue1"),
					Type:            pointers.Ref("microsoft.storage/storageaccounts/queues"),
					QueueProperties: &armstorage.ListQueueProperties{},
				},
			},
			expected: []AzureAsset{
				{
					Id:         "/subscriptions/<subid>/resourceGroups/<rgname>/providers/Microsoft.Storage/storageAccounts/<storageid>/queues/queue1",
					Name:       "queue1",
					Properties: map[string]any{},
					Extension: map[string]any{
						"storageAccountId":   "<storageid>",
						"storageAccountName": "<storageid>",
					},
					Type: "microsoft.storage/storageaccounts/queues",
				},
			},
			expectedError: false,
		},
	}

	log := testhelper.NewLogger(t)
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			provider := &storageAccountProvider{
				log: log,
				client: &storageAccountAzureClientWrapper{
					AssetQueues: func(_ context.Context, _ string, _ *arm.ClientOptions, _, _ string, _ *armstorage.QueueClientListOptions) ([]armstorage.QueueClientListResponse, error) {
						x := tc
						return []armstorage.QueueClientListResponse{response(x.input)}, nil
					},
				},
				diagnosticSettingsCache: cycle.NewCache[[]AzureAsset](log),
			}

			got, err := provider.ListStorageAccountQueues(t.Context(), []AzureAsset{
				{
					Type: "Storage Account",
					Id:   "<storageid>",
					Name: "<storageid>",
				},
			})
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.ElementsMatch(t, tc.expected, got)
		})
	}
}

func TestListStorageAccountQueueServices(t *testing.T) {
	response := func(v []*armstorage.QueueServiceProperties) armstorage.QueueServicesClientListResponse {
		return armstorage.QueueServicesClientListResponse{
			ListQueueServices: armstorage.ListQueueServices{
				Value: v,
			},
		}
	}

	tests := map[string]struct {
		input         []*armstorage.QueueServiceProperties
		expected      []AzureAsset
		expectedError bool
	}{
		"list queue services": {
			input: []*armstorage.QueueServiceProperties{
				{
					ID:                     pointers.Ref("/subscriptions/<subid>/resourceGroups/<rgname>/providers/Microsoft.Storage/storageAccounts/<storageid>/queueServices/queue1"),
					Name:                   pointers.Ref("queue1"),
					Type:                   pointers.Ref("microsoft.storage/storageaccounts/queueservices"),
					QueueServiceProperties: &armstorage.QueueServicePropertiesProperties{},
				},
			},
			expected: []AzureAsset{
				{
					Id:         "/subscriptions/<subid>/resourceGroups/<rgname>/providers/Microsoft.Storage/storageAccounts/<storageid>/queueServices/queue1",
					Name:       "queue1",
					Properties: map[string]any{},
					Extension: map[string]any{
						"storageAccountId":   "<storageid>",
						"storageAccountName": "<storageid>",
					},
					Type: "microsoft.storage/storageaccounts/queueservices",
				},
			},
			expectedError: false,
		},
	}

	log := testhelper.NewLogger(t)
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			provider := &storageAccountProvider{
				log: log,
				client: &storageAccountAzureClientWrapper{
					AssetQueueServices: func(_ context.Context, _ string, _ *arm.ClientOptions, _, _ string, _ *armstorage.QueueServicesClientListOptions) (armstorage.QueueServicesClientListResponse, error) {
						x := tc
						return response(x.input), nil
					},
				},
				diagnosticSettingsCache: cycle.NewCache[[]AzureAsset](log),
			}

			got, err := provider.ListStorageAccountQueueServices(t.Context(), []AzureAsset{
				{
					Type: "Storage Account",
					Id:   "<storageid>",
					Name: "<storageid>",
				},
			})
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.ElementsMatch(t, tc.expected, got)
		})
	}
}

<<<<<<< HEAD
=======
func TestListStorageAccountBlobContainers(t *testing.T) {
	response := func(v []*armstorage.ListContainerItem) armstorage.BlobContainersClientListResponse {
		return armstorage.BlobContainersClientListResponse{
			ListContainerItems: armstorage.ListContainerItems{
				Value: v,
			},
		}
	}

	tests := map[string]struct {
		input         []*armstorage.ListContainerItem
		expected      []AzureAsset
		expectedError bool
	}{
		"list blob services": {
			input: []*armstorage.ListContainerItem{
				{
					ID:   pointers.Ref("/subscriptions/<subid>/resourceGroups/<rgname>/providers/Microsoft.Storage/storageAccounts/<storageid>/blobContainers/blob1"),
					Name: pointers.Ref("blob1"),
					Type: pointers.Ref("microsoft.storage/storageaccounts/blobcontainers"),
					Properties: &armstorage.ContainerProperties{
						HasImmutabilityPolicy: pointers.Ref(false),
					},
				},
			},
			expected: []AzureAsset{
				{
					Id:   "/subscriptions/<subid>/resourceGroups/<rgname>/providers/Microsoft.Storage/storageAccounts/<storageid>/blobContainers/blob1",
					Name: "blob1",
					Properties: map[string]any{
						"hasImmutabilityPolicy": false,
					},
					Extension: map[string]any{
						"storageAccountId":   "<storageid>",
						"storageAccountName": "<storageid>",
					},
					Type: "microsoft.storage/storageaccounts/blobcontainers",
				},
			},
			expectedError: false,
		},
	}

	log := testhelper.NewLogger(t)
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			provider := &storageAccountProvider{
				log: log,
				client: &storageAccountAzureClientWrapper{
					AssetBlobContainers: func(_ context.Context, _ string, _ *arm.ClientOptions, _, _ string, _ *armstorage.BlobContainersClientListOptions) ([]armstorage.BlobContainersClientListResponse, error) {
						x := tc
						return []armstorage.BlobContainersClientListResponse{response(x.input)}, nil
					},
				},
				diagnosticSettingsCache: cycle.NewCache[[]AzureAsset](log),
			}

			got, err := provider.ListStorageAccountBlobContainers(t.Context(), []AzureAsset{
				{
					Type: "Storage Account",
					Id:   "<storageid>",
					Name: "<storageid>",
				},
			})
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.ElementsMatch(t, tc.expected, got)
		})
	}
}

>>>>>>> bf5dbb6e ([go] Bump Golang to v1.24.0 (#3279))
func TestListStorageAccountBlobServices(t *testing.T) {
	response := func(v []*armstorage.BlobServiceProperties) armstorage.BlobServicesClientListResponse {
		return armstorage.BlobServicesClientListResponse{
			BlobServiceItems: armstorage.BlobServiceItems{
				Value: v,
			},
		}
	}

	tests := map[string]struct {
		input         []*armstorage.BlobServiceProperties
		expected      []AzureAsset
		expectedError bool
	}{
		"list blob services": {
			input: []*armstorage.BlobServiceProperties{
				{
					ID:   pointers.Ref("/subscriptions/<subid>/resourceGroups/<rgname>/providers/Microsoft.Storage/storageAccounts/<storageid>/blobServices/blob1"),
					Name: pointers.Ref("blob1"),
					Type: pointers.Ref("microsoft.storage/storageaccounts/blobservices"),
					BlobServiceProperties: &armstorage.BlobServicePropertiesProperties{
						IsVersioningEnabled: pointers.Ref(true),
					},
				},
			},
			expected: []AzureAsset{
				{
					Id:   "/subscriptions/<subid>/resourceGroups/<rgname>/providers/Microsoft.Storage/storageAccounts/<storageid>/blobServices/blob1",
					Name: "blob1",
					Properties: map[string]any{
						"isVersioningEnabled": true,
					},
					Extension: map[string]any{
						"storageAccountId":   "<storageid>",
						"storageAccountName": "<storageid>",
					},
					Type: "microsoft.storage/storageaccounts/blobservices",
				},
			},
			expectedError: false,
		},
	}

	log := testhelper.NewLogger(t)
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			provider := &storageAccountProvider{
				log: log,
				client: &storageAccountAzureClientWrapper{
					AssetBlobServices: func(_ context.Context, _ string, _ *arm.ClientOptions, _, _ string, _ *armstorage.BlobServicesClientListOptions) ([]armstorage.BlobServicesClientListResponse, error) {
						x := tc
						return []armstorage.BlobServicesClientListResponse{response(x.input)}, nil
					},
				},
				diagnosticSettingsCache: cycle.NewCache[[]AzureAsset](log),
			}

			got, err := provider.ListStorageAccountBlobServices(t.Context(), []AzureAsset{
				{
					Type: "Storage Account",
					Id:   "<storageid>",
					Name: "<storageid>",
				},
			})
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.ElementsMatch(t, tc.expected, got)
		})
	}
}

<<<<<<< HEAD
=======
func TestListStorageAccountFileShares(t *testing.T) {
	response := func(v []*armstorage.FileShareItem) armstorage.FileSharesClientListResponse {
		return armstorage.FileSharesClientListResponse{
			FileShareItems: armstorage.FileShareItems{
				Value: v,
			},
		}
	}

	tests := map[string]struct {
		input         []*armstorage.FileShareItem
		expected      []AzureAsset
		expectedError bool
	}{
		"list file shares": {
			input: []*armstorage.FileShareItem{
				{
					ID:         pointers.Ref("/subscriptions/<subid>/resourceGroups/<rgname>/providers/Microsoft.Storage/storageAccounts/<storageid>/fileshares/fileshare1"),
					Name:       pointers.Ref("fileshare1"),
					Type:       pointers.Ref("microsoft.storage/storageaccounts/fileshares"),
					Properties: &armstorage.FileShareProperties{},
				},
			},
			expected: []AzureAsset{
				{
					Id:         "/subscriptions/<subid>/resourceGroups/<rgname>/providers/Microsoft.Storage/storageAccounts/<storageid>/fileshares/fileshare1",
					Name:       "fileshare1",
					Properties: map[string]any{},
					Extension: map[string]any{
						"storageAccountId":   "<storageid>",
						"storageAccountName": "<storageid>",
					},
					Type: "microsoft.storage/storageaccounts/fileshares",
				},
			},
			expectedError: false,
		},
	}

	log := testhelper.NewLogger(t)
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			provider := &storageAccountProvider{
				log: log,
				client: &storageAccountAzureClientWrapper{
					AssetFileShares: func(_ context.Context, _ string, _ *arm.ClientOptions, _, _ string, _ *armstorage.FileSharesClientListOptions) ([]armstorage.FileSharesClientListResponse, error) {
						x := tc
						return []armstorage.FileSharesClientListResponse{response(x.input)}, nil
					},
				},
				diagnosticSettingsCache: cycle.NewCache[[]AzureAsset](log),
			}

			got, err := provider.ListStorageAccountFileShares(t.Context(), []AzureAsset{
				{
					Type: "Storage Account",
					Id:   "<storageid>",
					Name: "<storageid>",
				},
			})
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.ElementsMatch(t, tc.expected, got)
		})
	}
}

func TestListStorageAccountTables(t *testing.T) {
	response := func(v []*armstorage.Table) armstorage.TableClientListResponse {
		return armstorage.TableClientListResponse{
			ListTableResource: armstorage.ListTableResource{Value: v},
		}
	}

	tests := map[string]struct {
		input         []*armstorage.Table
		expected      []AzureAsset
		expectedError bool
	}{
		"list tables": {
			input: []*armstorage.Table{
				{
					ID:              pointers.Ref("/subscriptions/<subid>/resourceGroups/<rgname>/providers/Microsoft.Storage/storageAccounts/<storageid>/tables/table1"),
					Name:            pointers.Ref("table1"),
					Type:            pointers.Ref("microsoft.storage/storageaccounts/tables"),
					TableProperties: &armstorage.TableProperties{},
				},
			},
			expected: []AzureAsset{
				{
					Id:         "/subscriptions/<subid>/resourceGroups/<rgname>/providers/Microsoft.Storage/storageAccounts/<storageid>/tables/table1",
					Name:       "table1",
					Properties: map[string]any{},
					Extension: map[string]any{
						"storageAccountId":   "<storageid>",
						"storageAccountName": "<storageid>",
					},
					Type: "microsoft.storage/storageaccounts/tables",
				},
			},
			expectedError: false,
		},
	}

	log := testhelper.NewLogger(t)
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			provider := &storageAccountProvider{
				log: log,
				client: &storageAccountAzureClientWrapper{
					AssetTables: func(_ context.Context, _ string, _ *arm.ClientOptions, _, _ string, _ *armstorage.TableClientListOptions) ([]armstorage.TableClientListResponse, error) {
						x := tc
						return []armstorage.TableClientListResponse{response(x.input)}, nil
					},
				},
				diagnosticSettingsCache: cycle.NewCache[[]AzureAsset](log),
			}

			got, err := provider.ListStorageAccountTables(t.Context(), []AzureAsset{
				{
					Type: "Storage Account",
					Id:   "<storageid>",
					Name: "<storageid>",
				},
			})
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.ElementsMatch(t, tc.expected, got)
		})
	}
}

func TestListStorageAccountFileServices(t *testing.T) {
	response := func(v []*armstorage.FileServiceProperties) armstorage.FileServicesClientListResponse {
		return armstorage.FileServicesClientListResponse{
			FileServiceItems: armstorage.FileServiceItems{
				Value: v,
			},
		}
	}

	tests := map[string]struct {
		input         []*armstorage.FileServiceProperties
		expected      []AzureAsset
		expectedError bool
	}{
		"list file services": {
			input: []*armstorage.FileServiceProperties{
				{
					ID:                    pointers.Ref("/subscriptions/<subid>/resourceGroups/<rgname>/providers/Microsoft.Storage/storageAccounts/<storageid>/fileServices/file1"),
					Name:                  pointers.Ref("file1"),
					Type:                  pointers.Ref("microsoft.storage/storageaccounts/fileservices"),
					FileServiceProperties: nil,
				},
			},
			expected: []AzureAsset{
				{
					Id:         "/subscriptions/<subid>/resourceGroups/<rgname>/providers/Microsoft.Storage/storageAccounts/<storageid>/fileServices/file1",
					Name:       "file1",
					Properties: nil,
					Extension: map[string]any{
						"storageAccountId":   "<storageid>",
						"storageAccountName": "<storageid>",
					},
					Type: "microsoft.storage/storageaccounts/fileservices",
				},
			},
			expectedError: false,
		},
	}

	log := testhelper.NewLogger(t)
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			provider := &storageAccountProvider{
				log: log,
				client: &storageAccountAzureClientWrapper{
					AssetFileServices: func(_ context.Context, _ string, _ *arm.ClientOptions, _, _ string, _ *armstorage.FileServicesClientListOptions) (armstorage.FileServicesClientListResponse, error) {
						x := tc
						return response(x.input), nil
					},
				},
				diagnosticSettingsCache: cycle.NewCache[[]AzureAsset](log),
			}

			got, err := provider.ListStorageAccountFileServices(t.Context(), []AzureAsset{
				{
					Type: "Storage Account",
					Id:   "<storageid>",
					Name: "<storageid>",
				},
			})
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.ElementsMatch(t, tc.expected, got)
		})
	}
}

>>>>>>> bf5dbb6e ([go] Bump Golang to v1.24.0 (#3279))
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
		expectedError            bool
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
			expectedError: false,
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
			expectedError: false,
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
			expectedError: false,
		},
	}

	for name, tc := range tests {
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

			got, err := provider.ListDiagnosticSettingsAssetTypes(t.Context(), cycle.Metadata{}, lo.Keys[string](tc.subscriptions))
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.ElementsMatch(t, tc.expected, got)
		})
	}
}
