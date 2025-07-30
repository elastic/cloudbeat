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

package azurefetcher

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/testutil"
	azurelib_inventory "github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

func TestStorageFetcher_Fetch(t *testing.T) {
	subscription := azurelib_inventory.AzureAsset{
		Name: "subscription_name",
	}
	storageAccount := azurelib_inventory.AzureAsset{
		Id:   "storage_account",
		Name: "storage_account",
	}
	azureBlobService := azurelib_inventory.AzureAsset{
		Id:          "blob_service",
		Name:        "blob_service",
		DisplayName: "blob_service",
	}
	azureQueueService := azurelib_inventory.AzureAsset{
		Id:          "queue_service",
		Name:        "queue_service",
		DisplayName: "queue_service",
	}
	azureQueue := azurelib_inventory.AzureAsset{
		Id:          "queue",
		Name:        "queue",
		DisplayName: "queue",
	}

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
<<<<<<< HEAD
=======
			inventory.AssetClassificationAzureStorageBlobContainer,
			azureBlobContainer.Id,
			azureBlobContainer.Name,
			inventory.WithRawAsset(azureBlobContainer),
			inventory.WithCloud(inventory.Cloud{
				AccountID:   "<tenant id>",
				Provider:    inventory.AzureCloudProvider,
				ServiceName: "Azure",
			}),
		),
		inventory.NewAssetEvent(
>>>>>>> 04b20493 ([Asset Inventory][Azure] Add missing `cloud.*` section information (#3470))
			inventory.AssetClassificationAzureStorageBlobService,
			[]string{azureBlobService.Id},
			azureBlobService.Name,
			inventory.WithRawAsset(azureBlobService),
<<<<<<< HEAD
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AzureCloudProvider,
				Service: &inventory.AssetCloudService{
					Name: "Azure",
				},
=======
			inventory.WithCloud(inventory.Cloud{
				AccountID:   "<tenant id>",
				Provider:    inventory.AzureCloudProvider,
				ServiceName: "Azure",
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureStorageFileService,
			azureFileService.Id,
			azureFileService.Name,
			inventory.WithRawAsset(azureFileService),
			inventory.WithCloud(inventory.Cloud{
				AccountID:   "<tenant id>",
				Provider:    inventory.AzureCloudProvider,
				ServiceName: "Azure",
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureStorageFileShare,
			azureFileShare.Id,
			azureFileShare.Name,
			inventory.WithRawAsset(azureFileShare),
			inventory.WithCloud(inventory.Cloud{
				AccountID:   "<tenant id>",
				Provider:    inventory.AzureCloudProvider,
				ServiceName: "Azure",
>>>>>>> 04b20493 ([Asset Inventory][Azure] Add missing `cloud.*` section information (#3470))
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureStorageQueueService,
			[]string{azureQueueService.Id},
			azureQueueService.Name,
			inventory.WithRawAsset(azureQueueService),
<<<<<<< HEAD
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AzureCloudProvider,
				Service: &inventory.AssetCloudService{
					Name: "Azure",
				},
=======
			inventory.WithCloud(inventory.Cloud{
				AccountID:   "<tenant id>",
				Provider:    inventory.AzureCloudProvider,
				ServiceName: "Azure",
>>>>>>> 04b20493 ([Asset Inventory][Azure] Add missing `cloud.*` section information (#3470))
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureStorageQueue,
			[]string{azureQueue.Id},
			azureQueue.Name,
			inventory.WithRawAsset(azureQueue),
<<<<<<< HEAD
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AzureCloudProvider,
				Service: &inventory.AssetCloudService{
					Name: "Azure",
				},
=======
			inventory.WithCloud(inventory.Cloud{
				AccountID:   "<tenant id>",
				Provider:    inventory.AzureCloudProvider,
				ServiceName: "Azure",
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureStorageTable,
			azureTable.Id,
			azureTable.Name,
			inventory.WithRawAsset(azureTable),
			inventory.WithCloud(inventory.Cloud{
				AccountID:   "<tenant id>",
				Provider:    inventory.AzureCloudProvider,
				ServiceName: "Azure",
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureStorageTableService,
			azureTableService.Id,
			azureTableService.Name,
			inventory.WithRawAsset(azureTableService),
			inventory.WithCloud(inventory.Cloud{
				AccountID:   "<tenant id>",
				Provider:    inventory.AzureCloudProvider,
				ServiceName: "Azure",
>>>>>>> 04b20493 ([Asset Inventory][Azure] Add missing `cloud.*` section information (#3470))
			}),
		),
	}

	// setup
	logger := clog.NewLogger("azurefetcher_test")
	provider := newMockStorageProvider(t)

	provider.EXPECT().ListSubscriptions(
		mock.Anything,
	).Return(
		[]azurelib_inventory.AzureAsset{subscription}, nil,
	)

	provider.EXPECT().ListStorageAccounts(
		mock.Anything, mock.Anything,
	).Return(
		[]azurelib_inventory.AzureAsset{storageAccount}, nil,
	)

	provider.EXPECT().ListStorageAccountBlobServices(
		mock.Anything, mock.Anything,
	).Return(
		[]azurelib_inventory.AzureAsset{azureBlobService}, nil,
	)

	provider.EXPECT().ListStorageAccountQueueServices(
		mock.Anything, mock.Anything,
	).Return(
		[]azurelib_inventory.AzureAsset{azureQueueService}, nil,
	)

	provider.EXPECT().ListStorageAccountQueues(
		mock.Anything, mock.Anything,
	).Return(
		[]azurelib_inventory.AzureAsset{azureQueue}, nil,
	)

	fetcher := newStorageFetcher(logger, provider)
	// test & compare
	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}
