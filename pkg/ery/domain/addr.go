package domain

import (
	"context"
	"net"
)

type Port uint16

type Addr struct {
	IP   string
	Port Port
}

type IPPool interface {
	Get(context.Context) (net.IP, error)
}

type PortPool interface {
	Get(context.Context) (Port, error)
}
