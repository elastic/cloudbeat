package clog

import (
	"context"
	"errors"
	"strings"

	"github.com/elastic/elastic-agent-libs/logp"
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

func (l *Logger) Named(name string) *Logger {
	return &Logger{l.Logger.Named(name)}
}

func (l *Logger) WithOptions(options ...logp.LogOption) *Logger {
	return &Logger{l.Logger.WithOptions(options...)}
}
func (l *Logger) With(args ...any) *Logger {
	return &Logger{l.Logger.With(args...)}
}

func NewLogger(selector string, options ...logp.LogOption) *Logger {
	logger := logp.NewLogger(selector).WithOptions(options...)
	return &Logger{logger}
}

func hasErrorType(errorType error, args ...any) bool {
	errorTypeStr := errorType.Error()
	for _, arg := range args {
		// Check if the error is of the same type
		if err, ok := arg.(error); ok && errors.Is(err, errorType) {
			return true
		}

		// Check if the error message contains the error type string
		if str, ok := arg.(string); ok && strings.Contains(str, errorTypeStr) {
			return true
		}
	}
	return false
}
