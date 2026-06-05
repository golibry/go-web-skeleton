//go:build ignore

package commands

import (
	"github.com/golibry/go-cli-command/cli"
	frameworkconfig "github.com/golibry/go-web-skeleton/framework/config"
	frameworkhttp "github.com/golibry/go-web-skeleton/framework/http"
	appregistry "{{MODULE_PATH}}/infrastructure/registry"
	approutes "{{MODULE_PATH}}/presentation/http"
)

func Registered(container *appregistry.Container) []cli.Command {
	commands := []cli.Command{
		frameworkhttp.NewCommand(frameworkhttp.Options{
			ServerConfig:   container.Config().HttpServer,
			Logger:         container.Logger(),
			RegisterRoutes: approutes.RegisterRoutes(container),
		}),
		&frameworkconfig.DebugCommand{Cfg: container.Config()},
	}

	return append(commands, migrationCommands(container)...)
}
