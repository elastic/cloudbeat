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
	"sync"
	"time"

	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/factory"
)

type dynamic struct {
	registry
	period  time.Duration
	updater UpdaterFunc
	lock    sync.RWMutex
}

type UpdaterFunc func() (factory.FetchersMap, error)

func NewDynamic(
	log *logp.Logger,
	period time.Duration,
	updater UpdaterFunc,
) Registry {
	a := &dynamic{
		registry: registry{
			log: log,
			reg: factory.FetchersMap{},
		},
		period:  period,
		updater: updater,
		lock:    sync.RWMutex{},
	}

	go a.scheduleUpdate()
	return a
}

func (a *dynamic) Keys() []string {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.registry.Keys()
}

func (a *dynamic) ShouldRun(key string) bool {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.registry.ShouldRun(key)
}

func (a *dynamic) Run(ctx context.Context, key string, metadata fetching.CycleMetadata) error {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.registry.Run(ctx, key, metadata)
}

func (a *dynamic) Stop() {
	a.lock.RLock()
	defer a.lock.RUnlock()

	a.registry.Stop()
}

func (a *dynamic) scheduleUpdate() {
	a.lock.Lock()
	defer a.lock.Unlock()
	time.AfterFunc(a.period, a.scheduleUpdate)

	err := a.doUpdate()
	if err != nil {
		a.log.Errorf("failed to update accounts: %v", err)
	}
}

func (a *dynamic) doUpdate() error {
	m, err := a.updater()
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	a.registry = registry{
		log: a.log,
		reg: m,
	}
	return nil
}
