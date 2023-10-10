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
	"strconv"
	"testing"

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
	mockAssets := map[string][]inventory.AzureAsset{
		inventory.ActivityLogAlertAssetType: {
			{
				Id:               "id1",
				Name:             "name1",
				Location:         "location1",
				Properties:       map[string]interface{}{"key1": "value1"},
				ResourceGroup:    "rg1",
				SubscriptionId:   "subId1",
				SubscriptionName: "subName1",
				TenantId:         "tenantId1",
				Type:             inventory.ActivityLogAlertAssetType,
				Sku:              "",
			},
			{
				Id:               "id2",
				Name:             "name2",
				Location:         "location2",
				Properties:       map[string]interface{}{"key2": "value2"},
				ResourceGroup:    "rg2",
				SubscriptionId:   "subId1",
				SubscriptionName: "subName1",
				TenantId:         "tenantId2",
				Type:             inventory.ActivityLogAlertAssetType,
				Sku:              "",
			},
		},
		inventory.BastionAssetType: {
			{
				Id:               "id3",
				Name:             "name3",
				Location:         "location3",
				Properties:       map[string]interface{}{"key3": "value3"},
				ResourceGroup:    "rg3",
				SubscriptionId:   "subId1",
				SubscriptionName: "subName1",
				TenantId:         "tenantId3",
				Type:             inventory.BastionAssetType,
				Sku:              "",
			},
		},
	}

	mockInventoryService := inventory.NewMockServiceAPI(s.T())
	mockInventoryService.EXPECT().
		ListAllAssetTypesByName(mock.Anything, mock.AnythingOfType("[]string")).
		RunAndReturn(func(ctx context.Context, types []string) ([]inventory.AzureAsset, error) {
			s.ElementsMatch(maps.Keys(mockAssets), types)

			var result []inventory.AzureAsset
			for _, tpe := range types {
				result = append(result, mockAssets[tpe]...)
			}
			return result, nil
		}).Once()

	fetcher := AzureBatchAssetFetcher{
		log:        testhelper.NewLogger(s.T()),
		resourceCh: s.resourceCh,
		provider:   mockInventoryService,
	}
	err := fetcher.Fetch(context.Background(), fetching.CycleMetadata{})
	s.Require().NoError(err)
	results := testhelper.CollectResources(s.resourceCh)

	s.Len(results, len(mockAssets))

	for assetType, expectedAssets := range mockAssets {
		result := findResult(results, assetType)
		s.Require().NotNil(result)

		s.Run(assetType, func() {
			assets := result.GetData().([]inventory.AzureAsset)
			s.Equal(expectedAssets, assets)

			meta, err := result.GetMetadata()
			s.Require().NoError(err)

			pair := AzureBatchAssets[assetType]
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
						"id":   expectedAssets[0].SubscriptionId,
						"name": expectedAssets[0].SubscriptionName,
					},
				},
			}, ecs)
		})
	}
}

func (s *AzureBatchAssetFetcherTestSuite) TestFetcher_Fetch_Batches() {
	var mockAssets []inventory.AzureAsset
	for i, variableFields := range []struct {
		sub string
		tpe string
	}{
		{
			// 0
			sub: "1",
			tpe: inventory.ActivityLogAlertAssetType,
		},
		{
			// 1
			sub: "1",
			tpe: inventory.ActivityLogAlertAssetType,
		},
		{
			// 2
			sub: "2",
			tpe: inventory.ActivityLogAlertAssetType,
		},
		{
			// 3
			sub: "3",
			tpe: inventory.BastionAssetType,
		},
		{
			// 4
			sub: "1",
			tpe: inventory.BastionAssetType,
		},
		{
			// 5
			sub: "2",
			tpe: inventory.ActivityLogAlertAssetType,
		},
		{
			// 6
			sub: "3",
			tpe: inventory.BastionAssetType,
		},
		{
			// 7
			sub: "4",
			tpe: inventory.BastionAssetType,
		},
	} {
		id := strconv.Itoa(i)
		mockAssets = append(mockAssets, inventory.AzureAsset{
			Id:             "id" + id,
			Name:           "name" + id,
			Location:       "loc" + id,
			Properties:     map[string]any{"key" + id: "value" + id},
			ResourceGroup:  "rg" + id,
			SubscriptionId: variableFields.sub,
			TenantId:       "tenant",
			Type:           variableFields.tpe,
			Sku:            "sku" + id,
		})
	}

	mockInventoryService := inventory.NewMockServiceAPI(s.T())
	mockInventoryService.EXPECT().
		ListAllAssetTypesByName(mock.Anything, mock.AnythingOfType("[]string")).
		Return(mockAssets, nil)
	fetcher := AzureBatchAssetFetcher{
		log:        testhelper.NewLogger(s.T()),
		resourceCh: s.resourceCh,
		provider:   mockInventoryService,
	}

	err := fetcher.Fetch(context.Background(), fetching.CycleMetadata{})
	s.Require().NoError(err)
	results := testhelper.CollectResources(s.resourceCh)

	s.Len(results, 5)
	s.ElementsMatch([]fetching.ResourceInfo{
		{ // sub 1
			Resource: &AzureBatchResource{
				Type:    fetching.MonitoringIdentity,
				SubType: fetching.AzureActivityLogAlertType,
				Assets:  []inventory.AzureAsset{mockAssets[0], mockAssets[1]},
			},
			CycleMetadata: fetching.CycleMetadata{Sequence: 0},
		},
		{ // sub 2
			Resource: &AzureBatchResource{
				Type:    fetching.MonitoringIdentity,
				SubType: fetching.AzureActivityLogAlertType,
				Assets:  []inventory.AzureAsset{mockAssets[2], mockAssets[5]},
			},
			CycleMetadata: fetching.CycleMetadata{Sequence: 0},
		},
		{ // sub 1
			Resource: &AzureBatchResource{
				Type:    fetching.CloudDns,
				SubType: fetching.AzureBastionType,
				Assets:  []inventory.AzureAsset{mockAssets[4]},
			},
			CycleMetadata: fetching.CycleMetadata{Sequence: 0},
		},
		{ // sub 3
			Resource: &AzureBatchResource{
				Type:    fetching.CloudDns,
				SubType: fetching.AzureBastionType,
				Assets:  []inventory.AzureAsset{mockAssets[3], mockAssets[6]},
			},
			CycleMetadata: fetching.CycleMetadata{Sequence: 0},
		},
		{ // sub 4
			Resource: &AzureBatchResource{
				Type:    fetching.CloudDns,
				SubType: fetching.AzureBastionType,
				Assets:  []inventory.AzureAsset{mockAssets[7]},
			},
			CycleMetadata: fetching.CycleMetadata{Sequence: 0},
		},
	}, results)
}

func findResult(results []fetching.ResourceInfo, assetType string) *fetching.ResourceInfo {
	for _, result := range results {
		if result.GetData().([]inventory.AzureAsset)[0].Type == assetType {
			return &result
		}
	}
	return nil
}
