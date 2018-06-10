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
	defaultAddr   = ":80"
	defaultScheme = "http"
)

// NewServer creates a reverse proxy server instance.
func NewServer(mappingRepo domain.MappingRepository) app.Server {
	return &server{
		mappingRepo: mappingRepo,
		addr:        defaultAddr,
		log:         zap.L().Named("proxy"),
	}
}

type server struct {
	mappingRepo domain.MappingRepository
	server      *http.Server
	addr        string
	log         *zap.Logger
}

func (s *server) Serve(ctx context.Context) error {
	s.server = &http.Server{
		Addr:    s.addr,
		Handler: &httputil.ReverseProxy{Director: s.handle},
	}

	var err error
	errCh := make(chan error, 1)
	go func() {
		s.log.Debug("starting proxy server...", zap.String("addr", s.addr))
		errCh <- errors.WithStack(s.server.ListenAndServe())
	}()

	select {
	case err = <-errCh:
		err = errors.WithStack(err)
	case <-ctx.Done():
		s.log.Debug("shutdowning proxy server...", zap.Error(ctx.Err()))
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
