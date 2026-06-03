package config

import (
	"io"

	"github.com/golibry/go-cli-command/cli"
)

type DebugCommand struct {
	cli.CommandWithoutFlags
	Cfg AppProvider
}

func (c *DebugCommand) Id() string {
	return "config:debug"
}

func (c *DebugCommand) Description() string {
	return "Debugs the configuration"
}

func (c *DebugCommand) Exec(writer io.Writer) error {
	_, _ = writer.Write([]byte("Config:\n"))
	_, _ = writer.Write([]byte(DebugAny(c.Cfg)))
	_, _ = writer.Write([]byte("\n"))
	return nil
}
