package mem

import (
	"context"
	"fmt"
	"sync"

	api_pb "github.com/srvc/ery/api"
	"github.com/srvc/ery/pkg/ery/domain"
)

type AppRepository struct {
	m      sync.Map
	ipPool domain.IPPool
}

var _ domain.AppRepository = (*AppRepository)(nil)

func NewAppRepository(ipPool domain.IPPool) *AppRepository {
	return &AppRepository{
		ipPool: ipPool,
	}
}

func (r *AppRepository) List(context.Context) ([]*api_pb.App, error) {
	apps := []*api_pb.App{}
	r.m.Range(func(_, v interface{}) bool {
		if app, ok := v.(*api_pb.App); ok {
			apps = append(apps, app)
		}
		return true
	})
	return apps, nil
}

func (r *AppRepository) GetByHostname(_ context.Context, hostname string) (*api_pb.App, error) {
	v, ok := r.m.Load(hostname)
	if ok {
		if app, ok := v.(*api_pb.App); ok {
			return app, nil
		}
	}
	return nil, fmt.Errorf("%s is not found", hostname)
}

func (r *AppRepository) Create(ctx context.Context, app *api_pb.App) error {
	if app.Ip == "" {
		ip, err := r.ipPool.Get(ctx)
		if err != nil {
			return err
		}
		app.Ip = ip.String()
	}
	r.m.Store(app.GetHostname(), app)
	return nil
}
