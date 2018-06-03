package ery

import "io"

// Config is a configuration object.
type Config struct {
	InReader             io.Reader
	OutWriter, ErrWriter io.Writer

	Version, Revision      string
	BuildDate, ReleaseType string

	API    APIConfig
	Daemon DaemonConfig
}

// APIConfig is a configuration object concerning in the API server.
type APIConfig struct {
	Hostname string
}

// DaemonConfig is a configuration object concerning in daemon.
type DaemonConfig struct {
	Name        string
	Description string
}
