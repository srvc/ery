package discovery

import "context"

// Server is an interface of servers listening tcp ports.
type Server interface {
	Serve() error
	Shutdown(ctx context.Context) error
	Addr() string
}
