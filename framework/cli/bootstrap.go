package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/golibry/go-cli-command/cli"
)

func NewRegistry(commands []cli.Command) (*cli.CommandsRegistry, error) {
	commandRegistry := cli.NewCommandsRegistry()

	for _, cmd := range commands {
		if err := commandRegistry.Register(cmd); err != nil {
			return nil, fmt.Errorf("failed to register command: %w", err)
		}
	}

	return commandRegistry, nil
}

func Bootstrap(
	logger *slog.Logger,
	availableCommands []cli.Command,
) {
	commandRegistry, err := NewRegistry(availableCommands)
	if err != nil {
		logger.Error("Failed to register commands", "error", err)
		os.Exit(1)
	}

	// Bootstrap and run the CLI application
	// os.Args[1: ] is mandatory to remove the program name from the args slice
	cli.Bootstrap(os.Args[1:], commandRegistry, os.Stdout, os.Exit)
}
