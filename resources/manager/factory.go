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

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/conditions"
	"github.com/elastic/cloudbeat/resources/fetching"
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
		logp.L().Warnf("fetcher %q factory method overwritten", name)
	}

	fa.m[name] = f
}

func (fa *factories) CreateFetcher(name string, c *common.Config) (fetching.Fetcher, error) {
	factory, ok := fa.m[name]
	if !ok {
		return nil, errors.New("fetcher factory could not be found")
	}

	return factory.Create(c)
}

func (fa *factories) RegisterFetchers(registry FetchersRegistry, cfg config.Config) error {
	parsedList, err := fa.parseConfigFetchers(cfg)
	if err != nil {
		return err
	}

	for _, p := range parsedList {
		c := fa.getConditions(p.name)
		err := registry.Register(p.name, p.f, c...)
		if err != nil {
			logp.L().Errorf("could not read register fetcher: %v", err)
		}
	}

	return nil
}

// TODO: Move conditions to factories and implement inside every factory
func (fa *factories) getConditions(name string) []fetching.Condition {
	c := make([]fetching.Condition, 0)
	switch name {
	case fetching.KubeAPIType:
		var condition fetching.Condition
		// TODO: Use fetcher's kubeconfig configuration
		client, err := kubernetes.GetKubernetesClient("", kubernetes.KubeClientOptions{})
		if err != nil {
			logp.L().Error("getConditions error in GetKubernetesClient: %v", err)
			condition = conditions.NewErrorCondition(fetching.KubeAPIType, err)
		} else {
			leaseProvider := conditions.NewLeaderLeaseProvider(context.Background(), client)
			condition = conditions.NewLeaseFetcherCondition(leaseProvider)
		}
		c = append(c, condition)
	}

	return c
}

type ParsedFetcher struct {
	name string
	f    fetching.Fetcher
}

func (fa *factories) parseConfigFetchers(cfg config.Config) ([]*ParsedFetcher, error) {
	arr := []*ParsedFetcher{}
	for _, fcfg := range cfg.Fetchers {
		p, err := fa.parseConfigFetcher(fcfg)
		if err != nil {
			return nil, err
		}

		arr = append(arr, p)
	}

	return arr, nil
}

func (fa *factories) parseConfigFetcher(fcfg *common.Config) (*ParsedFetcher, error) {
	gen := fetching.BaseFetcherConfig{}
	err := fcfg.Unpack(&gen)
	if err != nil {
		return nil, err
	}

	f, err := fa.CreateFetcher(gen.Name, fcfg)
	if err != nil {
		return nil, err
	}

	return &ParsedFetcher{gen.Name, f}, nil
}
