package app

import "context"

// Server is an interface of servers listening tcp ports.
type Server interface {
	Serve(context.Context) error
	Addr() string
}
