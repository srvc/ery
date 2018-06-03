package proxy

import (
	"context"
	"net/http"
	"net/http/httputil"

	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/domain"
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
	}
}

type server struct {
	mapper domain.Mapper
	server *http.Server
	addr   string
}

func (s *server) Serve(ctx context.Context) error {
	s.server = &http.Server{
		Addr:    s.addr,
		Handler: &httputil.ReverseProxy{Director: s.handle},
	}

	var err error
	errCh := make(chan error, 1)
	go func() { errCh <- s.server.ListenAndServe() }()

	select {
	case err = <-errCh:
		// do nothing
	case <-ctx.Done():
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
