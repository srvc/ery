package domain

import "context"

// MappingRepository is an interface for accessing <hostname>-<port> mappings.
type MappingRepository interface {
	List(ctx context.Context) ([]*Mapping, error)
	HasHost(ctx context.Context, host string) (bool, error)
	MapAddr(ctx context.Context, addr Addr) (Addr, error)
	Create(ctx context.Context, mapping *Mapping) error
	DeleteByHost(ctx context.Context, host string) error
	ListenEvent(ctx context.Context) (<-chan MappingEvent, <-chan error)
}

// MappingEvent contains event type and subject mapping.
type MappingEvent struct {
	Type MappingEventType
	Mapping
}

// MappingEventType represents mapping lifecycle events, such as "created".
type MappingEventType int

// Enum values of MappingEventType.
const (
	MappingEventCreated MappingEventType = iota
	MappingEventDestroyed
)
