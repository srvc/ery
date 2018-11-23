package local

import (
	"context"
	"net"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"

	"github.com/srvc/ery/pkg/domain"
	"github.com/srvc/ery/pkg/util/netutil"
)

// NewMappingRepository creates a new MappingRepository instance that can access local data.
func NewMappingRepository() domain.MappingRepository {
	return &mappingRepositoryImpl{
		eventEmitters: new(sync.Map),
	}
}

type mappingRepositoryImpl struct {
	mappingByHost     mappingByHost
	hosts             hosts
	eventEmitters     *sync.Map
	eventEmitterIDSeq uint64
}

func (r *mappingRepositoryImpl) List(ctx context.Context) ([]*domain.Mapping, error) {
	return r.mappingByHost.List(), nil
}

func (r *mappingRepositoryImpl) LookupIP(ctx context.Context, host string) (net.IP, bool) {
	return r.hosts.LookupIP(host)
}

func (r *mappingRepositoryImpl) MapAddr(ctx context.Context, addr domain.Addr) (domain.Addr, error) {
	if m, ok := r.mappingByHost.Get(addr.Host); ok {
		if got := m.Map(addr.Port); got.IsValid() {
			return got, nil
		}
	}
	return domain.Addr{}, errors.Errorf("%v is not found", addr)
}

func (r *mappingRepositoryImpl) Create(ctx context.Context, lAddr domain.Addr, rPort domain.Port) (domain.Addr, error) {
	m, ok := r.mappingByHost.Get(lAddr.Host)
	release := func() {}
	if ok {
		if _, ok = m.PortMap[lAddr.Port]; ok {
			return domain.Addr{}, errors.Errorf("%v has already been registered", lAddr.Host)
		}
	} else {
		m = &domain.Mapping{VirtualHost: lAddr.Host, PortMap: domain.PortMap{}}
		m.ProxyHost = r.hosts.GetIP(m.VirtualHost).String()
		release = func() { r.hosts.Delete(m.VirtualHost) }
	}

	if rPort == 0 {
		var err error
		rPort, err = netutil.GetFreePort(m.ProxyHost)
		if err != nil {
			release()
			return domain.Addr{}, errors.WithStack(err)
		}
	}
	m.PortMap[lAddr.Port] = rPort

	r.mappingByHost.Set(m.VirtualHost, m)

	r.emitEvent(domain.MappingEvent{
		Type:    domain.MappingEventCreated,
		Mapping: *m,
	})

	return domain.Addr{Host: m.ProxyHost, Port: rPort}, nil
}

func (r *mappingRepositoryImpl) DeleteByHost(ctx context.Context, host string) error {
	if m, ok := r.mappingByHost.Get(host); ok {
		r.mappingByHost.Delete(host)
		r.emitEvent(domain.MappingEvent{
			Type:    domain.MappingEventDestroyed,
			Mapping: *m,
		})
	}
	return nil
}

func (r *mappingRepositoryImpl) ListenEvent(ctx context.Context) (<-chan domain.MappingEvent, <-chan error) {
	evCh := make(chan domain.MappingEvent)
	errCh := make(chan error, 1)

	id := atomic.AddUint64(&r.eventEmitterIDSeq, 1)
	r.eventEmitters.Store(id, &mappingEventEmitter{
		id:    id,
		evCh:  evCh,
		errCh: errCh,
		ctx:   ctx,
	})

	return evCh, errCh
}

func (r *mappingRepositoryImpl) emitEvent(ev domain.MappingEvent) {
	var disposableIDs []uint64
	r.eventEmitters.Range(func(_, v interface{}) bool {
		if emitter, ok := v.(*mappingEventEmitter); !ok || !emitter.Emit(ev) {
			disposableIDs = append(disposableIDs, emitter.id)
		}
		return true
	})

	for _, id := range disposableIDs {
		r.eventEmitters.Delete(id)
	}
}

type mappingByHost struct {
	m sync.Map
}

func (m *mappingByHost) Get(host string) (out *domain.Mapping, ok bool) {
	var v interface{}
	if v, ok = m.m.Load(host); ok {
		out, ok = v.(*domain.Mapping)
	}
	return
}

func (m *mappingByHost) Set(host string, in *domain.Mapping) {
	m.m.Store(host, in)
}

func (m *mappingByHost) List() (out []*domain.Mapping) {
	m.m.Range(func(_, v interface{}) bool {
		if m, ok := v.(*domain.Mapping); ok {
			out = append(out, m)
		}
		return true
	})
	return
}

func (m *mappingByHost) Delete(host string) {
	m.m.Delete(host)
}

type hosts struct {
	m     sync.Map
	ipSet sync.Map
}

func (h *hosts) GetIP(host string) net.IP {
	if ip, ok := h.LookupIP(host); ok {
		return ip
	}
	for {
		ip := netutil.RandomLoopbackAddr()
		if _, ok := h.ipSet.Load(ip.String()); !ok {
			h.ipSet.Store(ip.String(), struct{}{})
			h.m.Store(host, ip)
			return ip
		}
	}
}

func (h *hosts) LookupIP(host string) (ip net.IP, ok bool) {
	var v interface{}
	if v, ok = h.m.Load(host); ok {
		ip, ok = v.(net.IP)
	}
	return
}

func (h *hosts) Delete(host string) {
	if ip, ok := h.LookupIP(host); ok {
		h.ipSet.Delete(ip.String())
	}
	h.m.Delete(host)
}

type mappingEventEmitter struct {
	id    uint64
	evCh  chan<- domain.MappingEvent
	errCh chan<- error
	ctx   context.Context
}

func (e *mappingEventEmitter) Emit(ev domain.MappingEvent) bool {
	select {
	case <-e.ctx.Done():
		e.errCh <- e.ctx.Err()
		return false
	default:
		e.evCh <- ev
		return true
	}
}
