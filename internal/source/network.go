package source

import "net"

// NetworkInterfaces lists the machine's network interfaces.
type NetworkInterfaces interface {
	Interfaces() ([]net.Interface, error)
}

// RealNetworkInterfaces reads the real network interfaces via the net
// package.
type RealNetworkInterfaces struct{}

func (RealNetworkInterfaces) Interfaces() ([]net.Interface, error) {
	return net.Interfaces()
}
