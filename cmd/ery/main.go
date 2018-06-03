package main

import (
	"log"
	"os"

	"github.com/srvc/ery/pkg/ery"
	"github.com/srvc/ery/pkg/ery/cmd"
	"github.com/srvc/ery/pkg/ery/di"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := createConfig()
	component := di.NewAppComponent(cfg)
	command := cmd.NewEryCommand(component)

	return command.Execute()
}

var (
	version, revision, builldDate, releaseType string
)

func createConfig() *ery.Config {
	return &ery.Config{
		InReader:  os.Stdin,
		OutWriter: os.Stdout,
		ErrWriter: os.Stderr,

		Version:     version,
		Revision:    revision,
		BuildDate:   builldDate,
		ReleaseType: releaseType,

		API: ery.APIConfig{
			Hostname: "api.ery.local",
		},
	}
}
