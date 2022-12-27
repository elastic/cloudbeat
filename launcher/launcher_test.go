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
	"testing"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/cmd/instance"
	"github.com/elastic/beats/v7/libbeat/common/reload"
	"github.com/elastic/beats/v7/libbeat/management"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
)

type beaterMock struct {
	cfg  *config.C
	done chan struct{}
}

func beaterMockCreator(b *beat.Beat, cfg *config.C) (beat.Beater, error) {
	return &beaterMock{
		cfg:  cfg,
		done: make(chan struct{}),
	}, nil
}

func (m *beaterMock) Run(b *beat.Beat) error {
	<-m.done
	return nil
}

func (m *beaterMock) Stop() {
	close(m.done)
}

type errorBeaterMock struct{}

func errorBeaterMockCreator(b *beat.Beat, cfg *config.C) (beat.Beater, error) {
	return &errorBeaterMock{}, nil
}

func errorBeaterCreator(b *beat.Beat, cfg *config.C) (beat.Beater, error) {
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

func (m *errorBeaterMock) Run(b *beat.Beat) error {
	time.Sleep(10 * time.Millisecond)
	return errors.New("some error")
}

func (m *errorBeaterMock) Stop() {
	panic("Error beater should not be stopped")
}

type panicBeaterMock struct{}

func panicBeaterMockCreator(b *beat.Beat, cfg *config.C) (beat.Beater, error) {
	return &panicBeaterMock{}, nil
}

func (m *panicBeaterMock) Run(b *beat.Beat) error {
	panic("panicBeaterMock panics")
}

func (m *panicBeaterMock) Stop() {
}

type reloaderMock struct {
	ch chan *config.C
}

func (m *reloaderMock) Channel() <-chan *config.C {
	return m.ch
}

func (m *reloaderMock) Stop() {
	close(m.ch)
}

type LauncherTestSuite struct {
	suite.Suite

	log  *logp.Logger
	opts goleak.Option
}

type launcherMocks struct {
	reloader *reloaderMock
	beat     *beat.Beat
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
	mocks := launcherMocks{}
	mocks.reloader = &reloaderMock{
		ch: make(chan *config.C),
	}
	mocks.beat = &beat.Beat{}
	return &mocks
}

func (s *LauncherTestSuite) MockBeatManager(mocks *launcherMocks) {
	settings := instance.Settings{
		Name:                  "some-beater",
		Version:               "version",
		DisableConfigResolver: true,
	}
	b, err := instance.NewInitializedBeat(settings)
	s.NoError(err)
	b.Manager, err = management.Factory(b.Config.Management)(b.Config.Management, reload.RegisterV2, b.Beat.Info.ID)
	s.NoError(err)
	mocks.beat = &b.Beat
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
			sut, err := New(s.log, mocks.reloader, beaterMockCreator, config.NewConfig())
			s.NoError(err)

			go func(ic incomingConfigs) {
				for _, c := range ic {
					time.Sleep(c.after)
					mocks.reloader.ch <- c.config
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
	sut, err := New(s.log, mocks.reloader, errorReloadBeaterCreator(), config.NewConfig())
	s.NoError(err)

	go func() {
		time.Sleep(40 * time.Millisecond)
		mocks.reloader.ch <- configErr
	}()

	err = sut.run()
	s.Error(err)
}

func (s *LauncherTestSuite) TestLauncherConfig() {
	validConfig := config.MustNewConfigFrom(mapstr.M{"config": mapstr.M{"v1": mapstr.M{"benchmark": "cis_k8s"}}})
	invalidConfig := config.MustNewConfigFrom(mapstr.M{"config": mapstr.M{"v1": mapstr.M{"benchmark": "cis_invalid"}}})

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
			sut, err := New(s.log, mocks.reloader, beaterMockCreator, config.NewConfig())
			s.NoError(err)

			mocks.reloader.ch = make(chan *config.C, len(tcase.configs))
			go func(ic incomingConfigs) {
				for _, c := range ic {
					time.Sleep(c.after)
					mocks.reloader.ch <- c.config
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

// TestLauncherErrorBeater should not call sut.Stop as the launcher should stop without callling it
func (s *LauncherTestSuite) TestLauncherErrorBeater() {
	mocks := s.InitMocks()
	sut, err := New(s.log, mocks.reloader, errorBeaterMockCreator, config.NewConfig())
	s.NoError(err)
	err = sut.run()
	s.Error(err)
}

// TestLauncherPanicBeater should not call sut.Stop as the launcher should stop without callling it
func (s *LauncherTestSuite) TestLauncherPanicBeater() {
	mocks := s.InitMocks()
	sut, err := New(s.log, mocks.reloader, panicBeaterMockCreator, config.NewConfig())
	s.NoError(err)
	err = sut.run()
	s.Error(err)
	s.ErrorContains(err, "panicBeaterMock panics")
}

func (s *LauncherTestSuite) TestLauncherUpdateAndStop() {
	mocks := s.InitMocks()
	sut, err := New(s.log, mocks.reloader, beaterMockCreator, config.NewConfig())
	s.NoError(err)
	go func() {
		mocks.reloader.ch <- config.NewConfig()
		sut.Stop()
	}()
	err = sut.run()
	s.NoError(err)
}

func (s *LauncherTestSuite) TestLauncherStopTwicePanics() {
	mocks := s.InitMocks()
	sut, err := New(s.log, mocks.reloader, beaterMockCreator, config.NewConfig())
	s.NoError(err)
	go func() {
		mocks.reloader.ch <- config.NewConfig()
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
	sut, err := New(s.log, mocks.reloader, errorBeaterCreator, config.NewConfig())
	s.NoError(err)
	err = sut.run()
	s.Error(err)
}
