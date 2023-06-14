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
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
	"time"

	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/elastic/cloudbeat/resources/fetchersManager/registry"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
)

type DelayFetcher struct {
	delay         time.Duration
	stopCalled    bool
	resourceCh    chan fetching.ResourceInfo
	wg            *sync.WaitGroup
	isRunningChan chan bool
	err           error
	execCounter   int
}

func newDelayFetcher(delay time.Duration, ch chan fetching.ResourceInfo, wg *sync.WaitGroup, isRunningChan chan bool) *DelayFetcher {
	return &DelayFetcher{delay, false, ch, wg, isRunningChan, nil, 0}
}

func (f *DelayFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) (err error) {
	f.execCounter++
	defer f.wg.Done()

	select {
	case <-ctx.Done():
		err = ctx.Err()
		f.err = err
		f.isRunningChan <- false
		return
	case <-time.After(f.delay):
		f.resourceCh <- fetching.ResourceInfo{
			Resource:      fetchers.ConfigResource{},
			CycleMetadata: cMetadata,
		}
		return nil
	}
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

func newPanicFetcher(message string, ch chan fetching.ResourceInfo, wg *sync.WaitGroup) fetching.Fetcher {
	return &PanicFetcher{message, false, ch, wg}
}

func (f *PanicFetcher) Fetch(context.Context, fetching.CycleMetadata) error {
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
	registry   registry.MockFetchersRegistry
	dataLayer  *Data
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
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
	s.registry = registry.MockFetchersRegistry{}
	s.wg = &sync.WaitGroup{}

	s.dataLayer = &Data{
		log:        s.log,
		timeout:    2 * time.Second,
		interval:   5 * time.Second,
		fetchers:   &s.registry,
		stop:       make(chan struct{}),
		stopNotice: make(chan time.Duration),
	}
}

func (s *DataTestSuite) TearDownTest() {
	// Verify no goroutines are leaking. Safest to keep this on top of the function.
	// Go defers are implemented as a LIFO stack. This should be the last one to run.
	goleak.VerifyNone(s.T(), s.opts)
}

//func (s *DataTestSuite) TestDataRun() {
//	fetcherCount := 10
//	interval := 10 * time.Second
//
//	s.wg.Add(fetcherCount)
//
//	fetchersManager.RegisterNFetchers(s.T(), s.registry, fetcherCount, s.resourceCh, s.wg)
//	d, err := NewData(s.log, interval, timeout, s.registry)
//	s.NoError(err)
//
//	stop := d.Run(s.ctx)
//	defer stop(context.Background(), time.Second)
//	s.wg.Wait() // waiting for all fetchers to complete
//
//	results := testhelper.CollectResources(s.resourceCh)
//	s.Equal(fetcherCount, len(results))
//}

func (s *DataTestSuite) TestDataRunPanic() {
	interval := 3 * time.Second
	fetcherMessage := "fetcher got panic"
	fetcherName := "panic_fetcher"
	mock1 := registry.MockFetchersRegistry{}
	mock1.EXPECT().Keys().Return([]string{fetcherName})
	mock1.EXPECT().ShouldRun(mock.Anything).Return(true)
	mock1.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Panic(fetcherMessage)
	mock1.EXPECT().Stop().Return()

	d := &Data{
		log:        s.log,
		timeout:    timeout,
		interval:   interval,
		fetchers:   &mock1,
		stop:       make(chan struct{}),
		stopNotice: make(chan time.Duration),
	}

	stop := d.Run(s.ctx)
	defer stop(context.Background(), time.Second)
	mock1.AssertNumberOfCalls(s.T(), "Keys", 1)
}

//func (s *DataTestSuite) TestDataRunTimeout() {
//	fetcherDelay := 4 * time.Second
//	interval := 5 * time.Second
//	fetcherName := "delay_fetcher"
//
//	s.registry.EXPECT().Keys().Return([]string{fetcherName})
//	s.registry.EXPECT().ShouldRun(mock.Anything).Return(true)
//	s.registry.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).WaitUntil(time.After(fetcherDelay))
//	s.registry.EXPECT().Stop().Return()
//
//	d, err := NewData(s.log, interval, timeout, s.registry)
//	s.NoError(err)
//
//	stop := d.Run(s.ctx)
//	s.NoError(err)
//	defer stop(s.ctx, time.Second)
//	s.registry.AssertNumberOfCalls(s.T(), "Run", 1)
//	//s.registry.AssertNumberOfCalls(s.T(), "Stop", 1)
//	//s.registry.AssertNumberOfCalls(s.T(), "Keys", 2)
//	//s.registry.AssertNumberOfCalls(s.T(), "ShouldRun", 1)
//
//}

//
//func (s *DataTestSuite) TestDataFetchSingleTimeout() {
//	fetcherDelay := 4 * time.Second
//	interval := 3 * time.Second
//	fetcherName := "timeout_fetcher"
//
//	s.wg.Add(1)
//	f := newDelayFetcher(fetcherDelay, s.resourceCh, s.wg, make(chan bool, 1))
//	err := s.registry.Register(fetcherName, f)
//	s.NoError(err)
//
//	d, err := NewData(s.log, interval, timeout, s.registry)
//	s.NoError(err)
//
//	err = d.fetchSingle(s.ctx, fetcherName, fetching.CycleMetadata{})
//	s.Error(err)
//}
//
//func (s *DataTestSuite) TestDataRunShouldNotRun() {
//	fetcherVal := 4
//	interval := 5 * time.Second
//	fetcherName := "not_run_fetcher"
//	fetcherConditionName := "false_condition"
//
//	f := fetchersManager.NewNumberFetcher(fetcherVal, s.resourceCh, s.wg)
//	c := fetchersManager.NewBoolFetcherCondition(false, fetcherConditionName)
//	err := s.registry.Register(fetcherName, f, c)
//	s.NoError(err)
//
//	d, err := NewData(s.log, interval, timeout, s.registry)
//	s.NoError(err)
//
//	stop := d.Run(s.ctx)
//	s.NoError(err)
//	defer stop(context.Background(), time.Second)
//
//	// Fetcher did not run, we can not wait for sync.done() to be called.
//	var results []fetching.ResourceInfo
//	select {
//	case result := <-s.resourceCh:
//		results = append(results, result)
//	case <-time.Tick(interval):
//		break
//	}
//
//	s.Empty(results)
//}
//
//func (s *DataTestSuite) TestDataStop() {
//	interval := 30 * time.Second
//	fetcherName := "run_fetcher"
//	fetcherConditionName := "true_condition"
//
//	isRunningChan := make(chan bool, 1)
//	f := newDelayFetcher(time.Minute, s.resourceCh, s.wg, isRunningChan)
//	c := fetchersManager.NewBoolFetcherCondition(true, fetcherConditionName)
//	err := s.registry.Register(fetcherName, f, c)
//	s.NoError(err)
//
//	d, err := NewData(s.log, interval, time.Second*5, s.registry)
//	s.NoError(err)
//
//	stop := d.Run(context.Background())
//	time.Sleep(1 * time.Second)
//	stop(context.Background(), time.Second)
//	time.Sleep(3 * time.Second)
//	s.True(f.stopCalled)
//	s.False(<-isRunningChan, "fetcher should not be running")
//	s.Equal(context.Canceled, f.err)
//}
//
//func (s *DataTestSuite) TestDataStopWithTimeout() {
//	interval := 30 * time.Second
//	fetcherName := "run_fetcher"
//	fetcherConditionName := "true_condition"
//
//	isRunningChan := make(chan bool, 1)
//	f := newDelayFetcher(time.Minute, s.resourceCh, s.wg, isRunningChan)
//	c := fetchersManager.NewBoolFetcherCondition(true, fetcherConditionName)
//	err := s.registry.Register(fetcherName, f, c)
//	s.NoError(err)
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
//	defer cancel()
//	d, err := NewData(s.log, interval, time.Second*5, s.registry)
//	s.NoError(err)
//
//	d.Run(ctx)
//	time.Sleep(2 * time.Second)
//	s.False(<-isRunningChan, "fetcher should not be running")
//	s.Equal(context.DeadlineExceeded, f.err)
//}
//
//func (s *DataTestSuite) TestDataStopWithGracefulShutdown() {
//	interval := 30 * time.Second
//	fetcherName := "run_fetcher"
//	fetcherConditionName := "true_condition"
//
//	isRunningChan := make(chan bool, 1)
//	f := newDelayFetcher(time.Minute, s.resourceCh, s.wg, isRunningChan)
//	c := fetchersManager.NewBoolFetcherCondition(true, fetcherConditionName)
//	err := s.registry.Register(fetcherName, f, c)
//	s.NoError(err)
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
//	defer cancel()
//	d, err := NewData(s.log, interval, time.Second*5, s.registry)
//	s.NoError(err)
//
//	stop := d.Run(ctx)
//	time.Sleep(2 * time.Second)
//
//	stop(ctx, time.Second)
//	time.Sleep(2 * time.Second)
//
//	s.False(<-isRunningChan, "fetcher should not be running")
//	s.Equal(context.Canceled, f.err)
//}
//
//func (s *DataTestSuite) TestDataStopWithNoticePeriod() {
//	fetcherName := "run_fetcher"
//	fetcherConditionName := "true_condition"
//
//	isRunningChan := make(chan bool, 1)
//	f := newDelayFetcher(time.Millisecond, s.resourceCh, s.wg, isRunningChan)
//	c := fetchersManager.NewBoolFetcherCondition(true, fetcherConditionName)
//	err := s.registry.Register(fetcherName, f, c)
//	s.NoError(err)
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
//	defer cancel()
//	d, err := NewData(s.log, 500*time.Millisecond, time.Second*5, s.registry)
//	s.NoError(err)
//
//	stop := d.Run(ctx)
//	time.Sleep(2 * time.Second)
//
//	stop(ctx, time.Second)
//	time.Sleep(2 * time.Second)
//	s.LessOrEqual(f.execCounter, 4)
//}
//
//func (s *DataTestSuite) TestDataDoubleStop() {
//	fetcherName := "run_fetcher"
//	fetcherConditionName := "true_condition"
//
//	isRunningChan := make(chan bool, 1)
//	f := newDelayFetcher(time.Millisecond, s.resourceCh, s.wg, isRunningChan)
//	c := fetchersManager.NewBoolFetcherCondition(false, fetcherConditionName)
//	err := s.registry.Register(fetcherName, f, c)
//	s.NoError(err)
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
//	defer cancel()
//	d, err := NewData(s.log, 500*time.Millisecond, time.Second*5, s.registry)
//	s.NoError(err)
//
//	stop := d.Run(ctx)
//	time.Sleep(2 * time.Second)
//
//	stop(ctx, time.Second)
//	stop(ctx, time.Second)
//}
