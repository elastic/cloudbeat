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

package clog

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/elastic/elastic-agent-libs/logp/logptest"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type LoggerTestSuite struct {
	suite.Suite
}

func TestLoggerTestSuite(t *testing.T) {
	s := new(LoggerTestSuite)

	suite.Run(t, s)
}

func (s *LoggerTestSuite) SetupSuite() {
}

func newObserverLogger(t *testing.T) (*Logger, *observer.ObservedLogs) {
	testLogger, observed := logptest.NewTestingLoggerWithObserver(t, "")
	return &Logger{Logger: testLogger.Named(t.Name())}, observed
}

func (s *LoggerTestSuite) TestErrorf() {
	tests := []struct {
		name          string
		logFunc       func(logger *Logger)
		expectedLevel zapcore.Level
	}{
		{
			name: "error arg",
			logFunc: func(l *Logger) {
				l.Errorf("failed: %v", context.Canceled)
			},
			expectedLevel: zap.WarnLevel,
		},
		{
			name: "wrapped error arg",
			logFunc: func(l *Logger) {
				l.Errorf("failed: %v", fmt.Errorf("wrap: %w", context.Canceled))
			},
			expectedLevel: zap.WarnLevel,
		},
		{
			name: "string arg",
			logFunc: func(l *Logger) {
				l.Errorf("failed: %s", context.Canceled.Error())
			},
			expectedLevel: zap.WarnLevel,
		},
		{
			name: "generic error stays at ERROR",
			logFunc: func(l *Logger) {
				l.Errorf("failed: %v", errors.New("something went wrong"))
			},
			expectedLevel: zap.ErrorLevel,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			logger, obs := newObserverLogger(s.T())
			tt.logFunc(logger)

			logs := obs.TakeAll()
			if s.Len(logs, 1) {
				s.Equal(tt.expectedLevel, logs[0].Level, "expected %s but got %s", tt.expectedLevel, logs[0].Level)
			}
		})
	}
}

func (s *LoggerTestSuite) TestError() {
	tests := []struct {
		name          string
		logFunc       func(logger *Logger)
		expectedLevel zapcore.Level
	}{
		{
			name: "error arg",
			logFunc: func(l *Logger) {
				l.Error(context.Canceled)
			},
			expectedLevel: zap.WarnLevel,
		},
		{
			name: "wrapped error arg",
			logFunc: func(l *Logger) {
				l.Error(fmt.Errorf("wrap: %w", context.Canceled))
			},
			expectedLevel: zap.WarnLevel,
		},
		{
			name: "string arg",
			logFunc: func(l *Logger) {
				l.Error(context.Canceled.Error())
			},
			expectedLevel: zap.WarnLevel,
		},
		{
			name: "generic error stays at ERROR",
			logFunc: func(l *Logger) {
				l.Error(errors.New("something went wrong"))
			},
			expectedLevel: zap.ErrorLevel,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			logger, obs := newObserverLogger(s.T())
			tt.logFunc(logger)

			logs := obs.TakeAll()
			if s.Len(logs, 1) {
				s.Equal(tt.expectedLevel, logs[0].Level, "expected %s but got %s", tt.expectedLevel, logs[0].Level)
			}
		})
	}
}
