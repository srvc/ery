package proxy

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/domain"
)

// NewManager creates a new server instance for managings proxy servers.
func NewManager(
	mappingRepo domain.MappingRepository,
	factory ServerFactory,
) app.Server {
	return &serverManager{
		mappingRepo:     mappingRepo,
		factory:         factory,
		cancellerByPort: new(sync.Map),
		log:             zap.L().Named("proxy"),
	}
}

type serverManager struct {
	mappingRepo     domain.MappingRepository
	factory         ServerFactory
	cancellerByPort *sync.Map
	log             *zap.Logger
}

func (m *serverManager) Serve(ctx context.Context) error {
	m.log.Debug("start listening mapping events")
	evCh, errCh := m.mappingRepo.ListenEvent(ctx)
	wg := new(sync.WaitGroup)

	defer func() {
		m.cancellerByPort = new(sync.Map)
	}()

	for {
		select {
		case ev := <-evCh:
			m.log.Debug("receive a mapping event", zap.Any("event", ev))
			switch ev.Type {
			case domain.MappingEventCreated:
				m.handleCreated(ctx, wg, ev)
			case domain.MappingEventDestroyed:
				m.handleDestroyed(ctx, wg, ev)
			}
		case err := <-errCh:
			return errors.WithStack(err)
		case <-ctx.Done():
			return errors.WithStack(ctx.Err())
		}
	}

	wg.Wait()

	return nil
}

func (m *serverManager) handleCreated(ctx context.Context, wg *sync.WaitGroup, ev domain.MappingEvent) {
	for cport := range ev.PortAddrMap {
		if v, ok := m.cancellerByPort.Load(cport); ok {
			if canceller, ok := v.(*canceller); ok {
				canceller.Add(1)
			}
		} else {
			wg.Add(1)
			canceller, cctx := cancellerWithContext(ctx)
			canceller.Add(1)
			m.cancellerByPort.Store(cport, canceller)
			go func(ctx context.Context, port domain.Port) {
				defer wg.Done()
				m.factory.CreateServer(port).Serve(ctx)
			}(cctx, cport)
		}
	}
}

func (m *serverManager) handleDestroyed(ctx context.Context, wg *sync.WaitGroup, ev domain.MappingEvent) {
	for cport := range ev.PortAddrMap {
		if v, ok := m.cancellerByPort.Load(cport); ok {
			if canceller, ok := v.(*canceller); ok {
				canceller.Done()
				if canceller.count == 0 {
					m.cancellerByPort.Delete(cport)
				}
			}
		}
	}
}

func cancellerWithContext(ctx context.Context) (*canceller, context.Context) {
	cctx, cancel := context.WithCancel(ctx)
	canceller := &canceller{
		cancel: cancel,
	}
	return canceller, cctx
}

type canceller struct {
	count  uint32
	cancel func()
}

func (c *canceller) Add(d uint32) {
	atomic.AddUint32(&c.count, d)
}

func (c *canceller) Done() {
	c.Add(^uint32(0))
	if c.count == 0 {
		c.cancel()
	}
}
