package ery

import "io"

// Config is a configuration object.
type Config struct {
	InReader             io.Reader
	OutWriter, ErrWriter io.Writer

	Version, Revision      string
	BuildDate, ReleaseType string

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
	Port uint16
}

// ProxyConfig is a configuration object concerning in the Proxy server.
type ProxyConfig struct {
	DefaultPort uint16
}

// DaemonConfig is a configuration object concerning in daemon.
type DaemonConfig struct {
	Name        string
	Description string
}
