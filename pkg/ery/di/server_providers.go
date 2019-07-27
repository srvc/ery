package di

import (
	"github.com/google/wire"

	"github.com/srvc/ery/pkg/app/api"
	"github.com/srvc/ery/pkg/app/container"
	"github.com/srvc/ery/pkg/app/dns"
	"github.com/srvc/ery/pkg/app/proxy"
	"github.com/srvc/ery/pkg/data/local"
	"github.com/srvc/ery/pkg/domain"
	"github.com/srvc/ery/pkg/ery"
)

func ProvideContainerWatcher(
	cfg *ery.Config,
	mappingRepo domain.MappingRepository,
	containerRepo domain.ContainerRepository,
) container.Watcher {
	return container.NewWatcher(
		mappingRepo,
		[]domain.ContainerRepository{
			containerRepo,
		},
		cfg.TLD,
		cfg.Package+".hostname",
	)
}

func ProvideLocalMappingRepository() domain.MappingRepository {
	return local.NewMappingRepository()
}

func ProvideLocalDockerContainerRepository() domain.ContainerRepository {
	return local.NewDockerContainerRepository()
}

var ServerSet = wire.NewSet(
	CommonSet,
	ServerApp{},
	api.NewServer,
	dns.NewServer,
	proxy.NewManager,
	proxy.NewFactory,
	ProvideContainerWatcher,
	ProvideLocalMappingRepository,
	ProvideLocalDockerContainerRepository,
)
