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

type AzureActivityLogsFetcherTestSuite struct {
	suite.Suite

	resourceCh chan fetching.ResourceInfo
}

func TestAzureActivityLogsFetcherTestSuite(t *testing.T) {
	s := new(AzureActivityLogsFetcherTestSuite)

	suite.Run(t, s)
}

func (s *AzureActivityLogsFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *AzureActivityLogsFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *AzureActivityLogsFetcherTestSuite) TestFetcher_Fetch() {
	ctx := context.Background()

	subId := "subId1"

	mockInventoryService := &inventory.MockServiceAPI{}
	mockAssets := []inventory.AzureAsset{
		{
			Id:             "id1",
			Name:           "name1",
			Location:       "location1",
			Properties:     map[string]interface{}{"key1": "value1"},
			ResourceGroup:  "rg1",
			SubscriptionId: subId,
			TenantId:       "tenantId1",
			Type:           inventory.ActivityLogAlertAssetType,
		},
		{
			Id:             "id2",
			Name:           "name2",
			Location:       "location2",
			Properties:     map[string]interface{}{"key2": "value2"},
			ResourceGroup:  "rg2",
			SubscriptionId: subId,
			TenantId:       "tenantId2",
			Type:           inventory.ActivityLogAlertAssetType,
		},
	}

	mockInventoryService.EXPECT().
		ListAllAssetTypesByName(mock.AnythingOfType("[]string")).
		Return(mockAssets, nil).Once()
	defer mockInventoryService.AssertExpectations(s.T())

	fetcher := AzureActivityLogsFetcher{
		log:        testhelper.NewLogger(s.T()),
		resourceCh: s.resourceCh,
		provider:   mockInventoryService,
	}
	err := fetcher.Fetch(ctx, fetching.CycleMetadata{})
	s.Require().NoError(err)
	results := testhelper.CollectResources(s.resourceCh)

	s.Require().Len(results, 1)

	activityLogs := results[0].GetData().(AzureActivityLogsAsset).ActivityLogs
	for index, r := range activityLogs {
		expected := mockAssets[index]
		s.Run(expected.Type, func() {
			s.Equal(expected, r)
		})
	}

	meta, err := results[0].GetMetadata()
	s.Require().NoError(err)
	exNameAndId := "azure-activity-log-alert-" + subId
	s.Equal(fetching.ResourceMetadata{
		ID:                  exNameAndId,
		Type:                AzureActivityLogsResourceTypes[mockAssets[0].Type],
		SubType:             "",
		Name:                exNameAndId,
		Region:              "",
		AwsAccountId:        "",
		AwsAccountAlias:     "",
		AwsOrganizationId:   "",
		AwsOrganizationName: "",
	}, meta)

}
