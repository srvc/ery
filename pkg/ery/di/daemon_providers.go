package di

import (
	"github.com/google/wire"
	"github.com/srvc/ery/pkg/app/daemon"
	"github.com/srvc/ery/pkg/ery"
)

func ProvideDaemonFactory(cfg *ery.Config) daemon.Factory {
	return daemon.NewFactory(cfg.Name, cfg.Summary)
}

var DaemonSet = wire.NewSet(
	CommonSet,
	DaemonApp{},
	ProvideDaemonFactory,
)
