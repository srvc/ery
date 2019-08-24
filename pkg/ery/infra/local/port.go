package local

import (
	"context"
	"sync"

	"github.com/srvc/ery"
	"github.com/srvc/ery/pkg/ery/domain"
	netutil "github.com/srvc/ery/pkg/util/net"
)

type PortPool struct {
	m sync.Mutex
}

var _ domain.PortPool = (*PortPool)(nil)

func NewPortPool() *PortPool {
	return &PortPool{}
}

func (p *PortPool) Get(ctx context.Context) (ery.Port, error) {
	p.m.Lock()
	defer p.m.Unlock()

	port, err := netutil.GetFreePort()
	if err != nil {
		return ery.Port(0), err
	}

	return ery.Port(port), nil
}
