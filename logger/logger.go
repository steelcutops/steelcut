package logger

import (
	"log/slog"
	"os"
)

type Logger interface {
	Info(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type StdLogger struct {
	internalLogger *slog.Logger
}

func New() Logger {
	l := slog.New(slog.NewTextHandler(os.Stderr, nil))
	return &StdLogger{internalLogger: l}
}

func (l *StdLogger) Info(msg string, args ...interface{}) {
	l.internalLogger.Info(msg, args...)
}

func (l *StdLogger) Debug(msg string, args ...interface{}) {
	l.internalLogger.Debug(msg, args...)
}

func (l *StdLogger) Warn(msg string, args ...interface{}) {
	l.internalLogger.Warn(msg, args...)
}

func (l *StdLogger) Error(msg string, args ...interface{}) {
	l.internalLogger.Error(msg, args...)
}
