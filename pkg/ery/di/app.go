package di

import (
	"net"
	"sync"

	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/app/api"
	"github.com/srvc/ery/pkg/app/daemon"
	"github.com/srvc/ery/pkg/app/dns"
	"github.com/srvc/ery/pkg/app/proxy"
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
	Mapper() domain.Mapper

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

	mapper         domain.Mapper
	initMapperOnce sync.Once

	apiServer, dnsServer, proxyServer                         app.Server
	initAPIServerOnce, initDNSServerOnce, initProxyServerOnce sync.Once

	daemonFactory         daemon.Factory
	initDaemonFactoryOnce sync.Once
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

func (c *appComponentImpl) Mapper() domain.Mapper {
	c.initMapperOnce.Do(func() {
		c.mapper = domain.NewMapper(c.LocalIP())
	})
	return c.mapper
}

func (c *appComponentImpl) APIServer() app.Server {
	c.initAPIServerOnce.Do(func() {
		c.apiServer = api.NewServer(c.Mapper(), c.Config().API.Hostname)
	})
	return c.apiServer
}

func (c *appComponentImpl) DNSServer() app.Server {
	c.initDNSServerOnce.Do(func() {
		c.dnsServer = dns.NewServer(c.Mapper(), c.LocalIP())
	})
	return c.dnsServer
}

func (c *appComponentImpl) ProxyServer() app.Server {
	c.initProxyServerOnce.Do(func() {
		c.proxyServer = proxy.NewServer(c.Mapper())
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
