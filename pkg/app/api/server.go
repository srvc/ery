package api

import (
	"context"
	"net"
	"net/http"
	"regexp"
	"strconv"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/domain"
)

var (
	hostnameserver = "api.discoverer.local"
	addrPortPat    = regexp.MustCompile(`\d+$`)
)

type server struct {
	mapper domain.Mapper
	server *http.Server
}

func NewServer(mapper domain.Mapper) app.Server {
	return &server{
		mapper: mapper,
	}
}

func (s *server) Serve() error {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}

	addr := lis.Addr().String()
	port, err := strconv.Atoi(string(addrPortPat.FindSubmatch([]byte(addr))[0]))
	if err != nil {
		return err
	}
	s.mapper.Add(uint32(port), hostnameserver)

	s.server = &http.Server{
		Handler: s.createHandler(),
	}

	err = s.server.Serve(lis)
	return err
}

func (s *server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *server) Addr() string {
	return s.server.Addr
}

func (s *server) createHandler() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.POST("/mapping", s.handlePostMappings)

	return e
}

func (s *server) handlePing(c echo.Context) error {
	c.String(http.StatusOK, "pong")
	return nil
}

func (s *server) handlePostMappings(c echo.Context) error {
	var req struct {
		Port      uint32   `json:"port" validate:"required"`
		Hostnames []string `json:"hostnames" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	for _, hostname := range req.Hostnames {
		s.mapper.Add(uint32(req.Port), hostname)
	}

	c.NoContent(http.StatusCreated)

	return nil
}
