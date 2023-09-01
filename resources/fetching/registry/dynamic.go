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
	"strings"
	"sync"
	"time"

	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
)

type dynamic struct {
	log        *logp.Logger
	registries map[string]Registry
	period     time.Duration
	updater    UpdaterFunc
	lock       sync.RWMutex
}

type UpdaterFunc func() (map[string]Registry, error)

func NewDynamic(
	log *logp.Logger,
	period time.Duration,
	updater UpdaterFunc,
) Registry {
	a := &dynamic{
		log:     log,
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

	var keys []string
	for account, reg := range a.registries {
		for _, key := range reg.Keys() {
			keys = append(keys, fmt.Sprintf("%s-%s", account, key))
		}
	}
	return keys
}

func (a *dynamic) ShouldRun(key string) bool {
	reg, regKey := a.get(key)
	if reg == nil {
		return false
	}
	return reg.ShouldRun(regKey)
}

func (a *dynamic) Run(ctx context.Context, key string, metadata fetching.CycleMetadata) error {
	reg, regKey := a.get(key)
	if reg == nil {
		return fmt.Errorf("could not find registry for key %s", key)
	}
	return reg.Run(ctx, regKey, metadata)
}

func (a *dynamic) Stop() {
	a.lock.RLock()
	defer a.lock.RUnlock()

	for _, r := range a.registries {
		r.Stop()
	}
}

func (a *dynamic) get(key string) (Registry, string) {
	a.lock.RLock()
	defer a.lock.RUnlock()

	parts := strings.SplitN(key, "-", 2)
	if len(parts) != 2 {
		a.log.Errorf("key %s is in wrong format", key)
		return nil, ""
	}

	reg, ok := a.registries[parts[0]]
	if !ok {
		return nil, ""
	}
	return reg, parts[1]
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
	newRegistries, err := a.updater()
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	a.registries = newRegistries
	return nil
}
