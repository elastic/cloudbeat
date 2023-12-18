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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/exp/maps"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/cycle"
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
		RunAndReturn(func(ctx context.Context, assetGroup string, _ []string) ([]inventory.AzureAsset, error) {
			return mockAssetGroups[assetGroup], nil
		})
	mockProvider.EXPECT().GetSubscriptions(mock.Anything, mock.Anything).Return(
		map[string]governance.Subscription{
			"subId": {
				FullyQualifiedID: "subId",
				ShortID:          "subId",
				DisplayName:      "subName",
				ManagementGroup: governance.ManagementGroup{
					FullyQualifiedID: "mgId",
					DisplayName:      "mgName",
				},
			},
		}, nil,
	).Once()
	mockProvider.EXPECT().
		ListDiagnosticSettingsAssetTypes(mock.Anything, cycle.Metadata{}, []string{"subId"}).
		Return(nil, nil)

	results, err := s.fetch(mockProvider, totalMockAssets)
	s.Require().NoError(err)
	for index, result := range results {
		expected := flatMockAssets[index]
		s.Run(expected.Type, func() {
			s.Equal(expected, result.GetData())

			meta, err := result.GetMetadata()
			s.Require().NoError(err)

			pair := AzureAssetTypeToTypePair[expected.Type]
			s.Equal(fetching.ResourceMetadata{
				ID:      expected.Id,
				Type:    pair.Type,
				SubType: pair.SubType,
				Name:    expected.Name,
				Region:  expected.Location,
				CloudAccountMetadata: fetching.CloudAccountMetadata{
					AccountId:        "subId",
					AccountName:      "subName",
					OrganisationId:   "mgId",
					OrganizationName: "mgName",
				},
			}, meta)

			ecs, err := result.GetElasticCommonData()
			s.Require().NoError(err)
			s.Empty(ecs)
		})
	}
}

func (s *AzureAssetsFetcherTestSuite) TestFetcher_Fetch_Errors() {
	asset := inventory.AzureAsset{
		Id:             "id",
		Name:           "name",
		SubscriptionId: "sub-id",
		Type:           inventory.DiskAssetType,
	}

	mockProvider := azurelib.NewMockProviderAPI(s.T())
	mockProvider.EXPECT().
		ListAllAssetTypesByName(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]string")).
		RunAndReturn(func(_ context.Context, assetGroup string, _ []string) ([]inventory.AzureAsset, error) {
			if assetGroup == AzureAssetGroups[0] {
				return []inventory.AzureAsset{asset}, nil
			}
			return nil, errors.New("some list asset error")
		})
	mockProvider.EXPECT().GetSubscriptions(mock.Anything, mock.Anything).Return(nil, errors.New("some get subscription error")).Once()
	mockProvider.EXPECT().
		ListDiagnosticSettingsAssetTypes(mock.Anything, cycle.Metadata{}, []string{}).
		Return(nil, nil)

	results, err := s.fetch(mockProvider, 1)
	s.Require().ErrorContains(err, "some list asset error")

	resource := results[0]
	s.Equal(asset, resource.GetData())
	metadata, err := resource.GetMetadata()
	s.Require().NoError(err)
	s.Equal(fetching.ResourceMetadata{
		ID:      "id",
		Type:    fetching.CloudCompute,
		SubType: fetching.AzureDiskType,
		Name:    "name",
		Region:  "",
		CloudAccountMetadata: fetching.CloudAccountMetadata{
			AccountId: "sub-id",
		},
	}, metadata)
}

func (s *AzureAssetsFetcherTestSuite) fetch(provider *azurelib.MockProviderAPI, expectedLength int) ([]fetching.ResourceInfo, error) {
	fetcher := AzureAssetsFetcher{
		log:        testhelper.NewLogger(s.T()),
		resourceCh: s.resourceCh,
		provider:   provider,
	}
	err := fetcher.Fetch(context.Background(), cycle.Metadata{})
	results := testhelper.CollectResources(s.resourceCh)
	s.Require().Len(results, expectedLength)
	return results, err
}

func TestAddUsedForActivityLogsFlag(t *testing.T) {
	tests := map[string]struct {
		inputAssets       []inventory.AzureAsset
		inputDiagSettings []inventory.AzureAsset
		expected          []inventory.AzureAsset
	}{
		"no storage account asset": {
			inputAssets:       []inventory.AzureAsset{{Id: "id_1", Type: inventory.DiskAssetType}},
			inputDiagSettings: []inventory.AzureAsset{{Properties: map[string]any{}}},
			expected:          []inventory.AzureAsset{{Id: "id_1", Type: inventory.DiskAssetType}},
		},
		"storage account asset not used for activity log": {
			inputAssets:       []inventory.AzureAsset{{Id: "id_1", Type: inventory.StorageAccountAssetType}},
			inputDiagSettings: []inventory.AzureAsset{{Properties: map[string]any{}}},
			expected:          []inventory.AzureAsset{{Id: "id_1", Type: inventory.StorageAccountAssetType}},
		},
		"storage account asset used for activity log": {
			inputAssets:       []inventory.AzureAsset{{Id: "id_1", Type: inventory.StorageAccountAssetType}},
			inputDiagSettings: []inventory.AzureAsset{{Properties: map[string]any{"storageAccountId": "id_1"}}},
			expected:          []inventory.AzureAsset{{Id: "id_1", Type: inventory.StorageAccountAssetType, Extension: map[string]any{"usedForActivityLogs": true}}},
		},
		"multiple storage account asset, one used for activity log": {
			inputAssets: []inventory.AzureAsset{
				{Id: "id_1", Type: inventory.StorageAccountAssetType},
				{Id: "id_2", Type: inventory.StorageAccountAssetType},
				{Id: "id_3", Type: inventory.StorageAccountAssetType},
			},
			inputDiagSettings: []inventory.AzureAsset{{Properties: map[string]any{"storageAccountId": "id_2"}}},
			expected: []inventory.AzureAsset{
				{Id: "id_1", Type: inventory.StorageAccountAssetType},
				{Id: "id_2", Type: inventory.StorageAccountAssetType, Extension: map[string]any{"usedForActivityLogs": true}},
				{Id: "id_3", Type: inventory.StorageAccountAssetType},
			},
		},
		"multiple storage account asset, two used for activity log": {
			inputAssets: []inventory.AzureAsset{
				{Id: "id_1", Type: inventory.StorageAccountAssetType},
				{Id: "id_2", Type: inventory.StorageAccountAssetType},
				{Id: "id_3", Type: inventory.StorageAccountAssetType},
			},
			inputDiagSettings: []inventory.AzureAsset{
				{Properties: map[string]any{"storageAccountId": "id_2"}},
				{Properties: map[string]any{"storageAccountId": "id_3"}},
			},
			expected: []inventory.AzureAsset{
				{Id: "id_1", Type: inventory.StorageAccountAssetType},
				{Id: "id_2", Type: inventory.StorageAccountAssetType, Extension: map[string]any{"usedForActivityLogs": true}},
				{Id: "id_3", Type: inventory.StorageAccountAssetType, Extension: map[string]any{"usedForActivityLogs": true}},
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			addUsedForActivityLogsFlag(tc.inputAssets, tc.inputDiagSettings)
			assert.Equal(t, tc.expected, tc.inputAssets)
		})
	}
}
