package local

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"

	"github.com/srvc/ery/pkg/domain"
)

// NewMappingRepository creates a new MappingRepository instance that can access local data.
func NewMappingRepository() domain.MappingRepository {
	return &mappingRepositoryImpl{
		mappingByHost: new(sync.Map),
		eventEmitters: new(sync.Map),
	}
}

type mappingRepositoryImpl struct {
	mappingByHost     *sync.Map
	eventEmitters     *sync.Map
	eventEmitterIDSeq uint64
}

func (r *mappingRepositoryImpl) List(ctx context.Context) ([]*domain.Mapping, error) {
	mappings := []*domain.Mapping{}
	r.mappingByHost.Range(func(_, v interface{}) bool {
		if m, ok := v.(*domain.Mapping); ok {
			mappings = append(mappings, m)
		}
		return true
	})
	return mappings, nil
}

func (r *mappingRepositoryImpl) HasHost(ctx context.Context, host string) (bool, error) {
	_, ok := r.mappingByHost.Load(host)
	return ok, nil
}

func (r *mappingRepositoryImpl) MapAddr(ctx context.Context, addr domain.Addr) (domain.Addr, error) {
	if v, ok := r.mappingByHost.Load(addr.Host); ok {
		if m, ok := v.(*domain.Mapping); ok {
			if got := m.Map(addr.Port); got.IsValid() {
				return got, nil
			}
		} else {
			r.DeleteByHost(ctx, addr.Host)
		}
	}
	return domain.Addr{}, errors.Errorf("%v is not found", addr)
}

func (r *mappingRepositoryImpl) Create(ctx context.Context, m *domain.Mapping) error {
	if _, ok := r.mappingByHost.Load(m.Host); ok {
		return errors.Errorf("%v has already been registered", m.Host)
	}

	r.mappingByHost.Store(m.Host, m)

	r.emitEvent(domain.MappingEvent{
		Type:    domain.MappingEventCreated,
		Mapping: *m,
	})

	return nil
}

func (r *mappingRepositoryImpl) DeleteByHost(ctx context.Context, host string) error {
	if v, ok := r.mappingByHost.Load(host); ok {
		r.mappingByHost.Delete(host)
		if m, ok := v.(*domain.Mapping); ok {
			r.emitEvent(domain.MappingEvent{
				Type:    domain.MappingEventDestroyed,
				Mapping: *m,
			})
		}
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
