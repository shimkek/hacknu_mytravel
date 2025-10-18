package logger

import (
	"go.uber.org/zap"
)

type Logger struct {
	zap *zap.SugaredLogger
}

func New(env string) *Logger {
	var l *zap.Logger
	var err error

	if env == "production" {
		l, err = zap.NewProduction()
	} else {
		l, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(err)
	}

	return &Logger{zap: l.Sugar()}
}

func (l *Logger) Info(msg string, args ...interface{}) {
	l.zap.Infof(msg, args...)
}

func (l *Logger) Error(msg string, args ...interface{}) {
	l.zap.Errorf(msg, args...)
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	l.zap.Debugf(msg, args...)
}

func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.zap.Fatalf(msg, args...)
}
