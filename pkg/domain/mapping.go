package domain

// PortMap is mapping of ports and addresses.
type PortMap map[Port]Port

// Mapping represents <hostname>:<port> - <local IP>:<port> map.
type Mapping struct {
	VirtualHost string  `json:"virtual_host"`
	ProxyHost   string  `json:"proxy_host"`
	PortMap     PortMap `json:"port_map"`
}

// Map returns an Addr mapped on the given port.
func (m *Mapping) Map(port Port) Addr {
	return Addr{Port: m.PortMap[port]}
}
