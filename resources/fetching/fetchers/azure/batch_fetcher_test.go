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

	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/exp/maps"

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
	mockInventoryService := &inventory.MockServiceAPI{}
	mockAssetGroups := make(map[string][]inventory.AzureAsset)
	totalMockAssets := 0
	var flatMockAssets []inventory.AzureAsset
	for _, assetGroup := range AzureBatchAssetGroups {
		var mockAssets []inventory.AzureAsset
		for _, assetType := range maps.Keys(AzureBatchAssets) {
			mockId := fmt.Sprintf("%s-%s", AzureBatchAssets[assetType].SubType, "subId1")
			mockAssets = append(mockAssets,
				inventory.AzureAsset{
					Id:               mockId,
					Name:             mockId,
					Location:         "location",
					Properties:       map[string]interface{}{"key": "value"},
					ResourceGroup:    "rg",
					SubscriptionId:   "subId1",
					SubscriptionName: "subName1",
					TenantId:         "tenantId",
					Type:             assetType,
					Sku:              "",
				},
			)
		}
		totalMockAssets += len(mockAssets)
		mockAssetGroups[assetGroup] = mockAssets
		flatMockAssets = append(flatMockAssets, mockAssets...)
	}

	mockInventoryService.EXPECT().
		ListAllAssetTypesByName(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]string")).
		RunAndReturn(func(ctx context.Context, assetGroup string, types []string) ([]inventory.AzureAsset, error) {
			return mockAssetGroups[assetGroup], nil
		})
	defer mockInventoryService.AssertExpectations(s.T())
	mockInventoryService.EXPECT().GetSubscriptions().Return(map[string]string{
		"subId1": "subName1",
	}).Once()

	fetcher := AzureBatchAssetFetcher{
		log:        testhelper.NewLogger(s.T()),
		resourceCh: s.resourceCh,
		provider:   mockInventoryService,
	}
	err := fetcher.Fetch(context.Background(), fetching.CycleMetadata{})
	s.Require().NoError(err)
	results := testhelper.CollectResources(s.resourceCh)

	s.Len(results, len(AzureBatchAssets))

	expectedAssetsGrouped := lo.GroupBy(flatMockAssets, func(asset inventory.AzureAsset) string {
		return asset.Id
	})

	for _, result := range results {
		expected := expectedAssetsGrouped[result.Resource.(*AzureBatchResource).Assets[0].Id]
		// All assets in the list have the same type and subtype
		s.Run(expected[0].Type, func() {
			s.Equal(expected, result.GetData())

			meta, err := result.GetMetadata()
			s.Require().NoError(err)

			pair := AzureBatchAssets[expected[0].Type]
			exNameAndId := fmt.Sprintf("%s-subId1", pair.SubType)
			s.Equal(fetching.ResourceMetadata{
				ID:                  exNameAndId,
				Type:                pair.Type,
				SubType:             pair.SubType,
				Name:                exNameAndId,
				Region:              "global",
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
						// All assets in the list have the same type and subtype
						"id":   expected[0].SubscriptionId,
						"name": expected[0].SubscriptionName,
					},
				},
			}, ecs)
		})
	}
}

func (s *AzureBatchAssetFetcherTestSuite) TestFetcher_Fetch_Subscriptions() {
	mockInventoryService := &inventory.MockServiceAPI{}
	mockAssetGroups := make(map[string][]inventory.AzureAsset)
	totalMockAssets := 0
	subMap := map[string]string{
		"subId1": "subName1",
		"subId2": "subName2",
		"subId3": "subName3",
		"subId4": "subName4",
	}
	for _, assetGroup := range AzureBatchAssetGroups {
		var mockAssets []inventory.AzureAsset
		for _, assetType := range maps.Keys(AzureBatchAssets) {
			for _, subKey := range maps.Keys(subMap) {
				mockId := fmt.Sprintf("%s-%s", AzureBatchAssets[assetType].SubType, subKey)
				mockAssets = append(mockAssets,
					inventory.AzureAsset{
						Id:               mockId,
						Name:             mockId,
						Location:         "location",
						Properties:       map[string]interface{}{"key": "value"},
						ResourceGroup:    "rg",
						SubscriptionId:   subKey,
						SubscriptionName: subMap[subKey],
						TenantId:         "tenantId",
						Type:             assetType,
						Sku:              "",
					},
				)
			}
		}
		totalMockAssets += len(mockAssets)
		mockAssetGroups[assetGroup] = mockAssets
	}

	mockInventoryService.EXPECT().
		ListAllAssetTypesByName(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]string")).
		RunAndReturn(func(ctx context.Context, assetGroup string, types []string) ([]inventory.AzureAsset, error) {
			return mockAssetGroups[assetGroup], nil
		})
	defer mockInventoryService.AssertExpectations(s.T())

	mockInventoryService.EXPECT().GetSubscriptions().Return(subMap).Once()

	fetcher := AzureBatchAssetFetcher{
		log:        testhelper.NewLogger(s.T()),
		resourceCh: s.resourceCh,
		provider:   mockInventoryService,
	}
	err := fetcher.Fetch(context.Background(), fetching.CycleMetadata{})
	s.Require().NoError(err)
	results := testhelper.CollectResources(s.resourceCh)

	s.Len(results, len(AzureBatchAssets)*len(subMap))

	expectedSubs := lo.GroupBy(results, func(result fetching.ResourceInfo) string {
		return result.Resource.(*AzureBatchResource).SubId
	})
	s.Len(expectedSubs, len(subMap))

	for _, subKey := range maps.Keys(expectedSubs) {
		typesPerSub := expectedSubs[subKey]
		s.Len(typesPerSub, len(AzureBatchAssets))
		for _, subTypeRes := range typesPerSub {
			assets := subTypeRes.Resource.(*AzureBatchResource).Assets
			s.Len(assets, len(AzureBatchAssetGroups))
			md, err := subTypeRes.Resource.GetMetadata()
			s.Require().NoError(err)
			s.Equal(md.ID, fmt.Sprintf("%s-%s", AzureBatchAssets[assets[0].Type].SubType, subKey))
			s.Equal(md.Name, fmt.Sprintf("%s-%s", AzureBatchAssets[assets[0].Type].SubType, subKey))
			s.Equal(md.SubType, AzureBatchAssets[assets[0].Type].SubType)
			s.Equal(md.Type, AzureBatchAssets[assets[0].Type].Type)
			s.Equal("global", md.Region)

			ecs, err := subTypeRes.Resource.GetElasticCommonData()
			s.Require().NoError(err)
			s.Equal(map[string]any{
				"cloud": map[string]any{
					"provider": "azure",
					"account": map[string]any{
						"id":   subKey,
						"name": subMap[subKey],
					},
				},
			}, ecs)
		}
	}
}
