package proxy

import (
	"github.com/srvc/ery/pkg/domain"
)

// ServerFactory is a factory object for creating proxy server instances.
type ServerFactory interface {
	CreateServer(addr domain.Addr) Server
}

// NewFactory creates a new ServerFactory instance.
func NewFactory(mappingRepo domain.MappingRepository) ServerFactory {
	return &serverFactory{
		mappingRepo: mappingRepo,
	}
}

type serverFactory struct {
	mappingRepo domain.MappingRepository
}

func (f *serverFactory) CreateServer(addr domain.Addr) Server {
	return newServerWithPort(f.mappingRepo, addr)
}
