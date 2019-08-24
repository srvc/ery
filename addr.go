package ery

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

type Port uint16

type Addr struct {
	IP   net.IP
	Port Port
}

var _ pflag.Value = (*Addr)(nil)

func (a *Addr) String() string {
	var ip string
	if a.IP != nil {
		ip = a.IP.String()
	}
	return fmt.Sprintf("%s:%d", ip, a.Port)
}

func (a *Addr) Set(in string) error {
	chunks := strings.Split(in, ":")

	if c := strings.Join(chunks[:len(chunks)-1], ":"); c != "" {
		if ip := net.ParseIP(c); ip != nil {
			a.IP = ip
		} else {
			return fmt.Errorf("invalid address format: %s", in)
		}
	}

	if c := chunks[len(chunks)-1]; c != "" {
		port, err := strconv.ParseUint(c, 10, 16)
		if err != nil {
			return err
		}
		a.Port = Port(port)
	}

	if a.Port == 0 {
		return fmt.Errorf("invalid address format: %s", in)
	}

	return nil
}

func (Addr) Type() string { return "Addr" }
