package source

import "net"

// NetworkInterfaces lists the machine's network interfaces and their
// addresses. Addrs is separate from Interfaces because net.Interface.Addrs
// itself makes a fresh OS query (it's not just a field read on the struct
// Interfaces already returned), so it needs its own seam to stay
// fixture-testable.
type NetworkInterfaces interface {
	Interfaces() ([]net.Interface, error)
	Addrs(iface net.Interface) ([]net.Addr, error)
}

// RealNetworkInterfaces reads the real network interfaces via the net
// package.
type RealNetworkInterfaces struct{}

func (RealNetworkInterfaces) Interfaces() ([]net.Interface, error) {
	return net.Interfaces()
}

func (RealNetworkInterfaces) Addrs(iface net.Interface) ([]net.Addr, error) {
	return iface.Addrs()
}
