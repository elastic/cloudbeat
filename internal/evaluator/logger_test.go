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

package evaluator

import (
	"testing"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/elastic-agent-libs/logp/logptest"
	"github.com/open-policy-agent/opa/v1/logging"
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

func newTestLogger(t *testing.T) (logging.Logger, *observer.ObservedLogs) {
	log := newLogger()

	observedCore, observedLogs := observer.New(zapcore.DebugLevel)
	testingLogger := logptest.NewTestingLogger(t, "opa", zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return observedCore
	}), zap.IncreaseLevel(log.(*logger).lvl))
	log.(*logger).log = &clog.Logger{Logger: testingLogger}
	return log, observedLogs
}

func (s *LoggerTestSuite) TestLogFormat() {
	logger, observedLogs := newTestLogger(s.T())
	logger.SetLevel(logging.Warn)
	logger.Warn("warn %s", "warn")
	logs := observedLogs.TakeAll()
	if s.Len(logs, 1) {
		s.assertLog(logs[0], zap.WarnLevel, "warn warn")
	}
}

func (s *LoggerTestSuite) TestLogFields() {
	logger, observedLogs := newTestLogger(s.T())
	logger.SetLevel(logging.Debug)
	logger = logger.WithFields(map[string]any{
		"key": "val",
	})

	logger.Debug("debug")
	logs := observedLogs.TakeAll()
	if s.Len(logs, 1) {
		s.assertLog(logs[0], zap.DebugLevel, "debug")
		s.Len(logs[0].Context, 1)
		s.Equal("val", logs[0].ContextMap()["key"])
	}
}

func (s *LoggerTestSuite) TestLogMultipleFields() {
	logger, observedLogs := newTestLogger(s.T())
	logger.SetLevel(logging.Debug)
	logger = logger.WithFields(map[string]any{
		"key1": "val1",
	})

	logger = logger.WithFields(map[string]any{
		"key2": "val2",
	})

	logger.Debug("debug")
	logs := observedLogs.TakeAll()
	if s.Len(logs, 1) {
		s.assertLog(logs[0], zap.DebugLevel, "debug")
		s.Len(logs[0].Context, 2)
		s.Equal("val1", logs[0].ContextMap()["key1"])
		s.Equal("val2", logs[0].ContextMap()["key2"])
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
	logger, observedLogs := newTestLogger(s.T())
	logger.SetLevel(logging.Debug)
	logger.Debug("debug")
	logs := observedLogs.TakeAll()
	if s.Len(logs, 1) {
		s.assertLog(logs[0], zap.DebugLevel, "debug")
	}

	logger.SetLevel(logging.Error)
	logger.Error("error")
	logs = observedLogs.TakeAll()
	if s.Len(logs, 1) {
		s.assertLog(logs[0], zap.ErrorLevel, "error")
	}

	logger.Info("info")
	logs = observedLogs.TakeAll()
	s.Empty(logs, 1)

	logger.SetLevel(logging.Info)
	logger.Info("info")
	logger.Error("error")
	logs = observedLogs.TakeAll()
	if s.Len(logs, 2) {
		s.assertLog(logs[0], zap.InfoLevel, "info")
		s.assertLog(logs[1], zap.ErrorLevel, "error")
	}

	logger.Debug("debug")
	logs = observedLogs.TakeAll()
	s.Empty(logs, 1)
}

func (s *LoggerTestSuite) assertLog(log observer.LoggedEntry, level zapcore.Level, message string) {
	s.Equal(level, log.Level)
	s.Equal(expectedLoggerName, log.LoggerName)
	s.Equal(message, log.Message)
}
