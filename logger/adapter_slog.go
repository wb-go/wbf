package logger

import (
	"context"
	"log/slog"
	"time"
)

// newSlogLogger creates a configured log/slog.Logger instance.
func newSlogLogger(appName, env string, cfg *GlobalConfig) *slog.Logger {
	level := toSlogLevel(cfg.Level)
	handler := slog.NewJSONHandler(cfg.GetWriter(), &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler).With(
		slog.String("service", appName),
		slog.String("env", env),
	)
}

// SlogAdapter implements the Logger interface using Go's standard log/slog package.
// It supports structured logging, context propagation, and group nesting.
type SlogAdapter struct {
	logger *slog.Logger
}

// NewSlogAdapter creates a new logger instance using log/slog with JSON encoding.
// It is pre-configured with service name and environment fields.
// File rotation and output options can be customized via functional options.
func NewSlogAdapter(appName, env string, opts ...Option) *SlogAdapter {
	cfg := defaultConfigs()
	for _, opt := range opts {
		opt(cfg)
	}
	return &SlogAdapter{
		logger: newSlogLogger(appName, env, cfg),
	}
}

// Debug logs a message at DebugLevel with the given key-value pairs.
func (a *SlogAdapter) Debug(msg string, args ...any) { a.logger.Debug(msg, args...) }

// Info logs a message at InfoLevel with the given key-value pairs.
func (a *SlogAdapter) Info(msg string, args ...any) { a.logger.Info(msg, args...) }

// Warn logs a message at WarnLevel with the given key-value pairs.
func (a *SlogAdapter) Warn(msg string, args ...any) { a.logger.Warn(msg, args...) }

// Error logs a message at ErrorLevel with the given key-value pairs.
func (a *SlogAdapter) Error(msg string, args ...any) { a.logger.Error(msg, args...) }

// Debugw logs a message at DebugLevel with structured key-value pairs (alias for Debug).
func (a *SlogAdapter) Debugw(msg string, keysAndValues ...any) { a.logger.Debug(msg, keysAndValues...) }

// Infow logs a message at InfoLevel with structured key-value pairs (alias for Info).
func (a *SlogAdapter) Infow(msg string, keysAndValues ...any) { a.logger.Info(msg, keysAndValues...) }

// Warnw logs a message at WarnLevel with structured key-value pairs (alias for Warn).
func (a *SlogAdapter) Warnw(msg string, keysAndValues ...any) { a.logger.Warn(msg, keysAndValues...) }

// Errorw logs a message at ErrorLevel with structured key-value pairs (alias for Error).
func (a *SlogAdapter) Errorw(msg string, keysAndValues ...any) { a.logger.Error(msg, keysAndValues...) }

// Ctx returns a new logger instance enriched with request_id from the context, if present.
// If no request_id is found, returns the original logger.
func (a *SlogAdapter) Ctx(ctx context.Context) Logger {
	requestID := GetRequestID(ctx)
	if requestID == "" {
		return a
	}

	return &SlogAdapter{logger: a.logger.With("request_id", requestID)}
}

// With returns a new logger instance with the given key-value pairs added to all subsequent logs.
func (a *SlogAdapter) With(args ...any) Logger {
	return &SlogAdapter{logger: a.logger.With(args...)}
}

// WithGroup creates a new logger with a named group prefix for all keys.
// This leverages slog's native group nesting support.
func (a *SlogAdapter) WithGroup(name string) Logger {
	return &SlogAdapter{logger: a.logger.WithGroup(name)}
}

// Log logs a message at the specified level with structured attributes.
// It checks if the level is enabled before constructing the log record to avoid unnecessary allocations.
func (a *SlogAdapter) Log(level Level, msg string, attrs ...Attr) {
	slogLevel := toSlogLevel(level)
	if !a.logger.Enabled(context.Background(), slogLevel) {
		return
	}
	a.logger.Log(context.Background(), slogLevel, msg, toSlogAttrs(attrs)...)
}

// LogAttrs logs a message at the specified level with structured attributes and context enrichment.
// It automatically injects request_id from the context if available.
func (a *SlogAdapter) LogAttrs(ctx context.Context, level Level, msg string, attrs ...Attr) {
	a.Ctx(ctx).Log(level, msg, attrs...)
}

// LogRequest logs an HTTP request with standard observability fields:
// method, path, status code, duration, and status class (1xx, 2xx, etc.).
// It automatically includes request_id from the context if present.
func (a *SlogAdapter) LogRequest(ctx context.Context, method, path string, status int, duration time.Duration) {
	a.Ctx(ctx).Info("request",
		"method", method,
		"path", path,
		"status", status,
		"duration", duration,
	)
}

// toSlogLevel converts a logger.Level to the corresponding slog.Level.
// Unknown levels default to LevelInfo.
func toSlogLevel(l Level) slog.Level {
	switch l {
	case DebugLevel:
		return slog.LevelDebug
	case InfoLevel:
		return slog.LevelInfo
	case WarnLevel:
		return slog.LevelWarn
	case ErrorLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// toSlogAttrs converts a slice of Attr to slog arguments (key-value pairs).
// Each Attr is converted to slog.Any(key, value).
func toSlogAttrs(attrs []Attr) []any {
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = slog.Any(attr.Key, attr.Value)
	}
	return args
}
