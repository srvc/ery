package local

import (
	"context"
	"net"
	"sync"

	"github.com/srvc/ery/pkg/ery/domain"
)

type PortPool struct {
	m sync.Mutex
}

var _ domain.PortPool = (*PortPool)(nil)

func NewPortPool() *PortPool {
	return &PortPool{}
}

func (p *PortPool) Get(ctx context.Context) (domain.Port, error) {
	p.m.Lock()
	defer p.m.Unlock()

	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return domain.Port(0), err
	}

	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return domain.Port(0), err
	}
	defer lis.Close()

	return domain.Port(lis.Addr().(*net.TCPAddr).Port), nil
}
