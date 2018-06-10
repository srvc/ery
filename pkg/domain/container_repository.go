package domain

import "context"

// ContainerRepository is an interface for accessing containers.
type ContainerRepository interface {
	ListenEvent(context.Context) (<-chan ContainerEvent, <-chan error)
}

// ContainerEvent contains event type and subject container meta data.
type ContainerEvent struct {
	Type      ContainerEventType
	Container Container
	Error     error
}

// ContainerEventType represents container lifecycle events, such as "created".
type ContainerEventType int

// Enum values of ContainerEventType.
const (
	ContainerEventCreated ContainerEventType = iota
	ContainerEventDestroyed
)
