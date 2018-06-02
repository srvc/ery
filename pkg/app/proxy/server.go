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
func NewServer(mapper domain.Mapper, addr string) app.Server {
	if addr != "" {
		addr = defaultAddr
	}
	return &server{
		mapper: mapper,
		addr:   addr,
	}
}

type server struct {
	mapper domain.Mapper
	server *http.Server
	addr   string
}

func (s *server) Serve() error {
	s.server = &http.Server{
		Addr:    s.addr,
		Handler: &httputil.ReverseProxy{Director: s.handle},
	}
	return s.server.ListenAndServe()
}

func (s *server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
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
