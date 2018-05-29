package discovery

import (
	"net"
	"sync"

	"github.com/srvc/ery/pkg/util/netutil"
)

// Mapper manages <hostname> - <local IP>:<port> mappings.
type Mapper interface {
	Lookup(host string) (targetHost string, ok bool)
	Add(port uint32, host string)
	Remove(port uint32)
	Clear()
}

// NewMapper creates a new Mapper instance.
func NewMapper(localIP net.IP) Mapper {
	m := &mapperImpl{
		localIP: localIP,
	}
	m.Clear()
	return m
}

type mapperImpl struct {
	hostnamesByPort *sync.Map
	hostTable       *sync.Map
	localIP         net.IP
}

func (m *mapperImpl) Lookup(host string) (targetHost string, ok bool) {
	if v, stored := m.hostTable.Load(host); stored {
		targetHost, ok = v.(string)
	}
	return
}

func (m *mapperImpl) Add(port uint32, host string) {
	m.appendHostname(port, host)
	m.hostTable.Store(host, netutil.HostAndPort(m.localIP.String(), port))
}

func (m *mapperImpl) Remove(port uint32) {
	for _, hostname := range m.getHostnamesByPort(port) {
		m.hostTable.Delete(hostname)
	}
	m.hostnamesByPort.Delete(port)
}

func (m *mapperImpl) Clear() {
	m.hostnamesByPort = new(sync.Map)
	m.hostTable = new(sync.Map)
}

func (m *mapperImpl) appendHostname(port uint32, hostname string) {
	m.hostnamesByPort.Store(port, append(m.getHostnamesByPort(port), hostname))
}

func (m *mapperImpl) getHostnamesByPort(port uint32) (hostnames []string) {
	if v, ok := m.hostnamesByPort.Load(port); ok {
		hostnames, _ = v.([]string)
	}
	return
}
