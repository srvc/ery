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

const (
	version = "0.0.1"
)

var (
	revision, buildDate string
)

func createConfig() *ery.Config {
	return &ery.Config{
		InReader:  os.Stdin,
		OutWriter: os.Stdout,
		ErrWriter: os.Stderr,

		Version:   version,
		Revision:  revision,
		BuildDate: buildDate,

		TLD:     "ery",
		Package: "tools.srvc.ery",

		DNS:   &ery.DNSConfig{},
		Proxy: &ery.ProxyConfig{},

		API: &ery.APIConfig{
			Hostname: "api.ery.local",
		},

		Daemon: &ery.DaemonConfig{
			Name:        "ery",
			Description: "Discover services in local",
		},
	}
}
