package mem

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"

	api_pb "github.com/srvc/ery/api"
	"github.com/srvc/ery/pkg/ery/domain"
	"go.uber.org/zap"
)

type AppRepository struct {
	sync.Mutex
	m          sync.Map
	byHostname sync.Map
	ipPool     domain.IPPool
	portPool   domain.PortPool
	log        *zap.Logger
}

var _ domain.AppRepository = (*AppRepository)(nil)

func NewAppRepository(
	ipPool domain.IPPool,
	portPool domain.PortPool,
) *AppRepository {
	return &AppRepository{
		ipPool:   ipPool,
		portPool: portPool,
		log:      zap.L().Named("mem"),
	}
}

func (r *AppRepository) List(context.Context) ([]*api_pb.App, error) {
	r.Lock()
	defer r.Unlock()

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
	r.Lock()
	defer r.Unlock()

	v, ok := r.byHostname.Load(hostname)
	if ok {
		v, ok := r.m.Load(v.(string))
		if ok {
			if app, ok := v.(*api_pb.App); ok {
				return app, nil
			}
		}
	}
	return nil, fmt.Errorf("%s is not found", hostname)
}

func (r *AppRepository) Create(ctx context.Context, app *api_pb.App) error {
	r.Lock()
	defer r.Unlock()

	if app.AppId == "" {
		k := make([]byte, 16)
		if _, err := rand.Read(k); err != nil {
			return err
		}
		app.AppId = fmt.Sprintf("%x", k)
	}
	if app.Ip == "" {
		ip, err := r.ipPool.Get(ctx)
		if err != nil {
			return err
		}
		app.Ip = ip.String()
	}
	for _, port := range app.GetPorts() {
		if port.GetInternalPort() == 0 {
			p, err := r.portPool.Get(ctx)
			if err != nil {
				return err
			}
			port.InternalPort = uint32(p)
		}
	}
	r.m.Store(app.GetAppId(), app)
	r.byHostname.Store(app.GetHostname(), app.GetAppId())

	r.log.Debug("registered a new app", zap.Any("app", app))

	return nil
}

func (r *AppRepository) Delete(_ context.Context, id string) error {
	r.Lock()
	defer r.Unlock()

	v, ok := r.m.Load(id)
	if !ok {
		return fmt.Errorf("%s is not found", id)
	}
	r.m.Delete(id)
	r.byHostname.Delete(v.(*api_pb.App).GetHostname())
	return nil
}
