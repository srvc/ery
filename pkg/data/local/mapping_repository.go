package local

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/srvc/ery/pkg/domain"
)

// NewMappingRepository creates a new MappingRepository instance that can access local data.
func NewMappingRepository() domain.MappingRepository {
	return &mappingRepositoryImpl{
		mappingByHost: new(sync.Map),
	}
}

type mappingRepositoryImpl struct {
	mappingByHost *sync.Map
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

	return nil
}

func (r *mappingRepositoryImpl) DeleteByHost(ctx context.Context, host string) error {
	r.mappingByHost.Delete(host)
	return nil
}
