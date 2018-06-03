package ery

import "io"

// Config is a configuration object.
type Config struct {
	InReader             io.Reader
	OutWriter, ErrWriter io.Writer

	Version, Revision      string
	BuildDate, ReleaseType string

	API APIConfig
}

// APIConfig is a configuration object concerning in the API server.
type APIConfig struct {
	Hostname string
}
