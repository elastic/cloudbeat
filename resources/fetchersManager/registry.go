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
	"fmt"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-libs/logp"
)

type FetchersRegistry interface {
	Register(fetcherName string, f fetching.Fetcher, c ...fetching.Condition) error
	Keys() []string
	ShouldRun(key string) bool
	Run(ctx context.Context, key string, metadata fetching.CycleMetadata) error
	Stop()
	RegisterFetchers(fetchers []*ParsedFetcher) error
}

type fetchersRegistry struct {
	log *logp.Logger
	reg map[string]registeredFetcher
}

type registeredFetcher struct {
	f fetching.Fetcher
	c []fetching.Condition
}

func NewFetcherRegistry(log *logp.Logger) FetchersRegistry {
	reg := &fetchersRegistry{
		log: log,
		reg: make(map[string]registeredFetcher),
	}
	return reg
}

// RegisterFetchers registers entire list of parsed fetchers
func (r *fetchersRegistry) RegisterFetchers(fetchers []*ParsedFetcher) error {
	for _, p := range fetchers {
		c, err := Factories.getConditions(r.log, p.name)
		if err != nil {
			return fmt.Errorf("RegisterFetchers error in getConditions for factory %s skipping Register due to: %v", p.name, err)
		}

		err = r.Register(p.name, p.f, c...)
		if err != nil {
			return fmt.Errorf("RegisterFetchers error in Register for factory %s skipping Register due to: %v", p.name, err)
		}
	}

	return nil
}

// Register registers a Fetcher implementation.
func (r *fetchersRegistry) Register(fetcherName string, f fetching.Fetcher, c ...fetching.Condition) error {
	r.log.Infof("Registering new fetcher: %s", fetcherName)
	if _, ok := r.reg[fetcherName]; ok {
		return fmt.Errorf("fetcher %s is already registered", fetcherName)
	}

	r.reg[fetcherName] = registeredFetcher{
		f: f,
		c: c,
	}

	return nil
}

func (r *fetchersRegistry) Keys() []string {
	var keys []string
	for k := range r.reg {
		keys = append(keys, k)
	}

	return keys
}

func (r *fetchersRegistry) ShouldRun(key string) bool {
	registered, ok := r.reg[key]
	if !ok {
		return false
	}

	for _, condition := range registered.c {
		if !condition.Condition() {
			r.log.Infof("Conditional fetcher %q should not run because %q", key, condition.Name())
			return false
		}
	}

	return true
}

func (r *fetchersRegistry) Run(ctx context.Context, key string, metadata fetching.CycleMetadata) error {
	registered, ok := r.reg[key]
	if !ok {
		return fmt.Errorf("fetcher %v not found", key)
	}

	return registered.f.Fetch(ctx, metadata)
}

func (r *fetchersRegistry) Stop() {
	for key, rfetcher := range r.reg {
		rfetcher.f.Stop()
		r.log.Infof("Fetcher for key %q stopped", key)
	}
}
