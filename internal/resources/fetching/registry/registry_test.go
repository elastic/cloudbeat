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

package registry

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	fetchers "github.com/elastic/cloudbeat/internal/resources/fetching/fetchers/aws"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
	"github.com/elastic/cloudbeat/internal/statushandler"
)

type registryTestSuite struct {
	suite.Suite

	fetchers   FetchersMap
	registry   Registry
	resourceCh chan fetching.ResourceInfo
	wg         *sync.WaitGroup
}

func TestRegistryTestSuite(t *testing.T) {
	s := new(registryTestSuite)

	suite.Run(t, s)
}

func (s *registryTestSuite) SetupTest() {
	s.fetchers = make(FetchersMap)
	s.registry = NewRegistry(testhelper.NewLogger(s.T()), WithFetchersMap(s.fetchers))
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
	s.wg = &sync.WaitGroup{}
}

func (s *registryTestSuite) registerFetcher(f fetching.Fetcher, key string, conditions ...fetching.Condition) {
	s.fetchers[key] = RegisteredFetcher{Fetcher: f, Condition: conditions}
}

func (s *registryTestSuite) TestKeys() {
	tests := []struct {
		key   string
		value int
	}{
		{
			"some_fetcher", 2,
		},
		{
			"other_fetcher", 4,
		},
		{
			"new_fetcher", 6,
		},
	}

	for i, test := range tests {
		f := newNumberFetcher(test.value, s.wg)
		s.registerFetcher(f, test.key)

		s.Len(s.registry.Keys(), i+1)
	}

	keys := s.registry.Keys()

	s.Contains(keys, "some_fetcher")
	s.Contains(keys, "other_fetcher")
	s.Contains(keys, "new_fetcher")
}

func (s *registryTestSuite) TestRunNotRegistered() {
	f := newNumberFetcher(1, s.wg)
	s.registerFetcher(f, "some-key")

	err := s.registry.Run(context.TODO(), "unknown", cycle.Metadata{})
	s.Require().Error(err)
}

func (s *registryTestSuite) TestRunRegistered() {
	f1 := newSyncNumberFetcher(1, s.resourceCh)
	s.registerFetcher(f1, "some-key-1")

	f2 := newSyncNumberFetcher(2, s.resourceCh)
	s.registerFetcher(f2, "some-key-2")

	f3 := newSyncNumberFetcher(3, s.resourceCh)
	s.registerFetcher(f3, "some-key-3")

	tests := []struct {
		key string
		res numberResource
	}{
		{
			"some-key-1", numberResource{Num: 1},
		},
		{
			"some-key-2", numberResource{Num: 2},
		},
		{
			"some-key-3", numberResource{Num: 3},
		},
	}

	for _, test := range tests {
		err := s.registry.Run(context.TODO(), test.key, cycle.Metadata{})
		results := testhelper.CollectResources(s.resourceCh)

		s.Require().NoError(err)
		s.Len(results, 1)
		s.Equal(test.res.Num, results[0].GetData())
	}
}

func (s *registryTestSuite) TestShouldRunNotRegistered() {
	f := newNumberFetcher(1, s.wg)
	s.registerFetcher(f, "some-key")

	res := s.registry.ShouldRun("unknown")
	s.False(res)
}

func (s *registryTestSuite) TestShouldRun() {
	conditionTrue := newBoolFetcherCondition(true, "always-fetcher-condition")
	conditionFalse := newBoolFetcherCondition(false, "never-fetcher-condition")

	tests := []struct {
		conditions []fetching.Condition
		expected   bool
	}{
		{
			[]fetching.Condition{}, true,
		},
		{
			[]fetching.Condition{conditionTrue}, true,
		},
		{
			[]fetching.Condition{conditionTrue, conditionTrue}, true,
		},
		{
			[]fetching.Condition{conditionTrue, conditionTrue, conditionFalse}, false,
		},
		{
			[]fetching.Condition{conditionFalse, conditionTrue, conditionTrue, conditionTrue, conditionTrue}, false,
		},
	}

	for _, test := range tests {
		f := newNumberFetcher(1, s.wg)
		s.registerFetcher(f, "some-key", test.conditions...)
		s.registry = NewRegistry(testhelper.NewLogger(s.T()), WithFetchersMap(s.fetchers))

		should := s.registry.ShouldRun("some-key")
		s.Equal(test.expected, should)
	}
}

type numberFetcher struct {
	num        int
	resourceCh chan fetching.ResourceInfo
	wg         *sync.WaitGroup
}
type syncNumberFetcher struct {
	num        int
	resourceCh chan fetching.ResourceInfo
}

func (f *syncNumberFetcher) Fetch(_ context.Context, cycleMetadata cycle.Metadata) error {
	f.resourceCh <- fetching.ResourceInfo{
		Resource:      numberResource{f.num},
		CycleMetadata: cycleMetadata,
	}

	return nil
}

func (f *syncNumberFetcher) Stop() {
}

