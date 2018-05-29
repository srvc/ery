package netutil

import "strconv"

// HostAndPort joins a hostname and a port.
func HostAndPort(host string, port uint32) string {
	return host + ":" + strconv.FormatUint(uint64(port), 10)
}
