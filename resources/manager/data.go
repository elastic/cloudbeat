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
	"fmt"
	"sync"
	"time"

	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources/fetching"
)

// Data maintains a cache that is updated by Fetcher implementations registered
// against it. It sends the cache to an output channel at the defined interval.
type Data struct {
	timeout  time.Duration
	interval time.Duration
	output   chan ResourceMap
	fetchers FetchersRegistry
	wg       *sync.WaitGroup
	stop     chan struct{}
}

// update is a single update sent from a worker to a manager.
type update struct {
	key string
	val []fetching.Resource
}

type fetcherResult struct {
	resources []fetching.Resource
	err       error
}

type ResourceMap map[string][]fetching.Resource

// NewData returns a new Data instance with the given interval.
func NewData(interval time.Duration, timeout time.Duration, fetchers FetchersRegistry) (*Data, error) {
	return &Data{
		timeout:  timeout,
		interval: interval,
		fetchers: fetchers,
		output:   make(chan ResourceMap),
		stop:     make(chan struct{}),
	}, nil
}

// Output returns the output channel.
func (d *Data) Output() <-chan ResourceMap {
	return d.output
}

// Run updates the cache using Fetcher implementations.
func (d *Data) Run(ctx context.Context) error {
	go d.fetchAndSleep(ctx)
	return nil
}

func (d *Data) fetchAndSleep(ctx context.Context) {
	d.fetchIteration(ctx)
	for {
		select {
		case <-d.stop:
			logp.L().Errorf("fetchers manager stopped")
			return
		case <-ctx.Done():
			logp.L().Errorf("fetcher manager canceled")
			return
		case <-time.After(d.interval):
			d.fetchIteration(ctx)
		}
	}
}

func (d *Data) fetchIteration(ctx context.Context) {
	d.wg = &sync.WaitGroup{}
	mu := sync.Mutex{}
	cycleData := make(ResourceMap)
	for _, key := range d.fetchers.Keys() {
		d.wg.Add(1)
		go func(k string) {
			defer d.wg.Done()
			val, err := d.fetchSingle(ctx, k)
			if err != nil {
				logp.L().Errorf("error running fetcher for key %q: %v", k, err)
			} else {
				mu.Lock()
				defer mu.Unlock()
				cycleData[k] = val
			}
		}(key)
	}

	d.wg.Wait()
	d.output <- cycleData
}

func (d *Data) fetchSingle(ctx context.Context, k string) ([]fetching.Resource, error) {
	if !d.fetchers.ShouldRun(k) {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	result := make(chan fetcherResult, 1)
	defer close(result)

	go func() {
		val, err := d.fetchProtected(ctx, k)
		result <- fetcherResult{val, err}
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("fetcher %s reached a timeout after %v seconds", k, d.timeout.Seconds())
	case res := <-result:
		return res.resources, res.err
	}
}

func (d *Data) fetchProtected(ctx context.Context, k string) (val []fetching.Resource, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("fetcher %s recovered from panic: %v", k, r)
		}
	}()

	val, err = d.fetchers.Run(ctx, k)
	return
}

// Stop cleans up Data resources gracefully.
func (d *Data) Stop(ctx context.Context) {
	d.fetchers.Stop(ctx)
	close(d.stop)
	d.wg.Wait()
	close(d.output)
}
