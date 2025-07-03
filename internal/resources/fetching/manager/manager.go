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

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/infra/observability"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/fetching/registry"
)

const scopeName = "github.com/elastic/cloudbeat/internal/resources/fetching/manager"

type Manager struct {
	log *clog.Logger

	// Duration of a single fetcher
	timeout time.Duration

	// Duration between two consecutive cycles
	interval time.Duration

	fetcherRegistry registry.Registry

	ctx    context.Context //nolint:containedctx
	cancel context.CancelFunc
}

func NewManager(ctx context.Context, log *clog.Logger, interval time.Duration, timeout time.Duration, fetchers registry.Registry) (*Manager, error) {
	ctx, cancel := context.WithCancel(ctx)

	return &Manager{
		log:             log,
		timeout:         timeout,
		interval:        interval,
		fetcherRegistry: fetchers,
		ctx:             ctx,
		cancel:          cancel,
	}, nil
}

// Run starts all configured fetchers to collect resources.
func (m *Manager) Run() {
	go m.fetchAndSleep(m.ctx)
}

func (m *Manager) Stop() {
	m.cancel()
	m.fetcherRegistry.Stop()
}

func (m *Manager) fetchAndSleep(ctx context.Context) {
	counter, err := observability.MeterFromContext(ctx, scopeName).Int64Counter("cloudbeat.fetcher.manager.cycles")
	if err != nil {
		m.log.Errorf("Failed to create fetcher manager cycles counter: %v", err)
	}

	// set immediate exec for first time run
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			m.log.Info("Fetchers manager canceled")
			return
		case <-timer.C:
			// update the interval
			timer.Reset(m.interval)
			counter.Add(ctx, 1)
			// this is blocking so the stop will not be called until all the fetchers are finished
			// in case there is a blocking fetcher it will halt (til the m.timeout)
			go m.fetchIteration(ctx)
		}
	}
}

// fetchIteration waits for all the registered fetchers and trigger them to fetch relevant resources.
// The function must not get called in parallel.
func (m *Manager) fetchIteration(ctx context.Context) {
	ctx, span := observability.StartSpan(
		ctx,
		scopeName,
		"Fetch Iteration",
		trace.WithAttributes(attribute.String("transaction.type", "request")),
	)
	defer span.End()
	logger := m.log.WithSpanContext(span.SpanContext())

	m.fetcherRegistry.Update()
	logger.Infof("Manager triggered fetching for %d fetchers", len(m.fetcherRegistry.Keys()))

	start := time.Now()

	seq := time.Now().Unix()
	logger.Infof("Cycle %d has started", seq)
	wg := &sync.WaitGroup{}
	for _, key := range m.fetcherRegistry.Keys() {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			err := m.fetchSingle(ctx, k, cycle.Metadata{Sequence: seq})
			if err != nil {
				logger.Errorf("Error running fetcher for key %s: %v", k, err)
			}
		}(key)
	}

	wg.Wait()
	logger.Infof("Manager finished waiting and sending data after %d milliseconds", time.Since(start).Milliseconds())
	logger.Infof("Cycle %d resource fetching has ended", seq)
}

func (m *Manager) fetchSingle(ctx context.Context, k string, cycleMetadata cycle.Metadata) error {
	if !m.fetcherRegistry.ShouldRun(k) {
		return nil
	}

	ctx, span := observability.StartSpan(ctx, scopeName, "Fetch "+k)
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	// The buffer is required to avoid go-routine leaks in a case a fetcher timed out
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		errCh <- m.fetchProtected(ctx, k, cycleMetadata)
	}()

	select {
	case <-ctx.Done():
		switch ctx.Err() {
		case context.DeadlineExceeded:
			return fmt.Errorf("fetcher %s reached a timeout after %v seconds", k, m.timeout.Seconds())
		case context.Canceled:
			return fmt.Errorf("fetcher %s %s", k, ctx.Err().Error())
		default:
			return fmt.Errorf("fetcher %s failed with an unknown error: %v", k, ctx.Err())
		}

	case err := <-errCh:
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}
}

// fetchProtected protect the fetching goroutine from getting panic
func (m *Manager) fetchProtected(ctx context.Context, k string, metadata cycle.Metadata) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("fetcher %s recovered from panic: %v", k, r)
		}
	}()

	return m.fetcherRegistry.Run(ctx, k, metadata)
}
