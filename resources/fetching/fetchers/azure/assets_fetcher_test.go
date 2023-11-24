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
	"golang.org/x/exp/maps"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/governance"
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

	mockAssetGroups := make(map[string][]inventory.AzureAsset)
	totalMockAssets := 0
	var flatMockAssets []inventory.AzureAsset
	for _, assetGroup := range AzureAssetGroups {
		var mockAssets []inventory.AzureAsset
		for _, assetType := range maps.Keys(AzureAssetTypeToTypePair) {
			mockAssets = append(mockAssets,
				inventory.AzureAsset{
					Id:             "id",
					Name:           "name",
					Location:       "location",
					Properties:     map[string]interface{}{"key": "value"},
					ResourceGroup:  "rg",
					SubscriptionId: "subId",
					TenantId:       "tenantId",
					Type:           assetType,
					Sku:            "",
				},
			)
		}
		totalMockAssets += len(mockAssets)
		mockAssetGroups[assetGroup] = mockAssets
		flatMockAssets = append(flatMockAssets, mockAssets...)
	}

	mockProvider := azurelib.NewMockProviderAPI(s.T())
	mockProvider.EXPECT().
		ListAllAssetTypesByName(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]string")).
		RunAndReturn(func(ctx context.Context, assetGroup string, types []string) ([]inventory.AzureAsset, error) {
			return mockAssetGroups[assetGroup], nil
		})
	mockProvider.EXPECT().GetSubscriptions(mock.Anything, mock.Anything).Return(
		map[string]governance.Subscription{
			"subId1": {
				ID:          "subId1",
				DisplayName: "subName1",
				MG: governance.ManagementGroup{
					ID:          "mgId1",
					DisplayName: "mgName1",
				},
			},
		}, nil,
	).Once()

	fetcher := AzureAssetsFetcher{
		log:        testhelper.NewLogger(s.T()),
		resourceCh: s.resourceCh,
		provider:   mockProvider,
	}
	err := fetcher.Fetch(ctx, fetching.CycleMetadata{})
	s.Require().NoError(err)
	results := testhelper.CollectResources(s.resourceCh)

	s.Require().Len(results, totalMockAssets)

	for index, result := range results {
		expected := flatMockAssets[index]
		s.Run(expected.Type, func() {
			s.Equal(expected, result.GetData())

			meta, err := result.GetMetadata()
			s.Require().NoError(err)

			pair := AzureAssetTypeToTypePair[expected.Type]
			s.Equal(fetching.ResourceMetadata{
				ID:                   expected.Id,
				Type:                 pair.Type,
				SubType:              pair.SubType,
				Name:                 expected.Name,
				Region:               expected.Location,
				CloudAccountMetadata: fetching.CloudAccountMetadata{}, // TODO: test thse
			}, meta)

			ecs, err := result.GetElasticCommonData()
			s.Require().NoError(err)
			s.Equal(map[string]any{
				"cloud": map[string]any{
					"provider": "azure",
				},
			}, ecs)
		})
	}
}
