package http

import (
	"io"

	"github.com/golibry/go-cli-command/cli"
)

type Command struct {
	cli.CommandWithoutFlags
	Options Options
}

func (c *Command) Id() string {
	return "http:start"
}

func (c *Command) Description() string {
	return "Starts the HTTP server"
}

func (c *Command) Exec(_ io.Writer) error {
	Start(c.Options)
	return nil
}
