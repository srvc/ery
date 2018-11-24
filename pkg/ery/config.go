package ery

import (
	"io"

	"github.com/srvc/ery/pkg/app/api"
	"github.com/srvc/ery/pkg/app/dns"
)

// Config is a configuration object.
type Config struct {
	InReader             io.Reader
	OutWriter, ErrWriter io.Writer
	WorkingDir           string

	Version             string
	Revision, BuildDate string

	TLD     string
	Package string

	API    api.Config
	DNS    dns.Config
	Daemon *DaemonConfig
}

// DaemonConfig is a configuration object concerning in daemon.
type DaemonConfig struct {
	Name        string
	Description string
}
