package domain

import "context"

// MappingRepository is an interface for accessing <hostname>-<port> mappings.
type MappingRepository interface {
	List(ctx context.Context) ([]*Mapping, error)
	HasHost(ctx context.Context, host string) (bool, error)
	MapAddr(ctx context.Context, addr Addr) (Addr, error)
	Create(ctx context.Context, mapping *Mapping) error
	DeleteByHost(ctx context.Context, host string) error
}
