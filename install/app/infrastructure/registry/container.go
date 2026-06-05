//go:build ignore

package registry

import (
	frameworkapp "github.com/golibry/go-web-skeleton/framework/app"
	appconfig "{{MODULE_PATH}}/config"
)

type Container = frameworkapp.Container[*appconfig.Config]

func NewContainer(cfg *appconfig.Config) (*Container, error) {
	return frameworkapp.NewContainerFromConfig(cfg)
}
