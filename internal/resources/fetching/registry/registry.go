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

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
)

type Registry interface {
	Keys() []string
	ShouldRun(key string) bool
	Run(ctx context.Context, key string, metadata cycle.Metadata) error
	Update()
	Stop()
}

type registry struct {
	log     *clog.Logger
	reg     FetchersMap
	updater UpdaterFunc
}

type Option func(r *registry)

type UpdaterFunc func() (FetchersMap, error)

func WithUpdater(fn UpdaterFunc) Option {
	return func(r *registry) {
		r.updater = fn
	}
}

func WithFetchersMap(f FetchersMap) Option {
	return func(r *registry) {
		r.reg = f
	}
}

func NewRegistry(log *clog.Logger, options ...Option) Registry {
	r := &registry{
		log: log,
	}
	for _, fn := range options {
		fn(r)
	}
	return r
}

func (r *registry) Keys() []string {
	keys := make([]string, 0, len(r.reg))
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

func (r *registry) Run(ctx context.Context, key string, metadata cycle.Metadata) error {
	registered, ok := r.reg[key]
	if !ok {
		return fmt.Errorf("fetcher %v not found", key)
	}

	return registered.Fetcher.Fetch(ctx, metadata)
}

func (r *registry) Update() {
	if r.updater == nil {
		return
	}
	fm, err := r.updater()
	if err != nil {
		r.log.Errorf("Failed to update registry: %v", err)
		return
	}
	r.reg = fm
}

func (r *registry) Stop() {
	for key, registered := range r.reg {
		registered.Fetcher.Stop()
		r.log.Infof("Fetcher for key %q stopped", key)
	}
}
