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
	"github.com/elastic/cloudbeat/resources/fetchersManager"
	"github.com/elastic/cloudbeat/resources/fetchersManager/factory"
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
	registry   FetchersRegistry
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
	s.registry = NewFetcherRegistry(s.log, s.fetchers)
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
		f := fetchersManager.NewNumberFetcher(test.value, nil, s.wg)
		fetchersManager.RegisterFetcher(s.fetchers, f, test.key, nil)

		s.Equal(i+1, len(s.registry.Keys()))
	}

	keys := s.registry.Keys()

	s.Contains(keys, "some_fetcher")
	s.Contains(keys, "other_fetcher")
	s.Contains(keys, "new_fetcher")
}

func (s *RegistryTestSuite) TestRunNotRegistered() {
	f := fetchersManager.NewNumberFetcher(1, nil, s.wg)
	fetchersManager.RegisterFetcher(s.fetchers, f, "some-key", nil)

	err := s.registry.Run(context.TODO(), "unknown", fetching.CycleMetadata{})
	s.Error(err)
}

func (s *RegistryTestSuite) TestRunRegistered() {
	f1 := fetchersManager.NewSyncNumberFetcher(1, s.resourceCh)
	fetchersManager.RegisterFetcher(s.fetchers, f1, "some-key-1", nil)

	f2 := fetchersManager.NewSyncNumberFetcher(2, s.resourceCh)
	fetchersManager.RegisterFetcher(s.fetchers, f2, "some-key-2", nil)

	f3 := fetchersManager.NewSyncNumberFetcher(3, s.resourceCh)
	fetchersManager.RegisterFetcher(s.fetchers, f3, "some-key-3", nil)

	var tests = []struct {
		key string
		res fetchersManager.NumberResource
	}{
		{
			"some-key-1", fetchersManager.NumberResource{Num: 1},
		},
		{
			"some-key-2", fetchersManager.NumberResource{Num: 2},
		},
		{
			"some-key-3", fetchersManager.NumberResource{Num: 3},
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
	f := fetchersManager.NewNumberFetcher(1, nil, s.wg)
	fetchersManager.RegisterFetcher(s.fetchers, f, "some-key", nil)

	res := s.registry.ShouldRun("unknown")
	s.False(res)
}

func (s *RegistryTestSuite) TestShouldRun() {
	conditionTrue := fetchersManager.NewBoolFetcherCondition(true, "always-fetcher-condition")
	conditionFalse := fetchersManager.NewBoolFetcherCondition(false, "never-fetcher-condition")

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
		f := fetchersManager.NewNumberFetcher(1, nil, s.wg)
		fetchersManager.RegisterFetcher(s.fetchers, f, "some-key", test.conditions)
		s.registry = NewFetcherRegistry(s.log, s.fetchers)

		should := s.registry.ShouldRun("some-key")
		s.Equal(test.expected, should)
	}
}
