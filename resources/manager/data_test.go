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
	"reflect"
	"testing"
	"time"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

type DelayFetcher struct {
	delay      time.Duration
	stopCalled bool
}

func newDelayFetcher(delay time.Duration) fetching.Fetcher {
	return &DelayFetcher{delay, false}
}

func (f *DelayFetcher) Fetch(ctx context.Context) ([]fetching.Resource, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("reached timeout")
	case <-time.After(f.delay):
		return fetchValue(int(f.delay.Seconds())), nil
	}
}

func (f *DelayFetcher) Stop() {
	f.stopCalled = true
}

type PanicFetcher struct {
	message    string
	stopCalled bool
}

func newPanicFetcher(message string) fetching.Fetcher {
	return &PanicFetcher{message, false}
}

func (f *PanicFetcher) Fetch(ctx context.Context) ([]fetching.Resource, error) {
	panic(f.message)
}

func (f *PanicFetcher) Stop() {
	f.stopCalled = true
}

const (
	fetcherCount = 10
)

func TestDataRun(t *testing.T) {
	timeout := time.Minute
	interval := 10 * time.Second
	opts := goleak.IgnoreCurrent()

	// Verify no goroutines are leaking. Safest to keep this on top of the function.
	// Go defers are implemented as a LIFO stack. This should be the last one to run.
	defer goleak.VerifyNone(t, opts)

	reg := NewFetcherRegistry()
	registerNFetchers(t, reg, fetcherCount)
	d, err := NewData(interval, timeout, reg)
	if err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = d.Run(ctx)
	if err != nil {
		return
	}
	defer d.Stop(ctx, cancel)

	o := d.Output()
	state := <-o

	if len(state) < fetcherCount {
		t.Errorf("expected %d keys but got %d", fetcherCount, len(state))
	}

	for i := 0; i < fetcherCount; i++ {
		key := fmt.Sprint(i)

		val, ok := state[key]
		if !ok {
			t.Errorf("expected key %s but not found", key)
		}

		if !reflect.DeepEqual(val, fetchValue(i)) {
			t.Errorf("expected key %s to have value %v but got %v", key, fetchValue(i), val)
		}
	}
}

func TestDataRunNotSync(t *testing.T) {
	iterations := 4
	timeout := time.Minute
	interval := 3 * time.Second
	fetcherDelay := 1 * time.Second
	fetcherName := "delay_fetcher"
	opts := goleak.IgnoreCurrent()

	// Verify no goroutines are leaking. Safest to keep this on top of the function.
	// Go defers are implemented as a LIFO stack. This should be the last one to run.
	defer goleak.VerifyNone(t, opts)
	f := newDelayFetcher(fetcherDelay)
	reg := NewFetcherRegistry()
	err := reg.Register(fetcherName, f)
	assert.NoError(t, err)

	d, err := NewData(interval, timeout, reg)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = d.Run(ctx)
	assert.NoError(t, err)
	defer d.Stop(ctx, cancel)

	aggregated := make([]ResourceMap, iterations)
	for i := 0; i < iterations; i++ {
		aggregated[i] = <-d.Output()
	}

	assert.Equal(t, iterations, len(aggregated))

	for i := 0; i < iterations; i++ {
		iterationResources := aggregated[i]
		assert.NotEmpty(t, iterationResources, "iteration %d failed", i)

		fetcherResources, ok := iterationResources[fetcherName]
		assert.True(t, ok, "iteration %d failed", i)

		assert.Equal(t, fetcherResources, fetchValue(int(fetcherDelay.Seconds())), "iteration %d failed", i)
	}
}

func TestDataRunPanic(t *testing.T) {
	iterations := 2
	timeout := time.Minute
	interval := 3 * time.Second
	fetcherMessage := "fetcher got panic"
	fetcherName := "panic_fetcher"
	opts := goleak.IgnoreCurrent()

	// Verify no goroutines are leaking. Safest to keep this on top of the function.
	// Go defers are implemented as a LIFO stack. This should be the last one to run.
	defer goleak.VerifyNone(t, opts)
	f := newPanicFetcher(fetcherMessage)
	reg := NewFetcherRegistry()
	err := reg.Register(fetcherName, f)
	assert.NoError(t, err)

	d, err := NewData(interval, timeout, reg)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = d.Run(ctx)
	assert.NoError(t, err)
	defer d.Stop(ctx, cancel)

	aggregated := make([]ResourceMap, iterations)
	for i := 0; i < iterations; i++ {
		aggregated[i] = <-d.Output()
	}

	assert.Equal(t, iterations, len(aggregated))

	for i := 0; i < iterations; i++ {
		iterationResources := aggregated[i]
		assert.Empty(t, iterationResources, "iteration %d failed", i)
	}
}

func TestDataRunFetchTimeout(t *testing.T) {
	timeout := 2 * time.Second
	fetcherDelay := 4 * time.Second
	interval := 5 * time.Second
	fetcherName := "delay_fetcher"
	opts := goleak.IgnoreCurrent()

	// Verify no goroutines are leaking. Safest to keep this on top of the function.
	// Go defers are implemented as a LIFO stack. This should be the last one to run.
	defer goleak.VerifyNone(t, opts)
	f := newDelayFetcher(fetcherDelay)
	reg := NewFetcherRegistry()
	err := reg.Register(fetcherName, f)
	assert.NoError(t, err)

	d, err := NewData(interval, timeout, reg)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = d.Run(ctx)
	assert.NoError(t, err)
	defer d.Stop(ctx, cancel)

	result := <-d.Output()
	assert.Empty(t, result)
}
