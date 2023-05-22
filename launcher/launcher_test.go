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

// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package launcher

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/management"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
)

var dummyBeaterName = "Dummybeat"

type beaterMock struct {
	cfg  *config.C
	done chan struct{}
}

func beaterMockCreator(_ *beat.Beat, cfg *config.C) (beat.Beater, error) {
	return &beaterMock{
		cfg:  cfg,
		done: make(chan struct{}),
	}, nil
}

func (m *beaterMock) Run(_ *beat.Beat) error {
	<-m.done
	return nil
}

func (m *beaterMock) Stop() {
	close(m.done)
}

type errorBeaterMock struct{}

func errorBeaterMockCreator(_ *beat.Beat, _ *config.C) (beat.Beater, error) {
	return &errorBeaterMock{}, nil
}

func errorBeaterCreator(_ *beat.Beat, _ *config.C) (beat.Beater, error) {
	return nil, errors.New("beater creation error")
}

func errorReloadBeaterCreator() func(b *beat.Beat, cfg *config.C) (beat.Beater, error) {
	shouldReturnError := false
	return func(b *beat.Beat, cfg *config.C) (beat.Beater, error) {
		if shouldReturnError {
			return errorBeaterCreator(b, cfg)
		}
		shouldReturnError = true
		return beaterMockCreator(b, cfg)
	}
}

func (m *errorBeaterMock) Run(_ *beat.Beat) error {
	time.Sleep(10 * time.Millisecond)
	return errors.New("some error")
}

func (m *errorBeaterMock) Stop() {
	panic("Error beater should not be stopped")
}

type panicBeaterMock struct{}

func panicBeaterMockCreator(_ *beat.Beat, _ *config.C) (beat.Beater, error) {
	return &panicBeaterMock{}, nil
}

func (m *panicBeaterMock) Run(_ *beat.Beat) error {
	panic("panicBeaterMock panics")
}

func (m *panicBeaterMock) Stop() {
}

type validatorMock struct {
	expected *config.C
}

func (v *validatorMock) Validate(cfg *config.C) error {
	var err error
	if !reflect.DeepEqual(cfg, v.expected) {
		err = fmt.Errorf("mock validation failed")
	}

	return err
}

type LauncherTestSuite struct {
	suite.Suite

	log  *logp.Logger
	opts goleak.Option
}

type launcherMocks struct {
	reloader   *MockReloader
	reloaderCh chan *config.C
	health     *MockHealth
	healthCh   chan error
	beat       *beat.Beat
	manager    *MockManager
	validator  Validator
}

func TestLauncherTestSuite(t *testing.T) {
	s := new(LauncherTestSuite)
	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}
	s.log = logp.NewLogger("cloudbeat_launcher_test_suite")

	s.opts = goleak.IgnoreCurrent()
	suite.Run(t, s)
}

func (s *LauncherTestSuite) InitMocks() *launcherMocks {
	mocks := launcherMocks{
		healthCh:   make(chan error),
		reloaderCh: make(chan *config.C, 10),
	}
	mocks.reloader = &MockReloader{}
	mocks.validator = &validatorMock{
		expected: config.MustNewConfigFrom(mapstr.M{"a": 1}),
	}
	mocks.health = &MockHealth{}

	mocks.manager = NewMockManager(s.T())
	mocks.beat = &beat.Beat{
		Manager: mocks.manager,
	}

	mocks.reloader.EXPECT().Channel().Return(mocks.reloaderCh)
	mocks.reloader.EXPECT().Stop()
	mocks.health.EXPECT().Channel().Return(mocks.healthCh)
	mocks.health.EXPECT().Stop()
	return &mocks
}

func (s *LauncherTestSuite) TearDownTest() {
	// Verify no goroutines are leaking. Safest to keep this on top of the function.
	// Go defers are implemented as a LIFO stack. This should be the last one to run.
	goleak.VerifyNone(s.T(), s.opts)
}

