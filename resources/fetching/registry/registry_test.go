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
	"github.com/elastic/cloudbeat/resources/fetching/factory"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"testing"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
	"sync"
)

type RegistryTestSuite struct {
	suite.Suite

	log        *logp.Logger
	fetchers   factory.FetchersMap
	registry   Registry
	resourceCh chan fetching.ResourceInfo
	wg         *sync.WaitGroup
}

func TestRegistryTestSuite(t *testing.T) {
	s := new(RegistryTestSuite)
	s.log = logp.NewLogger("cloudbeat_registry_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *RegistryTestSuite) SetupTest() {
	s.fetchers = make(factory.FetchersMap, 0)
	s.registry = NewRegistry(s.log, s.fetchers)
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
	s.wg = &sync.WaitGroup{}
}

func (s *RegistryTestSuite) TestKeys() {
	var tests = []struct {
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
		f := NewNumberFetcher(test.value, nil, s.wg)
		RegisterFetcher(s.fetchers, f, test.key, nil)

		s.Equal(i+1, len(s.registry.Keys()))
	}

	keys := s.registry.Keys()

	s.Contains(keys, "some_fetcher")
	s.Contains(keys, "other_fetcher")
	s.Contains(keys, "new_fetcher")
}

func (s *RegistryTestSuite) TestRunNotRegistered() {
	f := NewNumberFetcher(1, nil, s.wg)
	RegisterFetcher(s.fetchers, f, "some-key", nil)

	err := s.registry.Run(context.TODO(), "unknown", fetching.CycleMetadata{})
	s.Error(err)
}

func (s *RegistryTestSuite) TestRunRegistered() {
	f1 := NewSyncNumberFetcher(1, s.resourceCh)
	RegisterFetcher(s.fetchers, f1, "some-key-1", nil)

	f2 := NewSyncNumberFetcher(2, s.resourceCh)
	RegisterFetcher(s.fetchers, f2, "some-key-2", nil)

	f3 := NewSyncNumberFetcher(3, s.resourceCh)
	RegisterFetcher(s.fetchers, f3, "some-key-3", nil)

	var tests = []struct {
		key string
		res NumberResource
	}{
		{
			"some-key-1", NumberResource{Num: 1},
		},
		{
			"some-key-2", NumberResource{Num: 2},
		},
		{
			"some-key-3", NumberResource{Num: 3},
		},
	}

	for _, test := range tests {
		err := s.registry.Run(context.TODO(), test.key, fetching.CycleMetadata{})
		results := testhelper.CollectResources(s.resourceCh)

		s.NoError(err)
		s.Equal(1, len(results))
		s.Equal(test.res.Num, results[0].GetData())
	}
}

func (s *RegistryTestSuite) TestShouldRunNotRegistered() {
	f := NewNumberFetcher(1, nil, s.wg)
	RegisterFetcher(s.fetchers, f, "some-key", nil)

	res := s.registry.ShouldRun("unknown")
	s.False(res)
}

func (s *RegistryTestSuite) TestShouldRun() {
	conditionTrue := NewBoolFetcherCondition(true, "always-fetcher-condition")
	conditionFalse := NewBoolFetcherCondition(false, "never-fetcher-condition")

	var tests = []struct {
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
		f := NewNumberFetcher(1, nil, s.wg)
		RegisterFetcher(s.fetchers, f, "some-key", test.conditions)
		s.registry = NewRegistry(s.log, s.fetchers)

		should := s.registry.ShouldRun("some-key")
		s.Equal(test.expected, should)
	}
}

func RegisterFetcher(fMap factory.FetchersMap, f fetching.Fetcher, key string, condition []fetching.Condition) {
	fMap[key] = factory.RegisteredFetcher{Fetcher: f, Condition: condition}
}

type NumberFetcher struct {
	num        int
	stopCalled bool
	resourceCh chan fetching.ResourceInfo
	wg         *sync.WaitGroup
}
type syncNumberFetcher struct {
	num        int
	stopCalled bool
	resourceCh chan fetching.ResourceInfo
}

func (f *syncNumberFetcher) Fetch(_ context.Context, cMetadata fetching.CycleMetadata) error {
	f.resourceCh <- fetching.ResourceInfo{
		Resource:      NumberResource{f.num},
		CycleMetadata: cMetadata,
	}

	return nil
}

func (f *syncNumberFetcher) Stop() {
	f.stopCalled = true
}

func NewSyncNumberFetcher(num int, ch chan fetching.ResourceInfo) fetching.Fetcher {
	return &syncNumberFetcher{num, false, ch}
}

type NumberResource struct {
	Num int
}

func (res NumberResource) GetData() interface{} {
	return res.Num
}

func (res NumberResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:      "",
		Type:    "number",
		SubType: "number",
		Name:    "number",
	}, nil
}

func (res NumberResource) GetElasticCommonData() interface{} {
	return nil
}

func NewNumberFetcher(num int, ch chan fetching.ResourceInfo, wg *sync.WaitGroup) fetching.Fetcher {
	return &NumberFetcher{num, false, ch, wg}
}

func (f *NumberFetcher) Fetch(_ context.Context, cMetadata fetching.CycleMetadata) error {
	defer f.wg.Done()

	f.resourceCh <- fetching.ResourceInfo{
		Resource:      NumberResource{f.num},
		CycleMetadata: cMetadata,
	}

	return nil
}

func (f *NumberFetcher) Stop() {
	f.stopCalled = true
}

type boolFetcherCondition struct {
	val  bool
	name string
}

func NewBoolFetcherCondition(val bool, name string) fetching.Condition {
	return &boolFetcherCondition{val, name}
}

func (c *boolFetcherCondition) Condition() bool {
	return c.val
}

func (c *boolFetcherCondition) Name() string {
	return c.name
}
