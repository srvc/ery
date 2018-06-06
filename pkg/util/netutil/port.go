package netutil

import (
	"net"
)

// GetFreePort find free open port that is ready to use.
func GetFreePort() (int, error) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer lis.Close()

	return lis.Addr().(*net.TCPAddr).Port, nil
}
