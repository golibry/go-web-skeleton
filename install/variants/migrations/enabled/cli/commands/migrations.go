//go:build ignore

package commands

import (
	"github.com/golibry/go-cli-command/cli"
	frameworkmigrations "github.com/golibry/go-web-skeleton/framework/migrations"
	appregistry "{{MODULE_PATH}}/infrastructure/registry"
)

func migrationCommands(container *appregistry.Container) []cli.Command {
	return []cli.Command{
		frameworkmigrations.NewCommand(container.Config().Database),
	}
}
