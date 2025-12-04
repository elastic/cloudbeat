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

	"github.com/open-policy-agent/opa/v1/logging"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

// createTestLogger creates a logger for testing using newLoggerFromBase
// Uses a logger created with logptest.NewTestingLogger and an observer core
func createTestLogger(t *testing.T) (logging.Logger, *observer.ObservedLogs) {
	t.Helper()
	core, obs := testhelper.NewObserverLogger(t)
	opaLogger := newLoggerFromBase(core)

	return opaLogger, obs
}

func TestLogFormat(t *testing.T) {
	logger, obs := createTestLogger(t)
	logger.SetLevel(logging.Warn)
	logger.Warn("warn %s", "warn")
	logs := obs.TakeAll()
	if assert.Len(t, logs, 1) {
		assertLog(t, logs[0], zap.WarnLevel, "warn warn")
	}
}

func TestLogFields(t *testing.T) {
	logger, obs := createTestLogger(t)
	logger.SetLevel(logging.Debug)
	logger = logger.WithFields(map[string]any{
		"key": "val",
	})

	logger.Debug("debug")
	logs := obs.TakeAll()
	if assert.Len(t, logs, 1) {
		assertLog(t, logs[0], zap.DebugLevel, "debug")
		assert.Len(t, logs[0].Context, 1)
		assert.Equal(t, "val", logs[0].ContextMap()["key"])
	}
}

func TestLogMultipleFields(t *testing.T) {
	logger, obs := createTestLogger(t)
	logger.SetLevel(logging.Debug)
	logger = logger.WithFields(map[string]any{
		"key1": "val1",
	})

	logger = logger.WithFields(map[string]any{
		"key2": "val2",
	})

	logger.Debug("debug")
	logs := obs.TakeAll()
	if assert.Len(t, logs, 1) {
		assertLog(t, logs[0], zap.DebugLevel, "debug")
		assert.Len(t, logs[0].Context, 2)
		assert.Equal(t, "val1", logs[0].ContextMap()["key1"])
		assert.Equal(t, "val2", logs[0].ContextMap()["key2"])
	}
}

func TestLoggerGetLevel(t *testing.T) {
	logger, _ := createTestLogger(t)
	tests := []logging.Level{
		logging.Debug,
		logging.Info,
		logging.Warn,
		logging.Error,
	}

	for _, l := range tests {
		logger.SetLevel(l)
		assert.Equal(t, l, logger.GetLevel())
	}
}

func TestLoggerSetLevel(t *testing.T) {
	logger, obs := createTestLogger(t)
	logger.SetLevel(logging.Debug)
	logger.Debug("debug")
	logs := obs.TakeAll()
	if assert.Len(t, logs, 1) {
		assertLog(t, logs[0], zap.DebugLevel, "debug")
	}

	logger.SetLevel(logging.Error)
	logger.Error("error")
	logs = obs.TakeAll()
	if assert.Len(t, logs, 1) {
		assertLog(t, logs[0], zap.ErrorLevel, "error")
	}

	logger.Info("info")
	logs = obs.TakeAll()
	assert.Empty(t, logs)

	logger.SetLevel(logging.Info)
	logger.Info("info")
	logger.Error("error")
	logs = obs.TakeAll()
	if assert.Len(t, logs, 2) {
		assertLog(t, logs[0], zap.InfoLevel, "info")
		assertLog(t, logs[1], zap.ErrorLevel, "error")
	}

	logger.Debug("debug")
	logs = obs.TakeAll()
	assert.Empty(t, logs)
}

func assertLog(t *testing.T, log observer.LoggedEntry, level zapcore.Level, message string) {
	const expectedLoggerName = "opa"

	t.Helper()
	assert.Equal(t, level, log.Level)
	assert.Equal(t, t.Name()+"."+expectedLoggerName, log.LoggerName)
	assert.Equal(t, message, log.Message)
}
