package di

import (
	"github.com/srvc/ery/pkg/app/api"
	"github.com/srvc/ery/pkg/app/command"
	"github.com/srvc/ery/pkg/app/container"
	"github.com/srvc/ery/pkg/app/daemon"
	"github.com/srvc/ery/pkg/app/dns"
	"github.com/srvc/ery/pkg/app/proxy"
	"github.com/srvc/ery/pkg/domain"
)

type ServerApp struct {
	APIServer        api.Server
	DNSServer        dns.Server
	ProxyServer      proxy.Manager
	ContainerWatcher container.Watcher
}

type ClientApp struct {
	CommandRunner command.Runner
	MappingRepo   domain.MappingRepository
}

type DaemonApp struct {
	DaemonFactory daemon.Factory
}
