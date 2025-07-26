package main

import (
	"fmt"
	"os"

	"github.com/golibry/go-cli-command/cli"
	"github.com/golibry/go-web-skeleton/infrastructure/registry"
)

// main initializes the CLI application with a lightweight dependency injection container
// and sets up the command structure for all CLI commands.
func main() {
	// Initialize a lightweight CLI container (without a database)
	container, err := registry.NewContainer()
	if err != nil {
		_, _ = fmt.Fprintf(
			os.Stderr,
			"Could not start CLI application. Error building container registry: %s\n",
			err,
		)
		os.Exit(1)
	}

	// Ensure proper cleanup of resources on exit
	defer func() {
		if err := container.Close(); err != nil {
			container.Logger().Error("Failed to close container during shutdown", "error", err)
		}
	}()

	container.Logger().Info("Starting CLI application")

	// Create a command registry and register commands
	commandRegistry := cli.NewCommandsRegistry()

	for _, cmd := range availableCommands(container) {
		err := commandRegistry.Register(cmd)
		if err != nil {
			container.Logger().Error(
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
	// os.Args[1:] is mandatory to remove the program name from the args slice
	cli.Bootstrap(os.Args[1:], commandRegistry, os.Stdout, os.Exit)
}

// availableCommands returns a list of CLI commands available in the application.
func availableCommands(container *registry.Container) []cli.Command {
	return []cli.Command{
		// Add more commands here as needed
		// Example:
		// newServerCommand(container),
		// newMigrateCommand(container),
	}
}