func (s *LauncherTestSuite) TestWaitForUpdates() {
	configA := config.MustNewConfigFrom(mapstr.M{
		"common":    "A",
		"specificA": "a",
		"commonArr": []string{"a"},
	})

	configB := config.MustNewConfigFrom(mapstr.M{
		"common":    "B",
		"specificB": "b",
		"commonArr": []string{"b", "b"},
	})

	configC := config.MustNewConfigFrom(mapstr.M{
		"common":    "C",
		"specificC": "c",
		"commonArr": []string{"c", "c", "c"},
	})

	expected1 := config.MustNewConfigFrom(mapstr.M{
		"common":    "C",
		"specificA": "a",
		"specificB": "b",
		"specificC": "c",
		"commonArr": []string{"c", "c", "c"},
	})

	expected2 := config.MustNewConfigFrom(mapstr.M{
		"common":    "A",
		"specificA": "a",
		"specificB": "b",
		"commonArr": []string{"a"},
	})

	expected3 := config.MustNewConfigFrom(mapstr.M{
		"common":    "B",
		"specificA": "a",
		"specificB": "b",
		"specificC": "c",
		"commonArr": []string{"b", "b"},
	})

	expected4 := config.MustNewConfigFrom(mapstr.M{
		"common":    "A",
		"specificA": "a",
		"specificC": "c",
		"commonArr": []string{"a"},
	})

	type incomingConfigs []struct {
		after  time.Duration
		config *config.C
	}

	testcases := []struct {
		name     string
		delay    time.Duration
		configs  incomingConfigs
		expected *config.C
	}{
		{
			"no updates",
			100 * time.Millisecond,
			incomingConfigs{},
			config.NewConfig(),
		},
		{
			"single update",
			100 * time.Millisecond,
			incomingConfigs{
				{40 * time.Millisecond, configC},
			},
			configC,
		},
		{
			"multiple updates A B A C",
			100 * time.Millisecond,
			incomingConfigs{
				{1 * time.Millisecond, configA},
				{1 * time.Millisecond, configB},
				{1 * time.Millisecond, configA},
				{40 * time.Millisecond, configC},
			},
			expected1,
		},
		{
			"multiple updates A B A",
			100 * time.Millisecond,
			incomingConfigs{
				{1 * time.Millisecond, configA},
				{40 * time.Millisecond, configB},
				{1 * time.Millisecond, configA},
			},
			expected2,
		},
		{
			"multiple updates A C B",
			100 * time.Millisecond,
			incomingConfigs{
				{1 * time.Millisecond, configA},
				{1 * time.Millisecond, configC},
				{1 * time.Millisecond, configB},
			},
			expected3,
		},
		{
			"multiple updates C C A",
			100 * time.Millisecond,
			incomingConfigs{
				{1 * time.Millisecond, configC},
				{1 * time.Millisecond, configC},
				{1 * time.Millisecond, configA},
			},
			expected4,
		},
		{
			"multiple updates no delay A B A C",
			100 * time.Millisecond,
			incomingConfigs{
				{0, configA},
				{0, configB},
				{0, configA},
				{0, configC},
			},
			expected1,
		},
		{
			"no updates immediate stop",
			0,
			incomingConfigs{},
			config.NewConfig(),
		},
	}

	for _, tcase := range testcases {
		s.Run(tcase.name, func() {
			mocks := s.InitMocks()
			sut, err := New(s.log, dummyBeaterName, mocks.reloader, mocks.health, nil, beaterMockCreator, config.NewConfig())
			s.NoError(err)

			go func(ic incomingConfigs) {
				for _, c := range ic {
					time.Sleep(c.after)
					mocks.reloaderCh <- c.config
				}

				time.Sleep(tcase.delay)
				sut.Stop()
			}(tcase.configs)

			err = sut.run()
			s.NoError(err)
			beater, ok := sut.beater.(*beaterMock)
			s.True(ok)
			s.Equal(tcase.expected, beater.cfg)
		})
	}
}

// TestErrorWaitForUpdates should not call sut.Stop as the launcher should stop without callling it
func (s *LauncherTestSuite) TestErrorWaitForUpdates() {
	configErr := config.MustNewConfigFrom(mapstr.M{
		"error": "true",
	})

	mocks := s.InitMocks()
	sut, err := New(s.log, dummyBeaterName, mocks.reloader, mocks.health, nil, errorReloadBeaterCreator(), config.NewConfig())
	s.NoError(err)

	go func() {
		time.Sleep(40 * time.Millisecond)
		mocks.reloaderCh <- configErr
	}()

	err = sut.run()
	s.Error(err)
}

