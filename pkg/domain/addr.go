package domain

import "net"

// Addr contains ip and port.
type Addr struct {
	IP   net.IP
	Port uint32
}
