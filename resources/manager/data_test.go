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
	"testing"
	"time"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/stretchr/testify/suite"
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

type DataTestSuite struct {
	suite.Suite
	registry FetchersRegistry
	opts     goleak.Option
	ctx      context.Context
}

const timeout = 2 * time.Second

func TestDataTestSuite(t *testing.T) {
	suite.Run(t, new(DataTestSuite))
}

func (s *DataTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.opts = goleak.IgnoreCurrent()
	s.registry = NewFetcherRegistry()
}

func (s *DataTestSuite) TearDownTest() {
	// Verify no goroutines are leaking. Safest to keep this on top of the function.
	// Go defers are implemented as a LIFO stack. This should be the last one to run.
	goleak.VerifyNone(s.T(), s.opts)
}

func (s *DataTestSuite) TestDataRun() {
	fetcherCount := 10
	interval := 10 * time.Second

	registerNFetchers(s.T(), s.registry, fetcherCount)
	d, err := NewData(interval, timeout, s.registry)
	s.NoError(err)

	err = d.Run(s.ctx)
	s.NoError(err)

	defer d.Stop()

	o := d.Output()
	state := <-o

	s.Equal(fetcherCount, len(state))

	for i := 0; i < fetcherCount; i++ {
		key := fmt.Sprint(i)

		val, ok := state[key]
		s.True(ok)
		s.Equal(fetchValue(i), val)
	}
}

func (s *DataTestSuite) TestDataRunNotSync() {
	iterations := 4
	interval := 2 * time.Second

	fetcher1Name := "delay_fetcher"
	fetcher1Delay := 1 * time.Second

	fetcher2Name := "num_fetcher"
	fetcher2Value := 1

	f1 := newDelayFetcher(fetcher1Delay)
	err := s.registry.Register(fetcher1Name, f1)
	s.NoError(err)

	f2 := newNumberFetcher(fetcher2Value)
	err = s.registry.Register(fetcher2Name, f2)
	s.NoError(err)

	d, err := NewData(interval, timeout, s.registry)
	s.NoError(err)

	err = d.Run(s.ctx)
	s.NoError(err)
	defer d.Stop()

	aggregated := make([]ResourceMap, iterations)
	for i := 0; i < iterations; i++ {
		aggregated[i] = <-d.Output()
	}

	s.Equal(iterations, len(aggregated))

	for i := 0; i < iterations; i++ {
		iterationResources := aggregated[i]
		s.NotEmpty(iterationResources, "iteration %d failed", i)

		fetcherResources, ok := iterationResources[fetcher1Name]
		s.True(ok, "iteration %d failed", i)

		s.Equal(fetchValue(int(fetcher1Delay.Seconds())), fetcherResources, "iteration %d failed", i)

		fetcherResources, ok = iterationResources[fetcher2Name]
		s.True(ok, "iteration %d failed", i)

		s.Equal(fetchValue(fetcher2Value), fetcherResources, "iteration %d failed", i)
	}
}

func (s *DataTestSuite) TestDataRunPanic() {
	interval := 3 * time.Second
	fetcherMessage := "fetcher got panic"
	fetcherName := "panic_fetcher"

	f := newPanicFetcher(fetcherMessage)
	err := s.registry.Register(fetcherName, f)
	s.NoError(err)

	d, err := NewData(interval, timeout, s.registry)
	s.NoError(err)

	err = d.Run(s.ctx)
	s.NoError(err)
	defer d.Stop()

	result := <-d.Output()
	s.Empty(result)
}

func (s *DataTestSuite) TestDataFetchSinglePanic() {
	interval := 3 * time.Second
	fetcherMessage := "fetcher got panic"
	fetcherName := "panic_fetcher"

	f := newPanicFetcher(fetcherMessage)
	err := s.registry.Register(fetcherName, f)
	s.NoError(err)

	d, err := NewData(interval, timeout, s.registry)
	s.NoError(err)

	res, err := d.fetchSingle(s.ctx, fetcherName)
	s.Error(err)
	s.Nil(res)
}

func (s *DataTestSuite) TestDataRunTimeout() {
	fetcherDelay := 4 * time.Second
	interval := 5 * time.Second
	fetcherName := "delay_fetcher"

	f := newDelayFetcher(fetcherDelay)
	err := s.registry.Register(fetcherName, f)
	s.NoError(err)

	d, err := NewData(interval, timeout, s.registry)
	s.NoError(err)

	err = d.Run(s.ctx)
	s.NoError(err)
	defer d.Stop()

	result := <-d.Output()
	s.Empty(result)
}

func (s *DataTestSuite) TestDataFetchSingleTimeout() {
	fetcherDelay := 4 * time.Second
	interval := 3 * time.Second
	fetcherName := "timeout_fetcher"

	f := newDelayFetcher(fetcherDelay)
	err := s.registry.Register(fetcherName, f)
	s.NoError(err)

	d, err := NewData(interval, timeout, s.registry)
	s.NoError(err)

	res, err := d.fetchSingle(s.ctx, fetcherName)
	s.Error(err)
	s.Nil(res)
}

func (s *DataTestSuite) TestDataRunShouldNotRun() {
	fetcherVal := 4
	interval := 5 * time.Second
	fetcherName := "not_run_fetcher"
	fetcherConditionName := "false_condition"

	f := newNumberFetcher(fetcherVal)
	c := newBoolFetcherCondition(false, fetcherConditionName)
	err := s.registry.Register(fetcherName, f, c)
	s.NoError(err)

	d, err := NewData(interval, timeout, s.registry)
	s.NoError(err)

	err = d.Run(s.ctx)
	s.NoError(err)
	defer d.Stop()

	result := <-d.Output()
	s.Empty(result)
}
