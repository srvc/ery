package api

import (
	"context"
	"net"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/domain"
)

type server struct {
	mappingRepo domain.MappingRepository
	server      *http.Server
	hostname    string
	log         *zap.Logger
}

// NewServer creates an API server instance.
func NewServer(mappingRepo domain.MappingRepository, hostname string) app.Server {
	return &server{
		mappingRepo: mappingRepo,
		hostname:    hostname,
		log:         zap.L().Named("api"),
	}
}

func (s *server) Serve(ctx context.Context) error {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		return errors.WithStack(err)
	}

	addr := lis.Addr()
	err = s.mappingRepo.Create(ctx, &domain.Mapping{
		Host: s.hostname,
		PortAddrMap: domain.PortAddrMap{
			80: domain.LocalAddr(domain.Port(addr.(*net.TCPAddr).Port)),
		},
	})
	if err != nil {
		return errors.WithStack(err)
	}

	s.server = &http.Server{
		Handler: s.createHandler(),
	}

	errCh := make(chan error, 1)
	go func() {
		s.log.Debug("starting DNS server...", zap.Stringer("addr", addr), zap.String("hostname", s.hostname))
		errCh <- errors.WithStack(s.server.Serve(lis))
	}()

	select {
	case err = <-errCh:
		err = errors.WithStack(err)
	case <-ctx.Done():
		s.log.Debug("shutdowning API server...", zap.Error(ctx.Err()))
		s.server.Shutdown(context.Background())
		err = errors.WithStack(<-errCh)
	}

	return errors.WithStack(err)
}

func (s *server) Addr() string {
	return s.server.Addr
}

func (s *server) err(c echo.Context, code int, err error) {
	c.JSON(code, struct {
		Error string `json:"error"`
	}{Error: err.Error()})
}

func (s *server) createHandler() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/mappings", s.handleGetMappings)
	e.POST("/mappings", s.handlePostMappings)
	e.DELETE("/mappings/:host", s.handleDeleteMappings)

	return e
}

func (s *server) handlePostMappings(c echo.Context) error {
	var req *domain.Mapping

	if err := c.Bind(&req); err != nil {
		s.err(c, http.StatusBadRequest, err)
		return errors.WithStack(err)
	}

	err := s.mappingRepo.Create(c.Request().Context(), req)
	if err != nil {
		s.err(c, http.StatusInternalServerError, err)
		return errors.WithStack(err)
	}

	c.NoContent(http.StatusCreated)

	return nil
}

func (s *server) handleGetMappings(c echo.Context) error {
	resp := struct {
		Mappings []*domain.Mapping `json:"mappings"`
	}{}
	var err error

	resp.Mappings, err = s.mappingRepo.List(c.Request().Context())
	if err != nil {
		s.err(c, http.StatusInternalServerError, err)
		return errors.WithStack(err)
	}

	c.JSON(http.StatusOK, resp)

	return nil
}

func (s *server) handleDeleteMappings(c echo.Context) error {
	err := s.mappingRepo.DeleteByHost(c.Request().Context(), c.Param("host"))
	if err != nil {
		s.err(c, http.StatusInternalServerError, err)
		return errors.WithStack(err)
	}

	c.NoContent(http.StatusNoContent)
	return nil
}
