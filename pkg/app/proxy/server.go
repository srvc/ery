package proxy

import (
	"context"
	"net/http"
	"net/http/httputil"

	"github.com/pkg/errors"
	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/domain"
	"go.uber.org/zap"
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

func (s *server) Addr() string {
	return s.server.Addr
}

func (s *server) handle(req *http.Request) {
	host, err := s.mappingRepo.GetBySourceHost(req.Host)
	if err == nil {
		req.URL.Host = host
	} else {
		req.URL.Host = req.Host
	}
	req.URL.Scheme = defaultScheme
}
