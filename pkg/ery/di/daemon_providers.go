package di

import (
	"github.com/google/go-cloud/wire"
	"github.com/srvc/ery/pkg/app/daemon"
	"github.com/srvc/ery/pkg/ery"
)

func ProvideDaemonFactory(cfg *ery.DaemonConfig) daemon.Factory {
	return daemon.NewFactory(cfg.Name, cfg.Description)
}

var DaemonSet = wire.NewSet(
	CommonSet,
	DaemonApp{},
	ProvideDaemonFactory,
)
