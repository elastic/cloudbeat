package rerun

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/cmd/instance"
	"github.com/elastic/beats/v7/libbeat/common/reload"
	"github.com/elastic/beats/v7/libbeat/management"
	"github.com/elastic/elastic-agent-libs/config"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
)

type beaterMock struct {
	cfg    *agentconfig.C
	ctx    context.Context
	cancel context.CancelFunc
}

func beaterMockCreator(b *beat.Beat, cfg *agentconfig.C) (beat.Beater, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &beaterMock{
		cfg:    cfg,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (m *beaterMock) Run(b *beat.Beat) error {
	for {
		select {
		case <-m.ctx.Done():
			return nil
		}
	}
}

func (m *beaterMock) Stop() {
	m.cancel()
}

type errorBeaterMock struct{}

func errorBeaterMockCreator(b *beat.Beat, cfg *agentconfig.C) (beat.Beater, error) {
	return &errorBeaterMock{}, nil
}

func (m *errorBeaterMock) Run(b *beat.Beat) error {
	return errors.New("some error")
}

func (m *errorBeaterMock) Stop() {
}

type reloaderMock struct {
	ch chan *agentconfig.C
}

func (m *reloaderMock) Channel() <-chan *agentconfig.C {
	return m.ch
}

type validatorMock struct {
	expected *agentconfig.C
}

func (v *validatorMock) Validate(cfg *agentconfig.C) error {
	var err error
	if !reflect.DeepEqual(cfg, v.expected) {
		err = fmt.Errorf("mock validation failed")
	}

	return err
}

type StarterTestSuite struct {
	suite.Suite

	log  *logp.Logger
	opts goleak.Option
}

type starterMocks struct {
	ctx       context.Context
	cancel    context.CancelFunc
	reloader  *reloaderMock
	beat      *beat.Beat
	validator Validator
}

func TestStarterTestSuite(t *testing.T) {
	s := new(StarterTestSuite)
	s.log = logp.NewLogger("cloudbeat_starter_test_suite")
	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	s.opts = goleak.IgnoreCurrent()
	suite.Run(t, s)
}

func (s *StarterTestSuite) InitMocks() *starterMocks {
	mocks := starterMocks{}
	mocks.ctx, mocks.cancel = context.WithCancel(context.Background())
	mocks.reloader = &reloaderMock{
		ch: make(chan *agentconfig.C),
	}
	mocks.validator = &validatorMock{
		expected: agentconfig.MustNewConfigFrom(mapstr.M{"a": 1}),
	}
	mocks.beat = &beat.Beat{}
	return &mocks
}

func (s *StarterTestSuite) MockBeatManager(mocks *starterMocks) {
	settings := instance.Settings{
		Name:                  "some-beater",
		Version:               "version",
		DisableConfigResolver: true,
	}
	b, err := instance.NewInitializedBeat(settings)
	s.NoError(err)
	b.Manager, err = management.Factory(b.Config.Management)(b.Config.Management, reload.Register, b.Beat.Info.ID)
	s.NoError(err)
	mocks.beat = &b.Beat
}

func (s *StarterTestSuite) TearDownTest() {
	// Verify no goroutines are leaking. Safest to keep this on top of the function.
	// Go defers are implemented as a LIFO stack. This should be the last one to run.
	goleak.VerifyNone(s.T(), s.opts)
}

func (s *StarterTestSuite) TestWaitForUpdates() {
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
		configs  incomingConfigs
		expected *config.C
	}{
		{
			"no updates",
			incomingConfigs{},
			agentconfig.NewConfig(),
		},
		{
			"single update",
			incomingConfigs{
				{40 * time.Millisecond, configC},
			},
			configC,
		},
		{
			"multiple updates A B A C",
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
			incomingConfigs{
				{1 * time.Millisecond, configA},
				{40 * time.Millisecond, configB},
				{1 * time.Millisecond, configA},
			},
			expected2,
		},
		{
			"multiple updates A C B",
			incomingConfigs{
				{1 * time.Millisecond, configA},
				{1 * time.Millisecond, configC},
				{1 * time.Millisecond, configB},
			},
			expected3,
		},
		{
			"multiple updates C C A",
			incomingConfigs{
				{1 * time.Millisecond, configC},
				{1 * time.Millisecond, configC},
				{1 * time.Millisecond, configA},
			},
			expected4,
		},
	}

	for _, tcase := range testcases {
		s.Run(tcase.name, func() {
			mocks := s.InitMocks()
			sut, err := NewLauncher(mocks.ctx, s.log, mocks.reloader, nil, mocks.beat, beaterMockCreator, config.NewConfig())
			s.NoError(err)

			go func(ic incomingConfigs) {
				defer close(mocks.reloader.ch)

				for _, c := range ic {
					time.Sleep(c.after)
					mocks.reloader.ch <- c.config
				}

				time.Sleep(100 * time.Millisecond)
			}(tcase.configs)

			err = sut.run()
			s.Error(err)
			beater, ok := sut.beater.(*beaterMock)
			s.True(ok)
			s.Equal(tcase.expected, beater.cfg)
			sut.Stop()
			sut.wg.Wait()
		})
	}
}

func (s *StarterTestSuite) TestStarterErrorBeater() {
	mocks := s.InitMocks()
	sut, err := NewLauncher(mocks.ctx, s.log, mocks.reloader, nil, mocks.beat, errorBeaterMockCreator, config.NewConfig())
	err = sut.run()
	s.Error(err)
}

func (s *StarterTestSuite) TestStarterCancelBeater() {
	mocks := s.InitMocks()
	sut, err := NewLauncher(mocks.ctx, s.log, mocks.reloader, nil, mocks.beat, beaterMockCreator, config.NewConfig())
	go func() {
		time.Sleep(100 * time.Millisecond)
		mocks.cancel()
	}()
	err = sut.run()
	s.NoError(err)
}

func (s *StarterTestSuite) TestStarterValidator() {
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
			sut, err := NewLauncher(mocks.ctx, s.log, mocks.reloader, mocks.validator, mocks.beat, beaterMockCreator, config.NewConfig())
			s.NoError(err)

			mocks.reloader.ch = make(chan *agentconfig.C, len(tcase.configs))
			go func(ic incomingConfigs) {
				defer close(mocks.reloader.ch)

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
