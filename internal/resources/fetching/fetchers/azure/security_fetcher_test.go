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

	expectedPair := typePair{
		Type:    fetching.MonitoringIdentity,
		SubType: fetching.AzureSecurityContactsType,
	}

	tests := map[string]struct {
		mockSubscriptions   []governance.Subscription
		mockInventoryAssets []inventory.AzureAsset
		mockInventoryError  error
		cycleMetadata       cycle.Metadata
		expected            []fetching.ResourceInfo
		expectedMetaData    []fetching.ResourceMetadata
		expectedErr         bool
	}{
		"error": {
			mockSubscriptions:   []governance.Subscription{subscriptionOne},
			mockInventoryAssets: []inventory.AzureAsset{},
			mockInventoryError:  errors.New("mock error"),
			cycleMetadata:       newCycle(123),
			expected:            nil,
			expectedMetaData:    []fetching.ResourceMetadata{},
			expectedErr:         true,
		},

		"empty - batch should be returned": {
			mockSubscriptions:   []governance.Subscription{subscriptionOne},
			mockInventoryAssets: []inventory.AzureAsset{},
			mockInventoryError:  nil,
			cycleMetadata:       newCycle(123),
			expected: []fetching.ResourceInfo{
				{
					CycleMetadata: newCycle(123),
					Resource: &AzureBatchResource{
						typePair:     expectedPair,
						Subscription: subscriptionOne,
						Assets:       []inventory.AzureAsset{},
					},
				},
			},
			expectedMetaData: []fetching.ResourceMetadata{
				{
					ID:      "azure-security-contacts-sub1",
					Type:    "monitoring",
					SubType: fetching.AzureSecurityContactsType,
					Name:    "azure-security-contacts-sub1",
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
			mockInventoryAssets: []inventory.AzureAsset{
				{Id: "id1", Name: "name1", SubscriptionId: "sub1"},
				{Id: "id2", Name: "name2", SubscriptionId: "sub1"},
				{Id: "id3", Name: "name3", SubscriptionId: "sub2"},
			},
			mockInventoryError: nil,
			cycleMetadata:      newCycle(124),
			expected: []fetching.ResourceInfo{
				{
					CycleMetadata: newCycle(124),
					Resource: &AzureBatchResource{
						typePair:     expectedPair,
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
						typePair:     expectedPair,
						Subscription: subscriptionTwo,
						Assets: []inventory.AzureAsset{
							{Id: "id3", Name: "name3", SubscriptionId: "sub2"},
						},
					},
				},
			},
			expectedMetaData: []fetching.ResourceMetadata{
				{
					ID:      "azure-security-contacts-sub2",
					Type:    "monitoring",
					SubType: fetching.AzureSecurityContactsType,
					Name:    "azure-security-contacts-sub2",
					Region:  "global",
					CloudAccountMetadata: fetching.CloudAccountMetadata{
						AccountId:        "sub2",
						AccountName:      "subName2",
						OrganisationId:   "",
						OrganizationName: "",
					},
				},
				{
					ID:      "azure-security-contacts-sub1",
					Type:    "monitoring",
					SubType: fetching.AzureSecurityContactsType,
					Name:    "azure-security-contacts-sub1",
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
		tc := tc
		t.Run(name, func(t *testing.T) {
			// tear up
			m := azurelib.NewMockProviderAPI(t)

			subscriptions := map[string]governance.Subscription{}
			for _, s := range tc.mockSubscriptions {
				subscriptions[s.FullyQualifiedID] = s
			}

			m.EXPECT().GetSubscriptions(mock.Anything, mock.Anything).Return(subscriptions, nil).Once()

			perSub := lo.GroupBy(tc.mockInventoryAssets, func(item inventory.AzureAsset) string { return item.SubscriptionId })
			for _, sid := range tc.mockSubscriptions {
				l := perSub[sid.ShortID]
				if l == nil {
					l = []inventory.AzureAsset{}
				}
				m.EXPECT().ListSecurityContacts(mock.Anything, sid.ShortID).Return(l, tc.mockInventoryError).Once()
			}

			ch := make(chan fetching.ResourceInfo, 100)
			defer close(ch)

			fetcherUnderTest := NewAzureSecurityAssetFetcher(log, ch, m)

			// execute
			err := fetcherUnderTest.Fetch(context.Background(), tc.cycleMetadata)

			// verify
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			got := testhelper.CollectResources(ch)

			// sort to ensure assertion
			sortResourceInfoSlice(tc.expected)
			sortResourceInfoSlice(got)

			assert.Equal(t, tc.expected, got)

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
