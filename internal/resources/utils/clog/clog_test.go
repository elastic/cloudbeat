package clog

import (
	"context"
	"errors"
	"testing"

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type LoggerTestSuite struct {
	suite.Suite
}

func TestLoggerTestSuite(t *testing.T) {
	s := new(LoggerTestSuite)

	suite.Run(t, s)
}

func (s *LoggerTestSuite) SetupSuite() {
	err := logp.DevelopmentSetup(logp.ToObserverOutput())
	s.Require().NoError(err)
}

func (s *LoggerTestSuite) TestErrorfWithContextCanceled() {
	logger := NewLogger("test")

	err := context.Canceled
	logger.Errorf("some error: %s", err)         // error with context.Canceled
	logger.Errorf("some error: %s", err.Error()) // error string with context Canceled

	logs := logp.ObserverLogs().TakeAll()
	if s.Len(logs, 2) {
		s.Equal(zap.WarnLevel, logs[0].Level) // downgraded to warning
		s.Equal("some error: context canceled", logs[0].Message)

		s.Equal(zap.WarnLevel, logs[1].Level) // downgraded to warning
		s.Equal("some error: context canceled", logs[1].Message)
	}
}
func (s *LoggerTestSuite) TestLogErrorfWithoutContextCanceled() {
	logger := NewLogger("test")

	err := errors.New("oops")
	logger.Errorf("some error: %s", err)

	logs := logp.ObserverLogs().TakeAll()
	if s.Len(logs, 1) {
		s.Equal(zap.ErrorLevel, logs[0].Level)
		s.Equal("some error: oops", logs[0].Message)
	}
}
