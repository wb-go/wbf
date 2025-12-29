package logger

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// _argsPairs indicates that key-value arguments are expected in pairs.
	_argsPairs = 2
)

// ZapLogger holds a zap.Logger instance along with its sugared version and effective log level.
// It serves as an internal wrapper to avoid repeated sugar creation.
type ZapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
	level  zapcore.Level
}

// NewZapLogger creates a new zap.Logger configured with JSON encoding, structured fields,
// and the given application metadata. It supports file rotation and console output via options.
// The logger includes caller information and automatic stack traces for errors.
func NewZapLogger(appName, env string, opts ...Option) (*ZapLogger, error) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		FunctionKey:   zapcore.OmitKey,
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime:    zapcore.ISO8601TimeEncoder,
		EncodeCaller:  zapcore.ShortCallerEncoder,
	}

	cfg := defaultConfigs()
	for _, opt := range opts {
		opt(cfg)
	}

	zapLevel := toZapLevel(cfg.Level)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(cfg.GetWriter()),
		zapLevel,
	)

	l := zap.New(core,
		zap.Fields(
			zap.String("service", appName),
			zap.String("env", env),
		),
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)

	return &ZapLogger{
		logger: l,
		sugar:  l.Sugar(),
		level:  zapLevel,
	}, nil
}

// ZapAdapter implements the Logger interface using go.uber.org/zap as the underlying engine.
type ZapAdapter struct {
	zapLogger *ZapLogger
}

// NewZapAdapter creates a new logger adapter using zap.
// It returns an error if logger initialization fails (though currently it does not).
func NewZapAdapter(appName, env string, opts ...Option) (*ZapAdapter, error) {
	zl, err := NewZapLogger(appName, env, opts...)
	if err != nil {
		return nil, err
	}
	return &ZapAdapter{zapLogger: zl}, nil
}

// Debug logs a message at DebugLevel with the given key-value pairs.
func (a *ZapAdapter) Debug(msg string, args ...any) { a.zapLogger.sugar.Debugw(msg, args...) }

// Info logs a message at InfoLevel with the given key-value pairs.
func (a *ZapAdapter) Info(msg string, args ...any) { a.zapLogger.sugar.Infow(msg, args...) }

// Warn logs a message at WarnLevel with the given key-value pairs.
func (a *ZapAdapter) Warn(msg string, args ...any) { a.zapLogger.sugar.Warnw(msg, args...) }

// Error logs a message at ErrorLevel with the given key-value pairs.
func (a *ZapAdapter) Error(msg string, args ...any) { a.zapLogger.sugar.Errorw(msg, args...) }

// Debugw logs a message at DebugLevel with structured key-value pairs (alias for Debug).
func (a *ZapAdapter) Debugw(msg string, args ...any) { a.zapLogger.sugar.Debugw(msg, args...) }

// Infow logs a message at InfoLevel with structured key-value pairs (alias for Info).
func (a *ZapAdapter) Infow(msg string, args ...any) { a.zapLogger.sugar.Infow(msg, args...) }

// Warnw logs a message at WarnLevel with structured key-value pairs (alias for Warn).
func (a *ZapAdapter) Warnw(msg string, args ...any) { a.zapLogger.sugar.Warnw(msg, args...) }

// Errorw logs a message at ErrorLevel with structured key-value pairs (alias for Error).
func (a *ZapAdapter) Errorw(msg string, args ...any) { a.zapLogger.sugar.Errorw(msg, args...) }

// Ctx returns a new logger instance enriched with request_id from the context, if present.
// If no request_id is found, returns the original logger.
func (a *ZapAdapter) Ctx(ctx context.Context) Logger {
	requestID := GetRequestID(ctx)
	if requestID == "" {
		return a
	}

	newLogger := a.zapLogger.logger.With(zap.String("request_id", requestID))
	return &ZapAdapter{
		zapLogger: &ZapLogger{
			logger: newLogger,
			sugar:  newLogger.Sugar(),
			level:  a.zapLogger.level,
		},
	}
}

// With returns a new logger instance with the given key-value pairs added to all subsequent logs.
// Invalid key types default to "UNKNOWN".
func (a *ZapAdapter) With(args ...any) Logger {
	newLogger := a.zapLogger.logger.With(toZapFields(args)...)
	return &ZapAdapter{
		zapLogger: &ZapLogger{
			logger: newLogger,
			sugar:  newLogger.Sugar(),
			level:  a.zapLogger.level,
		},
	}
}

// WithGroup creates a new logger with a named group prefix for all keys.
// In zap, this is implemented using zap.Namespace.
func (a *ZapAdapter) WithGroup(name string) Logger {
	newLogger := a.zapLogger.logger.With(zap.Namespace(name))
	return &ZapAdapter{
		zapLogger: &ZapLogger{
			logger: newLogger,
			sugar:  newLogger.Sugar(),
			level:  a.zapLogger.level,
		},
	}
}

// Log logs a message at the specified level with structured attributes.
// It uses zap's Check/Write pattern for zero-allocation logging when the level is disabled.
func (a *ZapAdapter) Log(level Level, msg string, attrs ...Attr) {
	zapLevel := toZapLevel(level)
	if ce := a.zapLogger.logger.Check(zapLevel, msg); ce != nil {
		ce.Write(toZapFieldsFromAttrs(attrs)...)
	}
}

// LogAttrs logs a message at the specified level with structured attributes and context enrichment.
// It automatically injects request_id from the context if available.
func (a *ZapAdapter) LogAttrs(ctx context.Context, level Level, msg string, attrs ...Attr) {
	l := a.Ctx(ctx)
	l.Log(level, msg, attrs...)
}

// LogRequest logs an HTTP request with standard observability fields:
// method, path, status code, and duration.
// It automatically includes request_id from the context if present.
func (a *ZapAdapter) LogRequest(ctx context.Context, method, path string, status int, duration time.Duration) {
	a.Ctx(ctx).Info("request",
		"method", method,
		"path", path,
		"status", status,
		"duration", duration,
	)
}

// toZapLevel converts a logger.Level to the corresponding zapcore.Level.
// Unknown levels default to InfoLevel.
func toZapLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// toZapFields converts a slice of key-value pairs into zap.Fields.
// If the number of arguments is odd, a "<missing>" value is appended for the last key.
// Non-string keys are converted to "UNKNOWN".
func toZapFields(args []any) []zap.Field {
	if len(args)%2 != 0 {
		args = append(args, "<missing>")
	}
	fields := make([]zap.Field, 0, len(args)/_argsPairs)
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			key = "UNKNOWN"
		}
		fields = append(fields, zap.Any(key, args[i+1]))
	}
	return fields
}

// toZapFieldsFromAttrs converts a slice of Attr to zap.Fields.
func toZapFieldsFromAttrs(attrs []Attr) []zap.Field {
	fields := make([]zap.Field, 0, len(attrs))
	for _, a := range attrs {
		fields = append(fields, zap.Any(a.Key, a.Value))
	}
	return fields
}
