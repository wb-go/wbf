// Package config предоставляет управление конфигурацией с использованием Viper.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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

// Load читает конфигурацию из указанного файла и .env файла, если передан.
// Включает поддержку переменных окружения и флагов командной строки.
func (c *Config) Load(configFilePath, envFilePath, envPrefix string) error {
	if envFilePath != "" {
		err := godotenv.Load(envFilePath)
		if err != nil {
			return fmt.Errorf("failed to load .env file %s: %w", envFilePath, err)
		}
	}

	c.v.AutomaticEnv()

	if envPrefix != "" {
		c.v.SetEnvPrefix(envPrefix)
	}

	c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	c.v.SetConfigFile(configFilePath)

	err := c.v.ReadInConfig()
	if err != nil {
		return fmt.Errorf("failed to read config %s: %w", configFilePath, err)
	}

	c.v.BindPFlags(pflag.CommandLine)

	return nil
}

// DefineFlag позволяет объявлять флаги (короткий и длинный) и привязывать их к ключу конфигурации.
func (c *Config) DefineFlag(short, long, configKey string, defaultValue any, usage string) (err error) {
	switch v := defaultValue.(type) {
	case string:
		pflag.StringP(long, short, v, usage)
		err = c.v.BindPFlag(configKey, pflag.Lookup(long))
	case int:
		pflag.IntP(long, short, v, usage)
		err = c.v.BindPFlag(configKey, pflag.Lookup(long))
	case bool:
		pflag.BoolP(long, short, v, usage)
		err = c.v.BindPFlag(configKey, pflag.Lookup(long))
	case float64:
		pflag.Float64P(long, short, v, usage)
		err = c.v.BindPFlag(configKey, pflag.Lookup(long))
	case []string:
		pflag.StringSliceP(long, short, v, usage)
		err = c.v.BindPFlag(configKey, pflag.Lookup(long))
	case []int:
		pflag.IntSliceP(long, short, v, usage)
		err = c.v.BindPFlag(configKey, pflag.Lookup(long))
	case time.Duration:
		pflag.DurationP(long, short, v, usage)
		err = c.v.BindPFlag(configKey, pflag.Lookup(long))
	}
	return
}

// ParseFlags парсит объявленные флаги.
func (c *Config) ParseFlags() {
	pflag.Parse()
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

// SetDefault устанавливает значение по умолчанию для ключа.
func (c *Config) SetDefault(key string, value any) {
	c.v.SetDefault(key, value)
}
