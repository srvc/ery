package netutil

import "net"

// LocalhostIP detects an IP address of localhost.
func LocalhostIP() (localhost net.IP) {
	localhost = net.IPv4(127, 0, 0, 1)

	ips, _ := net.LookupIP("localhost")

	for _, ip := range ips {
		switch ip.String() {
		case "127.0.0.1", "::1":
			continue
		default:
			localhost = ip
			break
		}
	}

	return
}
