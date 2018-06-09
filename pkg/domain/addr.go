package domain

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

// Port represents port number.
type Port uint16

// PortFromString returns a Port object from the given string.
func PortFromString(str string) (Port, error) {
	port, err := strconv.ParseUint(str, 10, 16)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return Port(port), nil
}

// Addr contains ip and port.
type Addr struct {
	Host string
	Port Port
}

// NewAddr creates a new Addr object.
func NewAddr(host string, port Port) Addr {
	return Addr{
		Host: host,
		Port: port,
	}
}

// LocalAddr create an Addr object that points to localhost.
func LocalAddr(port Port) Addr {
	return Addr{
		Port: port,
	}
}

// HTTPAddr create an Addr object that points to HTTP port.
func HTTPAddr(host string) Addr {
	return Addr{
		Host: host,
		Port: Port(80),
	}
}

func (a *Addr) String() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

// IsValid returned true if the Addr object is valid.
func (a *Addr) IsValid() bool {
	return *a != Addr{}
}
