package main

import (
	"os"

	"github.com/golibry/go-web-skeleton/migrations/runner"
)

func main() {
	runner.RunMigrations(os.Args[1:], os.Exit, os.Stdout)
}
