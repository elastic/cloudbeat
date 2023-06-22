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
	"fmt"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/factory"
	"github.com/elastic/elastic-agent-libs/logp"
)

type Registry interface {
	Keys() []string
	ShouldRun(key string) bool
	Run(ctx context.Context, key string, metadata fetching.CycleMetadata) error
	Stop()
}

type registry struct {
	log *logp.Logger
	reg factory.FetchersMap
}

func NewRegistry(log *logp.Logger, f factory.FetchersMap) Registry {
	r := &registry{
		log: log,
		reg: f,
	}
	return r
}

func (r *registry) Keys() []string {
	var keys []string
	for k := range r.reg {
		keys = append(keys, k)
	}

	return keys
}

func (r *registry) ShouldRun(key string) bool {
	registered, ok := r.reg[key]
	if !ok {
		return false
	}

	for _, condition := range registered.Condition {
		if !condition.Condition() {
			r.log.Infof("Conditional fetcher %q should not run because %q", key, condition.Name())
			return false
		}
	}

	return true
}

func (r *registry) Run(ctx context.Context, key string, metadata fetching.CycleMetadata) error {
	registered, ok := r.reg[key]
	if !ok {
		return fmt.Errorf("fetcher %v not found", key)
	}

	return registered.Fetcher.Fetch(ctx, metadata)
}

func (r *registry) Stop() {
	for key, registered := range r.reg {
		registered.Fetcher.Stop()
		r.log.Infof("Fetcher for key %q stopped", key)
	}
}
