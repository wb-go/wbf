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

Например, если вы укажете `envPrefix = "MYAPP"` и ключ конфигурации `server.addr`, пакет будет искать переменную окружения:

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
	serverConfig serverConfig
	loggerConfig loggerConfig
}

type serverConfig struct {
	addr string
}

type loggerConfig struct {
	logLevel string
}

func NewAppConfig(configFilePath, envFilePath, envPrefix string) (*appConfig, error) {
	appConfig := &appConfig{}

	cfg := config.New()

	cfg.DefineFlag("a", "addr", "server.addr", ":7777", "Server address")

	cfg.ParseFlags()

	err := cfg.Load(configFilePath, envFilePath, envPrefix)
	if err != nil {
		return appConfig, fmt.Errorf("failed to load config: %w", err)
	}

	appConfig.serverConfig.addr = cfg.GetString("server.addr")
	appConfig.loggerConfig.logLevel = cfg.GetString("logger.level")

	return appConfig, nil
}
```
