package proxy

import (
	"context"
	"fmt"
	"sync"

	"github.com/docker/docker/client"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	api_pb "github.com/srvc/ery/api"
)

type Manager interface {
	Serve(context.Context) error
	AddProxy(context.Context, *api_pb.App) error
	DeleteProxy(ctx context.Context, appID string) error
}

type managerImpl struct {
	docker *client.Client
	addCh  chan *appServer
	m      sync.Map
	log    *zap.Logger
}

func NewManager(
	docker *client.Client,
) Manager {
	return &managerImpl{
		docker: docker,
		log:    zap.L().Named("proxy").Named("manager"),
	}
}

func (m *managerImpl) Serve(ctx context.Context) error {
	m.addCh = make(chan *appServer)

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		close(m.addCh)
	}()

	for s := range m.addCh {
		s := s
		m.m.Store(s.GetID(), s)
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer m.m.Delete(s.GetID())

			ctx, s.cancel = context.WithCancel(context.Background())

			m.log.Debug("proxy servers will start", zap.String("app_id", s.GetID()))
			err := s.Serve(ctx)
			m.log.Debug("proxy servers will shutdown", zap.String("app_id", s.GetID()))
			if err != nil {
				m.log.Warn("shutdown proxy servers", zap.Error(err), zap.String("app_id", s.GetID()))
			}
		}()
	}

	return nil
}

func (m *managerImpl) AddProxy(ctx context.Context, app *api_pb.App) error {
	appServer := &appServer{app: app}

	switch app.GetType() {
	case api_pb.App_TYPE_LOCAL:
		appServer.servers = append(
			appServer.servers,
			NewDockerServer(m.docker, app),
		)

	case api_pb.App_TYPE_DOCKER:
		// no-op

	case api_pb.App_TYPE_KUBERNETES:
		return fmt.Errorf("not yet implemented type: %s", app.GetType())
	default:
		return fmt.Errorf("uknown application type: %s", app.GetType())
	}

	m.addCh <- appServer

	return nil
}

func (m *managerImpl) DeleteProxy(ctx context.Context, appID string) error {
	v, ok := m.m.Load(appID)
	if !ok {
		return fmt.Errorf("app %s was not found", appID)
	}
	v.(*appServer).Shutdown()
	return nil
}

type appServer struct {
	app     *api_pb.App
	servers []interface {
		Serve(ctx context.Context) error
	}
	cancel func()
}

func (s *appServer) GetID() string { return s.app.GetAppId() }

func (s *appServer) Serve(ctx context.Context) error {
	ctx, s.cancel = context.WithCancel(ctx)
	defer s.cancel()

	eg, ctx := errgroup.WithContext(ctx)

	for _, s := range s.servers {
		s := s
		eg.Go(func() error { return s.Serve(ctx) })
	}

	return eg.Wait()
}

func (s *appServer) Shutdown() { s.cancel() }
