//go:build ignore

package main

import (
	"log"

	frameworkcli "github.com/golibry/go-web-skeleton/framework/cli"
	frameworkconfig "github.com/golibry/go-web-skeleton/framework/config"
	appcommands "{{MODULE_PATH}}/cli/commands"
	appconfig "{{MODULE_PATH}}/config"
	_ "{{MODULE_PATH}}/infrastructure/database"
	appregistry "{{MODULE_PATH}}/infrastructure/registry"
)

func main() {
	var cfg appconfig.Config
	if err := frameworkconfig.BuildCustom(&cfg); err != nil {
		log.Fatal(err)
	}

	container, err := appregistry.NewContainer(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = container.Close()
	}()

	frameworkcli.Bootstrap(container.Logger(), appcommands.Registered(container))
}
