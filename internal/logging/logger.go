package logging

import (
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Logger interface for structured logging
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

// LogrusLogger wraps logrus.Logger to implement our Logger interface
type LogrusLogger struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

// New creates a new logger with the specified level and format
func New(level, format string) Logger {
	logger := logrus.New()

	// Set log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	// Set format
	switch strings.ToLower(format) {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	default:
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	}

	// Set output
	logger.SetOutput(os.Stdout)

	return &LogrusLogger{
		logger: logger,
		entry:  logrus.NewEntry(logger),
	}
}

// NewBasic creates a basic logger for error scenarios
func NewBasic() Logger {
	return New("info", "text")
}

// NewWithOutput creates a logger with custom output
func NewWithOutput(level, format string, output io.Writer) Logger {
	logger := New(level, format)
	if ll, ok := logger.(*LogrusLogger); ok {
		ll.logger.SetOutput(output)
	}
	return logger
}

// Debug logs a debug message with optional fields
func (l *LogrusLogger) Debug(msg string, fields ...interface{}) {
	l.entry.WithFields(parseFields(fields...)).Debug(msg)
}

// Info logs an info message with optional fields
func (l *LogrusLogger) Info(msg string, fields ...interface{}) {
	l.entry.WithFields(parseFields(fields...)).Info(msg)
}

// Warn logs a warning message with optional fields
func (l *LogrusLogger) Warn(msg string, fields ...interface{}) {
	l.entry.WithFields(parseFields(fields...)).Warn(msg)
}

// Error logs an error message with optional fields
func (l *LogrusLogger) Error(msg string, fields ...interface{}) {
	l.entry.WithFields(parseFields(fields...)).Error(msg)
}

// Fatal logs a fatal message with optional fields and exits
func (l *LogrusLogger) Fatal(msg string, fields ...interface{}) {
	l.entry.WithFields(parseFields(fields...)).Fatal(msg)
}

// WithField returns a logger with a single field
func (l *LogrusLogger) WithField(key string, value interface{}) Logger {
	return &LogrusLogger{
		logger: l.logger,
		entry:  l.entry.WithField(key, value),
	}
}

// WithFields returns a logger with multiple fields
func (l *LogrusLogger) WithFields(fields map[string]interface{}) Logger {
	return &LogrusLogger{
		logger: l.logger,
		entry:  l.entry.WithFields(logrus.Fields(fields)),
	}
}

// parseFields converts key-value pairs to logrus.Fields
func parseFields(fields ...interface{}) logrus.Fields {
	result := make(logrus.Fields)

	for i := 0; i < len(fields)-1; i += 2 {
		if key, ok := fields[i].(string); ok {
			result[key] = fields[i+1]
		}
	}

	return result
}
