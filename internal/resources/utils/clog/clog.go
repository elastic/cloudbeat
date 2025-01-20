package clog

import (
	"context"
	"errors"
	"strings"

	"github.com/elastic/elastic-agent-libs/logp"
	"go.uber.org/zap"
)

type CustomLogger struct {
	*logp.Logger
}

type Logger interface {
	Debugf(template string, args ...any)
	Infof(template string, args ...any)
	Warnf(template string, args ...any)
	Errorf(template string, args ...any)
}

func (c *CustomLogger) Errorf(template string, args ...any) {
	// Downgrade context.Canceled errors to warning level
	if hasErrorType(context.Canceled, args...) {
		c.Warnf(template, args...)
		return
	}
	c.Logger.Errorf(template, args...)
}

func NewLogger(selector string, options ...logp.LogOption) *CustomLogger {
	options = append(options, zap.AddCallerSkip(1))
	logger := logp.NewLogger(selector).WithOptions(options...)

	l := &CustomLogger{logger}
	return l
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
