package di

import (
	"github.com/google/wire"

	"github.com/srvc/ery/pkg/app/api"
	"github.com/srvc/ery/pkg/app/dns"
	"github.com/srvc/ery/pkg/ery"
)

func ProvideAPIConfig(cfg *ery.Config) *api.Config { return &cfg.API }
func ProvideDNSConfig(cfg *ery.Config) *dns.Config { return &cfg.DNS }

var CommonSet = wire.NewSet(
	ProvideAPIConfig,
	ProvideDNSConfig,
)
