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
	"github.com/open-policy-agent/opa/v1/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/elastic/cloudbeat/internal/infra/clog"
)

type logger struct {
	log *clog.Logger
	lvl zap.AtomicLevel
}

var opaToZapLevelsMap = map[logging.Level]zapcore.Level{
	logging.Error: zap.ErrorLevel,
	logging.Warn:  zap.WarnLevel,
	logging.Info:  zap.InfoLevel,
	logging.Debug: zap.DebugLevel,
}

var zapToOpaLevelsMap = map[zapcore.Level]logging.Level{
	zap.ErrorLevel: logging.Error,
	zap.WarnLevel:  logging.Warn,
	zap.InfoLevel:  logging.Info,
	zap.DebugLevel: logging.Debug,
}

func (l *logger) Debug(fmt string, a ...any) {
	if l.lvl.Enabled(zapcore.DebugLevel) {
		l.log.Debugf(fmt, a...)
	}
}

func (l *logger) Info(fmt string, a ...any) {
	if l.lvl.Enabled(zapcore.InfoLevel) {
		l.log.Infof(fmt, a...)
	}
}

func (l *logger) Error(fmt string, a ...any) {
	if l.lvl.Enabled(zapcore.ErrorLevel) {
		l.log.Errorf(fmt, a...)
	}
}

func (l *logger) Warn(fmt string, a ...any) {
	if l.lvl.Enabled(zapcore.WarnLevel) {
		l.log.Warnf(fmt, a...)
	}
}

func (l *logger) WithFields(m map[string]any) logging.Logger {
	return &logger{
		log: l.log.With(mapToArray(m)...),
		lvl: l.lvl,
	}
}

func (l *logger) GetLevel() logging.Level {
	return toOpaLevel(l.lvl.Level())
}

func (l *logger) SetLevel(level logging.Level) {
	l.lvl.SetLevel(toZapLevel(level))
}

func mapToArray(m map[string]any) []any {
	ret := make([]any, 0, len(m))
	for k, v := range m {
		ret = append(ret, k, v)
	}

	return ret
}

// newLoggerFromBase creates an OPA logger from a base clog.Logger.
// This avoids using the global logger system and reuses the passed logger.
func newLoggerFromBase(baseLog *clog.Logger) logging.Logger {
	return &logger{
		log: baseLog.Named("opa").WithOptions(zap.AddCallerSkip(1)),
		lvl: zap.NewAtomicLevelAt(zapcore.LevelOf(baseLog.Core())),
	}
}

func toZapLevel(l logging.Level) zapcore.Level {
	if res, ok := opaToZapLevelsMap[l]; ok {
		return res
	}

	return zap.DebugLevel
}

func toOpaLevel(l zapcore.Level) logging.Level {
	if res, ok := zapToOpaLevelsMap[l]; ok {
		return res
	}

	return logging.Debug
}
