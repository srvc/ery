package api

import (
	"context"
	"net"
	"net/http"
	"regexp"
	"strconv"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/domain"
)

var addrPortPat = regexp.MustCompile(`\d+$`)

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

	addr := lis.Addr().String()
	port, err := strconv.Atoi(string(addrPortPat.FindSubmatch([]byte(addr))[0]))
	if err != nil {
		return errors.Wrapf(err, "failed to extract port number from address: %s", addr)
	}
	err = s.mappingRepo.Create(uint32(port), s.hostname)
	if err != nil {
		return errors.WithStack(err)
	}

	s.server = &http.Server{
		Handler: s.createHandler(),
	}

	errCh := make(chan error, 1)
	go func() {
		s.log.Debug("starting DNS server...", zap.String("addr", addr), zap.String("hostname", s.hostname))
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
	e.DELETE("/mappings/:port", s.handleDeleteMappings)

	return e
}

func (s *server) handlePostMappings(c echo.Context) error {
	var req struct {
		Port      uint32   `json:"port" validate:"required"`
		Hostnames []string `json:"hostnames" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		s.err(c, http.StatusBadRequest, err)
		return errors.WithStack(err)
	}

	err := s.mappingRepo.Create(uint32(req.Port), req.Hostnames...)
	if err != nil {
		s.err(c, http.StatusInternalServerError, err)
		return errors.WithStack(err)
	}

	c.NoContent(http.StatusCreated)

	return nil
}

func (s *server) handleGetMappings(c echo.Context) error {
	type Mapping struct {
		IP        string   `json:"ip"`
		Port      uint32   `json:"port"`
		Hostnames []string `json:"hostnames"`
	}
	type Response struct {
		Mappings []Mapping `json:"mappings"`
	}

	mappings, err := s.mappingRepo.List()
	if err != nil {
		s.err(c, http.StatusInternalServerError, err)
		return errors.WithStack(err)
	}

	resp := &Response{
		Mappings: make([]Mapping, 0, len(mappings)),
	}

	for _, m := range mappings {
		resp.Mappings = append(resp.Mappings, Mapping{
			IP:        m.IP.String(),
			Port:      m.Port,
			Hostnames: m.Hostnames,
		})
	}

	c.JSON(http.StatusOK, resp)

	return nil
}

func (s *server) handleDeleteMappings(c echo.Context) error {
	port, err := strconv.ParseUint(c.Param("port"), 10, 32)
	if err != nil {
		s.err(c, http.StatusBadRequest, err)
		return errors.WithStack(err)
	}

	err = s.mappingRepo.Delete(uint32(port))
	if err != nil {
		s.err(c, http.StatusInternalServerError, err)
		return errors.WithStack(err)
	}

	c.NoContent(http.StatusNoContent)
	return nil
}
