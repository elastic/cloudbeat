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
	"errors"
	"fmt"

	"github.com/elastic/cloudbeat/conf"
	"github.com/elastic/cloudbeat/resources/conditions"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
)

var Factories = newFactories()

type factories struct {
	m map[string]fetching.Factory
}

func newFactories() factories {
	return factories{m: make(map[string]fetching.Factory)}
}

func (fa *factories) ListFetcherFactory(name string, f fetching.Factory) {
	_, ok := fa.m[name]
	if ok {
		panic(fmt.Errorf("fetcher factory with name %q listed more than once", name))
	}

	fa.m[name] = f
}

func (fa *factories) CreateFetcher(log *logp.Logger, name string, c *config.C) (fetching.Fetcher, error) {
	factory, ok := fa.m[name]
	if !ok {
		return nil, errors.New("fetcher factory could not be found")
	}

	return factory.Create(log, c)
}

func (fa *factories) RegisterFetchers(log *logp.Logger, registry FetchersRegistry, cfg conf.Config) error {
	parsedList, err := fa.parseConfigFetchers(log, cfg)
	if err != nil {
		return err
	}

	for _, p := range parsedList {
		c, err := fa.getConditions(log, p.name)
		if err != nil {
			log.Errorf("RegisterFetchers error in getConditions for factory %s skipping Register due to: %v", p.name, err)
			continue
		}

		err = registry.Register(p.name, p.f, c...)
		if err != nil {
			log.Errorf("Could not read register fetcher: %v", err)
		}
	}

	return nil
}

// TODO: Move conditions to factories and implement inside every factory
func (fa *factories) getConditions(log *logp.Logger, name string) ([]fetching.Condition, error) {
	c := make([]fetching.Condition, 0)
	switch name {
	case fetching.KubeAPIType:
		// TODO: Use fetcher's kubeconfig configuration
		client, err := kubernetes.GetKubernetesClient("", kubernetes.KubeClientOptions{})
		if err != nil {
			log.Errorf("getConditions error in GetKubernetesClient: %v", err)
			return nil, err
		} else {
			leaseProvider := conditions.NewLeaderLeaseProvider(context.Background(), client)
			c = append(c, conditions.NewLeaseFetcherCondition(log, leaseProvider))
		}
	}

	return c, nil
}

type ParsedFetcher struct {
	name string
	f    fetching.Fetcher
}

func (fa *factories) parseConfigFetchers(log *logp.Logger, cfg conf.Config) ([]*ParsedFetcher, error) {
	arr := []*ParsedFetcher{}
	for _, fcfg := range cfg.Fetchers {
		p, err := fa.parseConfigFetcher(log, fcfg)
		if err != nil {
			return nil, err
		}

		arr = append(arr, p)
	}

	return arr, nil
}

func (fa *factories) parseConfigFetcher(log *logp.Logger, fcfg *config.C) (*ParsedFetcher, error) {
	gen := fetching.BaseFetcherConfig{}
	err := fcfg.Unpack(&gen)
	if err != nil {
		return nil, err
	}

	f, err := fa.CreateFetcher(log, gen.Name, fcfg)
	if err != nil {
		return nil, err
	}

	return &ParsedFetcher{gen.Name, f}, nil
}
