package proxy

import (
	"context"
	"errors"
	"fmt"
	"sync"

	api_pb "github.com/srvc/ery/api"
	"github.com/srvc/ery/pkg/ery/domain"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Manager interface {
	Serve(context.Context) error
	AddProxy(context.Context, *api_pb.App) error
	DeleteProxy(context.Context, *api_pb.App) error
}

type managerImpl struct {
	addCh chan *appServer
	m     sync.Map
	log   *zap.Logger
}

func NewManager() Manager {
	return &managerImpl{
		log: zap.L().Named("proxy").Named("manager"),
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

			m.log.Debug("proxy servers will start", zap.String("app_id", s.GetID()))
			err := s.Serve(ctx)
			if err != nil {
				m.log.Warn("shutdown proxy servers", zap.Error(err), zap.String("app_id", s.GetID()))
			}
		}()
	}

	return nil
}

func (m *managerImpl) AddProxy(ctx context.Context, app *api_pb.App) error {
	appServer := &appServer{app: app}

	for _, port := range app.GetPorts() {
		switch port.GetNetwork() {
		case api_pb.App_Port_TCP:
			appServer.servers = append(
				appServer.servers,
				NewTCPServer(
					&domain.Addr{IP: app.GetIp(), Port: domain.Port(port.GetExposedPort())},
					&domain.Addr{IP: "localhost", Port: domain.Port(port.GetInternalPort())},
				),
			)
		case api_pb.App_Port_UDP:
			return errors.New("not yet implemented")
		default:
			return fmt.Errorf("uknown network type: %s", port.GetNetwork())
		}
	}

	m.addCh <- appServer

	return nil
}

func (m *managerImpl) DeleteProxy(ctx context.Context, app *api_pb.App) error {
	v, ok := m.m.Load(app.GetAppId())
	if !ok {
		return fmt.Errorf("app %s was not found", app.GetAppId())
	}
	m.m.Delete(app.GetAppId())
	s, ok := v.(*appServer)
	if !ok {
		return errors.New("unknown value was found")
	}
	s.Shutdown()
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
