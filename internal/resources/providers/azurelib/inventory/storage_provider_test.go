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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
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
