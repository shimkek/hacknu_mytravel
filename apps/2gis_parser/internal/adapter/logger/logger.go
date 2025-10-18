package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	zap *zap.SugaredLogger
}

func New(env string) *Logger {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		// Configure prettier console output for development
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		config.EncoderConfig = zapcore.EncoderConfig{
			TimeKey:       "time",
			LevelKey:      "level",
			NameKey:       "logger",
			CallerKey:     "caller",
			FunctionKey:   zapcore.OmitKey,
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel:   zapcore.CapitalColorLevelEncoder, // Colorized level names
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format("15:04:05")) // Pretty time format HH:MM:SS
			},
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
		config.Encoding = "console"
		config.Development = true
		config.DisableStacktrace = true
	}

	l, err := config.Build()
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
