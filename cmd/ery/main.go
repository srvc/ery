package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/srvc/ery/pkg/ery"
	"github.com/srvc/ery/pkg/ery/cmd"
	"github.com/srvc/ery/pkg/util/cliutil"
)

func main() {
	var exitCode int
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		exitCode = 1
	}
	os.Exit(exitCode)
}

func run() error {
	defer cliutil.Close()

	cfg, err := createConfig()
	if err != nil {
		return errors.WithStack(err)
	}
	command := cmd.NewEryCommand(cfg)

	return errors.WithStack(command.Execute())
}

const (
	version = "0.0.1"
)

var (
	revision, buildDate string
)

func createConfig() (*ery.Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &ery.Config{
		InReader:   os.Stdin,
		OutWriter:  os.Stdout,
		ErrWriter:  os.Stderr,
		WorkingDir: wd,

		Version:   version,
		Revision:  revision,
		BuildDate: buildDate,

		TLD:     "ery",
		Package: "tools.srvc.ery",

		Daemon: &ery.DaemonConfig{
			Name:        "ery",
			Description: "Discover services in local",
		},
	}, nil
}
