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
	"fmt"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetching"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
)

var Factories = newFactories()

type factories struct {
	m map[string]fetching.Factory
}

type ParsedFetcher struct {
	name string
	f    fetching.Fetcher
}

func newFactories() factories {
	return factories{m: make(map[string]fetching.Factory)}
}

func (fa *factories) RegisterFactory(name string, f fetching.Factory) {
	_, ok := fa.m[name]
	if ok {
		panic(fmt.Errorf("fetcher factory with name %q is already registered", name))
	}

	fa.m[name] = f
}

func (fa *factories) CreateFetcher(log *logp.Logger, name string, c *agentconfig.C, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	factory, ok := fa.m[name]
	if !ok {
		return nil, fmt.Errorf("fetcher %s could not be found", name)
	}

	return factory.Create(log, c, ch)
}

func (fa *factories) ParseConfigFetchers(log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo) ([]*ParsedFetcher, error) {
	var arr []*ParsedFetcher

	for _, fcfg := range cfg.Fetchers {
		addCredentialsToFetcherConfiguration(log, cfg, fcfg)
		p, err := fa.parseConfigFetcher(log, fcfg, ch)
		if err != nil {
			return nil, err
		}

		arr = append(arr, p)
	}

	return arr, nil
}

func (fa *factories) parseConfigFetcher(log *logp.Logger, fcfg *agentconfig.C, ch chan fetching.ResourceInfo) (*ParsedFetcher, error) {
	gen := fetching.BaseFetcherConfig{}
	err := fcfg.Unpack(&gen)
	if err != nil {
		return nil, err
	}

	f, err := fa.CreateFetcher(log, gen.Name, fcfg, ch)
	if err != nil {
		return nil, err
	}

	return &ParsedFetcher{gen.Name, f}, nil
}

// addCredentialsToFetcherConfiguration adds the relevant credentials to the `fcfg`- the fetcher config
// This function takes the configuration file provided by the integration the `cfg` file
// and depending on the input type, extract the relevant credentials and add them to the fetcher config
func addCredentialsToFetcherConfiguration(log *logp.Logger, cfg *config.Config, fcfg *agentconfig.C) {
	if cfg.Type == config.InputTypeEks || cfg.Type == config.InputTypeAws {
		err := fcfg.Merge(cfg.AWSConfig)
		if err != nil {
			log.Errorf("Failed to merge aws configuration to fetcher configuration: %v", err)
		}
	}
}
