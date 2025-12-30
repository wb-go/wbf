package logger

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

// newZerologLogger creates a configured rs/zerolog.Logger instance.
func newZerologLogger(appName, env string, cfg *GlobalConfig) zerolog.Logger {
	level := toZerologLevel(cfg.Level)
	return zerolog.New(cfg.GetWriter()).Level(level).With().
		Timestamp().
		Str("service", appName).
		Str("env", env).
		Logger()
}

// ZerologAdapter implements the Logger interface using github.com/rs/zerolog as the underlying engine.
// It supports structured logging, context propagation (e.g., request_id), and log rotation.
type ZerologAdapter struct {
	logger zerolog.Logger
}

// NewZerologAdapter creates a new logger instance using zerolog.
// It is pre-configured with timestamp, service name, and environment fields.
// File rotation and output options can be customized via functional options.
func NewZerologAdapter(appName, env string, opts ...Option) *ZerologAdapter {
	cfg := defaultConfigs()
	for _, opt := range opts {
		opt(cfg)
	}
	return &ZerologAdapter{
		logger: newZerologLogger(appName, env, cfg),
	}
}

// Debug logs a message at DebugLevel with the given key-value pairs.
func (a *ZerologAdapter) Debug(msg string, args ...any) { a.logger.Debug().Fields(args).Msg(msg) }

// Info logs a message at InfoLevel with the given key-value pairs.
func (a *ZerologAdapter) Info(msg string, args ...any) { a.logger.Info().Fields(args).Msg(msg) }

// Warn logs a message at WarnLevel with the given key-value pairs.
func (a *ZerologAdapter) Warn(msg string, args ...any) { a.logger.Warn().Fields(args).Msg(msg) }

// Error logs a message at ErrorLevel with the given key-value pairs.
func (a *ZerologAdapter) Error(msg string, args ...any) { a.logger.Error().Fields(args).Msg(msg) }

// Debugw logs a message at DebugLevel with structured key-value pairs (alias for Debug).
func (a *ZerologAdapter) Debugw(msg string, kvs ...any) { a.logger.Debug().Fields(kvs).Msg(msg) }

// Infow logs a message at InfoLevel with structured key-value pairs (alias for Info).
func (a *ZerologAdapter) Infow(msg string, kvs ...any) { a.logger.Info().Fields(kvs).Msg(msg) }

// Warnw logs a message at WarnLevel with structured key-value pairs (alias for Warn).
func (a *ZerologAdapter) Warnw(msg string, kvs ...any) { a.logger.Warn().Fields(kvs).Msg(msg) }

// Errorw logs a message at ErrorLevel with structured key-value pairs (alias for Error).
func (a *ZerologAdapter) Errorw(msg string, kvs ...any) { a.logger.Error().Fields(kvs).Msg(msg) }

// Ctx returns a new logger instance enriched with request_id from the context, if present.
// If no request_id is found, returns the original logger.
func (a *ZerologAdapter) Ctx(ctx context.Context) Logger {
	requestID := GetRequestID(ctx)
	if requestID == "" {
		return a
	}
	return &ZerologAdapter{logger: a.logger.With().Str("request_id", requestID).Logger()}
}

// With returns a new logger instance with the given key-value pairs added to all subsequent logs.
func (a *ZerologAdapter) With(args ...any) Logger {
	return &ZerologAdapter{logger: a.logger.With().Fields(args).Logger()}
}

// WithGroup creates a new logger with a named group prefix for all keys.
// In zerolog, this is implemented as a dictionary field with the given name.
func (a *ZerologAdapter) WithGroup(name string) Logger {
	return &ZerologAdapter{logger: a.logger.With().Dict(name, zerolog.Dict()).Logger()}
}

// Log logs a message at the specified level with structured attributes.
// If the level is below the configured minimum, the log is silently dropped.
func (a *ZerologAdapter) Log(level Level, msg string, attrs ...Attr) {
	zlLevel := toZerologLevel(level)
	if zlLevel == zerolog.Disabled {
		return
	}

	event := a.logger.WithLevel(zlLevel)
	for _, attr := range attrs {
		event.Any(attr.Key, attr.Value)
	}
	event.Msg(msg)
}

// LogAttrs logs a message at the specified level with structured attributes and context enrichment.
// It automatically injects request_id from the context if available.
func (a *ZerologAdapter) LogAttrs(ctx context.Context, level Level, msg string, attrs ...Attr) {
	a.Ctx(ctx).Log(level, msg, attrs...)
}

// LogRequest logs an HTTP request with standard observability fields:
// method, path, status code, and duration.
// It automatically includes request_id from the context if present.
func (a *ZerologAdapter) LogRequest(ctx context.Context, method, path string, status int, duration time.Duration) {
	a.Ctx(ctx).Info("http request",
		"method", method,
		"path", path,
		"status", status,
		"duration", duration,
	)
}

// toZerologLevel converts a logger.Level to the corresponding zerolog.Level.
// Unknown levels default to InfoLevel.
func toZerologLevel(l Level) zerolog.Level {
	switch l {
	case DebugLevel:
		return zerolog.DebugLevel
	case InfoLevel:
		return zerolog.InfoLevel
	case WarnLevel:
		return zerolog.WarnLevel
	case ErrorLevel:
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}
