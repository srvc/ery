package di

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

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

	httpClient     *http.Client
	initHTTPClient sync.Once
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
		c.dnsServer = dns.NewServer(c.LocalMappingRepository(), domain.Port(c.Config().DNS.Port))
	})
	return c.dnsServer
}

func (c *appComponentImpl) ProxyServer() app.Server {
	c.initProxyServerOnce.Do(func() {
		mappingRepo := c.LocalMappingRepository()
		c.proxyServer = proxy.NewManager(mappingRepo, proxy.NewFactory(mappingRepo))
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
		c.localMappingRepo = local.NewMappingRepository(domain.Port(c.Config().Proxy.DefaultPort))
	})
	return c.localMappingRepo
}

func (c *appComponentImpl) RemoteMappingRepository() domain.MappingRepository {
	c.initRemoteMappingRepoOnce.Do(func() {
		url := &url.URL{Scheme: "http", Host: fmt.Sprintf("%s:%d", c.Config().API.Hostname, c.Config().Proxy.DefaultPort)}
		c.remoteMappingRepo = remote.NewMappingRepository(url, c.HTTPClient())
	})
	return c.remoteMappingRepo
}

func (c *appComponentImpl) LocalDockerContainerRepository() domain.ContainerRepository {
	c.initLocalDockerContainerRepoOnce.Do(func() {
		c.localDockerContainerRepo = local.NewDockerContainerRepository()
	})
	return c.localDockerContainerRepo
}

func (c *appComponentImpl) HTTPClient() *http.Client {
	c.initHTTPClient.Do(func() {
		dialerFunc := func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", fmt.Sprintf(":%d", c.Config().DNS.Port))
		}

		resolver := &net.Resolver{PreferGo: true, Dial: dialerFunc}
		dialer := net.Dialer{Resolver: resolver}

		transport := &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			Dial:                  dialer.Dial,
			DialContext:           dialer.DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		c.httpClient = &http.Client{Transport: transport}
	})
	return c.httpClient
}