func (s *LauncherTestSuite) TestLauncherValidator() {
	validConfig := config.MustNewConfigFrom(mapstr.M{"a": 1})
	invalidConfig := config.MustNewConfigFrom(mapstr.M{"a": 2})

	type incomingConfigs []struct {
		after  time.Duration
		config *config.C
	}

	testcases := []struct {
		name     string
		timeout  time.Duration
		configs  incomingConfigs
		expected *config.C
	}{
		{
			"no updates",
			5 * time.Millisecond,
			incomingConfigs{},
			nil,
		},
		{
			"valid update after timeout",
			5 * time.Millisecond,
			incomingConfigs{
				{10 * time.Millisecond, validConfig},
			},
			nil,
		},
		{
			"invalid update on time",
			10 * time.Millisecond,
			incomingConfigs{
				{5 * time.Millisecond, invalidConfig},
			},
			nil,
		},
		{
			"valid update on time",
			10 * time.Millisecond,
			incomingConfigs{
				{5 * time.Millisecond, validConfig},
			},
			validConfig,
		},
		{
			"invalid and later valid after timeout",
			40 * time.Millisecond,
			incomingConfigs{
				{1 * time.Millisecond, invalidConfig},
				{1 * time.Millisecond, invalidConfig},
				{1 * time.Millisecond, invalidConfig},
				{40 * time.Millisecond, validConfig},
			},
			nil,
		},
		{
			"valid and then more updates",
			40 * time.Millisecond,
			incomingConfigs{
				{1 * time.Millisecond, validConfig},
				{40 * time.Millisecond, invalidConfig},
				{1 * time.Millisecond, validConfig},
			},
			validConfig,
		},
		{
			"third update is valid on time",
			40 * time.Millisecond,
			incomingConfigs{
				{1 * time.Millisecond, invalidConfig},
				{1 * time.Millisecond, invalidConfig},
				{1 * time.Millisecond, validConfig},
			},
			validConfig,
		},
	}

	for _, tcase := range testcases {
		s.Run(tcase.name, func() {
			mocks := s.InitMocks()
			sut, err := New(s.log, dummyBeaterName, mocks.reloader, mocks.health, mocks.validator, beaterMockCreator, config.NewConfig())
			s.NoError(err)

			go func(ic incomingConfigs) {
				for _, c := range ic {
					time.Sleep(c.after)
					mocks.reloaderCh <- c.config
				}
			}(tcase.configs)

			cfg, err := sut.reconfigureWait(tcase.timeout)
			if tcase.expected == nil {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tcase.expected, cfg)
			}
		})
	}
}

func (s *LauncherTestSuite) TestHealthReporter() {
	errorStr := "some error"
	mocks := s.InitMocks()

	mocks.manager.EXPECT().Enabled().Return(true)
	mocks.manager.EXPECT().UpdateStatus(management.Degraded, errorStr)
	mocks.manager.EXPECT().UpdateStatus(management.Running, "")
	sut, err := New(s.log, dummyBeaterName, mocks.reloader, mocks.health, nil, beaterMockCreator, config.NewConfig())
	sut.beat = mocks.beat
	s.NoError(err)
	go func() {
		mocks.healthCh <- errors.New(errorStr)
		mocks.healthCh <- nil
		time.Sleep(time.Millisecond)

		sut.Stop()
	}()
	err = sut.run()
	s.NoError(err)
}

// TestLauncherErrorBeater should not call sut.Stop as the launcher should stop without callling it
func (s *LauncherTestSuite) TestLauncherErrorBeater() {
	mocks := s.InitMocks()
	sut, err := New(s.log, dummyBeaterName, mocks.reloader, mocks.health, nil, errorBeaterMockCreator, config.NewConfig())
	s.NoError(err)
	err = sut.run()
	s.Error(err)
}

// TestLauncherPanicBeater should not call sut.Stop as the launcher should stop without callling it
func (s *LauncherTestSuite) TestLauncherPanicBeater() {
	mocks := s.InitMocks()
	sut, err := New(s.log, dummyBeaterName, mocks.reloader, mocks.health, nil, panicBeaterMockCreator, config.NewConfig())
	s.NoError(err)
	err = sut.run()
	s.Error(err)
	s.ErrorContains(err, "panicBeaterMock panics")
}

func (s *LauncherTestSuite) TestLauncherUpdateAndStop() {
	mocks := s.InitMocks()
	sut, err := New(s.log, dummyBeaterName, mocks.reloader, mocks.health, nil, beaterMockCreator, config.NewConfig())
	s.NoError(err)
	go func() {
		mocks.reloaderCh <- config.NewConfig()
		sut.Stop()
	}()
	err = sut.run()
	s.NoError(err)
}

func (s *LauncherTestSuite) TestLauncherStopTwicePanics() {
	mocks := s.InitMocks()
	sut, err := New(s.log, dummyBeaterName, mocks.reloader, mocks.health, nil, beaterMockCreator, config.NewConfig())
	s.NoError(err)
	go func() {
		mocks.reloaderCh <- config.NewConfig()
		sut.Stop()
	}()
	err = sut.run()
	s.NoError(err)

	s.Panics(func() {
		sut.Stop()
	})
}

// TestLauncherErrorBeaterCreation should not call sut.Stop as the launcher should stop without callling it
func (s *LauncherTestSuite) TestLauncherErrorBeaterCreation() {
	mocks := s.InitMocks()
	sut, err := New(s.log, dummyBeaterName, mocks.reloader, mocks.health, nil, errorBeaterCreator, config.NewConfig())
	s.NoError(err)
	err = sut.run()
	s.Error(err)
}
