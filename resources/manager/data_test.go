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
	"sync"
	"testing"
	"time"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-libs/logp"
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
	wg         *sync.WaitGroup
}

type PanicFetcher struct {
	message    string
	stopCalled bool
	resourceCh chan fetching.ResourceInfo
	wg         *sync.WaitGroup
}

type DataFetcher struct {
	message    string
	stopCalled bool
	resourceCh chan fetching.ResourceInfo
	wg         *sync.WaitGroup
}

func newPanicFetcher(message string, ch chan fetching.ResourceInfo, wg *sync.WaitGroup) fetching.Fetcher {
	return &PanicFetcher{message, false, ch, wg}
}

func newDataFetcher(message string, ch chan fetching.ResourceInfo, wg *sync.WaitGroup) fetching.Fetcher {
	return &DataFetcher{message, false, ch, wg}
}

func (f *PanicFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	defer f.wg.Done()
	panic(f.message)
}

func (f *PanicFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	defer f.wg.Done()
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
	wg         *sync.WaitGroup
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
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
	s.wg = &sync.WaitGroup{}
}

func (s *DataTestSuite) TearDownTest() {
	// Verify no goroutines are leaking. Safest to keep this on top of the function.
	// Go defers are implemented as a LIFO stack. This should be the last one to run.
	goleak.VerifyNone(s.T(), s.opts)
}

func (s *DataTestSuite) TestDataRun() {
	fetcherCount := 10
	interval := 10 * time.Second

	s.wg.Add(fetcherCount)

	registerNFetchers(s.T(), s.registry, fetcherCount, s.resourceCh, s.wg)
	d, err := NewData(s.log, interval, timeout, s.registry)
	s.NoError(err)

	d.Run(s.ctx)
	defer d.Stop()
	s.wg.Wait() // waiting for all fetchers to complete

	results := testhelper.CollectResources(s.resourceCh)
	s.Equal(fetcherCount, len(results))
}

func (s *DataTestSuite) TestDataRunPanic() {
	interval := 3 * time.Second
	fetcherMessage := "fetcher got panic"
	fetcherName := "panic_fetcher"

	s.wg.Add(1)
	f := newPanicFetcher(fetcherMessage, s.resourceCh, s.wg)
	err := s.registry.Register(fetcherName, f)
	s.NoError(err)

	d, err := NewData(s.log, interval, timeout, s.registry)
	s.NoError(err)

	err = d.Run(s.ctx)
	s.NoError(err)
	defer d.Stop()

	s.wg.Wait()
	results := testhelper.CollectResources(s.resourceCh)

	s.Empty(results)
}

func (s *DataTestSuite) TestDataFetchSinglePanic() {
	interval := 3 * time.Second
	fetcherMessage := "fetcher got panic"
	fetcherName := "panic_fetcher"

	f := newPanicFetcher(fetcherMessage, s.resourceCh, nil)
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

	results := testhelper.CollectResources(s.resourceCh)

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

	err = d.fetchSingle(s.ctx, fetcherName, fetching.CycleMetadata{})
	s.Error(err)
}

func (s *DataTestSuite) TestDataRunShouldNotRun() {
	fetcherVal := 4
	interval := 5 * time.Second
	fetcherName := "not_run_fetcher"
	fetcherConditionName := "false_condition"

	s.wg.Add(1)
	f := newNumberFetcher(fetcherVal, s.resourceCh, s.wg)
	c := newBoolFetcherCondition(false, fetcherConditionName)
	err := s.registry.Register(fetcherName, f, c)
	s.NoError(err)

	d, err := NewData(s.log, interval, timeout, s.registry)
	s.NoError(err)

	err = d.Run(s.ctx)
	s.NoError(err)
	defer d.Stop()

	var results []fetching.ResourceInfo
	select {
	case result := <-s.resourceCh:
		results = append(results, result)
	case <-time.Tick(2 * time.Second):
		s.wg.Done()
	}

	s.Empty(results)
}
