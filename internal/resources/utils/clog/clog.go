package clog

import (
	"context"
	"errors"
	"strings"

	"github.com/elastic/elastic-agent-libs/logp"
	"go.uber.org/zap"
)

type Logger struct {
	*logp.Logger
}

func (c *Logger) Errorf(template string, args ...any) {
	// Downgrade context.Canceled errors to warning level
	if hasErrorType(context.Canceled, args...) {
		c.Warnf(template, args...)
		return
	}
	c.Logger.Errorf(template, args...)
}

func (c *Logger) Named(name string) *Logger {
	logger := c.Logger.Named(name)
	return &Logger{logger}
}

func (l *Logger) WithOptions(options ...logp.LogOption) *Logger {
	return &Logger{l.Logger.WithOptions(options...)}
}
func (l *Logger) With(args ...any) *Logger {
	return &Logger{l.Logger.With(args...)}
}

func NewLogger(selector string, options ...logp.LogOption) *Logger {
	options = append(options, zap.AddCallerSkip(1))
	logger := logp.NewLogger(selector).WithOptions(options...)
	l := &Logger{logger}
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

type lkey string

const loggerKey lkey = "cloudbeat_logger"

func WithLogger(ctx context.Context, name string) (context.Context, *Logger) {
	log := NewLogger(name)
	return context.WithValue(ctx, loggerKey, log), log
}

func GetLogger(ctx context.Context) *Logger {
	loggerValue := ctx.Value(loggerKey)
	if loggerValue == nil {
		log := NewLogger("cloudbeat")
		log.Warn("Context did not have logger key")
		return log
	}
	log, ok := loggerValue.(*Logger)
	if !ok {
		log = NewLogger("cloudbeat")
		log.Errorf("Unexpected logger type %T", loggerValue)
		return log
	}
	return log
}
