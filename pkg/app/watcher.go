package app

import "context"

// Watcher is an interface of watching some events.
type Watcher interface {
	Watch(context.Context) error
}
