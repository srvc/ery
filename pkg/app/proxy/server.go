package proxy

import (
	"context"
	"net/http"
	"net/http/httputil"

	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/domain"
	"go.uber.org/zap"
)

var (
	defaultAddr   = ":80"
	defaultScheme = "http"
)

// NewServer creates a reverse proxy server instance.
func NewServer(mapper domain.Mapper) app.Server {
	return &server{
		mapper: mapper,
		addr:   defaultAddr,
		log:    zap.L().Named("proxy"),
	}
}

type server struct {
	mapper domain.Mapper
	server *http.Server
	addr   string
	log    *zap.Logger
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
		errCh <- s.server.ListenAndServe()
	}()

	select {
	case err = <-errCh:
		// do nothing
	case <-ctx.Done():
		s.log.Debug("shutdowning proxy server...", zap.Error(ctx.Err()))
		s.server.Shutdown(context.TODO())
		err = <-errCh
	}

	return err
}

func (s *server) Addr() string {
	return s.server.Addr
}

func (s *server) handle(req *http.Request) {
	host, ok := s.mapper.Lookup(req.Host)
	if ok {
		req.URL.Host = host
	} else {
		req.URL.Host = req.Host
	}
	req.URL.Scheme = defaultScheme
}
