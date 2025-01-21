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
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/open-policy-agent/opa/v1/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	l.log.Debugf(fmt, a...)
}

func (l *logger) Info(fmt string, a ...any) {
	l.log.Infof(fmt, a...)
}

func (l *logger) Error(fmt string, a ...any) {
	l.log.Errorf(fmt, a...)
}

func (l *logger) Warn(fmt string, a ...any) {
	l.log.Warnf(fmt, a...)
}

func (l *logger) WithFields(m map[string]any) logging.Logger {
	return &logger{
		log: l.log.With(mapToArray(m)...),
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

func newLogger() logging.Logger {
	lvl := zap.NewAtomicLevelAt(logp.GetLevel())
	log := clog.NewLogger("opa").WithOptions(
		zap.IncreaseLevel(lvl),
		zap.AddCallerSkip(1),
	)

	return &logger{
		log: log,
		lvl: lvl,
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
