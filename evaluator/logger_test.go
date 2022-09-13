package evaluator

import (
	"testing"

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/open-policy-agent/opa/logging"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

const expectedLoggerName = "opa"

type LoggerTestSuite struct {
	suite.Suite
}

func TestLoggerTestSuite(t *testing.T) {
	s := new(LoggerTestSuite)

	suite.Run(t, s)
}

func (s *LoggerTestSuite) SetupSuite() {
	err := logp.DevelopmentSetup(logp.ToObserverOutput())
	s.NoError(err)
}

func (s *LoggerTestSuite) TestLogFormat() {
	logger := newLogger()
	logger.SetLevel(logging.Warn)
	logger.Warn("warn %s", "warn")
	logs := logp.ObserverLogs().TakeAll()
	if s.Len(logs, 1) {
		s.assertLog(logs[0], zap.WarnLevel, "warn warn")
	}
}

func (s *LoggerTestSuite) TestLogFields() {
	logger := newLogger()
	logger = logger.WithFields(map[string]interface{}{
		"key": "val",
	})

	logger.Debug("debug")
	logs := logp.ObserverLogs().TakeAll()
	if s.Len(logs, 1) {
		s.assertLog(logs[0], zap.DebugLevel, "debug")
		s.Equal("val", logs[0].ContextMap()["key"])
	}
}

func (s *LoggerTestSuite) TestLoggerGetLevel() {
	logger := newLogger()
	tests := []logging.Level{
		logging.Debug,
		logging.Info,
		logging.Warn,
		logging.Error,
	}

	for _, l := range tests {
		logger.SetLevel(l)
		s.Equal(l, logger.GetLevel())
	}
}

func (s *LoggerTestSuite) TestLoggerSetLevel() {
	logger := newLogger()
	logger.Debug("debug")
	logs := logp.ObserverLogs().TakeAll()
	if s.Len(logs, 1) {
		s.assertLog(logs[0], zap.DebugLevel, "debug")
	}

	logger.SetLevel(logging.Error)
	logger.Error("error")
	logs = logp.ObserverLogs().TakeAll()
	if s.Len(logs, 1) {
		s.assertLog(logs[0], zap.ErrorLevel, "error")
	}

	logger.Info("info")
	logs = logp.ObserverLogs().TakeAll()
	s.Empty(logs, 1)

	logger.SetLevel(logging.Info)
	logger.Info("info")
	logger.Error("error")
	logs = logp.ObserverLogs().TakeAll()
	if s.Len(logs, 2) {
		s.assertLog(logs[0], zap.InfoLevel, "info")
		s.assertLog(logs[1], zap.ErrorLevel, "error")
	}

	logger.Debug("debug")
	logs = logp.ObserverLogs().TakeAll()
	s.Empty(logs, 1)
}

func (s *LoggerTestSuite) assertLog(log observer.LoggedEntry, level zapcore.Level, message string) {
	s.Equal(level, log.Level)
	s.Equal(expectedLoggerName, log.LoggerName)
	s.Equal(message, log.Message)
}
