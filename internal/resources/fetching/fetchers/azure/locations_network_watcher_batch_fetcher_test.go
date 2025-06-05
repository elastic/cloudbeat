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
	"fmt"
	"sort"
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

type either[T any] struct {
	left  T
	right error
}

func eitherValue[T any](val T) either[T] {
	return either[T]{left: val, right: nil}
}

func eitherError[T any](err error) either[T] {
	var zero T
	return either[T]{left: zero, right: err}
}

func TestAzureLocationsNetworkWatcherAssetBatchFetcher_Fetch(t *testing.T) {
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

	tests := map[string]struct {
		subscriptions    either[[]governance.Subscription]
		locations        map[string]either[[]inventory.AzureAsset]
		networkWatchers  either[[]inventory.AzureAsset]
		cycleMetadata    cycle.Metadata
		expected         []fetching.ResourceInfo
		expectedMetaData []fetching.ResourceMetadata
		expectedErr      bool
	}{
		"Error on fetching subscription": {
			subscriptions:    eitherError[[]governance.Subscription](errors.New("failed to fetch subscriptions")),
			networkWatchers:  eitherValue([]inventory.AzureAsset{}),
			locations:        nil,
			cycleMetadata:    newCycle(899),
			expected:         nil,
			expectedMetaData: nil,
			expectedErr:      true,
		},

		"Error on fetching network watchers": {
			subscriptions:    eitherValue([]governance.Subscription{subscriptionOne}),
			networkWatchers:  eitherError[[]inventory.AzureAsset](errors.New("failed to fetch ARG")),
			locations:        make(map[string]either[[]inventory.AzureAsset]),
			cycleMetadata:    newCycle(898),
			expected:         nil,
			expectedMetaData: nil,
			expectedErr:      true,
		},

		"Error on fetching locations for subscription": {
			subscriptions: eitherValue([]governance.Subscription{subscriptionOne}),
			networkWatchers: eitherValue([]inventory.AzureAsset{
				azAssetNetworkWatcher("nw-1", "sub1", "brazilsouth"),
			}),
			locations: map[string]either[[]inventory.AzureAsset]{
				"sub1": eitherError[[]inventory.AzureAsset](errors.New("failed to fetch subscription")),
			},
			cycleMetadata:    newCycle(898),
			expected:         nil,
			expectedMetaData: nil,
			expectedErr:      true,
		},

		"Error on fetching locations for one subscription but success on other": {
			subscriptions: eitherValue([]governance.Subscription{subscriptionOne, subscriptionTwo}),
			networkWatchers: eitherValue([]inventory.AzureAsset{
				azAssetNetworkWatcher("nw-1", "sub1", "brazilsouth"),
			}),
			locations: map[string]either[[]inventory.AzureAsset]{
				"sub1": eitherValue([]inventory.AzureAsset{
					azAssetLocation("sub1", "brazilsouth"),
					azAssetLocation("sub1", "switzerlandnorth"),
				}),
				"sub2": eitherError[[]inventory.AzureAsset](errors.New("failed to fetch subscription")),
			},
			cycleMetadata: newCycle(898),
			expected: []fetching.ResourceInfo{
				batchedNetworkWatcherByLocationResourceInfo(
					newCycle(898), subscriptionOne, "brazilsouth", azAssetNetworkWatcher("nw-1", "sub1", "brazilsouth"),
				),
				batchedNetworkWatcherByLocationResourceInfo(
					newCycle(898), subscriptionOne, "switzerlandnorth",
				),
			},
			expectedMetaData: []fetching.ResourceMetadata{
				batchedNetworkWatcherByLocationMetadata(subscriptionOne, "brazilsouth"),
				batchedNetworkWatcherByLocationMetadata(subscriptionOne, "switzerlandnorth"),
			},
			expectedErr: true,
		},

		"Fetch locations for one subscription": {
			subscriptions: eitherValue([]governance.Subscription{subscriptionOne}),
			networkWatchers: eitherValue([]inventory.AzureAsset{
				azAssetNetworkWatcher("nw-1", "sub1", "brazilsouth"),
				azAssetNetworkWatcher("nw-2", "sub2", "switzerlandnorth"), // different subscription
				azAssetNetworkWatcher("nw-3", "sub2", "singapore"),        // different subscription
			}),
			locations: map[string]either[[]inventory.AzureAsset]{
				"sub1": eitherValue([]inventory.AzureAsset{
					azAssetLocation("sub1", "brazilsouth"),
					azAssetLocation("sub1", "switzerlandnorth"),
					azAssetLocation("sub1", "singapore"),
				}),
			},
			cycleMetadata: newCycle(898),
			expected: []fetching.ResourceInfo{
				batchedNetworkWatcherByLocationResourceInfo(
					newCycle(898), subscriptionOne, "brazilsouth", azAssetNetworkWatcher("nw-1", "sub1", "brazilsouth"),
				),
				batchedNetworkWatcherByLocationResourceInfo(
					newCycle(898), subscriptionOne, "switzerlandnorth",
				),
				batchedNetworkWatcherByLocationResourceInfo(
					newCycle(898), subscriptionOne, "singapore",
				),
			},
			expectedMetaData: []fetching.ResourceMetadata{
				batchedNetworkWatcherByLocationMetadata(subscriptionOne, "brazilsouth"),
				batchedNetworkWatcherByLocationMetadata(subscriptionOne, "switzerlandnorth"),
				batchedNetworkWatcherByLocationMetadata(subscriptionOne, "singapore"),
			},
			expectedErr: false,
		},

		"Fetch a location with multiple watchers for one subscription": {
			subscriptions: eitherValue([]governance.Subscription{subscriptionOne}),
			networkWatchers: eitherValue([]inventory.AzureAsset{
				azAssetNetworkWatcher("nw-1", "sub1", "brazilsouth"),
				azAssetNetworkWatcher("nw-2", "sub1", "brazilsouth"),
				azAssetNetworkWatcher("nw-3", "sub2", "brazilsouth"), // different subscription
			}),
			locations: map[string]either[[]inventory.AzureAsset]{
				"sub1": eitherValue([]inventory.AzureAsset{
					azAssetLocation("sub1", "brazilsouth"),
					azAssetLocation("sub1", "switzerlandnorth"),
				}),
			},
			cycleMetadata: newCycle(877),
			expected: []fetching.ResourceInfo{
				batchedNetworkWatcherByLocationResourceInfo(
					newCycle(877), subscriptionOne, "brazilsouth", azAssetNetworkWatcher("nw-1", "sub1", "brazilsouth"), azAssetNetworkWatcher("nw-2", "sub1", "brazilsouth"),
				),
				batchedNetworkWatcherByLocationResourceInfo(
					newCycle(877), subscriptionOne, "switzerlandnorth",
				),
			},
			expectedMetaData: []fetching.ResourceMetadata{
				batchedNetworkWatcherByLocationMetadata(subscriptionOne, "brazilsouth"),
				batchedNetworkWatcherByLocationMetadata(subscriptionOne, "switzerlandnorth"),
			},
			expectedErr: false,
		},

		"Fetch locations for multiple subscriptions": {
			subscriptions: eitherValue([]governance.Subscription{subscriptionOne, subscriptionTwo}),
			networkWatchers: eitherValue([]inventory.AzureAsset{
				azAssetNetworkWatcher("nw-1", "sub1", "brazilsouth"),
				azAssetNetworkWatcher("nw-2", "sub2", "switzerlandnorth"),
				azAssetNetworkWatcher("nw-3", "sub2", "singapore"),
			}),
			locations: map[string]either[[]inventory.AzureAsset]{
				"sub1": eitherValue([]inventory.AzureAsset{
					azAssetLocation("sub1", "brazilsouth"),
					azAssetLocation("sub1", "switzerlandnorth"),
					azAssetLocation("sub1", "singapore"),
				}),
				"sub2": eitherValue([]inventory.AzureAsset{
					azAssetLocation("sub2", "sweden"),
					azAssetLocation("sub2", "switzerlandnorth"),
					azAssetLocation("sub2", "singapore"),
				}),
			},
			cycleMetadata: newCycle(866),
			expected: []fetching.ResourceInfo{
				batchedNetworkWatcherByLocationResourceInfo(
					newCycle(866), subscriptionOne, "brazilsouth", azAssetNetworkWatcher("nw-1", "sub1", "brazilsouth"),
				),
				batchedNetworkWatcherByLocationResourceInfo(
					newCycle(866), subscriptionOne, "switzerlandnorth",
				),
				batchedNetworkWatcherByLocationResourceInfo(
					newCycle(866), subscriptionOne, "singapore",
				),
				batchedNetworkWatcherByLocationResourceInfo(
					newCycle(866), subscriptionTwo, "sweden",
				),
				batchedNetworkWatcherByLocationResourceInfo(
					newCycle(866), subscriptionTwo, "switzerlandnorth", azAssetNetworkWatcher("nw-2", "sub2", "switzerlandnorth"),
				),
				batchedNetworkWatcherByLocationResourceInfo(
					newCycle(866), subscriptionTwo, "singapore", azAssetNetworkWatcher("nw-3", "sub2", "singapore"),
				),
			},
			expectedMetaData: []fetching.ResourceMetadata{
				batchedNetworkWatcherByLocationMetadata(subscriptionOne, "brazilsouth"),
				batchedNetworkWatcherByLocationMetadata(subscriptionOne, "switzerlandnorth"),
				batchedNetworkWatcherByLocationMetadata(subscriptionOne, "singapore"),
				batchedNetworkWatcherByLocationMetadata(subscriptionTwo, "sweden"),
				batchedNetworkWatcherByLocationMetadata(subscriptionTwo, "switzerlandnorth"),
				batchedNetworkWatcherByLocationMetadata(subscriptionTwo, "singapore"),
			},
			expectedErr: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			m := azurelib.NewMockProviderAPI(t)

			subscriptions := map[string]governance.Subscription{}
			for _, s := range tc.subscriptions.left {
				subscriptions[s.FullyQualifiedID] = s
			}

			m.EXPECT().GetSubscriptions(mock.Anything, mock.Anything).Return(subscriptions, tc.subscriptions.right).Once()
			for sub, locations := range tc.locations {
				m.EXPECT().ListLocations(mock.Anything, sub).Return(locations.left, locations.right).Once()
			}

			if len(tc.networkWatchers.left) > 0 || tc.networkWatchers.right != nil {
				m.EXPECT().ListAllAssetTypesByName(mock.Anything, inventory.AssetGroupResources, []string{inventory.NetworkWatchersAssetType}).Return(tc.networkWatchers.left, tc.networkWatchers.right).Once()
			}

			ch := make(chan fetching.ResourceInfo, 100)
			defer close(ch)
			f := NewAzureLocationsNetworkWatcherAssetBatchFetcher(log, ch, m)

			err := f.Fetch(t.Context(), tc.cycleMetadata)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			got := testhelper.CollectResources(ch)

			// sort to ensure assertion
			sortResourceNetworkWatcherByLocationResource(tc.expected)
			sortResourceNetworkWatcherByLocationResource(got)

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

func sortResourceNetworkWatcherByLocationResource(r []fetching.ResourceInfo) {
	for idx := range r {
		abr, _ := (&r[idx]).Resource.(*NetworkWatchersBatchedByLocationResource)
		sort.Slice(abr.NetworkWatchers, func(i, j int) bool { return abr.NetworkWatchers[i].Id > abr.NetworkWatchers[j].Id })
	}

	sort.Slice(r, func(i, j int) bool {
		mi, _ := (&r[i]).Resource.GetMetadata()
		mj, _ := (&r[j]).Resource.GetMetadata()
		return mi.ID > mj.ID
	})
}

func batchedNetworkWatcherByLocationResourceInfo(cycle cycle.Metadata, subscription governance.Subscription, location string, networkWatchers ...inventory.AzureAsset) fetching.ResourceInfo {
	return fetching.ResourceInfo{
		CycleMetadata: cycle,
		Resource: &NetworkWatchersBatchedByLocationResource{
			typePair: typePair{
				SubType: fetching.AzureNetworkWatchersType,
				Type:    fetching.MonitoringIdentity,
			},
			Subscription:    subscription,
			Location:        azAssetLocation(subscription.ShortID, location),
			NetworkWatchers: networkWatchers,
		},
	}
}

func batchedNetworkWatcherByLocationMetadata(subscription governance.Subscription, location string) fetching.ResourceMetadata {
	id := fmt.Sprintf("%s-%s-%s", fetching.AzureNetworkWatchersType, location, subscription.ShortID)
	return fetching.ResourceMetadata{
		ID:                   id,
		Name:                 id,
		Type:                 fetching.MonitoringIdentity,
		SubType:              fetching.AzureNetworkWatchersType,
		Region:               location,
		CloudAccountMetadata: subscription.GetCloudAccountMetadata(),
	}
}

func azAssetLocation(subId, location string) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:             location,
		Name:           location,
		Type:           "location",
		Location:       location,
		SubscriptionId: subId,
	}
}

func azAssetNetworkWatcher(id, subId, location string) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:             id,
		Name:           id,
		Type:           "microsoft.network/networkwatchers",
		Location:       location,
		SubscriptionId: subId,
	}
}
