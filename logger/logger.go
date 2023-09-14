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
	SetLevel(level slog.Level)
}

type StdLogger struct {
	internalLogger *slog.Logger
	logLevel       *slog.LevelVar
}

func New() Logger {
	logLevel := &slog.LevelVar{} // Default to INFO by slog
	opts := slog.HandlerOptions{
		Level: logLevel,
	}
	l := slog.New(slog.NewTextHandler(os.Stderr, &opts))
	return &StdLogger{
		internalLogger: l,
		logLevel:       logLevel,
	}
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

func (l *StdLogger) SetLevel(level slog.Level) {
	l.logLevel.Set(level)
}
