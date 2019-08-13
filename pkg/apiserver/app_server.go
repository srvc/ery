package apiserver

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/izumin5210/grapi/pkg/grapiserver"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api_pb "github.com/srvc/ery/api"
)

// AppServiceServer is a composite interface of api_pb.AppServiceServer and grapiserver.Server.
type AppServiceServer interface {
	api_pb.AppServiceServer
	grapiserver.Server
}

// NewAppServiceServer creates a new AppServiceServer instance.
func NewAppServiceServer() AppServiceServer {
	return &appServiceServerImpl{}
}

type appServiceServerImpl struct {
}

func (s *appServiceServerImpl) ListApps(ctx context.Context, req *api_pb.ListAppsRequest) (*api_pb.ListAppsResponse, error) {
	// TODO: Not yet implemented.
	return nil, status.Error(codes.Unimplemented, "TODO: You should implement it!")
}

func (s *appServiceServerImpl) GetApp(ctx context.Context, req *api_pb.GetAppRequest) (*api_pb.App, error) {
	// TODO: Not yet implemented.
	return nil, status.Error(codes.Unimplemented, "TODO: You should implement it!")
}

func (s *appServiceServerImpl) CreateApp(ctx context.Context, req *api_pb.CreateAppRequest) (*api_pb.App, error) {
	// TODO: Not yet implemented.
	return nil, status.Error(codes.Unimplemented, "TODO: You should implement it!")
}

func (s *appServiceServerImpl) UpdateApp(ctx context.Context, req *api_pb.UpdateAppRequest) (*api_pb.App, error) {
	// TODO: Not yet implemented.
	return nil, status.Error(codes.Unimplemented, "TODO: You should implement it!")
}

func (s *appServiceServerImpl) DeleteApp(ctx context.Context, req *api_pb.DeleteAppRequest) (*empty.Empty, error) {
	// TODO: Not yet implemented.
	return nil, status.Error(codes.Unimplemented, "TODO: You should implement it!")
}
