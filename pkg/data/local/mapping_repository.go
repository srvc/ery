package local

import (
	"net"
	"sync"

	"github.com/pkg/errors"

	"github.com/srvc/ery/pkg/domain"
	"github.com/srvc/ery/pkg/util/netutil"
)

// NewMappingRepository creates a new MappingRepository instance that can access local data.
func NewMappingRepository(localIP net.IP) domain.MappingRepository {
	m := &mappingRepositoryImpl{
		localIP: localIP,
	}
	m.clear()
	return m
}

type mappingRepositoryImpl struct {
	hostnamesByPort *sync.Map
	hostTable       *sync.Map
	localIP         net.IP
}

func (m *mappingRepositoryImpl) List() ([]*domain.Mapping, error) {
	mappings := []*domain.Mapping{}
	m.hostnamesByPort.Range(func(k, v interface{}) bool {
		mappings = append(mappings, &domain.Mapping{
			Addr: domain.Addr{
				IP:   m.localIP,
				Port: k.(uint32),
			},
			Hostnames: v.([]string),
		})
		return true
	})
	return mappings, nil
}

func (m *mappingRepositoryImpl) GetBySourceHost(host string) (targetHost string, err error) {
	if v, stored := m.hostTable.Load(host); stored {
		var ok bool
		targetHost, ok = v.(string)
		if !ok {
			err = errors.Errorf("%s is not found", host)
		}
	}
	return
}

func (m *mappingRepositoryImpl) Create(port uint32, hosts ...string) error {
	if len(hosts) == 0 {
		return errors.Errorf(":%d needs to map at least 1 hosts", port)
	}
	target := netutil.HostAndPort(m.localIP.String(), port)
	for _, host := range hosts {
		m.appendHostname(port, host)
		m.hostTable.Store(host, target)
	}
	return nil
}

func (m *mappingRepositoryImpl) Delete(port uint32) error {
	for _, hostname := range m.getHostnamesByPort(port) {
		m.hostTable.Delete(hostname)
	}
	m.hostnamesByPort.Delete(port)
	return nil
}

func (m *mappingRepositoryImpl) DeleteAll() error {
	m.clear()
	return nil
}

func (m *mappingRepositoryImpl) appendHostname(port uint32, hostname string) {
	m.hostnamesByPort.Store(port, append(m.getHostnamesByPort(port), hostname))
}

func (m *mappingRepositoryImpl) getHostnamesByPort(port uint32) (hostnames []string) {
	if v, ok := m.hostnamesByPort.Load(port); ok {
		hostnames, _ = v.([]string)
	}
	return
}

func (m *mappingRepositoryImpl) clear() {
	m.hostnamesByPort = new(sync.Map)
	m.hostTable = new(sync.Map)
}
