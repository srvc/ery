package domain

// Mapping represents <hostname> - <local IP>:<port> map.
type Mapping struct {
	Addr
	Hostnames []string
}
