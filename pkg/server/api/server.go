package api

import (
	"context"
	"fmt"

	"github.com/izumin5210/grapi/pkg/grapiserver"
	"go.uber.org/zap"

	api_pb "github.com/srvc/ery/api"
	"github.com/srvc/ery/pkg/ery/domain"
	"github.com/srvc/ery/pkg/server/api/internal"
)

var (
	defaultNetwork = "udp"
	defaultPort    = 53
)

type Server struct {
	appRepo domain.AppRepository

	server *grapiserver.Engine
	log    *zap.Logger
}

func NewServer(
	appRepo domain.AppRepository,
) *Server {
	return &Server{
		appRepo: appRepo,
		log:     zap.L().Named("api"),
	}
}

func (s *Server) Serve(ctx context.Context) error {
	app := &api_pb.App{
		Name:     "srvc.tools/ery/api",
		Hostname: "api.ery.local",
		Type:     api_pb.App_TYPE_LOCAL,
		Ip:       "127.0.0.1",
	}
	err := s.appRepo.Create(ctx, app)
	if err != nil {
		return nil
	}
	s.server = grapiserver.New(
		grapiserver.WithGrpcAddr(defaultNetwork, fmt.Sprintf("%s:%d", app.GetIp(), defaultPort)),
		grapiserver.WithServers(
			internal.NewAppServiceServer(s.appRepo),
		),
	)
	return s.server.ServeContext(ctx)
}