func newSyncNumberFetcher(num int, ch chan fetching.ResourceInfo) fetching.Fetcher {
	return &syncNumberFetcher{num: num, resourceCh: ch}
}

type numberResource struct {
	Num int
}

func (res numberResource) GetData() any {
	return res.Num
}

func (res numberResource) GetIds() []string {
	return nil
}

func (res numberResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:      "",
		Type:    "number",
		SubType: "number",
		Name:    "number",
	}, nil
}

func (res numberResource) GetElasticCommonData() (map[string]any, error) {
	return nil, nil
}

func newNumberFetcher(num int, wg *sync.WaitGroup) fetching.Fetcher {
	return &numberFetcher{num: num, wg: wg}
}

func (f *numberFetcher) Fetch(_ context.Context, cycleMetadata cycle.Metadata) error {
	defer f.wg.Done()

	f.resourceCh <- fetching.ResourceInfo{
		Resource:      numberResource{f.num},
		CycleMetadata: cycleMetadata,
	}

	return nil
}

func (f *numberFetcher) Stop() {
}

type boolFetcherCondition struct {
	val  bool
	name string
}

func newBoolFetcherCondition(val bool, name string) fetching.Condition {
	return &boolFetcherCondition{val, name}
}

func (c *boolFetcherCondition) Condition() bool {
	return c.val
}

func (c *boolFetcherCondition) Name() string {
	return c.name
}

func Test_registry_Update(t *testing.T) {
	emptyFn := func(t *testing.T, r Registry) {
		assert.Empty(t, r.Keys())
		assert.False(t, r.ShouldRun("some-key"))
	}
	count := 0

	tp := t
	tests := []struct {
		name    string
		updater UpdaterFunc
		testFn  func(t *testing.T, r Registry)
	}{
		{
			name:    "check nil",
			updater: nil,
			testFn:  emptyFn,
		},
		{
			name: "error",
			updater: func(context.Context) (FetchersMap, error) {
				return nil, errors.New("some-error")
			},
			testFn: emptyFn,
		},
		{
			name: "success after fail",
			updater: func(context.Context) (FetchersMap, error) {
				switch count {
				case 0:
					count++
					return nil, errors.New("some-error")
				case 1:
					count++
					return FetchersMap{"fetcher": newMockFetcher(tp, nil, 1)}, nil
				default:
					panic("unexpected count")
				}
			},
			testFn: func(t *testing.T, r Registry) {
				emptyFn(t, r) // empty at beginning because of error
				r.Update(t.Context())
				assert.Len(t, r.Keys(), 1)
				require.NoError(t, r.Run(t.Context(), "fetcher", cycle.Metadata{}))
				assert.Panics(t, func() { r.Update(t.Context()) })
			},
		},
		{
			name: "fail after success",
			updater: func(context.Context) (FetchersMap, error) {
				switch count {
				case 0:
					count++
					return FetchersMap{"fetcher": newMockFetcher(tp, nil, 2)}, nil
				case 1:
					count++
					return nil, errors.New("some-error")
				default:
					panic("unexpected count")
				}
			},
			testFn: func(t *testing.T, r Registry) {
				assert.Len(t, r.Keys(), 1)
				require.NoError(t, r.Run(t.Context(), "fetcher", cycle.Metadata{}))

				r.Update(t.Context()) // update fails, registry remains as is
				assert.Len(t, r.Keys(), 1)
				require.NoError(t, r.Run(t.Context(), "fetcher", cycle.Metadata{}))

				assert.Panics(t, func() { r.Update(t.Context()) })
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp = t
			count = 0

			r := NewRegistry(testhelper.NewLogger(t), WithUpdater(tt.updater))
			defer r.Stop()
			r.Update(t.Context())
			require.NotNil(t, tt.testFn)
			tt.testFn(t, r)
		})
	}
}

func newMockFetcher(t *testing.T, err error, times int) RegisteredFetcher {
	m := fetching.NewMockFetcher(t)
	m.EXPECT().Stop().Once()
	m.EXPECT().Fetch(mock.Anything, mock.Anything).Return(err).Times(times)
	return RegisteredFetcher{Fetcher: m}
}

func Test_cleanTypeOf(t *testing.T) {
	tests := []struct {
		val  any
		want string
	}{
		{
			val:  nil,
			want: "<nil>",
		},
		{
			val:  fetching.MockFetcher{},
			want: "fetching.MockFetcher",
		},
		{
			val:  new(fetching.MockFetcher),
			want: "fetching.MockFetcher",
		},
		{
			val:  to.Ptr(to.Ptr(to.Ptr(to.Ptr(to.Ptr(to.Ptr(fetchers.NewLoggingFetcher(nil, nil, nil, nil, nil, statushandler.NOOP{}))))))),
			want: "fetchers.LoggingFetcher",
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T", tt.val), func(t *testing.T) {
			assert.Equalf(t, tt.want, cleanTypeOf(tt.val), "cleanTypeOf(%v)", tt.val)
		})
	}
}
