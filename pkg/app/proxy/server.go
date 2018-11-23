package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/domain"
	"github.com/srvc/ery/pkg/util/netutil"
)

var (
	defaultScheme = "http"
)

func newServerWithPort(mappingRepo domain.MappingRepository, addr domain.Addr) app.Server {
	return &server{
		mappingRepo: mappingRepo,
		addr:        addr,
		log:         zap.L().Named("proxy"),
	}
}

type server struct {
	mappingRepo domain.MappingRepository
	server      *http.Server
	addr        domain.Addr
	log         *zap.Logger
}

func (s *server) Serve(ctx context.Context) error {
	s.server = &http.Server{
		Addr:    s.addr.String(),
		Handler: &httputil.ReverseProxy{Director: s.handle},
	}

	var err error
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("starting proxy server...", zap.Stringer("addr", &s.addr))
		errCh <- errors.WithStack(s.server.ListenAndServe())
	}()

	select {
	case err = <-errCh:
		err = errors.WithStack(err)
		s.log.Info("shutdowning proxy server...", zap.Error(err), zap.Stringer("addr", &s.addr))
	case <-ctx.Done():
		s.log.Info("shutdowning proxy server...", zap.Error(ctx.Err()), zap.Stringer("addr", &s.addr))
		s.server.Shutdown(context.TODO())
		err = errors.WithStack(<-errCh)
	}

	return errors.WithStack(err)
}

func (s *server) handle(req *http.Request) {
	req.URL.Scheme = defaultScheme

	hostAndPort := strings.SplitN(req.Host, ":", 2)
	addr := domain.HTTPAddr(hostAndPort[0])
	if len(hostAndPort) == 2 {
		var err error
		addr.Port, err = domain.PortFromString(hostAndPort[1])
		if err != nil {
			return
		}
	}

	outAddr, err := s.mappingRepo.MapAddr(req.Context(), addr)
	if err != nil {
		return
	}
	if outAddr.Host == "" {
		outAddr.Host = s.localhost()
	}
	req.URL.Host = fmt.Sprintf("%s:%d", outAddr.Host, outAddr.Port)
}

func (s *server) localhost() string {
	return netutil.LocalIP().String() // FIXME
}
