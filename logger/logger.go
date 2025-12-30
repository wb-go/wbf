// Package logger provides a unified, structured logging interface with support for multiple
// underlying logging engines (Zap, slog, zerolog, logrus). It enables consistent logging
// across services while allowing runtime engine selection and structured attribute handling.
package logger

import (
	"context"
	"time"
)

// Level represents the severity of a log record.
type Level int

// Engine represents a supported underlying logging implementation.
type Engine string

const (
	// ZapEngine selects the github.com/uber-go/zap logger.
	ZapEngine Engine = "zap"
	// SlogEngine selects the stdlib log/slog logger (Go 1.21+).
	SlogEngine Engine = "slog"
	// ZerologEngine selects the github.com/rs/zerolog logger.
	ZerologEngine Engine = "zerolog"
	// LogrusEngine selects the github.com/sirupsen/logrus logger.
	LogrusEngine Engine = "logrus"

	// DebugLevel is the most verbose level, typically used for development.
	DebugLevel Level = iota - 4
	// InfoLevel is the default logging level for general operational information.
	InfoLevel
	// WarnLevel indicates unexpected or unusual events that are not errors.
	WarnLevel
	// ErrorLevel indicates serious errors that require attention.
	ErrorLevel
)

// Attr represents a key-value pair for structured logging.
type Attr struct {
	Key   string
	Value any
}

// Logger defines a unified interface for structured logging across multiple engines.
// It supports context-aware logging, attribute-based records, and HTTP request tracing.
type Logger interface {
	// Debug logs a message at DebugLevel.
	Debug(msg string, args ...any)
	// Info logs a message at InfoLevel.
	Info(msg string, args ...any)
	// Warn logs a message at WarnLevel.
	Warn(msg string, args ...any)
	// Error logs a message at ErrorLevel.
	Error(msg string, args ...any)

	// Debugw logs a message with structured key-value pairs at DebugLevel.
	Debugw(msg string, keysAndValues ...any)
	// Infow logs a message with structured key-value pairs at InfoLevel.
	Infow(msg string, keysAndValues ...any)
	// Warnw logs a message with structured key-value pairs at WarnLevel.
	Warnw(msg string, keysAndValues ...any)
	// Errorw logs a message with structured key-value pairs at ErrorLevel.
	Errorw(msg string, keysAndValues ...any)

	// Ctx returns a new logger instance enriched with values from the provided context
	// (e.g., request_id, trace_id).
	Ctx(ctx context.Context) Logger
	// With returns a new logger instance with the given key-value pairs added to all subsequent logs.
	With(args ...any) Logger
	// WithGroup creates a new logger with a named group prefix for all keys (where supported by the engine).
	WithGroup(name string) Logger

	// LogRequest logs an HTTP request with standard observability fields:
	// method, path, status code, and duration.
	LogRequest(ctx context.Context, method, path string, status int, duration time.Duration)

	// Log logs a message at the specified level with structured attributes.
	Log(level Level, msg string, attrs ...Attr)
	// LogAttrs logs a message at the specified level with structured attributes and context enrichment.
	LogAttrs(ctx context.Context, level Level, msg string, attrs ...Attr)
}

// InitLogger initializes a logger instance for the given engine, application name, and environment.
// It applies optional configuration via functional options.
// Returns an error only for engines that require explicit initialization (e.g., Zap).
func InitLogger(engine Engine, appName, env string, opts ...Option) (Logger, error) {
	switch engine {
	case ZapEngine:
		return NewZapAdapter(appName, env, opts...)
	case SlogEngine:
		return NewSlogAdapter(appName, env, opts...), nil
	case ZerologEngine:
		return NewZerologAdapter(appName, env, opts...), nil
	case LogrusEngine:
		return NewLogrusAdapter(appName, env, opts...), nil
	default:
		return NewSlogAdapter(appName, env, opts...), nil
	}
}

// String returns the string representation of the log level.
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// String creates a string attribute for structured logging.
func String(key string, value string) Attr {
	return Attr{Key: key, Value: value}
}

// Int creates an int attribute for structured logging.
func Int(key string, value int) Attr {
	return Attr{Key: key, Value: value}
}

// Int8 creates an int8 attribute for structured logging.
func Int8(key string, value int8) Attr {
	return Attr{Key: key, Value: value}
}

// Int16 creates an int16 attribute for structured logging.
func Int16(key string, value int16) Attr {
	return Attr{Key: key, Value: value}
}

// Int32 creates an int32 attribute for structured logging.
func Int32(key string, value int32) Attr {
	return Attr{Key: key, Value: value}
}

// Int64 creates an int64 attribute for structured logging.
func Int64(key string, value int64) Attr {
	return Attr{Key: key, Value: value}
}

// Uint creates a uint attribute for structured logging.
func Uint(key string, value uint) Attr {
	return Attr{Key: key, Value: value}
}

// Uint8 creates a uint8 attribute for structured logging.
func Uint8(key string, value uint8) Attr {
	return Attr{Key: key, Value: value}
}

// Uint16 creates a uint16 attribute for structured logging.
func Uint16(key string, value uint16) Attr {
	return Attr{Key: key, Value: value}
}

// Uint32 creates a uint32 attribute for structured logging.
func Uint32(key string, value uint32) Attr {
	return Attr{Key: key, Value: value}
}

// Uint64 creates a uint64 attribute for structured logging.
func Uint64(key string, value uint64) Attr {
	return Attr{Key: key, Value: value}
}

// Bool creates a bool attribute for structured logging.
func Bool(key string, value bool) Attr {
	return Attr{Key: key, Value: value}
}

// Time creates a time.Time attribute for structured logging.
func Time(key string, value time.Time) Attr {
	return Attr{Key: key, Value: value}
}

// Any creates an attribute with an arbitrary value for structured logging.
func Any(key string, value any) Attr {
	return Attr{Key: key, Value: value}
}

// Slice creates an attribute with a slice of any type for structured logging.
func Slice[T any](key string, value []T) Attr {
	return Attr{Key: key, Value: value}
}
