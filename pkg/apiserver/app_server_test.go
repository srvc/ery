package apiserver

import (
	"context"
	"testing"

	api_pb "github.com/srvc/ery/api"
)

func Test_AppServiceServer_ListApps(t *testing.T) {
	svr := NewAppServiceServer()

	ctx := context.Background()
	req := &api_pb.ListAppsRequest{}

	resp, err := svr.ListApps(ctx, req)

	t.SkipNow()

	if err != nil {
		t.Errorf("returned an error %v", err)
	}

	if resp == nil {
		t.Error("response should not nil")
	}
}

func Test_AppServiceServer_GetApp(t *testing.T) {
	svr := NewAppServiceServer()

	ctx := context.Background()
	req := &api_pb.GetAppRequest{}

	resp, err := svr.GetApp(ctx, req)

	t.SkipNow()

	if err != nil {
		t.Errorf("returned an error %v", err)
	}

	if resp == nil {
		t.Error("response should not nil")
	}
}

func Test_AppServiceServer_CreateApp(t *testing.T) {
	svr := NewAppServiceServer()

	ctx := context.Background()
	req := &api_pb.CreateAppRequest{}

	resp, err := svr.CreateApp(ctx, req)

	t.SkipNow()

	if err != nil {
		t.Errorf("returned an error %v", err)
	}

	if resp == nil {
		t.Error("response should not nil")
	}
}

func Test_AppServiceServer_UpdateApp(t *testing.T) {
	svr := NewAppServiceServer()

	ctx := context.Background()
	req := &api_pb.UpdateAppRequest{}

	resp, err := svr.UpdateApp(ctx, req)

	t.SkipNow()

	if err != nil {
		t.Errorf("returned an error %v", err)
	}

	if resp == nil {
		t.Error("response should not nil")
	}
}

func Test_AppServiceServer_DeleteApp(t *testing.T) {
	svr := NewAppServiceServer()

	ctx := context.Background()
	req := &api_pb.DeleteAppRequest{}

	resp, err := svr.DeleteApp(ctx, req)

	t.SkipNow()

	if err != nil {
		t.Errorf("returned an error %v", err)
	}

	if resp == nil {
		t.Error("response should not nil")
	}
}
