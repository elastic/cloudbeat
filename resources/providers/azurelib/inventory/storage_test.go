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
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/stretchr/testify/require"
)

func newBlobServicesClientListResponse(items ...*armstorage.BlobServiceProperties) armstorage.BlobServicesClientListResponse {
	return armstorage.BlobServicesClientListResponse{
		BlobServiceItems: armstorage.BlobServiceItems{
			Value: items,
		},
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
