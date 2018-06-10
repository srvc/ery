package di

import (
	"sync"

	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/app/api"
	"github.com/srvc/ery/pkg/app/container"
	"github.com/srvc/ery/pkg/app/daemon"
	"github.com/srvc/ery/pkg/app/dns"
	"github.com/srvc/ery/pkg/app/proxy"
	"github.com/srvc/ery/pkg/data/local"
	"github.com/srvc/ery/pkg/data/remote"
	"github.com/srvc/ery/pkg/domain"
	"github.com/srvc/ery/pkg/ery"
)

// AppComponent is an interface to provide accessors for dependencies.
type AppComponent interface {
	// cli context
	Config() *ery.Config

	// domain
	LocalMappingRepository() domain.MappingRepository
	RemoteMappingRepository() domain.MappingRepository

	// app
	APIServer() app.Server
	DNSServer() app.Server
	ProxyServer() app.Server
	ContainerWatcher() app.Watcher
	DaemonFactory() daemon.Factory
}

// NewAppComponent creates a new AppComponent instance.
func NewAppComponent(cfg *ery.Config) AppComponent {
	return &appComponentImpl{
		config: cfg,
	}
}

type appComponentImpl struct {
	config *ery.Config

	apiServer, dnsServer, proxyServer                         app.Server
	initAPIServerOnce, initDNSServerOnce, initProxyServerOnce sync.Once

	containerWatcher         app.Watcher
	initContainerWatcherOnce sync.Once

	daemonFactory         daemon.Factory
	initDaemonFactoryOnce sync.Once

	localMappingRepo                 domain.MappingRepository
	initLocalMappingRepoOnce         sync.Once
	remoteMappingRepo                domain.MappingRepository
	initRemoteMappingRepoOnce        sync.Once
	localDockerContainerRepo         domain.ContainerRepository
	initLocalDockerContainerRepoOnce sync.Once
}

func (c *appComponentImpl) Config() *ery.Config {
	return c.config
}

func (c *appComponentImpl) APIServer() app.Server {
	c.initAPIServerOnce.Do(func() {
		c.apiServer = api.NewServer(c.LocalMappingRepository(), c.Config().API.Hostname)
	})
	return c.apiServer
}

func (c *appComponentImpl) DNSServer() app.Server {
	c.initDNSServerOnce.Do(func() {
		c.dnsServer = dns.NewServer(c.LocalMappingRepository())
	})
	return c.dnsServer
}

func (c *appComponentImpl) ProxyServer() app.Server {
	c.initProxyServerOnce.Do(func() {
		c.proxyServer = proxy.NewServer(c.LocalMappingRepository())
	})
	return c.proxyServer
}

func (c *appComponentImpl) ContainerWatcher() app.Watcher {
	c.initContainerWatcherOnce.Do(func() {
		c.containerWatcher = container.NewWatcher(
			c.LocalMappingRepository(),
			[]domain.ContainerRepository{
				c.LocalDockerContainerRepository(),
			},
			c.Config().TLD,
			c.Config().Package+".hostname",
		)
	})
	return c.containerWatcher
}

func (c *appComponentImpl) DaemonFactory() daemon.Factory {
	c.initDaemonFactoryOnce.Do(func() {
		cfg := c.Config().Daemon
		c.daemonFactory = daemon.NewFactory(cfg.Name, cfg.Description)
	})
	return c.daemonFactory
}

func (c *appComponentImpl) LocalMappingRepository() domain.MappingRepository {
	c.initLocalMappingRepoOnce.Do(func() {
		c.localMappingRepo = local.NewMappingRepository()
	})
	return c.localMappingRepo
}

func (c *appComponentImpl) RemoteMappingRepository() domain.MappingRepository {
	c.initRemoteMappingRepoOnce.Do(func() {
		c.remoteMappingRepo = remote.NewMappingRepository("http://" + c.Config().API.Hostname)
	})
	return c.remoteMappingRepo
}

func (c *appComponentImpl) LocalDockerContainerRepository() domain.ContainerRepository {
	c.initLocalDockerContainerRepoOnce.Do(func() {
		c.localDockerContainerRepo = local.NewDockerContainerRepository()
	})
	return c.localDockerContainerRepo
}
