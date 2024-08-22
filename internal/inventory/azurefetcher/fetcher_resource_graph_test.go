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

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/testutil"
	azurelib_inventory "github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

func TestResourceGraphFetcher_Fetch(t *testing.T) {
	appService := azurelib_inventory.AzureAsset{
		Id:          "/subscriptions/<id>/resourceGroups/<name>/providers/Microsoft.Web/sites/<name2>",
		Name:        "<name2>",
		DisplayName: "<name2>",
		TenantId:    "<tenant id>",
	}
	disk := azurelib_inventory.AzureAsset{
		Id:          "/subscriptions/<id>/resourceGroups/<name>/providers/Microsoft.Compute/disks/<name2>",
		Name:        "<name2>",
		DisplayName: "<name2>",
		TenantId:    "<tenant id>",
	}

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureAppService,
			[]string{appService.Id},
			appService.Name,
			inventory.WithRawAsset(appService),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AzureCloudProvider,
				Account: inventory.AssetCloudAccount{
					Id: "<tenant id>",
				},
				Service: &inventory.AssetCloudService{
					Name: "Azure",
				},
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureDisk,
			[]string{disk.Id},
			disk.Name,
			inventory.WithRawAsset(disk),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AzureCloudProvider,
				Account: inventory.AssetCloudAccount{
					Id: "<tenant id>",
				},
				Service: &inventory.AssetCloudService{
					Name: "Azure",
				},
			}),
		),
	}

	// setup
	logger := logp.NewLogger("azurefetcher_test")
	provider := newMockResourceGraphProvider(t)

	provider.EXPECT().ListAllAssetTypesByName(
		mock.Anything, mock.Anything, []string{azurelib_inventory.WebsitesAssetType},
	).Return(
		[]azurelib_inventory.AzureAsset{appService}, nil,
	)

	provider.EXPECT().ListAllAssetTypesByName(
		mock.Anything, mock.Anything, []string{azurelib_inventory.DiskAssetType},
	).Return(
		[]azurelib_inventory.AzureAsset{disk}, nil,
	)

	provider.EXPECT().ListAllAssetTypesByName(
		mock.Anything, mock.Anything, mock.Anything,
	).Return(
		[]azurelib_inventory.AzureAsset{}, nil,
	)

	fetcher := newResourceGraphFetcher(logger, provider)
	// test & compare
	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}
