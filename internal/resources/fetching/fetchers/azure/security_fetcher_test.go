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
	"errors"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/governance"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestAzureSecurityAssetFetcher(t *testing.T) {
	log := testhelper.NewLogger(t)

	subscriptionOne := governance.Subscription{
		FullyQualifiedID: "sub1",
		ShortID:          "sub1",
		DisplayName:      "subName1",
		ManagementGroup: governance.ManagementGroup{
			FullyQualifiedID: "",
			DisplayName:      "",
		},
	}

	subscriptionTwo := governance.Subscription{
		FullyQualifiedID: "sub2",
		ShortID:          "sub2",
		DisplayName:      "subName2",
		ManagementGroup: governance.ManagementGroup{
			FullyQualifiedID: "",
			DisplayName:      "",
		},
	}

	expectedPairContactsType := typePair{
		Type:    fetching.MonitoringIdentity,
		SubType: fetching.AzureSecurityContactsType,
	}

	expectedPairAutoProvisioningSettingsType := typePair{
		Type:    fetching.MonitoringIdentity,
		SubType: fetching.AzureAutoProvisioningSettingsType,
	}

	type mockInventoryResponse struct {
		MockAssets []inventory.AzureAsset
		MockError  error
	}

	// mockInventoryAssets per subscription id
	type mockInventoryAssets struct {
		MockSecurityContacts                 map[string]mockInventoryResponse
		MockSecurityAutoProvisioningSettings map[string]mockInventoryResponse
	}

	tests := map[string]struct {
		mockSubscriptions   []governance.Subscription
		mockInventoryAssets mockInventoryAssets
		cycleMetadata       cycle.Metadata
		expected            []fetching.ResourceInfo
		expectedMetaData    []fetching.ResourceMetadata
		expectedErr         bool
	}{
		"error": {
			mockSubscriptions: []governance.Subscription{subscriptionOne},
			mockInventoryAssets: mockInventoryAssets{
				MockSecurityContacts: map[string]mockInventoryResponse{
					subscriptionOne.ShortID: {
						MockAssets: []inventory.AzureAsset{},
						MockError:  errors.New("mock error"),
					},
				},
				MockSecurityAutoProvisioningSettings: map[string]mockInventoryResponse{
					subscriptionOne.ShortID: {
						MockAssets: []inventory.AzureAsset{},
						MockError:  errors.New("mock error"),
					},
				},
			},
			cycleMetadata:    newCycle(123),
			expected:         nil,
			expectedMetaData: []fetching.ResourceMetadata{},
			expectedErr:      true,
		},

		"empty - batch should be returned": {
			mockSubscriptions: []governance.Subscription{subscriptionOne},
			mockInventoryAssets: mockInventoryAssets{
				MockSecurityContacts: map[string]mockInventoryResponse{
					"sub1": {
						MockAssets: []inventory.AzureAsset{},
						MockError:  nil,
					},
				},
				MockSecurityAutoProvisioningSettings: map[string]mockInventoryResponse{
					"sub1": {
						MockAssets: []inventory.AzureAsset{},
						MockError:  nil,
					},
				},
			},
			cycleMetadata: newCycle(123),
			expected: []fetching.ResourceInfo{
				{
					CycleMetadata: newCycle(123),
					Resource: &AzureBatchResource{
						typePair:     expectedPairContactsType,
						Subscription: subscriptionOne,
						Assets:       []inventory.AzureAsset{},
					},
				},
				{
					CycleMetadata: newCycle(123),
					Resource: &AzureBatchResource{
						typePair:     expectedPairAutoProvisioningSettingsType,
						Subscription: subscriptionOne,
						Assets:       []inventory.AzureAsset{},
					},
				},
			},
			expectedMetaData: []fetching.ResourceMetadata{
				{
					ID:      fetching.AzureSecurityContactsType + "-sub1",
					Type:    "monitoring",
					SubType: fetching.AzureSecurityContactsType,
					Name:    fetching.AzureSecurityContactsType + "-sub1",
					Region:  "global",
					CloudAccountMetadata: fetching.CloudAccountMetadata{
						AccountId:        "sub1",
						AccountName:      "subName1",
						OrganisationId:   "",
						OrganizationName: "",
					},
				},
				{
					ID:      fetching.AzureAutoProvisioningSettingsType + "-sub1",
					Type:    "monitoring",
					SubType: fetching.AzureAutoProvisioningSettingsType,
					Name:    fetching.AzureAutoProvisioningSettingsType + "-sub1",
					Region:  "global",
					CloudAccountMetadata: fetching.CloudAccountMetadata{
						AccountId:        "sub1",
						AccountName:      "subName1",
						OrganisationId:   "",
						OrganizationName: "",
					},
				},
			},
			expectedErr: false,
		},

		"2 subs": {
			mockSubscriptions: []governance.Subscription{subscriptionOne, subscriptionTwo},
			mockInventoryAssets: mockInventoryAssets{
				MockSecurityContacts: map[string]mockInventoryResponse{
					"sub1": {
						MockAssets: []inventory.AzureAsset{
							{Id: "id1", Name: "name1", SubscriptionId: "sub1"},
							{Id: "id2", Name: "name2", SubscriptionId: "sub1"},
						},
						MockError: nil,
					},
					"sub2": {
						MockAssets: []inventory.AzureAsset{
							{Id: "id3", Name: "name3", SubscriptionId: "sub2"},
						},
						MockError: nil,
					},
				},
				MockSecurityAutoProvisioningSettings: map[string]mockInventoryResponse{
					"sub1": {
						MockAssets: []inventory.AzureAsset{},
						MockError:  nil,
					},
					"sub2": {
						MockAssets: []inventory.AzureAsset{
							{Id: "id4", Name: "name4", SubscriptionId: "sub2"},
						},
						MockError: nil,
					},
				},
			},
			cycleMetadata: newCycle(124),
			expected: []fetching.ResourceInfo{
				{
					CycleMetadata: newCycle(124),
					Resource: &AzureBatchResource{
						typePair:     expectedPairContactsType,
						Subscription: subscriptionOne,
						Assets: []inventory.AzureAsset{
							{Id: "id1", Name: "name1", SubscriptionId: "sub1"},
							{Id: "id2", Name: "name2", SubscriptionId: "sub1"},
						},
					},
				},
				{
					CycleMetadata: newCycle(124),
					Resource: &AzureBatchResource{
						typePair:     expectedPairContactsType,
						Subscription: subscriptionTwo,
						Assets: []inventory.AzureAsset{
							{Id: "id3", Name: "name3", SubscriptionId: "sub2"},
						},
					},
				},
				{
					CycleMetadata: newCycle(124),
					Resource: &AzureBatchResource{
						typePair:     expectedPairAutoProvisioningSettingsType,
						Subscription: subscriptionOne,
						Assets:       []inventory.AzureAsset{},
					},
				},
				{
					CycleMetadata: newCycle(124),
					Resource: &AzureBatchResource{
						typePair:     expectedPairAutoProvisioningSettingsType,
						Subscription: subscriptionTwo,
						Assets: []inventory.AzureAsset{
							{Id: "id4", Name: "name4", SubscriptionId: "sub2"},
						},
					},
				},
			},
			expectedMetaData: []fetching.ResourceMetadata{
				{
					ID:      fetching.AzureSecurityContactsType + "-sub2",
					Type:    "monitoring",
					SubType: fetching.AzureSecurityContactsType,
					Name:    fetching.AzureSecurityContactsType + "-sub2",
					Region:  "global",
					CloudAccountMetadata: fetching.CloudAccountMetadata{
						AccountId:        "sub2",
						AccountName:      "subName2",
						OrganisationId:   "",
						OrganizationName: "",
					},
				},
				{
					ID:      fetching.AzureSecurityContactsType + "-sub1",
					Type:    "monitoring",
					SubType: fetching.AzureSecurityContactsType,
					Name:    fetching.AzureSecurityContactsType + "-sub1",
					Region:  "global",
					CloudAccountMetadata: fetching.CloudAccountMetadata{
						AccountId:        "sub1",
						AccountName:      "subName1",
						OrganisationId:   "",
						OrganizationName: "",
					},
				},
				{
					ID:      fetching.AzureAutoProvisioningSettingsType + "-sub2",
					Type:    "monitoring",
					SubType: fetching.AzureAutoProvisioningSettingsType,
					Name:    fetching.AzureAutoProvisioningSettingsType + "-sub2",
					Region:  "global",
					CloudAccountMetadata: fetching.CloudAccountMetadata{
						AccountId:        "sub2",
						AccountName:      "subName2",
						OrganisationId:   "",
						OrganizationName: "",
					},
				},
				{
					ID:      fetching.AzureAutoProvisioningSettingsType + "-sub1",
					Type:    "monitoring",
					SubType: fetching.AzureAutoProvisioningSettingsType,
					Name:    fetching.AzureAutoProvisioningSettingsType + "-sub1",
					Region:  "global",
					CloudAccountMetadata: fetching.CloudAccountMetadata{
						AccountId:        "sub1",
						AccountName:      "subName1",
						OrganisationId:   "",
						OrganizationName: "",
					},
				},
			},
			expectedErr: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// tear up
			m := azurelib.NewMockProviderAPI(t)

			subscriptions := map[string]governance.Subscription{}
			for _, s := range tc.mockSubscriptions {
				subscriptions[s.FullyQualifiedID] = s
			}

			m.EXPECT().GetSubscriptions(mock.Anything, mock.Anything).Return(subscriptions, nil).Once()

			for subID, mockResponse := range tc.mockInventoryAssets.MockSecurityContacts {
				m.EXPECT().ListSecurityContacts(mock.Anything, subID).Return(mockResponse.MockAssets, mockResponse.MockError).Once()
			}

			for subID, mockResponse := range tc.mockInventoryAssets.MockSecurityAutoProvisioningSettings {
				m.EXPECT().ListAutoProvisioningSettings(mock.Anything, subID).Return(mockResponse.MockAssets, mockResponse.MockError).Once()
			}

			ch := make(chan fetching.ResourceInfo, 100)
			defer close(ch)

			fetcherUnderTest := NewAzureSecurityAssetFetcher(log, ch, m)

			// execute
			err := fetcherUnderTest.Fetch(t.Context(), tc.cycleMetadata)

			// verify
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			got := testhelper.CollectResources(ch)

			// sort to ensure assertion
			sortResourceInfoSlice(t, tc.expected)
			sortResourceInfoSlice(t, got)

			assert.ElementsMatch(t, tc.expected, got, "ResourceInfo slice mismatch")

			for _, resourceInfo := range got {
				elm, err := resourceInfo.GetElasticCommonData()
				require.NoError(t, err)
				assert.Equal(t, map[string]any(nil), elm)
			}

			metadata := lo.Map(got, func(item fetching.ResourceInfo, _ int) fetching.ResourceMetadata {
				metadata, err := item.GetMetadata()
				require.NoError(t, err)
				return metadata
			})
			assert.ElementsMatch(t, tc.expectedMetaData, metadata)
		})
	}
}
