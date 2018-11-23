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
		mappingRepo: mappingRepo,
		factory:     factory,
		log:         zap.L().Named("proxy"),
	}
}

type serverManager struct {
	mappingRepo domain.MappingRepository
	factory     ServerFactory
	cancellers  cancellers
	log         *zap.Logger
}

func (m *serverManager) Serve(ctx context.Context) error {
	m.log.Debug("start listening mapping events")
	evCh, errCh := m.mappingRepo.ListenEvent(ctx)
	wg := new(sync.WaitGroup)

	defer func() {
		m.cancellers = cancellers{}
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
			m.log.Debug("stop listening mapping events")
			return errors.WithStack(ctx.Err())
		}
	}

	wg.Wait()

	return nil
}

func (m *serverManager) handleCreated(ctx context.Context, wg *sync.WaitGroup, ev domain.MappingEvent) {
	for cport := range ev.PortMap {
		addr := domain.Addr{Host: ev.ProxyHost, Port: cport}
		if c, ok := m.cancellers.Get(addr); ok {
			c.Add(1)
		} else {
			wg.Add(1)
			c, cctx := cancellerWithContext(ctx)
			c.Add(1)
			m.cancellers.Set(addr, c)
			go func(ctx context.Context, addr domain.Addr) {
				defer wg.Done()
				m.factory.CreateServer(addr).Serve(ctx)
			}(cctx, addr)
		}
	}
}

func (m *serverManager) handleDestroyed(ctx context.Context, wg *sync.WaitGroup, ev domain.MappingEvent) {
	for cport := range ev.PortMap {
		addr := domain.Addr{Host: ev.ProxyHost, Port: cport}
		if c, ok := m.cancellers.Get(addr); ok {
			c.Done()
			if c.count == 0 {
				m.cancellers.Delete(addr)
			}
		}
	}
}

type cancellers struct {
	byAddr sync.Map
}

func (cs *cancellers) Get(addr domain.Addr) (c *canceller, ok bool) {
	var v interface{}
	if v, ok = cs.byAddr.Load(addr.String()); ok {
		c, ok = v.(*canceller)
	}
	return
}

func (cs *cancellers) Set(addr domain.Addr, c *canceller) {
	cs.byAddr.Store(addr.String(), c)
}

func (cs *cancellers) Delete(addr domain.Addr) {
	cs.byAddr.Delete(addr.String())
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
