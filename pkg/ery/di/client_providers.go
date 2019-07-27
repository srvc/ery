package di

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/google/wire"
	"github.com/spf13/afero"

	"github.com/srvc/ery/pkg/app/api"
	"github.com/srvc/ery/pkg/app/command"
	"github.com/srvc/ery/pkg/app/dns"
	"github.com/srvc/ery/pkg/data/remote"
	"github.com/srvc/ery/pkg/domain"
	"github.com/srvc/ery/pkg/ery"
)

func ProvideCommandRunner(cfg *ery.Config, mappingRepo domain.MappingRepository) command.Runner {
	return command.NewRunner(
		afero.NewOsFs(),
		mappingRepo,
		cfg.API.Port,
		cfg.WorkingDir,
		cfg.OutWriter,
		cfg.ErrWriter,
		cfg.InReader,
	)
}

func ProvideRemoteMappingRepository(url *url.URL, httpClient *http.Client) domain.MappingRepository {
	return remote.NewMappingRepository(url, httpClient)
}

func ProvideAPIServerURL(cfg *api.Config) *url.URL {
	return &url.URL{Scheme: "http", Host: fmt.Sprintf("%s:%d", cfg.Hostname, cfg.Port)}
}

func ProvideHTTPClient(cfg *dns.Config) *http.Client {
	dialerFunc := func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{}
		return d.DialContext(ctx, "udp", fmt.Sprintf(":%d", cfg.Port))
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

	return &http.Client{Transport: transport}
}

var ClientSet = wire.NewSet(
	CommonSet,
	ClientApp{},
	ProvideCommandRunner,
	ProvideRemoteMappingRepository,
	ProvideAPIServerURL,
	ProvideHTTPClient,
)
