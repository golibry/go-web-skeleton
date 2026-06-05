//go:build ignore

package commands

import (
	"github.com/golibry/go-cli-command/cli"
	appregistry "{{MODULE_PATH}}/infrastructure/registry"
)

func migrationCommands(*appregistry.Container) []cli.Command {
	return nil
}
