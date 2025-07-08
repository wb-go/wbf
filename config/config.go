package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	v *viper.Viper
}

func New() *Config {
	v := viper.New()
	return &Config{v: v}
}

func (c *Config) Load(path string) error {
	c.v.SetConfigFile(path)
	return c.v.ReadInConfig()
}

func (c *Config) GetString(key string) string {
	return c.v.GetString(key)
}
