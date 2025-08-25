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
	azureBlobContainer := azurelib_inventory.AzureAsset{
		Id:          "blob_container",
		Name:        "blob_container",
		DisplayName: "blob_container",
	}
	azureBlobService := azurelib_inventory.AzureAsset{
		Id:          "blob_service",
		Name:        "blob_service",
		DisplayName: "blob_service",
	}
	azureFileService := azurelib_inventory.AzureAsset{
		Id:          "file_service",
		Name:        "file_service",
		DisplayName: "file_service",
	}
	azureFileShare := azurelib_inventory.AzureAsset{
		Id:          "file_share",
		Name:        "file_share",
		DisplayName: "file_share",
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
	azureTable := azurelib_inventory.AzureAsset{
		Id:          "table",
		Name:        "table",
		DisplayName: "table",
	}
	azureTableService := azurelib_inventory.AzureAsset{
		Id:          "table_service",
		Name:        "table_service",
		DisplayName: "table_service",
	}

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureStorageBlobContainer,
			azureBlobContainer.Id,
			azureBlobContainer.Name,
			inventory.WithRawAsset(azureBlobContainer),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				ServiceName: "Azure Storage",
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureStorageBlobService,
			azureBlobService.Id,
			azureBlobService.Name,
			inventory.WithRawAsset(azureBlobService),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				ServiceName: "Azure Storage",
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureStorageFileService,
			azureFileService.Id,
			azureFileService.Name,
			inventory.WithRawAsset(azureFileService),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				ServiceName: "Azure Storage",
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureStorageFileShare,
			azureFileShare.Id,
			azureFileShare.Name,
			inventory.WithRawAsset(azureFileShare),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				ServiceName: "Azure Storage",
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureStorageQueueService,
			azureQueueService.Id,
			azureQueueService.Name,
			inventory.WithRawAsset(azureQueueService),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				ServiceName: "Azure Storage",
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureStorageQueue,
			azureQueue.Id,
			azureQueue.Name,
			inventory.WithRawAsset(azureQueue),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				ServiceName: "Azure Storage",
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureStorageTable,
			azureTable.Id,
			azureTable.Name,
			inventory.WithRawAsset(azureTable),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				ServiceName: "Azure Storage",
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureStorageTableService,
			azureTableService.Id,
			azureTableService.Name,
			inventory.WithRawAsset(azureTableService),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				ServiceName: "Azure Storage",
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

	provider.EXPECT().ListStorageAccountBlobContainers(
		mock.Anything, mock.Anything,
	).Return(
		[]azurelib_inventory.AzureAsset{azureBlobContainer}, nil,
	)

	provider.EXPECT().ListStorageAccountBlobServices(
		mock.Anything, mock.Anything,
	).Return(
		[]azurelib_inventory.AzureAsset{azureBlobService}, nil,
	)

	provider.EXPECT().ListStorageAccountFileServices(
		mock.Anything, mock.Anything,
	).Return(
		[]azurelib_inventory.AzureAsset{azureFileService}, nil,
	)

	provider.EXPECT().ListStorageAccountFileShares(
		mock.Anything, mock.Anything,
	).Return(
		[]azurelib_inventory.AzureAsset{azureFileShare}, nil,
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

	provider.EXPECT().ListStorageAccountTables(
		mock.Anything, mock.Anything,
	).Return(
		[]azurelib_inventory.AzureAsset{azureTable}, nil,
	)

	provider.EXPECT().ListStorageAccountTableServices(
		mock.Anything, mock.Anything,
	).Return(
		[]azurelib_inventory.AzureAsset{azureTableService}, nil,
	)

	fetcher := newStorageFetcher(logger, "<tenant id>", provider)
	// test & compare
	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}
