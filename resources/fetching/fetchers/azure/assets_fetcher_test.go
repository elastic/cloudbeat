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

	"github.com/samber/lo"
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
	fetcher := AzureAssetsFetcher{
		log:        testhelper.NewLogger(s.T()),
		resourceCh: s.resourceCh,
		provider:   mockInventoryService,
	}

	mockAssets := []inventory.AzureAsset{
		{
			Id:             "id1",
			Name:           "name1",
			Location:       "location1",
			Properties:     map[string]interface{}{"key1": "value1"},
			ResourceGroup:  "rg1",
			SubscriptionId: "subId1",
			TenantId:       "tenantId1",
			Type:           inventory.VirtualMachineAssetType,
		},
		{
			Id:             "id2",
			Name:           "name2",
			Location:       "location2",
			Properties:     map[string]interface{}{"key2": "value2"},
			ResourceGroup:  "rg2",
			SubscriptionId: "subId2",
			TenantId:       "tenantId2",
			Type:           inventory.StorageAccountAssetType,
		},
	}

	mockInventoryService.On("ListAllAssetTypesByName", mock.MatchedBy(func(assets []string) bool {
		return true
	})).Return(
		mockAssets, nil,
	)

	err := fetcher.Fetch(ctx, fetching.CycleMetadata{})
	s.NoError(err)
	results := testhelper.CollectResources(s.resourceCh)

	s.Equal(len(AzureResourceTypes), len(results))

	lo.ForEach(results, func(r fetching.ResourceInfo, index int) {
		data := r.GetData()
		s.NotNil(data)
		resource := data.(inventory.AzureAsset)
		s.NotEmpty(resource)
		s.Equal(mockAssets[index].Id, resource.Id)
		s.Equal(mockAssets[index].Name, resource.Name)
		s.Equal(mockAssets[index].Location, resource.Location)
		s.Equal(mockAssets[index].Properties, resource.Properties)
		s.Equal(mockAssets[index].ResourceGroup, resource.ResourceGroup)
		s.Equal(mockAssets[index].SubscriptionId, resource.SubscriptionId)
		s.Equal(mockAssets[index].TenantId, resource.TenantId)
		s.Equal(mockAssets[index].Type, resource.Type)
		meta, err := r.GetMetadata()
		s.NoError(err)
		s.NotNil(meta)
		s.NoError(err)
		s.NotEmpty(meta)
		s.Equal(mockAssets[index].Id, meta.ID)
		s.Equal(AzureResourceTypes[mockAssets[index].Type], meta.Type)
		s.Equal("", meta.SubType)
		s.Equal(mockAssets[index].Name, meta.Name)
		s.Equal(mockAssets[index].Location, meta.Region)
	})
}
