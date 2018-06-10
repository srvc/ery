package di

import (
	"net"
	"sync"

	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/app/api"
	"github.com/srvc/ery/pkg/app/daemon"
	"github.com/srvc/ery/pkg/app/dns"
	"github.com/srvc/ery/pkg/app/proxy"
	"github.com/srvc/ery/pkg/data/local"
	"github.com/srvc/ery/pkg/data/remote"
	"github.com/srvc/ery/pkg/domain"
	"github.com/srvc/ery/pkg/ery"
	"github.com/srvc/ery/pkg/util/netutil"
)

// AppComponent is an interface to provide accessors for dependencies.
type AppComponent interface {
	// cli context
	Config() *ery.Config
	LocalIP() net.IP

	// domain
	LocalMappingRepository() domain.MappingRepository
	RemoteMappingRepository() domain.MappingRepository

	// app
	APIServer() app.Server
	DNSServer() app.Server
	ProxyServer() app.Server
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

	localIP         net.IP
	initLocalIPOnce sync.Once

	apiServer, dnsServer, proxyServer                         app.Server
	initAPIServerOnce, initDNSServerOnce, initProxyServerOnce sync.Once

	daemonFactory         daemon.Factory
	initDaemonFactoryOnce sync.Once

	localMappingRepo          domain.MappingRepository
	initLocalMappingRepoOnce  sync.Once
	remoteMappingRepo         domain.MappingRepository
	initRemoteMappingRepoOnce sync.Once
}

func (c *appComponentImpl) Config() *ery.Config {
	return c.config
}

func (c *appComponentImpl) LocalIP() net.IP {
	c.initLocalIPOnce.Do(func() {
		c.localIP = netutil.LocalhostIP()
	})
	return c.localIP
}

func (c *appComponentImpl) APIServer() app.Server {
	c.initAPIServerOnce.Do(func() {
		c.apiServer = api.NewServer(c.LocalMappingRepository(), c.Config().API.Hostname)
	})
	return c.apiServer
}

func (c *appComponentImpl) DNSServer() app.Server {
	c.initDNSServerOnce.Do(func() {
		c.dnsServer = dns.NewServer(c.LocalMappingRepository(), c.LocalIP())
	})
	return c.dnsServer
}

func (c *appComponentImpl) ProxyServer() app.Server {
	c.initProxyServerOnce.Do(func() {
		c.proxyServer = proxy.NewServer(c.LocalMappingRepository())
	})
	return c.proxyServer
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
