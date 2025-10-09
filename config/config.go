// Package config предоставляет управление конфигурацией с использованием Viper.
package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	ErrShortFlagLength = errors.New("short flag must be one character")
	ErrUnsupportedFlag = errors.New("unsupported flag type")
	ErrFlagNotFound    = errors.New("flag not found")
	ErrLoadEnvFile     = errors.New("failed to load env file")
	ErrLoadConfigFile  = errors.New("failed to load config file")
)

// Config оборачивает экземпляр конфигурации Viper.
type Config struct {
	v *viper.Viper
}

// New создает новый экземпляр Config.
func New() *Config {
	v := viper.New()
	return &Config{v: v}
}

// LoadEnvFiles загружает один или несколько файлов .env в os.Environ().
func (c *Config) LoadEnvFiles(paths ...string) error {
	for _, path := range paths {
		if err := godotenv.Load(path); err != nil {
			return fmt.Errorf("%w %s: %v", ErrLoadEnvFile, path, err)
		}
	}
	return nil
}

// LoadConfigFiles загружает и объединяет несколько файлов конфигурации.
func (c *Config) LoadConfigFiles(paths ...string) error {
	for _, cfgPath := range paths {
		c.v.SetConfigFile(cfgPath)
		if err := c.v.MergeInConfig(); err != nil {
			return fmt.Errorf("%w %s: %v", ErrLoadConfigFile, cfgPath, err)
		}
	}
	return nil
}

// EnableEnv включает автоматическую загрузку переменных окружения.
// envPrefix (если задан) используется как префикс для всех ключей.
func (c *Config) EnableEnv(envPrefix string) {
	c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if envPrefix != "" {
		c.v.SetEnvPrefix(envPrefix)
	}
	c.v.AutomaticEnv()
}

// DefineFlag позволяет объявлять флаги (короткий и длинный) и привязывать их к ключу конфигурации.
func (c *Config) DefineFlag(short, long, configKey string, defaultValue any, usage string) error {
	if len(short) > 1 {
		return fmt.Errorf("%w: got %s", ErrShortFlagLength, short)
	}
	switch v := defaultValue.(type) {
	case string:
		pflag.StringP(long, short, v, usage)
	case int:
		pflag.IntP(long, short, v, usage)
	case bool:
		pflag.BoolP(long, short, v, usage)
	case float64:
		pflag.Float64P(long, short, v, usage)
	case []string:
		pflag.StringSliceP(long, short, v, usage)
	case []int:
		pflag.IntSliceP(long, short, v, usage)
	case time.Duration:
		pflag.DurationP(long, short, v, usage)
	default:
		return fmt.Errorf("%w: %T", ErrUnsupportedFlag, v)
	}

	flag := pflag.Lookup(long)
	if flag == nil {
		return fmt.Errorf("%w %q", ErrFlagNotFound, long)
	}

	return c.v.BindPFlag(configKey, flag)
}

// ParseFlags парсит объявленные флаги.
func (c *Config) ParseFlags() error {
	pflag.Parse()
	return c.v.BindPFlags(pflag.CommandLine)
}

// GetString получает строковое значение из конфигурации по ключу.
func (c *Config) GetString(key string) string {
	return c.v.GetString(key)
}

// GetInt получает целочисленное значение из конфигурации по ключу.
func (c *Config) GetInt(key string) int {
	return c.v.GetInt(key)
}

// GetBool получает логическое значение из конфигурации по ключу.
func (c *Config) GetBool(key string) bool {
	return c.v.GetBool(key)
}

// GetFloat64 получает вещественное значение из конфигурации по ключу.
func (c *Config) GetFloat64(key string) float64 {
	return c.v.GetFloat64(key)
}

// GetTime получает значение времени из конфигурации по ключу.
func (c *Config) GetTime(key string) time.Time {
	return c.v.GetTime(key)
}

// GetDuration получает значение продолжительности из конфигурации по ключу.
func (c *Config) GetDuration(key string) time.Duration {
	return c.v.GetDuration(key)
}

// GetStringSlice получает срез строк из конфигурации по ключу.
func (c *Config) GetStringSlice(key string) []string {
	return c.v.GetStringSlice(key)
}

// GetIntSlice получает срез целых чисел из конфигурации по ключу.
func (c *Config) GetIntSlice(key string) []int {
	return c.v.GetIntSlice(key)
}

// Unmarshal позволяет распаковать конфигурацию в структуру.
func (c *Config) Unmarshal(rawVal any, opts ...viper.DecoderConfigOption) error {
	return c.v.Unmarshal(rawVal, opts...)
}

// UnmarshalKey позволяет распаковать часть конфигурации по ключу в структуру.
func (c *Config) UnmarshalKey(key string, rawVal any, opts ...viper.DecoderConfigOption) error {
	return c.v.UnmarshalKey(key, rawVal, opts...)
}

// UnmarshalExact позволяет строго распаковать конфигурацию в структуру.
// Вернёт ошибку, если в файле есть ключи, которых нет в структуре.
func (c *Config) UnmarshalExact(rawVal any, opts ...viper.DecoderConfigOption) error {
	return c.v.UnmarshalExact(rawVal, opts...)
}

// SetDefault устанавливает значение по умолчанию для ключа.
func (c *Config) SetDefault(key string, value any) {
	c.v.SetDefault(key, value)
}
