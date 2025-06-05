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

package governance

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func Test_provider_GetSubscriptions(t *testing.T) {
	ctx := t.Context()
	err1 := errors.New("some error 1")
	err2 := errors.New("some error 2")
	assets := []inventory.AzureAsset{
		generateManagementGroupAsset(1),
		generateManagementGroupAsset(2),
		generateManagementGroupAsset(3),
		generateManagementGroupAsset(4),
		generateManagementGroupAsset(5),
		generateSubscriptionAsset(1, 1),
		generateSubscriptionAsset(2, 1),
		generateSubscriptionAsset(3, 4),
		generateSubscriptionAsset(4, 5),
		generateSubscriptionAsset(5, 4),
		generateSubscriptionAsset(6, 2),
	}
	rand.Shuffle(len(assets), func(i, j int) { // shuffle assets, as order shouldn't matter
		assets[i], assets[j] = assets[j], assets[i]
	})
	expectedSubscriptions := map[string]Subscription{
		"sub-id-1": generateSubscription(1, 1),
		"sub-id-2": generateSubscription(2, 1),
		"sub-id-3": generateSubscription(3, 4),
		"sub-id-4": generateSubscription(4, 5),
		"sub-id-5": generateSubscription(5, 4),
		"sub-id-6": generateSubscription(6, 2),
	}

	t.Run("no assets", func(t *testing.T) {
		p := NewProvider(testhelper.NewLogger(t), mockClient(t, nil, nil))
		subs, err := p.GetSubscriptions(ctx, cycle.Metadata{Sequence: 1})
		require.NoError(t, err)
		assert.Equal(t, map[string]Subscription{}, subs)

		subsSame, err := p.GetSubscriptions(ctx, cycle.Metadata{Sequence: 1})
		require.NoError(t, err)
		assert.Equal(t, subs, subsSame)
	})

	t.Run("error on first call", func(t *testing.T) {
		p := NewProvider(testhelper.NewLogger(t), mockClient(t, nil, err1))
		_, err := p.GetSubscriptions(ctx, cycle.Metadata{Sequence: 10})
		require.ErrorIs(t, err, err1)

		p.(*provider).client = mockClient(t, nil, err2)
		_, err = p.GetSubscriptions(ctx, cycle.Metadata{Sequence: 1})
		require.ErrorIs(t, err, err2)

		p.(*provider).client = mockClient(t, assets, nil)
		got, err := p.GetSubscriptions(ctx, cycle.Metadata{Sequence: 1})
		require.NoError(t, err)
		assert.Equal(t, expectedSubscriptions, got)
	})

	t.Run("error on later call", func(t *testing.T) {
		p := NewProvider(testhelper.NewLogger(t), mockClient(t, assets, nil))
		got, err := p.GetSubscriptions(ctx, cycle.Metadata{Sequence: 1})
		require.NoError(t, err)
		assert.Equal(t, expectedSubscriptions, got)

		p.(*provider).client = mockClient(t, nil, err1)
		got, err = p.GetSubscriptions(ctx, cycle.Metadata{Sequence: 10})
		require.NoError(t, err)
		assert.Equal(t, expectedSubscriptions, got)
	})

	t.Run("lock", func(t *testing.T) {
		testhelper.SkipLong(t)

		var firstRun atomic.Bool
		m := inventory.NewMockResourceGraphProviderAPI(t)
		m.EXPECT().
			ListAllAssetTypesByName(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(func(_ context.Context, _ string, _ []string) ([]inventory.AzureAsset, error) {
				if firstRun.CompareAndSwap(false, true) {
					time.Sleep(100 * time.Millisecond)
					return assets, nil
				}
				return nil, err1
			})
		p := NewProvider(testhelper.NewLogger(t), m)

		var wg sync.WaitGroup
		wg.Add(1)
		defer wg.Wait()
		go func() {
			defer wg.Done()

			got, err := p.GetSubscriptions(ctx, cycle.Metadata{Sequence: 1})
			assert.NoError(t, err)
			assert.Equal(t, expectedSubscriptions, got)
		}()

		got, err := p.GetSubscriptions(ctx, cycle.Metadata{Sequence: 1})
		require.NoError(t, err)
		assert.Equal(t, expectedSubscriptions, got)
	})
}

func mockClient(t *testing.T, assets []inventory.AzureAsset, err error) inventory.ResourceGraphProviderAPI {
	t.Helper()
	client := inventory.NewMockResourceGraphProviderAPI(t)
	client.EXPECT().
		ListAllAssetTypesByName(mock.Anything, mock.Anything, mock.Anything).
		Return(assets, err).
		Once()
	return client
}

func generateManagementGroup(id int) ManagementGroup {
	return ManagementGroup{
		FullyQualifiedID: fmtField("mg-id", id),
		DisplayName:      fmtField("mg-display-name", id),
	}
}

func generateSubscription(id int, parentId int) Subscription {
	return Subscription{
		FullyQualifiedID: fmtField("/subscriptions/sub-id", id),
		ShortID:          fmtField("sub-id", id),
		DisplayName:      fmtField("sub-display-name", id),
		ManagementGroup:  generateManagementGroup(parentId),
	}
}

func generateManagementGroupAsset(id int) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:          fmtField("mg-id", id),
		Name:        fmtField("mg-name", id),
		DisplayName: fmtField("mg-display-name", id),
		Location:    "location",
		Properties:  nil,
		TenantId:    "tenant-id",
		Type:        "microsoft.management/managementgroups",
	}
}

func generateSubscriptionAsset(id int, parentId int) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:       fmtField("/subscriptions/sub-id", id),
		Name:     fmtField("sub-display-name", id),
		Location: "location",
		Properties: map[string]any{
			"managementGroupAncestorsChain": []any{
				map[string]any{
					"displayName": fmtField("mg-display-name", parentId),
					"name":        fmtField("mg-name", parentId),
				},
			},
		},
		SubscriptionId: fmtField("sub-id", id),
		TenantId:       "tenant-id",
		Type:           "microsoft.resources/subscriptions",
	}
}

func fmtField(s string, i int) string {
	return fmt.Sprintf("%s-%d", s, i)
}
