package netutil

import (
	"net"

	"github.com/srvc/ery/pkg/domain"
)

// GetFreePort find free open port that is ready to use.
func GetFreePort(host string) (domain.Port, error) {
	lis, err := net.Listen("tcp", host+":0")
	if err != nil {
		return 0, err
	}
	defer lis.Close()

	return domain.Port(lis.Addr().(*net.TCPAddr).Port), nil
}
