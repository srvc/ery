package netutil

import (
	"math/rand"
	"net"
	"time"

	"go.uber.org/zap"
)

// LocalIP detects an IP address of localhost.
func LocalIP() (localhost net.IP) {
	localhost = net.IPv4(127, 0, 0, 1)

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		zap.L().Warn("failed to retrieve interface addresses", zap.Error(err))
		return
	}

	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		default:
			continue
		}
		if !ip.IsLoopback() && ip.To4() != nil {
			localhost = ip
			return
		}
	}

	return
}

func RandomLoopbackAddr() net.IP {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return net.IPv4(127, 0, 0, byte(1+r.Intn(255)))
}
