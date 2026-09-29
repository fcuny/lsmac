package collect

import (
	"errors"
	"net"
	"reflect"
	"testing"
)

type fakeNetworkInterfaces struct {
	ifaces      []net.Interface
	err         error
	addrsByName map[string][]net.Addr
	addrsErr    error
}

func (f fakeNetworkInterfaces) Interfaces() ([]net.Interface, error) {
	return f.ifaces, f.err
}

func (f fakeNetworkInterfaces) Addrs(iface net.Interface) ([]net.Addr, error) {
	if f.addrsErr != nil {
		return nil, f.addrsErr
	}
	return f.addrsByName[iface.Name], nil
}

func mustParseCIDR(t *testing.T, s string) net.Addr {
	t.Helper()
	ip, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		t.Fatalf("net.ParseCIDR(%q): %v", s, err)
	}
	ipNet.IP = ip
	return ipNet
}

func TestCollectIO(t *testing.T) {
	ni := fakeNetworkInterfaces{
		ifaces: []net.Interface{
			{Name: "lo0", Flags: net.FlagUp | net.FlagLoopback},
			{Name: "en0", HardwareAddr: net.HardwareAddr{0xc4, 0x35, 0xd9, 0x89, 0x5c, 0x6c}, Flags: net.FlagUp | net.FlagBroadcast},
		},
		addrsByName: map[string][]net.Addr{
			"lo0": {mustParseCIDR(t, "127.0.0.1/8")},
			"en0": {mustParseCIDR(t, "192.168.1.161/24"), mustParseCIDR(t, "fe80::1cd4:5590:d53a:7c08/64")},
		},
	}

	got, err := CollectIO(ni)
	if err != nil {
		t.Fatalf("CollectIO() error = %v", err)
	}

	want := IO{Interfaces: []NetworkInterface{
		{Name: "lo0", IsUp: true, IsLoopback: true, Addrs: []string{"127.0.0.1/8"}},
		{Name: "en0", HardwareAddr: "c4:35:d9:89:5c:6c", IsUp: true, Addrs: []string{"192.168.1.161/24", "fe80::1cd4:5590:d53a:7c08/64"}},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CollectIO() = %+v, want %+v", got, want)
	}
}

func TestCollectIOAddrsErrorIsNonFatal(t *testing.T) {
	ni := fakeNetworkInterfaces{
		ifaces:   []net.Interface{{Name: "en0", Flags: net.FlagUp}},
		addrsErr: errors.New("transient interface lookup failure"),
	}

	got, err := CollectIO(ni)
	if err != nil {
		t.Fatalf("CollectIO() error = %v, want no error when only Addrs fails", err)
	}
	if len(got.Interfaces) != 1 || got.Interfaces[0].Addrs != nil {
		t.Errorf("CollectIO() = %+v, want one interface with no addresses", got)
	}
}

func TestCollectIOError(t *testing.T) {
	ni := fakeNetworkInterfaces{err: errors.New("netlink unavailable")}

	if _, err := CollectIO(ni); err == nil {
		t.Fatal("CollectIO() error = nil, want error propagated from NetworkInterfaces")
	}
}
