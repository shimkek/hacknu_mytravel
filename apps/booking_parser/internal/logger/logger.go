package logger

import (
	"log"
	"os"
)

type Logger struct {
	debugEnabled bool
}

func New() *Logger {
	return &Logger{
		debugEnabled: os.Getenv("DEBUG") == "true",
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if l.debugEnabled {
		log.Printf("[DEBUG] "+format, args...)
	}
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	log.Fatalf("[FATAL] "+format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}
