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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/testutil"
	azurelib_inventory "github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
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
	vm := azurelib_inventory.AzureAsset{
		Id:          "/vm",
		DisplayName: "<name3>",
		TenantId:    "<tenant id>",
		Properties: map[string]any{
			"extended": map[string]any{
				"instanceView": map[string]any{
					"computerName": "localhost",
				},
			},
			"hardwareProfile": map[string]any{
				"vmSize": "xlarge",
			},
		},
	}

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureAppService,
			appService.Id,
			appService.Name,
			inventory.WithRawAsset(appService),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				AccountID:   "<tenant id>",
				ServiceName: "Azure App Services",
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureDisk,
			disk.Id,
			disk.Name,
			inventory.WithRawAsset(disk),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				AccountID:   "<tenant id>",
				ServiceName: "Azure Storage",
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureVirtualMachine,
			vm.Id,
			vm.DisplayName,
			inventory.WithRawAsset(vm),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AzureCloudProvider,
				AccountID:   "<tenant id>",
				ServiceName: "Azure Virtual Machines",
				MachineType: "xlarge",
				InstanceID:  "/vm",
			}),
			inventory.WithHost(inventory.Host{
				ID:   vm.Id,
				Name: "localhost",
				Type: "xlarge",
			}),
		),
	}

	// setup
	logger := testhelper.NewLogger(t)
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
		mock.Anything, mock.Anything, []string{azurelib_inventory.VirtualMachineAssetType},
	).Return(
		[]azurelib_inventory.AzureAsset{vm}, nil,
	)

	provider.EXPECT().ListAllAssetTypesByName(
		mock.Anything, mock.Anything, mock.Anything,
	).Return(
		[]azurelib_inventory.AzureAsset{}, nil,
	)

	fetcher := newResourceGraphFetcher(logger, "<tenant id>", provider)
	// test & compare
	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}

func TestHelperMethod_UnpackVMProperties(t *testing.T) {
	cases := []struct {
		name                   string
		input                  string
		expectedComputerName   string
		expectedVmSize         string
		expectedErrorUnpacking bool
	}{
		{
			name:  "no overlap with vmProperties",
			input: `{"this is not a matching map": "nope"}`,
		},
		{
			name:                   "types do not match vmProperties but keys do",
			input:                  `{"extended": {"instanceView": "this is wrong"}}`,
			expectedErrorUnpacking: true,
		},
		{
			name:                 "missing data does not break the flow",
			input:                `{"extended": {"instanceView": {"computerName": "localhost"}}}`,
			expectedComputerName: "localhost",
			expectedVmSize:       "",
		},
		{
			name:                 "happy path",
			input:                `{"extended": {"instanceView": {"computerName": "localhost"}}, "hardwareProfile": {"vmSize": "xlarge"}}`,
			expectedComputerName: "localhost",
			expectedVmSize:       "xlarge",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := map[string]any{}
			err := json.Unmarshal([]byte(tc.input), &m)
			require.NoError(t, err)
			got := tryUnpackingVMProperties(m)

			if tc.expectedErrorUnpacking {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, tc.expectedComputerName, got.Extended.InstanceView.ComputerName)
				assert.Equal(t, tc.expectedVmSize, got.HardwareProfile.VmSize)
			}
		})
	}
}
