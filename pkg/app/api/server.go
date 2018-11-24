package api

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/domain"
	"github.com/srvc/ery/pkg/util/echoutil"
)

// Config is a configuration object concerning in the API server.
type Config struct {
	Hostname string
	Port     domain.Port
}

type server struct {
	*Config
	mappingRepo domain.MappingRepository
	server      *http.Server
	log         *zap.Logger
}

// NewServer creates an API server instance.
func NewServer(mappingRepo domain.MappingRepository, cfg *Config) app.Server {
	return &server{
		Config:      cfg,
		mappingRepo: mappingRepo,
		log:         zap.L().Named("api"),
	}
}

func (s *server) Serve(ctx context.Context) error {
	lAddr := domain.Addr{Host: s.Hostname, Port: s.Port}
	rAddr, err := s.mappingRepo.Create(ctx, lAddr, 0)
	if err != nil {
		return errors.WithStack(err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", rAddr.Port))
	if err != nil {
		return errors.WithStack(err)
	}

	s.server = &http.Server{
		Handler: s.createHandler(),
	}

	errCh := make(chan error, 1)
	go func() {
		s.log.Info("starting API server...", zap.Stringer("src_addr", &lAddr), zap.Stringer("dest_addr", &rAddr), zap.String("hostname", s.Hostname))
		errCh <- errors.WithStack(s.server.Serve(lis))
	}()

	select {
	case err = <-errCh:
		err = errors.WithStack(err)
	case <-ctx.Done():
		s.log.Info("shutdowning API server...", zap.Error(ctx.Err()))
		s.server.Shutdown(context.Background())
		err = errors.WithStack(<-errCh)
	}

	return errors.WithStack(err)
}

func (s *server) err(c echo.Context, code int, err error) {
	c.JSON(code, struct {
		Error string `json:"error"`
	}{Error: err.Error()})
}

func (s *server) createHandler() http.Handler {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(echoutil.ZapLoggerMiddleware(s.log))

	e.GET("/mappings", s.handleGetMappings)
	e.POST("/mappings", s.handlePostMappings)
	e.DELETE("/mappings/:host", s.handleDeleteMappings)

	return e
}

func (s *server) handlePostMappings(c echo.Context) error {
	var req domain.Addr

	if err := c.Bind(&req); err != nil {
		s.err(c, http.StatusBadRequest, err)
		return errors.WithStack(err)
	}

	resp, err := s.mappingRepo.Create(c.Request().Context(), req, 0)
	if err != nil {
		s.err(c, http.StatusInternalServerError, err)
		return errors.WithStack(err)
	}

	c.JSON(http.StatusCreated, resp)

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
