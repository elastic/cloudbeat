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
	"github.com/gofrs/uuid"
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
	fetchers FetchersRegistry
	wg       *sync.WaitGroup
	stop     chan struct{}
}

// NewData returns a new Data instance.
// interval is the duration the manager wait between two consecutive cycles.
// timeout is the maximum duration the manager wait for a single fetcher to return results.
func NewData(log *logp.Logger, interval time.Duration, timeout time.Duration, fetchers FetchersRegistry) (*Data, error) {
	return &Data{
		log:      log,
		timeout:  timeout,
		interval: interval,
		fetchers: fetchers,
		stop:     make(chan struct{}),
	}, nil
}

// Run starts all configured fetchers to collect resources.
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

// fetchIteration waits for all the registered fetchers and trigger them to fetch relevant resources.
// The function must not get called in parallel.
func (d *Data) fetchIteration(ctx context.Context) {
	d.log.Infof("Manager triggered fetching for %d fetchers", len(d.fetchers.Keys()))

	d.wg = &sync.WaitGroup{}
	start := time.Now()

	cycleId, _ := uuid.NewV4()
	cycleMetadata := fetching.CycleMetadata{CycleId: cycleId}
	d.log.Infof("Cycle %s has started", cycleId.String())

	for _, key := range d.fetchers.Keys() {
		d.wg.Add(1)
		go func(k string) {
			defer d.wg.Done()
			err := d.fetchSingle(ctx, k, cycleMetadata)
			if err != nil {
				d.log.Errorf("Error running fetcher for key %s: %v", k, err)
			}
		}(key)
	}

	d.wg.Wait()
	d.log.Infof("Manager finished waiting and sending data after %d milliseconds", time.Since(start).Milliseconds())
	d.log.Infof("Cycle %s resource fetching has ended", cycleId.String())
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
		return fmt.Errorf("fetcher %s reached a timeout after %v seconds", k, d.timeout.Seconds())
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

// Stop cleans up Data resources gracefully.
func (d *Data) Stop() {
	d.fetchers.Stop()
	close(d.stop)
	d.wg.Wait()
}
