package cli

import (
	"log/slog"
	"os"

	"github.com/golibry/go-cli-command/cli"
)

func Bootstrap(
	logger *slog.Logger,
	availableCommands []cli.Command,
) {
	// Create a command registry and register commands
	commandRegistry := cli.NewCommandsRegistry()

	for _, cmd := range availableCommands {
		err := commandRegistry.Register(cmd)
		if err != nil {
			logger.Error(
				"Failed to register command",
				"command",
				cmd.Id(),
				"error",
				err,
			)
			os.Exit(1)
		}
	}

	// Bootstrap and run the CLI application
	// os.Args[1: ] is mandatory to remove the program name from the args slice
	cli.Bootstrap(os.Args[1:], commandRegistry, os.Stdout, os.Exit)
}
