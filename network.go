package ery

import (
	"fmt"
	"strings"
)

type Network int

const (
	_ Network = iota
	NetworkTCP
	NetworkUDP
)

func (n *Network) Set(in string) error {
	switch strings.ToUpper(in) {
	case NetworkTCP.String():
		*n = NetworkTCP
	case NetworkUDP.String():
		*n = NetworkUDP
	default:
		return fmt.Errorf("unknown network type: %s", in)
	}
	return nil
}

func (*Network) Type() string { return "Network" }
