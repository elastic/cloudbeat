package evaluator

import (
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/open-policy-agent/opa/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger struct {
	log *logp.Logger
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

func (l *logger) Debug(fmt string, a ...interface{}) {
	l.log.Debugf(fmt, a...)
}

func (l *logger) Info(fmt string, a ...interface{}) {
	l.log.Infof(fmt, a...)
}

func (l *logger) Error(fmt string, a ...interface{}) {
	l.log.Errorf(fmt, a...)
}

func (l *logger) Warn(fmt string, a ...interface{}) {
	l.log.Warnf(fmt, a...)
}

func (l *logger) WithFields(m map[string]interface{}) logging.Logger {
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

func mapToArray(m map[string]interface{}) []interface{} {
	var ret []interface{}
	for k, v := range m {
		ret = append(ret, k, v)
	}

	return ret
}

func newLogger() logging.Logger {
	lvl := zap.NewAtomicLevelAt(zapcore.DebugLevel)
	log := logp.NewLogger("opa").WithOptions(
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
