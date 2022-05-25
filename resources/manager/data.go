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
	log      *logp.Logger
	timeout  time.Duration
	interval time.Duration
	output   chan ResourceMap
	fetchers FetchersRegistry
	wg       *sync.WaitGroup
	stop     chan struct{}
}

type fetcherResult struct {
	resources []fetching.Resource
	err       error
}

type ResourceMap map[string][]fetching.Resource

// NewData returns a new Data instance.
// interval is the duration the manager wait between two consequtive cycles.
// timeout is the maximum duration the manager wait for a single fetcher to return results.
func NewData(log *logp.Logger, interval time.Duration, timeout time.Duration, fetchers FetchersRegistry) (*Data, error) {
	return &Data{
		log:      log,
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
	// Happens once in a lifetime of cloudbeat and then enters the loop
	d.fetchIteration(ctx)
	for {
		select {
		case <-d.stop:
			d.log.Info("Fetchers manager stopped.")
			return
		case <-ctx.Done():
			d.log.Info("Fetchers manager canceled.")
			return
		case <-time.After(d.interval):
			d.fetchIteration(ctx)
		}
	}
}

// fetchIteration waits for all the registered fetchers and sends all the resources on the output channel.
// The function must not get called in parallel.
func (d *Data) fetchIteration(ctx context.Context) {
	d.log.Infof("Manager triggered fetching for %d fetchers", len(d.fetchers.Keys()))

	d.wg = &sync.WaitGroup{}
	mu := sync.Mutex{}
	cycleData := make(ResourceMap)
	start := time.Now()

	for _, key := range d.fetchers.Keys() {
		d.wg.Add(1)
		go func(k string) {
			defer d.wg.Done()
			val, err := d.fetchSingle(ctx, k)
			if err != nil {
				d.log.Errorf("Error running fetcher for key %s: %v", k, err)
			} else if val != nil {
				d.log.Debugf("Fetcher %s finished and found %d values", k, len(val))
				mu.Lock()
				defer mu.Unlock()
				cycleData[k] = val
			}
		}(key)
	}

	d.wg.Wait()
	d.log.Infof("Manager finished waiting and sending data after %d milliseconds", time.Since(start).Milliseconds())
	d.output <- cycleData
}

func (d *Data) fetchSingle(ctx context.Context, k string) ([]fetching.Resource, error) {
	if !d.fetchers.ShouldRun(k) {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	// The buffer is required to avoid go-routine leaks in a case a fetcher timed out
	result := make(chan fetcherResult, 1)

	go func() {
		val, err := d.fetchProtected(ctx, k)
		result <- fetcherResult{val, err}
		close(result)
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("fetcher %s reached a timeout after %v seconds", k, d.timeout.Seconds())
	case res := <-result:
		return res.resources, res.err
	}
}

// fetchProtected protect the fetching goroutine from getting panic
func (d *Data) fetchProtected(ctx context.Context, k string) (val []fetching.Resource, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("fetcher %s recovered from panic: %v", k, r)
		}
	}()

	val, err = d.fetchers.Run(ctx, k)
	return val, err
}

// Stop cleans up Data resources gracefully.
func (d *Data) Stop() {
	d.fetchers.Stop()
	close(d.stop)
	d.wg.Wait()
	close(d.output)
}
