// +build wireinject

package di

import (
	"github.com/google/go-cloud/wire"

	"github.com/srvc/ery/pkg/ery"
)

func NewServerApp(cfg *ery.Config) *ServerApp {
	wire.Build(ServerSet)
	return nil
}

func NewClientApp(cfg *ery.Config) *ClientApp {
	wire.Build(ClientSet)
	return nil
}

func NewDaemonApp(cfg *ery.Config) *DaemonApp {
	wire.Build(DaemonSet)
	return nil
}
