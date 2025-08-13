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

package cycle

import (
	"context"
	"sync"

	"github.com/elastic/cloudbeat/internal/infra/clog"
)

// Cache is a thread-safe generic struct that you can use to cache values for the current cycle. On a new cycle,
// determined by cycle metadata, the callback function is called and a new value is initialized. If the callback fails
// and an old value exists, it is re-used.
type Cache[T any] struct {
	log       *clog.Logger
	lastCycle Metadata

	cachedValue T
	mu          sync.RWMutex
}

func NewCache[T any](log *clog.Logger) *Cache[T] {
	return &Cache[T]{
		log:       log.Named("cycle.cache"),
		lastCycle: Metadata{Sequence: -1},
	}
}

func (c *Cache[T]) GetValue(ctx context.Context, cycle Metadata, fetch func(context.Context) (T, error)) (T, error) {
	if !c.needsUpdate(cycle) {
		return c.cachedValue, nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.needsUpdateLocked(cycle) { // check again with write lock.
		return c.cachedValue, nil
	}

	result, err := fetch(ctx)
	if err != nil {
		if c.lastCycle.Sequence < 0 {
			return result, err
		}
		c.log.Errorf(ctx, "Failed to renew, using cached value: %v", err)
	} else {
		c.cachedValue = result
		c.lastCycle = cycle
	}
	return c.cachedValue, nil
}

func (c *Cache[T]) needsUpdate(cycle Metadata) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.needsUpdateLocked(cycle)
}

func (c *Cache[T]) needsUpdateLocked(cycle Metadata) bool {
	return c.lastCycle.Sequence < 0 || c.lastCycle.Sequence < cycle.Sequence
}
