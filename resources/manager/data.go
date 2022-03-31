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
	"sync"
	"time"

	"github.com/elastic/cloudbeat/resources/fetching"

	"github.com/elastic/beats/v7/libbeat/logp"
)

// Data maintains a cache that is updated by Fetcher implementations registered
// against it. It sends the cache to an output channel at the defined interval.
type Data struct {
	interval  time.Duration
	output    chan ResourceMap
	cycleData ResourceMap
	fetchers  FetchersRegistry
	wg        *sync.WaitGroup
}

// update is a single update sent from a worker to a manager.
type update struct {
	key string
	val []fetching.Resource
}

type ResourceMap map[string][]fetching.Resource

// NewData returns a new Data instance with the given interval.
func NewData(interval time.Duration, fetchers FetchersRegistry) (*Data, error) {
	return &Data{
		interval:  interval,
		output:    make(chan ResourceMap),
		cycleData: make(ResourceMap),
		fetchers:  fetchers,
	}, nil
}

// Output returns the output channel.
func (d *Data) Output() <-chan ResourceMap {
	return d.output
}

// Run updates the cache using Fetcher implementations.
func (d *Data) Run(ctx context.Context) error {
	updates := make(chan update)

	var wg sync.WaitGroup
	d.wg = &wg

	for _, key := range d.fetchers.Keys() {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			d.fetchWorker(ctx, updates, k)
		}(key)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		d.fetchManager(ctx, updates)
	}()

	return nil
}

func (d *Data) fetchWorker(ctx context.Context, updates chan update, k string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if !d.fetchers.ShouldRun(k) {
				break
			}

			val, err := d.fetchers.Run(ctx, k)
			if err != nil {
				logp.L().Errorf("error running fetcher for key %q: %v", k, err)
			}

			updates <- update{k, val}
		}
		// Go to sleep in each iteration.
		time.Sleep(d.interval)
	}
}

func (d *Data) fetchManager(ctx context.Context, updates chan update) {
	ticker := time.NewTicker(d.interval)

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			// Generate input ID?
			// Send aggregated data at cycle tick
			d.output <- d.cycleData
			d.cycleData = make(ResourceMap)

		// Aggregate fetcher's data into cycle data map
		case u := <-updates:
			d.cycleData[u.key] = u.val
		}
	}
}

// Stop cleans up Data resources gracefully.
func (d *Data) Stop(ctx context.Context, cancel context.CancelFunc) {
	cancel()

	d.fetchers.Stop(ctx)
	d.wg.Wait()

	close(d.output)
}
