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

package fetchers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
)

type AzureAssetsFetcherTestSuite struct {
	suite.Suite

	resourceCh chan fetching.ResourceInfo
}

func TestAzureAssetsFetcherTestSuite(t *testing.T) {
	s := new(AzureAssetsFetcherTestSuite)

	suite.Run(t, s)
}

func (s *AzureAssetsFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *AzureAssetsFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *AzureAssetsFetcherTestSuite) TestFetcher_Fetch() {
	ctx := context.Background()

	mockInventoryService := &inventory.MockServiceAPI{}
	var mockAssets []inventory.AzureAsset
	for _, assetType := range []string{
		inventory.ClassicStorageAccountAssetType,
		inventory.ClassicVirtualMachineAssetType,
		inventory.DiskAssetType,
		inventory.DocumentDBDatabaseAccountAssetType,
		inventory.MySQLDBAssetType,
		inventory.NetworkWatchersAssetType,
		inventory.NetworkWatchersFlowLogAssetType,
		inventory.PostgreSQLDBAssetType,
		inventory.SQLServersAssetType,
		inventory.StorageAccountAssetType,
		inventory.VirtualMachineAssetType,
		inventory.WebsitesAssetType,
		inventory.VaultAssetType,
	} {
		mockAssets = append(mockAssets,
			inventory.AzureAsset{
				Id:               "id",
				Name:             "name",
				Location:         "location",
				Properties:       map[string]interface{}{"key": "value"},
				ResourceGroup:    "rg",
				SubscriptionId:   "subId",
				SubscriptionName: "subName",
				TenantId:         "tenantId",
				Type:             assetType,
				Sku:              "",
			},
		)
	}
	mockInventoryService.EXPECT().
		ListAllAssetTypesByName(mock.Anything, mock.AnythingOfType("[]string")).
		Return(mockAssets, nil).Once()
	defer mockInventoryService.AssertExpectations(s.T())

	fetcher := AzureAssetsFetcher{
		log:        testhelper.NewLogger(s.T()),
		resourceCh: s.resourceCh,
		provider:   mockInventoryService,
	}
	err := fetcher.Fetch(ctx, fetching.CycleMetadata{})
	s.Require().NoError(err)
	results := testhelper.CollectResources(s.resourceCh)

	s.Require().Len(results, len(AzureAssetTypeToTypePair))
	s.Require().Len(results, len(mockAssets))

	for index, result := range results {
		expected := mockAssets[index]
		s.Run(expected.Type, func() {
			s.Equal(expected, result.GetData())

			meta, err := result.GetMetadata()
			s.Require().NoError(err)
			pair := AzureAssetTypeToTypePair[expected.Type]
			s.Equal(fetching.ResourceMetadata{
				ID:                  expected.Id,
				Type:                pair.Type,
				SubType:             pair.SubType,
				Name:                expected.Name,
				Region:              expected.Location,
				AwsAccountId:        "",
				AwsAccountAlias:     "",
				AwsOrganizationId:   "",
				AwsOrganizationName: "",
			}, meta)

			ecs, err := result.GetElasticCommonData()
			s.Require().NoError(err)
			s.Equal(map[string]any{
				"cloud": map[string]any{
					"provider": "azure",
					"account": map[string]any{
						"id":   expected.SubscriptionId,
						"name": expected.SubscriptionName,
					},
				},
			}, ecs)
		})
	}
}
