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

package manager

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"sync"
)

type numberFetcher struct {
	num        int
	stopCalled bool
	resourceCh chan fetching.ResourceInfo
	wg         *sync.WaitGroup
}

type NumberResource struct {
	Num int
}

func newNumberFetcher(num int, ch chan fetching.ResourceInfo, wg *sync.WaitGroup) fetching.Fetcher {
	return &numberFetcher{num, false, ch, wg}
}

func (f *numberFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	defer f.wg.Done()

	f.resourceCh <- fetching.ResourceInfo{
		Resource:      fetchValue(f.num),
		CycleMetadata: cMetadata,
	}

	return nil
}

func (f *numberFetcher) Stop() {
	f.stopCalled = true
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

func fetchValue(num int) fetching.Resource {
	return NumberResource{num}
}

func registerNFetchers(t *testing.T, reg FetchersRegistry, n int, ch chan fetching.ResourceInfo, wg *sync.WaitGroup) {
	for i := 0; i < n; i++ {
		key := fmt.Sprint(i)
		err := reg.Register(key, newNumberFetcher(i, ch, wg))
		assert.NoError(t, err)
	}
}

type RegistryTestSuite struct {
	suite.Suite

	log        *logp.Logger
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
	s.registry = NewFetcherRegistry(s.log)
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
		f := newNumberFetcher(test.value, nil, s.wg)
		err := s.registry.Register(test.key, f, nil)
		s.Nil(err)

		s.Equal(i+1, len(s.registry.Keys()))
	}

	keys := s.registry.Keys()

	s.Contains(keys, "some_fetcher")
	s.Contains(keys, "other_fetcher")
	s.Contains(keys, "new_fetcher")
}

func (s *RegistryTestSuite) TestRegisterDuplicateKey() {
	f := newNumberFetcher(1, nil, s.wg)
	err := s.registry.Register("some-key", f, nil)
	s.NoError(err)

	err = s.registry.Register("some-key", f, nil)
	s.Error(err)
}

func (s *RegistryTestSuite) TestRegister10() {
	count := 10
	registerNFetchers(s.T(), s.registry, count, nil, s.wg)
	s.Equal(count, len(s.registry.Keys()))
}

func (s *RegistryTestSuite) TestRunNotRegistered() {
	f := newNumberFetcher(1, nil, s.wg)
	err := s.registry.Register("some-key", f, nil)
	s.NoError(err)

	err = s.registry.Run(context.TODO(), "unknown", fetching.CycleMetadata{})
	s.Error(err)
}

func (s *RegistryTestSuite) TestRunRegistered() {
	f1 := newSyncNumberFetcher(1, s.resourceCh)
	err := s.registry.Register("some-key-1", f1, nil)
	s.NoError(err)

	f2 := newSyncNumberFetcher(2, s.resourceCh)
	err = s.registry.Register("some-key-2", f2, nil)
	s.NoError(err)

	f3 := newSyncNumberFetcher(3, s.resourceCh)
	err = s.registry.Register("some-key-3", f3, nil)
	s.NoError(err)

	var tests = []struct {
		key string
		res NumberResource
	}{
		{
			"some-key-1", NumberResource{1},
		},
		{
			"some-key-2", NumberResource{2},
		},
		{
			"some-key-3", NumberResource{3},
		},
	}

	for _, test := range tests {
		err = s.registry.Run(context.TODO(), test.key, fetching.CycleMetadata{})
		results := testhelper.CollectResources(s.resourceCh)

		s.NoError(err)
		s.Equal(1, len(results))
		s.Equal(test.res.Num, results[0].GetData())
	}
}

func (s *RegistryTestSuite) TestShouldRunNotRegistered() {
	f := newNumberFetcher(1, nil, s.wg)
	err := s.registry.Register("some-key", f, nil)
	s.NoError(err)

	res := s.registry.ShouldRun("unknown")
	s.False(res)
}

func (s *RegistryTestSuite) TestShouldRun() {
	conditionTrue := newBoolFetcherCondition(true, "always-fetcher-condition")
	conditionFalse := newBoolFetcherCondition(false, "never-fetcher-condition")

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
		s.registry = NewFetcherRegistry(s.log)
		f := newNumberFetcher(1, nil, s.wg)
		err := s.registry.Register("some-key", f, test.conditions...)
		s.NoError(err)

		should := s.registry.ShouldRun("some-key")
		s.Equal(test.expected, should)
	}
}

func (res NumberResource) GetData() interface{} {
	return res.Num
}

func (res NumberResource) GetMetadata() fetching.ResourceMetadata {
	return fetching.ResourceMetadata{
		ID:      "",
		Type:    "number",
		SubType: "number",
		Name:    "number",
	}
}

func (res NumberResource) GetElasticCommonData() interface{} {
	return nil
}
