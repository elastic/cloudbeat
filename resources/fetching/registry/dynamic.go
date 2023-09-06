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
	"sync"
	"time"

	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/factory"
)

type dynamic struct {
	registry
	log     *logp.Logger
	period  time.Duration
	updater UpdaterFunc

	lock    sync.RWMutex
	running bool
	done    chan struct{}
}

type UpdaterFunc func() (factory.FetchersMap, error)

func NewDynamic(
	log *logp.Logger,
	period time.Duration,
	updater UpdaterFunc,
) Registry {
	d := &dynamic{
		log:     log,
		period:  period,
		updater: updater,
		lock:    sync.RWMutex{},
		done:    make(chan struct{}, 1),
	}

	return d
}

func (d *dynamic) Keys() []string {
	d.ensureRunning()
	defer d.lock.RUnlock()

	return d.registry.Keys()
}

func (d *dynamic) ShouldRun(key string) bool {
	d.ensureRunning()
	defer d.lock.RUnlock()

	return d.registry.ShouldRun(key)
}

func (d *dynamic) Run(ctx context.Context, key string, metadata fetching.CycleMetadata) error {
	d.ensureRunning()
	defer d.lock.RUnlock()

	return d.registry.Run(ctx, key, metadata)
}

func (d *dynamic) Stop() {
	if !d.setRunning(false) {
		return
	}
	defer d.lock.Unlock()

	d.done <- struct{}{}

	d.registry.Stop()
	d.registry = registry{}
}

func (d *dynamic) ensureRunning() {
	defer d.lock.RLock() // return object locked and ready to use registry functions
	if !d.setRunning(true) {
		return
	}

	d.doUpdateLocked() // first update
	d.lock.Unlock()

	go func() {
		timer := time.NewTimer(d.period)
		for {
			select {
			case <-timer.C:
				d.doUpdate()
				timer.Reset(d.period)
			case <-d.done:
				return
			}
		}
	}()
}

func (d *dynamic) setRunning(newValue bool) bool {
	// Optimization: try with RLock first
	if d.isRunning() == newValue {
		return false
	}
	// Need to re-check after Lock(), otherwise there is a race condition
	d.lock.Lock()
	if d.running == newValue {
		d.lock.Unlock()
		return false
	}
	d.running = newValue
	return true // return object locked
}

func (d *dynamic) isRunning() bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.running
}

func (d *dynamic) doUpdate() {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.doUpdateLocked()
}

func (d *dynamic) doUpdateLocked() {
	m, err := d.updater()
	if err != nil {
		d.log.Errorf("failed to update accounts: %v", err)
		return
	}

	d.registry = registry{
		log: d.log,
		reg: m,
	}
}
