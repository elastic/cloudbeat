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
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
)

type AzureBatchAssetFetcherTestSuite struct {
	suite.Suite

	resourceCh chan fetching.ResourceInfo
}

func TestAzureBatchAssetFetcherTestSuite(t *testing.T) {
	s := new(AzureBatchAssetFetcherTestSuite)

	suite.Run(t, s)
}

func (s *AzureBatchAssetFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *AzureBatchAssetFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *AzureBatchAssetFetcherTestSuite) TestFetcher_Fetch() {
	ctx := context.Background()

	mockInventoryService := &inventory.MockServiceAPI{}
	mockAssets := map[string][]inventory.AzureAsset{
		inventory.ActivityLogAlertAssetType: {
			{
				Id:             "id1",
				Name:           "name1",
				Location:       "location1",
				Properties:     map[string]interface{}{"key1": "value1"},
				ResourceGroup:  "rg1",
				SubscriptionId: "subId1",
				TenantId:       "tenantId1",
				Type:           inventory.ActivityLogAlertAssetType,
			},
			{
				Id:             "id2",
				Name:           "name2",
				Location:       "location2",
				Properties:     map[string]interface{}{"key2": "value2"},
				ResourceGroup:  "rg2",
				SubscriptionId: "subId1",
				TenantId:       "tenantId2",
				Type:           inventory.ActivityLogAlertAssetType,
			},
		},
		inventory.BastionAssetType: {
			{
				Id:             "id3",
				Name:           "name3",
				Location:       "location3",
				Properties:     map[string]interface{}{"key3": "value3"},
				ResourceGroup:  "rg3",
				SubscriptionId: "subId1",
				TenantId:       "tenantId3",
				Type:           inventory.BastionAssetType,
			},
		},
	}

	mockInventoryService.EXPECT().
		ListAllAssetTypesByName(mock.AnythingOfType("[]string")).
		RunAndReturn(func(types []string) ([]inventory.AzureAsset, error) {
			s.Require().Len(types, 1)
			mockAssetsList, ok := mockAssets[types[0]]
			s.Require().True(ok)
			return mockAssetsList, nil
		}).Times(len(mockAssets))
	defer mockInventoryService.AssertExpectations(s.T())

	fetcher := AzureBatchAssetFetcher{
		log:        testhelper.NewLogger(s.T()),
		resourceCh: s.resourceCh,
		provider:   mockInventoryService,
	}
	err := fetcher.Fetch(ctx, fetching.CycleMetadata{})
	s.Require().NoError(err)
	results := testhelper.CollectResources(s.resourceCh)

	s.Require().Len(results, len(mockAssets))

	idx := 0
	for assetType, expectedAssets := range mockAssets {
		s.Run(assetType, func() {
			assets := results[idx].GetData().([]inventory.AzureAsset)
			s.Equal(expectedAssets, assets)

			meta, err := results[idx].GetMetadata()
			s.Require().NoError(err)

			resourceType := AzureBatchAssets[assetType]
			exNameAndId := fmt.Sprintf("%s-subId1", resourceType)
			s.Equal(fetching.ResourceMetadata{
				ID:                  exNameAndId,
				Type:                resourceType,
				SubType:             "",
				Name:                exNameAndId,
				Region:              "",
				AwsAccountId:        "",
				AwsAccountAlias:     "",
				AwsOrganizationId:   "",
				AwsOrganizationName: "",
			}, meta)
		})
		idx++
	}
}
