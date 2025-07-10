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

package preset

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/fetching/registry"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestNewCisAwsOrganizationFetchers_Leak(t *testing.T) {
	testhelper.SkipLong(t)

	t.Run("drain", func(t *testing.T) {
		subtest(t, true)
	})
	t.Run("no drain", func(t *testing.T) {
		subtest(t, false)
	})
}

func subtest(t *testing.T, drain bool) { //revive:disable-line:flag-parameter
	const (
		nAccounts           = 5
		nFetchers           = 33
		resourcesPerAccount = 111
	)

	var accounts []AwsAccount
	for i := 0; i < nAccounts; i++ {
		accounts = append(accounts, AwsAccount{
			Identity: cloud.Identity{
				Account:      fmt.Sprintf("account-%d", i),
				AccountAlias: fmt.Sprintf("alias-%d", i),
			},
		})
	}

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	ctx, cancel := context.WithCancel(t.Context())

	factory := mockFactory(nAccounts,
		func(_ context.Context, _ *clog.Logger, _ aws.Config, ch chan fetching.ResourceInfo, _ *cloud.Identity, _ ErrorPublisher) registry.FetchersMap {
			if drain {
				// create some resources if we are testing for that
				go func() {
					for i := 0; i < resourcesPerAccount; i++ {
						ch <- fetching.ResourceInfo{
							Resource:      mockResource(),
							CycleMetadata: cycle.Metadata{Sequence: int64(i)},
						}
					}
				}()
			}

			fm := registry.FetchersMap{}
			for i := 0; i < nFetchers; i++ {
				fm[fmt.Sprintf("fetcher-%d", i)] = registry.RegisteredFetcher{}
			}
			return fm
		},
	)

	mep := NewMockErrorPublisher(t)

	rootCh := make(chan fetching.ResourceInfo)
	fetcherMap := newCisAwsOrganizationFetchers(ctx, testhelper.NewLogger(t), rootCh, mep, accounts, nil, factory)
	assert.Lenf(t, fetcherMap, nAccounts, "Correct amount of maps")

	if drain {
		expectedResources := nAccounts * resourcesPerAccount
		resources := testhelper.CollectResourcesWithTimeout(rootCh, expectedResources, 1*time.Second)
		assert.Lenf(
			t,
			resources,
			expectedResources,
			"Correct amount of resources fetched",
		)
		defer func() {
			assert.Emptyf(
				t,
				testhelper.CollectResourcesWithTimeout(rootCh, 1, 100*time.Millisecond),
				"Channel not drained",
			)
		}()

		nameCounts := make(map[string]int)
		for _, resource := range resources {
			assert.NotNil(t, resource.GetData())
			cd, err := resource.GetElasticCommonData()
			require.NoError(t, err)
			assert.NotNil(t, cd)
			metadata, err := resource.GetMetadata()
			require.NotNil(t, metadata)
			require.NoError(t, err)
			assert.Equal(t, "some-region", metadata.Region)
			assert.NotEqual(t, "some-id", metadata.AccountId)
			assert.NotEqual(t, "some-alias", metadata.AccountName)
			nameCounts[metadata.AccountId]++
			nameCounts[metadata.AccountName]++
		}
		assert.Len(t, nameCounts, 2*nAccounts)
		for _, v := range nameCounts {
			assert.Equal(t, resourcesPerAccount, v)
		}
	}

	cancel()
}

func TestNewCisAwsOrganizationFetchers_LeakContextDone(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	ctx, cancel := context.WithCancel(t.Context())

	mep := NewMockErrorPublisher(t)

	newCisAwsOrganizationFetchers(
		ctx,
		testhelper.NewLogger(t),
		make(chan fetching.ResourceInfo),
		mep,
		[]AwsAccount{{
			Identity: cloud.Identity{
				Account:      "1",
				AccountAlias: "account",
			},
		}},
		nil,
		mockFactory(1,
			func(_ context.Context, _ *clog.Logger, _ aws.Config, ch chan fetching.ResourceInfo, _ *cloud.Identity, _ ErrorPublisher) registry.FetchersMap {
				ch <- fetching.ResourceInfo{
					Resource:      mockResource(),
					CycleMetadata: cycle.Metadata{Sequence: 1},
				}

				return registry.FetchersMap{"fetcher": registry.RegisteredFetcher{}}
			},
		),
	)

	cancel()
}

func TestNewCisAwsOrganizationFetchers_CloseChannel(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	mep := NewMockErrorPublisher(t)

	newCisAwsOrganizationFetchers(
		t.Context(),
		testhelper.NewLogger(t),
		make(chan fetching.ResourceInfo),
		mep,
		[]AwsAccount{{
			Identity: cloud.Identity{
				Account:      "1",
				AccountAlias: "account",
			},
		}},
		nil,
		mockFactory(1,
			func(_ context.Context, _ *clog.Logger, _ aws.Config, ch chan fetching.ResourceInfo, _ *cloud.Identity, _ ErrorPublisher) registry.FetchersMap {
				defer close(ch)
				return registry.FetchersMap{"fetcher": registry.RegisteredFetcher{}}
			},
		),
	)
}

func TestNewCisAwsOrganizationFetchers_Cache(t *testing.T) {
	cache := map[string]registry.FetchersMap{
		"1": {"fetcher": registry.RegisteredFetcher{}},
		"3": {"fetcher": registry.RegisteredFetcher{}},
	}

	mep := NewMockErrorPublisher(t)

	m := newCisAwsOrganizationFetchers(
		t.Context(),
		testhelper.NewLogger(t),
		make(chan fetching.ResourceInfo),
		mep,
		[]AwsAccount{
			{
				Identity: cloud.Identity{
					Account:      "1",
					AccountAlias: "account",
				},
			},
			{
				Identity: cloud.Identity{
					Account:      "2",
					AccountAlias: "account2",
				},
			},
		},
		cache,
		mockFactory(1,
			func(_ context.Context, _ *clog.Logger, _ aws.Config, _ chan fetching.ResourceInfo, identity *cloud.Identity, ep ErrorPublisher) registry.FetchersMap {
				assert.Equal(t, "2", identity.Account)
				return registry.FetchersMap{"fetcher": registry.RegisteredFetcher{}}
			},
		),
	)
	assert.Len(t, cache, 2)
	assert.Len(t, m, 2)
	assert.NotContains(t, cache, "3")
}

func mockResource() *fetching.MockResource {
	m := fetching.MockResource{}
	m.EXPECT().GetData().Return(struct{}{}).Once()
	m.EXPECT().GetMetadata().Return(fetching.ResourceMetadata{
		Region: "some-region",
		CloudAccountMetadata: fetching.CloudAccountMetadata{
			AccountId:   "some-id",
			AccountName: "some-alias",
		},
	}, nil).Once()
	m.EXPECT().GetElasticCommonData().Return(map[string]any{}, nil).Once()
	return &m
}

func mockFactory(times int, f awsFactory) awsFactory {
	factory := mockAwsFactory{}
	factory.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(f).Times(times)
	return factory.Execute
}
