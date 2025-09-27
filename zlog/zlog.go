// Package zlog предоставляет логирование с использованием Zerolog.
package zlog

import (
	"os"

	"github.com/rs/zerolog"
)

// Logger глобальный экземпляр логгера.
var Logger zerolog.Logger

// Init инициализирует глобальный логгер.
func Init() {
	Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

// InitConsole инициализирует глобальный логгер с ConsoleWriter, цветной и удобный для чтения через Stdout.
func InitConsole() {
	Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
	}).With().
		Timestamp().
		Logger().
		Level(zerolog.TraceLevel)
}

// SetLevel устанавливает уровень логирования для глобального логгера Logger.
//
// Параметры:
//
//	logLevelStr — строковое представление уровня логирования:
//	"trace", "debug", "info", "warn", "error", "fatal", "panic".
func SetLevel(logLevelStr string) error {
	logLevel, err := zerolog.ParseLevel(logLevelStr)
	if err != nil {
		return err
	}

	Logger = Logger.Level(logLevel)

	return nil
}
