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
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"

	"github.com/elastic/cloudbeat/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/resources/fetching/registry"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
)

type ManagerTestSuite struct {
	suite.Suite
	registry *registry.MockRegistry
	opts     goleak.Option
}

const timeout = 2 * time.Second

func TestManagerTestSuite(t *testing.T) {
	s := new(ManagerTestSuite)

	suite.Run(t, s)
}

func (s *ManagerTestSuite) SetupTest() {
	s.opts = goleak.IgnoreCurrent()
	s.registry = &registry.MockRegistry{}
}

func (s *ManagerTestSuite) TearDownTest() {
	// Verify no goroutines are leaking. Safest to keep this on top of the function.
	// Go defers are implemented as a LIFO stack. This should be the last one to run.
	goleak.VerifyNone(s.T(), s.opts)
}

func (s *ManagerTestSuite) TestManagerRun() {
	interval := 5 * time.Second
	fetcherName := "test_fetcher"

	s.registry.EXPECT().Keys().Return([]string{fetcherName, fetcherName, fetcherName, fetcherName, fetcherName}).Twice()
	s.registry.EXPECT().ShouldRun(mock.Anything).Return(true).Times(5)
	s.registry.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(5)
	s.registry.EXPECT().Update().Once()
	s.registry.EXPECT().Stop().Once()

	m, err := NewManager(context.Background(), testhelper.NewLogger(s.T()), interval, timeout, s.registry)
	s.Require().NoError(err)

	m.Run()
	waitForACycleToEnd(interval)
	m.Stop()

	s.registry.AssertExpectations(s.T())
}

func (s *ManagerTestSuite) TestManagerRunPanic() {
	interval := 3 * time.Second
	fetcherMessage := "fetcher got panic"
	fetcherName := "panic_fetcher"

	s.registry.EXPECT().Keys().Return([]string{fetcherName}).Twice()
	s.registry.EXPECT().ShouldRun(mock.Anything).Return(true).Once()
	s.registry.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Panic(fetcherMessage).Once()
	s.registry.EXPECT().Update().Once()
	s.registry.EXPECT().Stop().Once()

	m, err := NewManager(context.Background(), testhelper.NewLogger(s.T()), interval, timeout, s.registry)
	s.Require().NoError(err)

	m.Run()
	waitForACycleToEnd(interval)
	m.Stop()

	s.registry.AssertExpectations(s.T())
}

func (s *ManagerTestSuite) TestManagerRunTimeout() {
	fetcherDelay := 4 * time.Second
	interval := 5 * time.Second
	fetcherName := "delay_fetcher"

	s.registry.EXPECT().Keys().Return([]string{fetcherName}).Twice()
	s.registry.EXPECT().ShouldRun(mock.Anything).Return(true).Once()
	s.registry.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).WaitUntil(time.After(fetcherDelay)).Once()
	s.registry.EXPECT().Update().Once()
	s.registry.EXPECT().Stop().Once()

	m, err := NewManager(context.Background(), testhelper.NewLogger(s.T()), interval, timeout, s.registry)
	s.Require().NoError(err)

	m.Run()
	waitForACycleToEnd(interval)

	m.Stop()

	s.registry.AssertExpectations(s.T())
}

func (s *ManagerTestSuite) TestManagerFetchSingleTimeout() {
	fetcherDelay := 4 * time.Second
	interval := 3 * time.Second
	fetcherName := "timeout_fetcher"

	s.registry.EXPECT().ShouldRun(mock.Anything).Return(true).Once()
	s.registry.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Call.Return(func(ctx context.Context, key string, metadata cycle.Metadata) {
		select {
		case <-ctx.Done():
			return
		case <-time.After(fetcherDelay):
			return
		}
	}).Once()

	m, err := NewManager(context.Background(), testhelper.NewLogger(s.T()), interval, timeout, s.registry)
	s.Require().NoError(err)

	err = m.fetchSingle(context.Background(), fetcherName, cycle.Metadata{})
	s.Require().Error(err)
	s.registry.AssertExpectations(s.T())
}

func (s *ManagerTestSuite) TestManagerRunShouldNotRun() {
	interval := 5 * time.Second
	fetcherName := "not_run_fetcher"

	s.registry.EXPECT().Keys().Return([]string{fetcherName}).Twice()
	s.registry.EXPECT().ShouldRun(mock.Anything).Return(false).Once()
	s.registry.EXPECT().Update().Once()
	s.registry.EXPECT().Stop().Once()

	d, err := NewManager(context.Background(), testhelper.NewLogger(s.T()), interval, timeout, s.registry)
	s.Require().NoError(err)

	d.Run()
	waitForACycleToEnd(interval)
	d.Stop()
	s.registry.AssertExpectations(s.T())
}

func (s *ManagerTestSuite) TestManagerStop() {
	interval := 30 * time.Second
	fetcherName := "run_fetcher"

	s.registry.EXPECT().Keys().Return([]string{fetcherName}).Twice()
	s.registry.EXPECT().ShouldRun(mock.Anything).Return(true).Once()
	s.registry.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	s.registry.EXPECT().Update().Once()
	s.registry.EXPECT().Stop().Once()

	m, err := NewManager(context.Background(), testhelper.NewLogger(s.T()), interval, time.Second*5, s.registry)
	s.Require().NoError(err)

	m.Run()
	waitForACycleToEnd(2 * time.Second)
	m.Stop()
	time.Sleep(2 * time.Second)

	s.registry.AssertExpectations(s.T())
	s.Require().EqualError(context.Canceled, m.ctx.Err().Error())
}

func (s *ManagerTestSuite) TestManagerStopWithTimeout() {
	interval := 30 * time.Second
	fetcherName := "run_fetcher"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	s.registry.EXPECT().Keys().Return([]string{fetcherName}).Twice()
	s.registry.EXPECT().ShouldRun(mock.Anything).Return(true).Once()
	s.registry.EXPECT().Run(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	s.registry.EXPECT().Update().Once()

	m, err := NewManager(ctx, testhelper.NewLogger(s.T()), interval, time.Second*5, s.registry)
	s.Require().NoError(err)

	m.Run()
	time.Sleep(3 * time.Second)
	s.Require().EqualError(context.DeadlineExceeded, ctx.Err().Error())
	s.registry.AssertExpectations(s.T())
}

func waitForACycleToEnd(interval time.Duration) {
	time.Sleep(interval - 1*time.Second)
}
