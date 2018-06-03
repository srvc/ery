package di

import (
	"io"
	"net"
	"sync"

	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/app/api"
	"github.com/srvc/ery/pkg/app/dns"
	"github.com/srvc/ery/pkg/app/proxy"
	"github.com/srvc/ery/pkg/domain"
	"github.com/srvc/ery/pkg/ery"
	"github.com/srvc/ery/pkg/util/netutil"
)

// AppComponent is an interface to provide accessors for dependencies.
type AppComponent interface {
	// cli context
	InReader() io.Reader
	OutWriter() io.Writer
	ErrWriter() io.Writer
	LocalIP() net.IP

	// domain
	Mapper() domain.Mapper

	// app
	APIServer() app.Server
	DNSServer() app.Server
	ProxyServer() app.Server
}

// NewAppComponent creates a new AppComponent instance.
func NewAppComponent(cfg *ery.Config) AppComponent {
	return &appComponentImpl{
		Config: cfg,
	}
}

type appComponentImpl struct {
	*ery.Config

	localIP         net.IP
	initLocalIPOnce sync.Once

	mapper         domain.Mapper
	initMapperOnce sync.Once

	apiServer, dnsServer, proxyServer                         app.Server
	initAPIServerOnce, initDNSServerOnce, initProxyServerOnce sync.Once
}

func (c *appComponentImpl) InReader() io.Reader {
	return c.Config.InReader
}

func (c *appComponentImpl) OutWriter() io.Writer {
	return c.Config.OutWriter
}

func (c *appComponentImpl) ErrWriter() io.Writer {
	return c.Config.ErrWriter
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
		c.apiServer = api.NewServer(c.Mapper())
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
