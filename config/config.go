// Package config предоставляет управление конфигурацией с использованием Viper.
package config

import (
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

// Load читает конфигурацию из указанного файла.
func (c *Config) Load(path string) error {
	c.v.SetConfigFile(path)
	return c.v.ReadInConfig()
}

// GetString получает строковое значение из конфигурации по ключу.
func (c *Config) GetString(key string) string {
	return c.v.GetString(key)
}

func (c *Config) GetInt(key string) int {
	return c.v.GetInt(key)
}

func (c *Config) Unmarshal(rawVal any, opts ...viper.DecoderConfigOption) error {
	return c.v.Unmarshal(rawVal, opts...)
}

func (c *Config) SetDefault(key string, value any) {
	c.v.SetDefault(key, value)
}
