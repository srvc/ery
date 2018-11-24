package ery

import (
	"io"

	"github.com/srvc/ery/pkg/domain"
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

	API    *APIConfig
	DNS    *DNSConfig
	Proxy  *ProxyConfig
	Daemon *DaemonConfig
}

// APIConfig is a configuration object concerning in the API server.
type APIConfig struct {
	Hostname string
}

// DNSConfig is a configuration object concerning in the DNS server.
type DNSConfig struct {
	Port domain.Port
}

// ProxyConfig is a configuration object concerning in the Proxy server.
type ProxyConfig struct {
	DefaultPort domain.Port
}

// DaemonConfig is a configuration object concerning in daemon.
type DaemonConfig struct {
	Name        string
	Description string
}
