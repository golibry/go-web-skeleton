//go:build ignore

package config

import basecfg "github.com/golibry/go-web-skeleton/framework/config"

type Config struct {
	App        basecfg.App        `validate:"required"`
	Log        basecfg.Log        `validate:"required"`
	Database   basecfg.Database   `validate:"required"`
	HttpServer basecfg.HttpServer `validate:"required"`
}

func (c *Config) AppRef() *basecfg.App {
	return &c.App
}

func (c *Config) LogConfig() basecfg.Log {
	return c.Log
}

func (c *Config) DatabaseConfig() *basecfg.Database {
	return &c.Database
}
