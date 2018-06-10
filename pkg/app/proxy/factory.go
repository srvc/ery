package proxy

import (
	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/domain"
)

// ServerFactory is a factory object for creating proxy server instances.
type ServerFactory interface {
	CreateServer(port domain.Port) app.Server
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

func (f *serverFactory) CreateServer(port domain.Port) app.Server {
	return newServerWithPort(f.mappingRepo, port)
}
