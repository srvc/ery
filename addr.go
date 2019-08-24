package ery

import (
	"fmt"
	"net"
)

type Port uint16

type Addr struct {
	IP   net.IP
	Port Port
}

func (a *Addr) String() string {
	var ip string
	if a.IP != nil {
		ip = a.IP.String()
	}
	return fmt.Sprintf("%s:%d", ip, a.Port)
}
