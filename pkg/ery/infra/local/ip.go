package local

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"
	"sync"

	"github.com/izumin5210/execx"
	"github.com/srvc/ery/pkg/ery/domain"
)

type IPPool struct {
	loopback net.Interface
	lastByte byte
	m        sync.Mutex
}

var _ domain.IPPool = (*IPPool)(nil)

func NewIPPool() (*IPPool, error) {
	lb, err := getLoopbackInterface()
	if err != nil {
		return nil, err
	}
	return &IPPool{
		loopback: lb,
		lastByte: 2,
	}, nil
}

func (g *IPPool) Get(ctx context.Context) (net.IP, error) {
	g.m.Lock()
	defer g.m.Unlock()

	ip := net.IPv4(127, 0, 3, g.lastByte)
	g.lastByte++

	var args []string

	switch runtime.GOOS {
	case "darwin":
		args = []string{g.loopback.Name, "alias", ip.String(), "up"}

	case "linux":
		args = []string{g.loopback.Name, ip.String(), "up"}

	default:
		return net.IP{}, fmt.Errorf("unsupported os: %s", runtime.GOOS)
	}

	err := execx.CommandContext(ctx, "ifconfig", args...).Run()
	if err != nil {
		return net.IP{}, err
	}

	return ip, nil
}

func getLoopbackInterface() (net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return net.Interface{}, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			return iface, nil
		}
	}
	return net.Interface{}, errors.New("failed to find loopback interface")
}
