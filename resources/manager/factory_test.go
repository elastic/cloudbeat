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
	"testing"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/stretchr/testify/suite"
)

type FactoriesTestSuite struct {
	suite.Suite
	F factories
}

type numberFetcherFactory struct {
}

func (n *numberFetcherFactory) Create(c *common.Config) (fetching.Fetcher, error) {
	x, _ := c.Int("num", -1)
	return &numberFetcher{int(x), false}, nil
}

func numberConfig(number int) *common.Config {
	c := common.NewConfig()
	c.SetInt("num", -1, int64(number))
	return c
}

func TestFactoriesTestSuite(t *testing.T) {
	suite.Run(t, new(FactoriesTestSuite))
}

func (s *FactoriesTestSuite) SetupTest() {
	s.F = newFactories()
}

func (s *FactoriesTestSuite) TestListFetcher() {
	var tests = []struct {
		key string
	}{
		{"duplicate_fetcher"},
		{"duplicate_fetcher"},
		{"other_fetcher"},
		{"new_fetcher"},
	}

	for _, test := range tests {
		s.F.ListFetcherFactory(test.key, &numberFetcherFactory{})
	}

	s.Contains(s.F.m, "duplicate_fetcher")
	s.Contains(s.F.m, "other_fetcher")
	s.Contains(s.F.m, "new_fetcher")
}

func (s *FactoriesTestSuite) TestCreateFetcher() {
	var tests = []struct {
		key   string
		value int
	}{
		{"duplicate_fetcher", 1},
		{"duplicate_fetcher", 2},
		{"other_fetcher", 4},
		{"new_fetcher", 6},
	}

	for _, test := range tests {
		s.F.ListFetcherFactory(test.key, &numberFetcherFactory{})
		c := numberConfig(test.value)
		f, err := s.F.CreateFetcher(test.key, c)
		s.NoError(err)
		res, err := f.Fetch(context.TODO())
		s.NoError(err)

		s.Equal(1, len(res))
		s.Equal(test.value, res[0].GetData())
	}
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
		reg := NewFetcherRegistry()
		numCfg := numberConfig(test.value)
		numCfg.SetString("name", -1, test.key)
		conf := config.DefaultConfig
		conf.Fetchers = append(conf.Fetchers, numCfg)
		err := s.F.RegisterFetchers(reg, conf)
		s.NoError(err)
		s.Equal(1, len(reg.Keys()))

		res, err := reg.Run(context.Background(), test.key)
		s.NoError(err)
		s.Equal(test.value, res[0].GetData())

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
		reg := NewFetcherRegistry()
		numCfg := numberConfig(test.value)
		numCfg.SetString("name", -1, test.key)
		conf := config.DefaultConfig
		conf.Fetchers = append(conf.Fetchers, numCfg)
		err := s.F.RegisterFetchers(reg, conf)
		s.Error(err)
	}
}
