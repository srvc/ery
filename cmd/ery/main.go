package main

import (
	"log"
	"os"

	"github.com/pkg/errors"
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

	return errors.WithStack(command.Execute())
}

var (
	version, revision, buildDate, releaseType string
)

func createConfig() *ery.Config {
	return &ery.Config{
		InReader:  os.Stdin,
		OutWriter: os.Stdout,
		ErrWriter: os.Stderr,

		Version:     version,
		Revision:    revision,
		BuildDate:   buildDate,
		ReleaseType: releaseType,

		TLD:     "ery",
		Package: "tools.srvc.ery",

		API: ery.APIConfig{
			Hostname: "api.ery.local",
		},

		Daemon: ery.DaemonConfig{
			Name:        "ery",
			Description: "Discover services in local",
		},
	}
}
