package main

import (
	"log"
	"os"

	"github.com/srvc/ery/pkg/ery/cmd"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	command := cmd.NewEryCommand(
		os.Stdin,
		os.Stdout,
		os.Stderr,
	)

	return command.Execute()
}
