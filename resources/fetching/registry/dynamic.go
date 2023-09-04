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
	d := &dynamic{
		registry: registry{
			log: log,
			reg: factory.FetchersMap{},
		},
		period:  period,
		updater: updater,
		lock:    sync.RWMutex{},
	}

	go d.scheduleUpdate()
	return d
}

func (d *dynamic) Keys() []string {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.registry.Keys()
}

func (d *dynamic) ShouldRun(key string) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.registry.ShouldRun(key)
}

func (d *dynamic) Run(ctx context.Context, key string, metadata fetching.CycleMetadata) error {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.registry.Run(ctx, key, metadata)
}

func (d *dynamic) Stop() {
	d.lock.RLock()
	defer d.lock.RUnlock()

	d.registry.Stop()
}

func (d *dynamic) scheduleUpdate() {
	d.lock.Lock()
	defer d.lock.Unlock()
	time.AfterFunc(d.period, d.scheduleUpdate)

	err := d.doUpdate()
	if err != nil {
		d.log.Errorf("failed to update accounts: %v", err)
	}
}

func (d *dynamic) doUpdate() error {
	m, err := d.updater()
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	d.registry = registry{
		log: d.log,
		reg: m,
	}
	return nil
}
