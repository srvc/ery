package domain

// PortAddrMap is mapping of ports and addresses.
type PortAddrMap map[Port]Addr

// Mapping represents <hostname>:<port> - <local IP>:<port> map.
type Mapping struct {
	Host        string      `json:"host"`
	PortAddrMap PortAddrMap `json:"map"`
}

// Map returns an Addr mapped on the given port.
func (m *Mapping) Map(port Port) Addr {
	return m.PortAddrMap[port]
}
