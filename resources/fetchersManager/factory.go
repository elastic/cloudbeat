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
	"errors"
	"fmt"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/conditions"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
)

var Factories = newFactories()

type factories struct {
	m map[string]fetching.Factory
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
		return nil, errors.New("fetcher factory could not be found")
	}

	return factory.Create(log, c, ch)
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
		}
		leaseProvider := conditions.NewLeaderLeaseProvider(context.Background(), client)
		c = append(c, conditions.NewLeaseFetcherCondition(log, leaseProvider))
	}

	return c, nil
}

type ParsedFetcher struct {
	name string
	f    fetching.Fetcher
}

func (fa *factories) ParseConfigFetchers(log *logp.Logger, cfg config.Config, ch chan fetching.ResourceInfo) ([]*ParsedFetcher, error) {
	var arr []*ParsedFetcher

	fetchers := fa.loadFetchers(cfg)
	for _, fcfg := range fetchers {
		p, err := fa.parseConfigFetcher(log, fcfg, ch)
		if err != nil {
			return nil, err
		}

		arr = append(arr, p)
	}

	return arr, nil
}

func (fa *factories) loadFetchers(cfg config.Config) []*agentconfig.C {
	var fetchers []*agentconfig.C
	if cfg.Type == config.InputTypeEKS {
		fetchers = cfg.Fetchers.EKS
	} else {
		fetchers = cfg.Fetchers.Vanilla
	}
	return fetchers
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
