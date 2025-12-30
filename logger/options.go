package logger

import (
	"io"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Default values for log file rotation.
const (
	_defaultMaxSize    = 100 // in megabytes
	_defaultMaxBackups = 7   // number of backup files
	_defaultMaxAge     = 30  // in days
)

// GlobalConfig holds configuration parameters for logger initialization,
// including log level, application metadata, and file rotation settings.
type GlobalConfig struct {
	// Level is the minimum severity level for log records to be output.
	Level Level
	// AppName identifies the application for observability purposes.
	AppName string
	// Env specifies the deployment environment (e.g., "prod", "staging", "dev").
	Env string
	// Filename is the path to the log file. If empty, file logging is disabled.
	Filename string
	// MaxSize is the maximum size in megabytes of the log file before rotation.
	MaxSize int
	// MaxBackups is the maximum number of old log files to retain.
	MaxBackups int
	// MaxAge is the maximum number of days to retain old log files.
	MaxAge int
	// Compress determines whether rotated log files are gzipped.
	Compress bool
	// Stdout enables logging to standard output in addition to file logging.
	Stdout bool
}

// Option represents a functional configuration option for the logger.
type Option func(*GlobalConfig)

// defaultConfigs returns a GlobalConfig with safe default values for production use.
func defaultConfigs() *GlobalConfig {
	return &GlobalConfig{
		Level:      InfoLevel,
		MaxSize:    _defaultMaxSize,
		MaxBackups: _defaultMaxBackups,
		MaxAge:     _defaultMaxAge,
		Compress:   true,
		Stdout:     true,
	}
}

// WithLevel sets the minimum log level for the logger.
func WithLevel(l Level) Option {
	return func(c *GlobalConfig) { c.Level = l }
}

// WithRotation configures file-based log rotation with the given parameters.
// The filename must be a valid path. MaxSize is in megabytes, MaxAge in days.
func WithRotation(filename string, maxSize, maxBackups, maxAge int) Option {
	return func(c *GlobalConfig) {
		c.Filename = filename
		c.MaxSize = maxSize
		c.MaxBackups = maxBackups
		c.MaxAge = maxAge
	}
}

// GetWriter returns an io.Writer that combines stdout and file logging as configured.
// If both are enabled, logs are written to both destinations simultaneously using io.MultiWriter.
// File rotation is handled by lumberjack.Logger.
func (c *GlobalConfig) GetWriter() io.Writer {
	var writers []io.Writer
	if c.Stdout {
		writers = append(writers, os.Stdout)
	}
	if c.Filename != "" {
		writers = append(writers, &lumberjack.Logger{
			Filename:   c.Filename,
			MaxSize:    c.MaxSize,
			MaxBackups: c.MaxBackups,
			MaxAge:     c.MaxAge,
			Compress:   c.Compress,
		})
	}
	return io.MultiWriter(writers...)
}
