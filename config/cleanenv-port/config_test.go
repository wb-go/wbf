package cleanenvport_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cleanenvport "github.com/wb-go/wbf/config/cleanenv-port"
)

type (
	testConfigStructure struct {
		Server   ServerCfg   `yaml:"server"`
		Database DatabaseCfg `yaml:"database"`
		Logger   LoggerCfg   `yaml:"logger"`
	}

	ServerCfg struct {
		Host string `yaml:"host" env:"SERVER_HOST" validate:"required"`
		Port int    `yaml:"port" env:"SERVER_PORT" validate:"required,min=1,max=65535"`
	}

	DatabaseCfg struct {
		DSN             string        `yaml:"dsn" env:"DATABASE_DSN" validate:"required"`
		MaxOpenConns    int           `yaml:"max_open_conns" env:"DATABASE_MAX_OPEN_CONNS" validate:"min=1"`
		MaxIdleConns    int           `yaml:"max_idle_conns" env:"DATABASE_MAX_IDLE_CONNS" validate:"min=1"`
		ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"DATABASE_CONN_MAX_LIFETIME"`
	}

	LoggerCfg struct {
		Level string `yaml:"level" env:"LOG_LEVEL" validate:"required,oneof=debug info warn error"`
	}
)

func TestLoadPath_ValidConfig(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(`
server:
  host: "localhost"
  port: 8080
database:
  dsn: "postgres://..."
  max_open_conns: 10
  max_idle_conns: 5
  conn_max_lifetime: "1h"
logger:
  level: "info"
`))
	require.NoError(t, err)
	tmpFile.Close()

	var cfg testConfigStructure
	err = cleanenvport.LoadPath(tmpFile.Name(), &cfg)
	require.NoError(t, err)
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "postgres://...", cfg.Database.DSN)
	assert.Equal(t, "info", cfg.Logger.Level)
}

func TestLoadPath_ValidationFailed(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(`
server:
  host: ""
  port: 0
database:
  dsn: "postgres://..."
  max_open_conns: 10
  max_idle_conns: 5
  conn_max_lifetime: "1h"
logger:
  level: "info"
`))
	require.NoError(t, err)
	tmpFile.Close()

	var cfg testConfigStructure
	err = cleanenvport.LoadPath(tmpFile.Name(), &cfg)
	require.Error(t, err)
	assert.ErrorIs(t, err, cleanenvport.ErrConfigValidation)
}
