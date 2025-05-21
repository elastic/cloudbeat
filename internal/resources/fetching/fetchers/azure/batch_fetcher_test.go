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
	"maps"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/governance"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
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
	mockAssetGroups := make(map[string][]inventory.AzureAsset)
	var flatMockAssets []inventory.AzureAsset
	for _, assetGroup := range AzureBatchAssetGroups {
		var mockAssets []inventory.AzureAsset
		for assetType := range maps.Keys(AzureBatchAssets) {
			mockId := fmt.Sprintf("%s-%s", AzureBatchAssets[assetType].SubType, "subId1")
			mockAssets = append(mockAssets,
				inventory.AzureAsset{
					Id:             mockId,
					Name:           mockId,
					Location:       "location",
					Properties:     map[string]any{"key": "value"},
					ResourceGroup:  "rg",
					SubscriptionId: "subId1",
					TenantId:       "tenantId",
					Type:           assetType,
					Sku:            map[string]any{"key": "value"},
					Identity:       map[string]any{"key": "value"},
				},
			)
		}
		mockAssetGroups[assetGroup] = mockAssets
		flatMockAssets = append(flatMockAssets, mockAssets...)
	}

	mockProvider := azurelib.NewMockProviderAPI(s.T())
	mockProvider.EXPECT().
		ListAllAssetTypesByName(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]string")).
		RunAndReturn(func(_ context.Context, assetGroup string, _ []string) ([]inventory.AzureAsset, error) {
			return mockAssetGroups[assetGroup], nil
		})
	mockProvider.EXPECT().GetSubscriptions(mock.Anything, mock.Anything).Return(
		map[string]governance.Subscription{
			"subId1": {
				FullyQualifiedID: "subId1",
				ShortID:          "subId1",
				DisplayName:      "subName1",
				ManagementGroup: governance.ManagementGroup{
					FullyQualifiedID: "mgId1",
					DisplayName:      "mgName1",
				},
			},
		}, nil,
	).Once()

	fetcher := AzureBatchAssetFetcher{
		log:        testhelper.NewLogger(s.T()),
		resourceCh: s.resourceCh,
		provider:   mockProvider,
	}
	t := s.T()
	err := fetcher.Fetch(t.Context(), cycle.Metadata{})
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
				ID:      exNameAndId,
				Type:    pair.Type,
				SubType: pair.SubType,
				Name:    exNameAndId,
				Region:  "global",
				CloudAccountMetadata: fetching.CloudAccountMetadata{
					AccountId:        "subId1",
					AccountName:      "subName1",
					OrganisationId:   "mgId1",
					OrganizationName: "mgName1",
				},
			}, meta)

			ecs, err := result.GetElasticCommonData()
			s.Require().NoError(err)
			s.Empty(ecs)
		})
	}
}

func (s *AzureBatchAssetFetcherTestSuite) TestFetcher_Fetch_Subscriptions() {
	mockAssetGroups := make(map[string][]inventory.AzureAsset)

	subMap := make(map[string]governance.Subscription)
	for subId := 1; subId <= 4; subId++ {
		subIdStr := fmt.Sprintf("subId%d", subId)
		subMap[subIdStr] = governance.Subscription{
			FullyQualifiedID: subIdStr,
			ShortID:          subIdStr,
			DisplayName:      fmt.Sprintf("subName%d", subId),
			ManagementGroup: governance.ManagementGroup{
				FullyQualifiedID: fmt.Sprintf("mgId%d", subId),
				DisplayName:      fmt.Sprintf("mgName%d", subId),
			},
		}
	}

	for _, assetGroup := range AzureBatchAssetGroups {
		var mockAssets []inventory.AzureAsset
		for assetType := range maps.Keys(AzureBatchAssets) {
			for subKey := range maps.Keys(subMap) {
				mockId := fmt.Sprintf("%s-%s", AzureBatchAssets[assetType].SubType, subKey)
				mockAssets = append(mockAssets,
					inventory.AzureAsset{
						Id:             mockId,
						Name:           mockId,
						Location:       "location",
						Properties:     map[string]any{"key": "value"},
						ResourceGroup:  "rg",
						SubscriptionId: subKey,
						TenantId:       "tenantId",
						Type:           assetType,
						Sku:            map[string]any{"key": "value"},
						Identity:       map[string]any{"key": "value"},
					},
				)
			}
		}
		mockAssetGroups[assetGroup] = mockAssets
	}

	mockProvider := azurelib.NewMockProviderAPI(s.T())
	mockProvider.EXPECT().
		ListAllAssetTypesByName(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]string")).
		RunAndReturn(func(_ context.Context, assetGroup string, _ []string) ([]inventory.AzureAsset, error) {
			return mockAssetGroups[assetGroup], nil
		})
	mockProvider.EXPECT().GetSubscriptions(mock.Anything, mock.Anything).Return(subMap, nil).Once()

	fetcher := AzureBatchAssetFetcher{
		log:        testhelper.NewLogger(s.T()),
		resourceCh: s.resourceCh,
		provider:   mockProvider,
	}
	t := s.T()
	err := fetcher.Fetch(t.Context(), cycle.Metadata{})
	s.Require().NoError(err)
	results := testhelper.CollectResources(s.resourceCh)

	s.Len(results, len(AzureBatchAssets)*len(subMap))

	expectedSubs := lo.GroupBy(results, func(result fetching.ResourceInfo) string {
		return result.Resource.(*AzureBatchResource).Subscription.FullyQualifiedID
	})
	s.Len(expectedSubs, len(subMap))

	for subKey := range maps.Keys(expectedSubs) {
		typesPerSub := expectedSubs[subKey]
		s.Len(typesPerSub, len(AzureBatchAssets))
		for _, subTypeRes := range typesPerSub {
			assets := subTypeRes.Resource.(*AzureBatchResource).Assets
			s.Len(assets, len(AzureBatchAssetGroups))
			metadata, err := subTypeRes.Resource.GetMetadata()
			s.Require().NoError(err)

			expectedSub := subMap[subKey]
			s.Equal(fetching.ResourceMetadata{
				ID:      fmt.Sprintf("%s-%s", AzureBatchAssets[assets[0].Type].SubType, subKey),
				Type:    AzureBatchAssets[assets[0].Type].Type,
				SubType: AzureBatchAssets[assets[0].Type].SubType,
				Name:    fmt.Sprintf("%s-%s", AzureBatchAssets[assets[0].Type].SubType, subKey),
				Region:  "global",
				CloudAccountMetadata: fetching.CloudAccountMetadata{
					AccountId:        expectedSub.FullyQualifiedID,
					AccountName:      expectedSub.DisplayName,
					OrganisationId:   expectedSub.ManagementGroup.FullyQualifiedID,
					OrganizationName: expectedSub.ManagementGroup.DisplayName,
				},
			}, metadata)

			ecs, err := subTypeRes.Resource.GetElasticCommonData()
			s.Require().NoError(err)
			s.Empty(ecs)
		}
	}
}
