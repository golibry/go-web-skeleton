package registry

import (
	"fmt"

	"github.com/golibry/go-web-skeleton/config"
)

type ConfigService struct {
	config config.Config
}

func NewConfigService() (*ConfigService, error) {
	cfg, err := config.BuildConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build configuration: %w", err)
	}
	return &ConfigService{config: cfg}, nil
}

func (c *ConfigService) Config() config.Config {
	return c.config
}