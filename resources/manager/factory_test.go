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
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/elastic-agent-libs/logp"
	"testing"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetching"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/stretchr/testify/suite"
)

type syncNumberFetcher struct {
	num        int
	stopCalled bool
	resourceCh chan fetching.ResourceInfo
}

func newSyncNumberFetcher(num int, ch chan fetching.ResourceInfo) fetching.Fetcher {
	return &syncNumberFetcher{num, false, ch}
}

func (f *syncNumberFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.resourceCh <- fetching.ResourceInfo{
		Resource:      fetchValue(f.num),
		CycleMetadata: cMetadata,
	}

	return nil
}

func (f *syncNumberFetcher) Stop() {
	f.stopCalled = true
}

type FactoriesTestSuite struct {
	suite.Suite

	log        *logp.Logger
	F          factories
	resourceCh chan fetching.ResourceInfo
}

type numberFetcherFactory struct{}

func (n *numberFetcherFactory) Create(log *logp.Logger, c *agentconfig.C, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	x, _ := c.Int("num", -1)
	return &syncNumberFetcher{int(x), false, ch}, nil
}

func numberConfig(number int) *agentconfig.C {
	c := agentconfig.NewConfig()
	err := c.SetInt("num", -1, int64(number))
	if err != nil {
		logp.L().Errorf("Could not set number config: %v", err)
		return nil
	}
	return c
}

func TestFactoriesTestSuite(t *testing.T) {
	s := new(FactoriesTestSuite)
	s.log = logp.NewLogger("cloudbeat_factories_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *FactoriesTestSuite) SetupTest() {
	s.F = newFactories()
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *FactoriesTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *FactoriesTestSuite) TestListFetcher() {
	var tests = []struct {
		key string
	}{
		{"some_fetcher"},
		{"other_fetcher"},
		{"new_fetcher"},
	}

	for _, test := range tests {
		s.F.ListFetcherFactory(test.key, &numberFetcherFactory{})
	}

	s.Contains(s.F.m, "some_fetcher")
	s.Contains(s.F.m, "other_fetcher")
	s.Contains(s.F.m, "new_fetcher")
}

func (s *FactoriesTestSuite) TestCreateFetcher() {
	var tests = []struct {
		key   string
		value int
	}{
		{"some_fetcher", 1},
		{"other_fetcher", 4},
		{"new_fetcher", 6},
	}

	for _, test := range tests {
		s.F.ListFetcherFactory(test.key, &numberFetcherFactory{})
		c := numberConfig(test.value)

		f, err := s.F.CreateFetcher(s.log, test.key, c, s.resourceCh)
		s.NoError(err)
		err = f.Fetch(context.TODO(), fetching.CycleMetadata{})
		results := testhelper.CollectResources(s.resourceCh)

		s.Equal(1, len(results))
		s.Nil(err)
		s.Equal(test.value, results[0].GetData())
	}
}

func (s *FactoriesTestSuite) TestCreateFetcherCollision() {
	var tests = []struct {
		key string
	}{
		{"some_fetcher"},
		{"some_fetcher"},
	}

	s.Panics(func() {
		for _, test := range tests {
			s.F.ListFetcherFactory(test.key, &numberFetcherFactory{})
		}
	})
}

func (s *FactoriesTestSuite) TestRegisterFetchers() {
	var tests = []struct {
		key   string
		value int
	}{
		{"new_fetcher", 6},
		{"other_fetcher", 4},
	}

	for _, test := range tests {
		s.F = newFactories()
		s.F.ListFetcherFactory(test.key, &numberFetcherFactory{})
		reg := NewFetcherRegistry(s.log)
		numCfg := numberConfig(test.value)
		err := numCfg.SetString("name", -1, test.key)
		if err != nil {
			logp.L().Errorf("Could not set name: %v", err)
			return
		}
		conf := config.DefaultConfig
		conf.Fetchers = append(conf.Fetchers, numCfg)
		err = s.F.RegisterFetchers(s.log, reg, conf, s.resourceCh)

		s.NoError(err)
		s.Equal(1, len(reg.Keys()))

		err = reg.Run(context.Background(), test.key, fetching.CycleMetadata{})
		results := testhelper.CollectResources(s.resourceCh)

		s.NoError(err)
		s.Equal(test.value, results[0].Resource.GetData())
	}
}

func (s *FactoriesTestSuite) TestRegisterNotFoundFetchers() {
	var tests = []struct {
		key   string
		value int
	}{
		{"not_found_fetcher", 42},
	}

	for _, test := range tests {
		reg := NewFetcherRegistry(s.log)
		numCfg := numberConfig(test.value)
		err := numCfg.SetString("name", -1, test.key)
		if err != nil {
			logp.L().Errorf("Could not set name: %v", err)
			return
		}
		conf := config.DefaultConfig
		conf.Fetchers = append(conf.Fetchers, numCfg)
		err = s.F.RegisterFetchers(s.log, reg, conf, s.resourceCh)
		s.Error(err)
	}
}
