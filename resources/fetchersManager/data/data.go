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

package data

import (
	"context"
	"fmt"
	"github.com/elastic/cloudbeat/resources/fetchersManager/registry"
	"sync"
	"time"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-libs/logp"
)

// Data maintains a cache that is updated by Fetcher implementations registered
// against it. It sends the cache to an output channel at the defined interval.
type Data struct {
	log *logp.Logger

	// Duration of a single fetcher
	timeout time.Duration

	// Duration between two consecutive cycles
	interval time.Duration

	fetchers registry.FetchersRegistry

	ctx    context.Context
	cancel context.CancelFunc
}

func NewData(ctx context.Context, log *logp.Logger, interval time.Duration, timeout time.Duration, fetchers registry.FetchersRegistry) (*Data, error) {
	ctx, cancel := context.WithCancel(ctx)

	return &Data{
		log:      log,
		timeout:  timeout,
		interval: interval,
		fetchers: fetchers,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// Run starts all configured fetchers to collect resources.
func (d *Data) Run() {
	go d.fetchAndSleep(d.ctx)
}

func (d *Data) Stop() {
	d.cancel()
	d.fetchers.Stop()
}

func (d *Data) fetchAndSleep(ctx context.Context) {

	// set immediate exec for first time run
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			d.log.Info("Fetchers manager canceled")
			return
		case <-timer.C:
			// update the interval
			timer.Reset(d.interval)
			// this is blocking so the stop will not be called until all the fetchers are finished
			// in case there is a blocking fetcher it will halt (til the d.timeout)
			go d.fetchIteration(ctx)
		}
	}
}

// fetchIteration waits for all the registered fetchers and trigger them to fetch relevant resources.
// The function must not get called in parallel.
func (d *Data) fetchIteration(ctx context.Context) {
	d.log.Infof("Manager triggered fetching for %d fetchers", len(d.fetchers.Keys()))

	start := time.Now()

	seq := time.Now().Unix()
	d.log.Infof("Cycle %d has started", seq)
	wg := &sync.WaitGroup{}
	for _, key := range d.fetchers.Keys() {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			err := d.fetchSingle(ctx, k, fetching.CycleMetadata{Sequence: seq})
			if err != nil {
				d.log.Errorf("Error running fetcher for key %s: %v", k, err)
			}
		}(key)
	}

	wg.Wait()
	d.log.Infof("Manager finished waiting and sending data after %d milliseconds", time.Since(start).Milliseconds())
	d.log.Infof("Cycle %d resource fetching has ended", seq)
}

func (d *Data) fetchSingle(ctx context.Context, k string, cycleMetadata fetching.CycleMetadata) error {
	if !d.fetchers.ShouldRun(k) {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	// The buffer is required to avoid go-routine leaks in a case a fetcher timed out
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		errCh <- d.fetchProtected(ctx, k, cycleMetadata)
	}()

	select {
	case <-ctx.Done():
		switch ctx.Err() {
		case context.DeadlineExceeded:
			return fmt.Errorf("fetcher %s reached a timeout after %v seconds", k, d.timeout.Seconds())
		case context.Canceled:
			return fmt.Errorf("fetcher %s was canceled", k)
		default:
			return fmt.Errorf("fetcher %s failed with an unknown error: %v", k, ctx.Err())
		}

	case err := <-errCh:
		return err
	}
}

// fetchProtected protect the fetching goroutine from getting panic
func (d *Data) fetchProtected(ctx context.Context, k string, metadata fetching.CycleMetadata) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("fetcher %s recovered from panic: %v", k, r)
		}
	}()

	return d.fetchers.Run(ctx, k, metadata)
}
