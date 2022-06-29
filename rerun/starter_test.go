package rerun

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
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

type StarterTestSuite struct {
	suite.Suite

	log  *logp.Logger
	opts goleak.Option
}

type starterMocks struct {
	ctx      context.Context
	cancel   context.CancelFunc
	reloader *reloaderMock
	beat     *beat.Beat
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
	mocks.beat = &beat.Beat{}
	return &mocks
}

func (s *StarterTestSuite) TearDownTest() {
	// Verify no goroutines are leaking. Safest to keep this on top of the function.
	// Go defers are implemented as a LIFO stack. This should be the last one to run.
	goleak.VerifyNone(s.T(), s.opts)
}

func (s *StarterTestSuite) TestWaitForUpdates() {
	configA, err := config.NewConfigFrom(mapstr.M{
		"common":    "A",
		"specificA": "a",
		"commonArr": []string{"a"},
	})
	s.NoError(err)

	configB, err := config.NewConfigFrom(mapstr.M{
		"common":    "B",
		"specificB": "b",
		"commonArr": []string{"b", "b"},
	})
	s.NoError(err)

	configC, err := config.NewConfigFrom(mapstr.M{
		"common":    "C",
		"specificC": "c",
		"commonArr": []string{"c", "c", "c"},
	})
	s.NoError(err)

	expected1, err := config.NewConfigFrom(mapstr.M{
		"common":    "C",
		"specificA": "a",
		"specificB": "b",
		"specificC": "c",
		"commonArr": []string{"c", "c", "c"},
	})
	s.NoError(err)

	expected2, err := config.NewConfigFrom(mapstr.M{
		"common":    "A",
		"specificA": "a",
		"specificB": "b",
		"commonArr": []string{"a"},
	})
	s.NoError(err)

	expected3, err := config.NewConfigFrom(mapstr.M{
		"common":    "B",
		"specificA": "a",
		"specificB": "b",
		"specificC": "c",
		"commonArr": []string{"b", "b"},
	})
	s.NoError(err)

	expected4, err := config.NewConfigFrom(mapstr.M{
		"common":    "A",
		"specificA": "a",
		"specificC": "c",
		"commonArr": []string{"a"},
	})
	s.NoError(err)

	type incomingConfigs []struct {
		after  time.Duration
		config *config.C
	}

	testcases := []struct {
		timeout  time.Duration
		configs  incomingConfigs
		expected *config.C
	}{
		{
			5 * time.Millisecond,
			incomingConfigs{},
			agentconfig.NewConfig(),
		},
		{
			40 * time.Millisecond,
			incomingConfigs{
				{40 * time.Millisecond, configC},
			},
			configC,
		},
		{
			40 * time.Millisecond,
			incomingConfigs{
				{1 * time.Millisecond, configA},
				{1 * time.Millisecond, configB},
				{1 * time.Millisecond, configA},
				{40 * time.Millisecond, configC},
			},
			expected1,
		},
		{
			40 * time.Millisecond,
			incomingConfigs{
				{1 * time.Millisecond, configA},
				{40 * time.Millisecond, configB},
				{1 * time.Millisecond, configA},
			},
			expected2,
		},
		{
			40 * time.Millisecond,
			incomingConfigs{
				{1 * time.Millisecond, configA},
				{1 * time.Millisecond, configC},
				{1 * time.Millisecond, configB},
			},
			expected3,
		},
		{
			40 * time.Millisecond,
			incomingConfigs{
				{1 * time.Millisecond, configC},
				{1 * time.Millisecond, configC},
				{1 * time.Millisecond, configA},
			},
			expected4,
		},
	}

	for _, tcase := range testcases {
		mocks := s.InitMocks()
		sut, err := NewStarter(mocks.ctx, s.log, mocks.reloader, nil, mocks.beat, beaterMockCreator, config.NewConfig())
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
	}
}

func (s *StarterTestSuite) TestStarterErrorBeater() {
	mocks := s.InitMocks()
	sut, err := NewStarter(mocks.ctx, s.log, mocks.reloader, nil, mocks.beat, errorBeaterMockCreator, config.NewConfig())
	err = sut.run()
	s.Error(err)
}

func (s *StarterTestSuite) TestStarterCancelBeater() {
	mocks := s.InitMocks()
	sut, err := NewStarter(mocks.ctx, s.log, mocks.reloader, nil, mocks.beat, beaterMockCreator, config.NewConfig())
	go func() {
		time.Sleep(100 * time.Millisecond)
		mocks.cancel()
	}()
	err = sut.run()
	s.NoError(err)
}
