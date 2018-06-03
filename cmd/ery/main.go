package main

import (
	"log"
	"os"

	"github.com/srvc/ery/pkg/ery/cmd"
	"github.com/srvc/ery/pkg/ery/di"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	command := cmd.NewEryCommand(
		di.NewAppComponent(
			os.Stdin,
			os.Stdout,
			os.Stderr,
		),
	)

	return command.Execute()
}
