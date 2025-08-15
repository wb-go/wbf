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
