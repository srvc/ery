package domain

import (
	"context"
	"net"

	"github.com/srvc/ery"
)

type IPPool interface {
	Get(context.Context) (net.IP, error)
}

type PortPool interface {
	Get(context.Context) (ery.Port, error)
}
