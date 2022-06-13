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
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"testing"
	"time"

	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
)

type DelayFetcher struct {
	delay      time.Duration
	stopCalled bool
	resourceCh chan fetching.ResourceInfo
}

func newDelayFetcher(delay time.Duration, ch chan fetching.ResourceInfo) fetching.Fetcher {
	return &DelayFetcher{delay, false, ch}
}

func (f *DelayFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	select {
	case <-ctx.Done():
		fmt.Errorf("reached timeout")
	case <-time.After(f.delay):
		f.resourceCh <- fetching.ResourceInfo{
			Resource:      fetchValue(int(f.delay.Seconds())),
			CycleMetadata: cMetadata,
		}
	}

	return nil
}

func (f *DelayFetcher) Stop() {
	f.stopCalled = true
}

type PanicFetcher struct {
	message    string
	stopCalled bool
	resourceCh chan fetching.ResourceInfo
}

func newPanicFetcher(message string, ch chan fetching.ResourceInfo) fetching.Fetcher {
	return &PanicFetcher{message, false, ch}
}

func (f *PanicFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	panic(f.message)
}

func (f *PanicFetcher) Stop() {
	f.stopCalled = true
}

type DataTestSuite struct {
	suite.Suite

	ctx        context.Context
	log        *logp.Logger
	registry   FetchersRegistry
	opts       goleak.Option
	resourceCh chan fetching.ResourceInfo
}

const timeout = 2 * time.Second

func TestDataTestSuite(t *testing.T) {
	s := new(DataTestSuite)
	s.log = logp.NewLogger("cloudbeat_data_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *DataTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.opts = goleak.IgnoreCurrent()
	s.registry = NewFetcherRegistry(s.log)
	s.resourceCh = make(chan fetching.ResourceInfo)
}

func (s *DataTestSuite) TearDownTest() {
	// Verify no goroutines are leaking. Safest to keep this on top of the function.
	// Go defers are implemented as a LIFO stack. This should be the last one to run.
	goleak.VerifyNone(s.T(), s.opts)
}

func (s *DataTestSuite) TestDataRun() {
	fetcherCount := 10
	interval := 10 * time.Second

	registerNFetchers(s.T(), s.registry, fetcherCount, s.resourceCh)
	d, err := NewData(s.log, interval, timeout, s.registry)
	s.NoError(err)

	d.Run(s.ctx)
	defer d.Stop()

	results := testhelper.WaitForResources(s.resourceCh, fetcherCount, 2)
	s.Equal(fetcherCount, len(results))
}

func (s *DataTestSuite) TestDataRunPanic() {
	interval := 3 * time.Second
	fetcherMessage := "fetcher got panic"
	fetcherName := "panic_fetcher"

	f := newPanicFetcher(fetcherMessage, s.resourceCh)
	err := s.registry.Register(fetcherName, f)
	s.NoError(err)

	d, err := NewData(s.log, interval, timeout, s.registry)
	s.NoError(err)

	err = d.Run(s.ctx)
	s.NoError(err)
	defer d.Stop()

	results := testhelper.WaitForResources(s.resourceCh, 1, 2)

	s.Empty(results)
}

func (s *DataTestSuite) TestDataFetchSinglePanic() {
	interval := 3 * time.Second
	fetcherMessage := "fetcher got panic"
	fetcherName := "panic_fetcher"

	f := newPanicFetcher(fetcherMessage, s.resourceCh)
	err := s.registry.Register(fetcherName, f)
	s.NoError(err)

	d, err := NewData(s.log, interval, timeout, s.registry)
	s.NoError(err)

	err = d.fetchSingle(s.ctx, fetcherName, fetching.CycleMetadata{})
	s.Error(err)
}

func (s *DataTestSuite) TestDataRunTimeout() {
	fetcherDelay := 4 * time.Second
	interval := 5 * time.Second
	fetcherName := "delay_fetcher"

	f := newDelayFetcher(fetcherDelay, s.resourceCh)
	err := s.registry.Register(fetcherName, f)
	s.NoError(err)

	d, err := NewData(s.log, interval, timeout, s.registry)
	s.NoError(err)

	err = d.Run(s.ctx)
	s.NoError(err)
	defer d.Stop()

	results := testhelper.WaitForResources(s.resourceCh, 1, 3 /* timeout + 1 second */)

	s.Empty(results)
}

func (s *DataTestSuite) TestDataFetchSingleTimeout() {
	fetcherDelay := 4 * time.Second
	interval := 3 * time.Second
	fetcherName := "timeout_fetcher"

	f := newDelayFetcher(fetcherDelay, s.resourceCh)
	err := s.registry.Register(fetcherName, f)
	s.NoError(err)

	d, err := NewData(s.log, interval, timeout, s.registry)
	s.NoError(err)

	errCh := make(chan error)
	defer close(errCh)
	go func() {
		errCh <- d.fetchSingle(s.ctx, fetcherName, fetching.CycleMetadata{})
	}()

	s.Error(<-errCh)
}

func (s *DataTestSuite) TestDataRunShouldNotRun() {
	fetcherVal := 4
	interval := 5 * time.Second
	fetcherName := "not_run_fetcher"
	fetcherConditionName := "false_condition"

	f := newNumberFetcher(fetcherVal, s.resourceCh)
	c := newBoolFetcherCondition(false, fetcherConditionName)
	err := s.registry.Register(fetcherName, f, c)
	s.NoError(err)

	d, err := NewData(s.log, interval, timeout, s.registry)
	s.NoError(err)

	err = d.Run(s.ctx)
	s.NoError(err)
	defer d.Stop()

	results := testhelper.WaitForResources(s.resourceCh, 1, 2)
	
	s.Empty(results)
}
