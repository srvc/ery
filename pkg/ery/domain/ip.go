package domain

import (
	"context"
	"net"
)

type IPPool interface {
	Get(context.Context) (net.IP, error)
}
