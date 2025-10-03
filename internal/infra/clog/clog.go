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
	"strings"

	"github.com/elastic/elastic-agent-libs/logp"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Logger struct {
	*logp.Logger
}

func (l *Logger) Errorf(template string, args ...any) {
	// Downgrade context.Canceled errors to warning level
	if hasErrorType(context.Canceled, args...) {
		l.Warnf(template, args...)
		return
	}
	l.Logger.Errorf(template, args...)
}

func (l *Logger) Error(args ...any) {
	// Downgrade context.Canceled errors to warning level
	if hasErrorType(context.Canceled, args...) {
		l.Warn(args...)
		return
	}
	l.Logger.Error(args...)
}

func (l *Logger) Named(name string) *Logger {
	return &Logger{l.Logger.Named(name)}
}

func (l *Logger) WithOptions(options ...logp.LogOption) *Logger {
	return &Logger{l.Logger.WithOptions(options...)}
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{l.Logger.With(args...)}
}

func (l *Logger) WithSpanContext(spanCtx trace.SpanContext) *Logger {
	newLogger := l
	if spanCtx.HasSpanID() {
		newLogger = newLogger.With("span.id", spanCtx.SpanID().String())
	}
	if spanCtx.HasTraceID() {
		newLogger = newLogger.With("trace.id", spanCtx.TraceID().String())
	}
	return newLogger
}

func NewLogger(selector string, options ...logp.LogOption) *Logger {
	options = append(options, zap.AddCallerSkip(1))
	logger := logp.NewLogger(selector).WithOptions(options...)
	return &Logger{logger}
}

func hasErrorType(errorType error, args ...any) bool {
	errorTypeStr := errorType.Error()
	for _, arg := range args {
		// Check if the error is of the same type
		if err, ok := arg.(error); ok && (errors.Is(err, errorType) || strings.Contains(err.Error(), errorTypeStr)) {
			return true
		}

		// Check if the error message contains the error type string
		if str, ok := arg.(string); ok && strings.Contains(str, errorTypeStr) {
			return true
		}
	}
	return false
}
