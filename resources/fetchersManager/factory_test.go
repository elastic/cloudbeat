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

package fetchersManager

import (
	"context"
	"testing"

	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/elastic-agent-libs/logp"

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
	resourceCh chan fetching.ResourceInfo
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
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *FactoriesTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *FactoriesTestSuite) TestRegisterFetchers() {
	tests := []struct {
		key             string
		value           int
		integrationType string
	}{
		{"new_fetcher", 6, ""},
		{"new_fetcher", 6, "cloudbeat/cis_k8s"},
		{"other_fetcher", 4, ""},
		{"other_fetcher", 4, "cloudbeat/cis_k8s"},
	}

	for _, test := range tests {
		numCfg := numberConfig(test.value)
		err := numCfg.SetString("name", -1, test.key)
		s.NoError(err, "Could not set name: %v", err)

		conf := &config.Config{
			Benchmark: test.integrationType,
		}
		conf.Fetchers = []*agentconfig.C{numCfg}

		parsedList, err := ParseConfigFetchers(s.log, conf, s.resourceCh, map[string]fetching.Fetcher{
			test.key: newSyncNumberFetcher(test.value, s.resourceCh),
		})
		s.NoError(err)

		reg := NewFetcherRegistry(s.log)
		err = reg.RegisterFetchers(parsedList, nil)
		s.NoError(err)
		s.Equal(1, len(reg.Keys()))

		err = reg.Run(context.Background(), test.key, fetching.CycleMetadata{})
		results := testhelper.CollectResources(s.resourceCh)

		s.NoError(err)
		s.NotEmpty(results)
		s.Equal(test.value, results[0].Resource.GetData())
	}
}

func (s *FactoriesTestSuite) TestRegisterNotFoundFetchers() {
	tests := []struct {
		key   string
		value int
	}{
		{"not_found_fetcher", 42},
	}

	for _, test := range tests {
		conf := &config.Config{}
		numCfg := numberConfig(test.value)
		err := numCfg.SetString("name", -1, test.key)
		s.NoError(err, "Could not set name: %v", err)

		conf.Fetchers = []*agentconfig.C{numCfg}

		_, err = ParseConfigFetchers(s.log, conf, s.resourceCh, map[string]fetching.Fetcher{})
		s.Error(err)
	}
}

func (s *FactoriesTestSuite) TestRegisterFromFullConfig() {
	tests := []struct {
		config string
	}{
		{
			`
not_data_yaml:
  activated_rules:
    cis_k8s:
      - a
fetchers:
  - name: process
`,
		},
		{
			`
not_data_yaml:
  activated_rules:
    cis_k8s:
      - a
fetchers:
  - name: aws-eks
`,
		},
	}

	for _, test := range tests {
		cfg, err := agentconfig.NewConfigFrom(test.config)
		s.NoError(err)
		c, err := config.New(cfg)
		s.NoError(err)

		reg := NewFetcherRegistry(s.log)
		var fetcher config.Fetcher
		err = c.Fetchers[0].Unpack(&fetcher)
		s.NoError(err)

		parsedList, err := ParseConfigFetchers(s.log, c, s.resourceCh, map[string]fetching.Fetcher{
			fetcher.Name: &syncNumberFetcher{-1, false, s.resourceCh},
		})
		s.Equal(fetcher.Name, parsedList[0].name)
		s.NoError(err)

		err = reg.RegisterFetchers(parsedList, nil)
		s.NoError(err)
	}
}
