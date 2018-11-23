package domain

import (
	"context"
	"net"
)

// MappingRepository is an interface for accessing <hostname>-<port> mappings.
type MappingRepository interface {
	List(ctx context.Context) ([]*Mapping, error)
	LookupIP(ctx context.Context, host string) (net.IP, bool)
	MapAddr(ctx context.Context, addr Addr) (Addr, error)
	Create(ctx context.Context, lAddr Addr, rPort Port) (Addr, error)
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
