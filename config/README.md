# Config

`config` — пакет для управления конфигурацией приложений с использованием [Viper](https://github.com/spf13/viper), поддержки `.env` файлов через [godotenv](https://github.com/joho/godotenv) и флагов командной строки через [pflag](https://github.com/spf13/pflag).

---

## Особенности

- Загрузка конфигурации из YAML/JSON/TOML файлов.
- Поддержка `.env` файлов.
- Автоматическое чтение переменных окружения.
- Определение и парсинг флагов командной строки.
- Получение значений различных типов: `string`, `int`, `bool`, `float64`, `time.Duration`, срезы и т.д.
- Распаковка конфигурации в структуры.
- Значения по умолчанию.
- Поддержка префикса для переменных окружения (`envPrefix`).

---

## Приоритет значений конфигурации

При определении значения для ключа конфигурации используется следующий приоритет (от высокого к низкому):

1. **Флаги командной строки** (переданные при запуске приложения)
2. **Переменные окружения** (с учётом префикса, если задан)
3. **Значения из конфигурационного файла** (`.yaml`, `.json`, `.toml` и т.д.)
4. **Значения по умолчанию**, установленные через `SetDefault`

> Это означает, что флаги командной строки всегда переопределяют переменные окружения и значения из файла, а переменные окружения — значения из файла.

---

## Использование `envPrefix`

`envPrefix` позволяет задать префикс для всех переменных окружения.  

Например, если вы укажете `envPrefix = "APP"` и ключ конфигурации `server.addr`, пакет будет искать переменную окружения:

```bash
APP_SERVER_ADDR
```

## Пример использования пакета `config`

```go
package app

import (
	"fmt"

	"github.com/wb-go/wbf/config"
)

type appConfig struct {
	serverConfig   serverConfig
	loggerConfig   loggerConfig
	postgresConfig postgresConfig
}

type serverConfig struct {
	addr string
}

type loggerConfig struct {
	logLevel string
}

type postgresConfig struct {
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
	port			int
}

func NewAppConfig() (*appConfig, error) {
	envFilePath := "./.env-example"
	appConfigFilePath := "./config-example1.yaml"
	postgresConfigFilePath := "./config-example2.yaml"

	cfg := config.New()

	// Загрузка .env файлов
	if err := cfg.LoadEnvFiles(envFilePath); err != nil {
		return nil, fmt.Errorf("failed to load env files: %w", err)
	}

	// Включение поддержки переменных окружения
	cfg.EnableEnv("")

	// Загрузка файлов конфигурации
	if err := cfg.LoadConfigFiles(appConfigFilePath, postgresConfigFilePath); err != nil {
		return nil, fmt.Errorf("failed to load config files: %w", err)
	}

	// Определение флагов командной строки
	cfg.DefineFlag("p", "srvport", "transport.http.port", 7777, "HTTP server port")
	if err := cfg.ParseFlags(); err != nil {
		return nil, fmt.Errorf("failed to pars flags: %w", err)
	}

	// Распаковка в структуру
	var appConfig *appConfig
	appConfig.serverConfig.addr = cfg.GetString("server.addr")
	appConfig.loggerConfig.logLevel = cfg.GetString("logger.level")
	appConfig.loggerConfig.logLevel = cfg.GetString("logger.level")
	appConfig.postgresConfig.maxOpenConns = cfg.GetInt("postgres.max_open_conns")
	appConfig.postgresConfig.maxOpenConns = cfg.GetInt("postgres.max_idle_conns")
	appConfig.postgresConfig.maxOpenConns = cfg.GetDuration("postgres.conn_max_lifetime")
	appConfig.postgresConfig.port = cfg.GetInt("postgres.port") // из переменной окружения (из файла .env)

	return appConfig, nil
}
```
