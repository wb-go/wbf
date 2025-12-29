package logger

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

func newLogrusLogger(appName, env string, cfg *GlobalConfig) *logrus.Entry {
	l := logrus.New()
	l.SetOutput(cfg.GetWriter())
	l.SetLevel(toLogrusLevel(cfg.Level))

	return l.WithFields(logrus.Fields{
		"service": appName,
		"env":     env,
	})
}

// LogrusAdapter implements the Logger interface using github.com/sirupsen/logrus as the underlying engine.
// It supports structured logging and context propagation, though some features (like WithGroup)
// have limited support due to logrus's design.
type LogrusAdapter struct {
	entry *logrus.Entry
}

// NewLogrusAdapter creates a new logger instance using logrus.
// It is pre-configured with service name and environment fields.
// File rotation and output options can be customized via functional options.
func NewLogrusAdapter(appName, env string, opts ...Option) *LogrusAdapter {
	cfg := defaultConfigs()
	for _, opt := range opts {
		opt(cfg)
	}
	return &LogrusAdapter{
		entry: newLogrusLogger(appName, env, cfg),
	}
}

// Debug logs a message at DebugLevel with the given key-value pairs.
// Note: logrus does not natively support message + structured args in a single call,
// so the message is treated as part of the structured data.
func (a *LogrusAdapter) Debug(msg string, args ...any) {
	allArgs := append([]any{"msg", msg}, args...)
	a.entry.Debug(allArgs...)
}

// Info logs a message at InfoLevel with the given key-value pairs.
// Note: logrus does not natively support message + structured args in a single call,
// so the message is treated as part of the structured data.
func (a *LogrusAdapter) Info(msg string, args ...any) {
	allArgs := append([]any{"msg", msg}, args...)
	a.entry.Info(allArgs...)
}

// Warn logs a message at WarnLevel with the given key-value pairs.
// Note: logrus does not natively support message + structured args in a single call,
// so the message is treated as part of the structured data.
func (a *LogrusAdapter) Warn(msg string, args ...any) {
	allArgs := append([]any{"msg", msg}, args...)
	a.entry.Warn(allArgs...)
}

// Error logs a message at ErrorLevel with the given key-value pairs.
// Note: logrus does not natively support message + structured args in a single call,
// so the message is treated as part of the structured data.
func (a *LogrusAdapter) Error(msg string, args ...any) {
	allArgs := append([]any{"msg", msg}, args...)
	a.entry.Error(allArgs...)
}

// Debugw logs a message at DebugLevel with structured key-value pairs.
func (a *LogrusAdapter) Debugw(msg string, kvs ...any) { a.With(kvs...).Debug(msg) }

// Infow logs a message at InfoLevel with structured key-value pairs.
func (a *LogrusAdapter) Infow(msg string, kvs ...any) { a.With(kvs...).Info(msg) }

// Warnw logs a message at WarnLevel with structured key-value pairs.
func (a *LogrusAdapter) Warnw(msg string, kvs ...any) { a.With(kvs...).Warn(msg) }

// Errorw logs a message at ErrorLevel with structured key-value pairs.
func (a *LogrusAdapter) Errorw(msg string, kvs ...any) { a.With(kvs...).Error(msg) }

// Ctx returns a new logger instance enriched with request_id from the context, if present.
// If no request_id is found, returns the original logger.
func (a *LogrusAdapter) Ctx(ctx context.Context) Logger {
	requestID := GetRequestID(ctx)
	if requestID == "" {
		return a
	}
	return &LogrusAdapter{
		entry: a.entry.WithField("request_id", requestID),
	}
}

// With returns a new logger instance with the given key-value pairs added to all subsequent logs.
// Only string keys are supported; non-string keys are silently ignored.
// If the number of arguments is odd, the last key is ignored.
func (a *LogrusAdapter) With(args ...any) Logger {
	if len(args) == 0 {
		return a
	}

	fields := make(logrus.Fields)
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue
		}
		if i+1 < len(args) {
			fields[key] = args[i+1]
		}
	}

	return &LogrusAdapter{
		entry: a.entry.WithFields(fields),
	}
}

// WithGroup has no effect in logrus, as it does not support hierarchical field grouping.
// This method returns the original logger unchanged.
func (a *LogrusAdapter) WithGroup(_ string) Logger {
	return a
}

// Log logs a message at the specified level with structured attributes.
func (a *LogrusAdapter) Log(level Level, msg string, attrs ...Attr) {
	fields := make(logrus.Fields)
	for _, attr := range attrs {
		fields[attr.Key] = attr.Value
	}

	a.entry.WithFields(fields).Log(toLogrusLevel(level), msg)
}

// LogAttrs logs a message at the specified level with structured attributes and context enrichment.
// It automatically injects request_id from the context if available.
func (a *LogrusAdapter) LogAttrs(ctx context.Context, level Level, msg string, attrs ...Attr) {
	a.Ctx(ctx).Log(level, msg, attrs...)
}

// LogRequest logs an HTTP request with standard observability fields:
// method, path, status code, and duration.
// It automatically includes request_id from the context if present.
func (a *LogrusAdapter) LogRequest(ctx context.Context, method, path string, status int, duration time.Duration) {
	a.Ctx(ctx).With(
		"method", method,
		"path", path,
		"status", status,
		"duration", duration,
	).Info("http request")
}

// toLogrusLevel converts a logger.Level to the corresponding logrus.Level.
// Unknown levels default to InfoLevel.
func toLogrusLevel(l Level) logrus.Level {
	switch l {
	case DebugLevel:
		return logrus.DebugLevel
	case InfoLevel:
		return logrus.InfoLevel
	case WarnLevel:
		return logrus.WarnLevel
	case ErrorLevel:
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}
