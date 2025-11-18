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
	"testing"

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

func (s *LoggerTestSuite) TestErrorfWithContextCanceled() {
	observedCore, observedLogs := observer.New(zapcore.DebugLevel)
	logger := NewLogger("test", zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return observedCore
	}))

	err := context.Canceled
	logger.Errorf("some error: %s", err)         // error with context.Canceled
	logger.Errorf("some error: %s", err.Error()) // error string with context Canceled

	logs := observedLogs.TakeAll()
	if s.Len(logs, 2) {
		s.Equal(zap.WarnLevel, logs[0].Level) // downgraded to warning
		s.Equal("some error: context canceled", logs[0].Message)

		s.Equal(zap.WarnLevel, logs[1].Level) // downgraded to warning
		s.Equal("some error: context canceled", logs[1].Message)
	}
}

func (s *LoggerTestSuite) TestLogErrorfWithoutContextCanceled() {
	observedCore, observedLogs := observer.New(zapcore.DebugLevel)
	logger := NewLogger("test", zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return observedCore
	}))

	err := errors.New("oops")
	logger.Errorf("some error: %s", err)

	logs := observedLogs.TakeAll()
	if s.Len(logs, 1) {
		s.Equal(zap.ErrorLevel, logs[0].Level)
		s.Equal("some error: oops", logs[0].Message)
	}
}
