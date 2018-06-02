package domain

import "net"

type Addr struct {
	IP   net.IP
	Port uint32
}

type Mapping struct {
	Addr
	Hostnames []string
}
