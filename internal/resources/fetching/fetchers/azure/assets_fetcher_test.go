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
		for assetType := range maps.Keys(AzureAssetTypeToTypePair) {
			mockAssets = append(mockAssets,
				inventory.AzureAsset{
					Id:             "id",
					Name:           "name",
					Location:       "location",
					Properties:     map[string]any{"key": "value"},
					ResourceGroup:  "rg",
					SubscriptionId: "subId",
					TenantId:       "tenantId",
					Type:           assetType,
					Sku:            map[string]any{"key": "value"},
					Identity:       map[string]any{"key": "value"},
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
		RunAndReturn(func(_ context.Context, assetGroup string, _ []string) ([]inventory.AzureAsset, error) {
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
	).Twice()

	// since we have storage account asset those the storage account enricher will be called and so these two provider functions.
	mockProvider.EXPECT().
		ListDiagnosticSettingsAssetTypes(mock.Anything, cycle.Metadata{}, []string{"subId"}).
		Return(nil, nil).
		Once()
	storageAccounts := lo.Filter(flatMockAssets, func(item inventory.AzureAsset, _ int) bool {
		return item.Type == inventory.StorageAccountAssetType
	})
	storageAccountsSubscriptionsIds := lo.Uniq(lo.Map(storageAccounts, func(item inventory.AzureAsset, _ int) string {
		return item.SubscriptionId
	}))
	mockProvider.EXPECT().
		ListStorageAccountBlobServices(mock.Anything, storageAccounts).
		Return(nil, nil)
	mockProvider.EXPECT().
		ListStorageAccountsBlobDiagnosticSettings(mock.Anything, storageAccounts).
		Return(nil, nil)
	mockProvider.EXPECT().
		ListStorageAccountsTableDiagnosticSettings(mock.Anything, storageAccounts).
		Return(nil, nil)
	mockProvider.EXPECT().
		ListStorageAccountsQueueDiagnosticSettings(mock.Anything, storageAccounts).
		Return(nil, nil)
	mockProvider.EXPECT().
		ListStorageAccounts(mock.Anything, storageAccountsSubscriptionsIds).
		Return(nil, nil)

	vaults := lo.Filter(flatMockAssets, func(item inventory.AzureAsset, _ int) bool {
		return item.Type == inventory.VaultAssetType
	})
	for _, v := range vaults {
		mockProvider.EXPECT().
			ListKeyVaultKeys(mock.Anything, v).
			Return(nil, nil)
		mockProvider.EXPECT().
			ListKeyVaultSecrets(mock.Anything, v).
			Return(nil, nil)
		mockProvider.EXPECT().
			ListKeyVaultDiagnosticSettings(mock.Anything, v).
			Return(nil, nil)
	}

	// since we have app service asset we need to mock the enricher
	appServices := lo.Filter(flatMockAssets, func(item inventory.AzureAsset, _ int) bool {
		return item.Type == inventory.WebsitesAssetType
	})
	for _, as := range appServices {
		mockProvider.EXPECT().
			GetAppServiceAuthSettings(mock.Anything, as).
			Return(nil, nil)
		mockProvider.EXPECT().
			GetAppServiceSiteConfig(mock.Anything, as).
			Return(nil, nil)
	}

	// since we have sql server asset we need to mock the enricher
	mockProvider.EXPECT().
		ListSQLEncryptionProtector(mock.Anything, "subId", "rg", "name").
		Return(nil, nil)
	mockProvider.EXPECT().
		GetSQLBlobAuditingPolicies(mock.Anything, "subId", "rg", "name").
		Return(nil, nil)
	mockProvider.EXPECT().
		ListSQLTransparentDataEncryptions(mock.Anything, "subId", "rg", "name").
		Return(nil, nil)
	mockProvider.EXPECT().
		ListSQLAdvancedThreatProtectionSettings(mock.Anything, "subId", "rg", "name").
		Return(nil, nil)
	mockProvider.EXPECT().
		ListSQLFirewallRules(mock.Anything, "subId", "rg", "name").
		Return(nil, nil)

	// since we have postgresql asset we need to mock the enricher
	mockProvider.EXPECT().
		ListSinglePostgresConfigurations(mock.Anything, "subId", "rg", "name").
		Return(nil, nil)
	mockProvider.EXPECT().
		ListFlexiblePostgresConfigurations(mock.Anything, "subId", "rg", "name").
		Return(nil, nil)
	mockProvider.EXPECT().
		ListSinglePostgresFirewallRules(mock.Anything, "subId", "rg", "name").
		Return(nil, nil)
	mockProvider.EXPECT().
		ListFlexiblePostgresFirewallRules(mock.Anything, "subId", "rg", "name").
		Return(nil, nil)

	// since we have mysql  asset we need to mock the enricher
	mockProvider.EXPECT().
		GetFlexibleTLSVersionConfiguration(mock.Anything, "subId", "rg", "name").
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
			switch expected.Type {
			case inventory.VirtualMachineAssetType:
				{
					s.GreaterOrEqual(len(ecs), 1)
					s.Contains(ecs, "host.name")
				}
			case inventory.RoleDefinitionsType:
				{
					s.Len(ecs, 2)
					s.Contains(ecs, "user.effective.id")
					s.Contains(ecs, "user.effective.name")
				}
			default:
				s.Empty(ecs)
			}
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
	mockProvider.EXPECT().
		GetSubscriptions(mock.Anything, mock.Anything).
		Return(nil, errors.New("some get subscription error")).
		Twice()

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

func (s *AzureAssetsFetcherTestSuite) TestFetcher_Fetch_VM_Hostname_Empty() {
	mockAssetGroups := make(map[string][]inventory.AzureAsset)
	totalMockAssets := 0
	var flatMockAssets []inventory.AzureAsset
	for _, assetGroup := range AzureAssetGroups {
		var mockAssets []inventory.AzureAsset
		propertyVariations := [](map[string]any){
			map[string]any{"definitelyNotOsProfile": "nope"},
			map[string]any{"osProfile": map[string]any{"notComputerName": "nope"}},
			map[string]any{"osProfile": map[string]any{"computerName": 42}},
		}
		for _, properties := range propertyVariations {
			mockAssets = append(mockAssets,
				inventory.AzureAsset{
					Id:             "id",
					Name:           "name",
					Location:       "location",
					Properties:     properties,
					ResourceGroup:  "rg",
					SubscriptionId: "subId",
					TenantId:       "tenantId",
					Type:           inventory.VirtualMachineAssetType,
					Sku:            map[string]any{"key": "value"},
					Identity:       map[string]any{"key": "value"},
				},
			)
		}
		totalMockAssets += len(mockAssets)
		mockAssetGroups[assetGroup] = mockAssets
		flatMockAssets = append(flatMockAssets, mockAssets...)
	}

	mockProvider := azurelib.NewMockProviderAPI(s.T())
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
	).Twice()
	mockProvider.EXPECT().
		ListDiagnosticSettingsAssetTypes(mock.Anything, cycle.Metadata{}, []string{"subId"}).
		Return(nil, nil).
		Once()
	mockProvider.EXPECT().
		ListAllAssetTypesByName(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]string")).
		RunAndReturn(func(_ context.Context, assetGroup string, _ []string) ([]inventory.AzureAsset, error) {
			return mockAssetGroups[assetGroup], nil
		})

	results, err := s.fetch(mockProvider, totalMockAssets)
	s.Require().NoError(err)
	for index, result := range results {
		expected := flatMockAssets[index]
		s.Run(expected.Type, func() {
			s.Equal(expected, result.GetData())

			ecs, err := result.GetElasticCommonData()
			s.Require().NoError(err)
			s.Contains(ecs, "host.name")
			s.NotContains(ecs, "host.hostname", "ECS data should not contain host.hostname for incomplete properties")
		})
	}
}

func (s *AzureAssetsFetcherTestSuite) TestFetcher_Fetch_VM_Hostname_OK() {
	mockAssetGroups := make(map[string][]inventory.AzureAsset)
	totalMockAssets := 0
	var flatMockAssets []inventory.AzureAsset
	for _, assetGroup := range AzureAssetGroups {
		var mockAssets []inventory.AzureAsset
		mockAssets = append(mockAssets,
			inventory.AzureAsset{
				Id:             "id",
				Name:           "name",
				Location:       "location",
				Properties:     map[string]any{"osProfile": map[string]any{"computerName": "hostname"}},
				ResourceGroup:  "rg",
				SubscriptionId: "subId",
				TenantId:       "tenantId",
				Type:           inventory.VirtualMachineAssetType,
				Sku:            map[string]any{"key": "value"},
				Identity:       map[string]any{"key": "value"},
			},
		)
		totalMockAssets += len(mockAssets)
		mockAssetGroups[assetGroup] = mockAssets
		flatMockAssets = append(flatMockAssets, mockAssets...)
	}

	mockProvider := azurelib.NewMockProviderAPI(s.T())
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
	).Twice()
	mockProvider.EXPECT().
		ListDiagnosticSettingsAssetTypes(mock.Anything, cycle.Metadata{}, []string{"subId"}).
		Return(nil, nil).
		Once()
	mockProvider.EXPECT().
		ListAllAssetTypesByName(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]string")).
		RunAndReturn(func(_ context.Context, assetGroup string, _ []string) ([]inventory.AzureAsset, error) {
			return mockAssetGroups[assetGroup], nil
		})

	results, err := s.fetch(mockProvider, totalMockAssets)
	s.Require().NoError(err)
	for index, result := range results {
		expected := flatMockAssets[index]
		s.Run(expected.Type, func() {
			s.Equal(expected, result.GetData())

			ecs, err := result.GetElasticCommonData()
			s.Require().NoError(err)
			s.Contains(ecs, "host.name")
			s.Equal("name", ecs["host.name"])
			s.Contains(ecs, "host.hostname")
			s.Equal("hostname", ecs["host.hostname"])
		})
	}
}

func (s *AzureAssetsFetcherTestSuite) fetch(provider *azurelib.MockProviderAPI, expectedLength int) ([]fetching.ResourceInfo, error) {
	fetcher := AzureAssetsFetcher{
		log:        testhelper.NewLogger(s.T()),
		resourceCh: s.resourceCh,
		provider:   provider,
		enrichers:  initEnrichers(provider),
	}
	err := fetcher.Fetch(t.Context(), cycle.Metadata{})
	results := testhelper.CollectResources(s.resourceCh)
	s.Require().Len(results, expectedLength)
	return results, err
}
