package collect

import (
	"errors"
	"net"
	"testing"
)

type fakeNetworkInterfaces struct {
	ifaces []net.Interface
	err    error
}

func (f fakeNetworkInterfaces) Interfaces() ([]net.Interface, error) {
	return f.ifaces, f.err
}

func TestCollectIO(t *testing.T) {
	ni := fakeNetworkInterfaces{ifaces: []net.Interface{
		{Name: "lo0", Flags: net.FlagUp | net.FlagLoopback},
		{Name: "en0", HardwareAddr: net.HardwareAddr{0xc4, 0x35, 0xd9, 0x89, 0x5c, 0x6c}, Flags: net.FlagUp | net.FlagBroadcast},
	}}

	got, err := CollectIO(ni)
	if err != nil {
		t.Fatalf("CollectIO() error = %v", err)
	}

	want := IO{Interfaces: []NetworkInterface{
		{Name: "lo0", IsUp: true, IsLoopback: true},
		{Name: "en0", HardwareAddr: "c4:35:d9:89:5c:6c", IsUp: true},
	}}
	if len(got.Interfaces) != len(want.Interfaces) {
		t.Fatalf("CollectIO() = %+v, want %+v", got, want)
	}
	for i := range want.Interfaces {
		if got.Interfaces[i] != want.Interfaces[i] {
			t.Errorf("Interfaces[%d] = %+v, want %+v", i, got.Interfaces[i], want.Interfaces[i])
		}
	}
}

func TestCollectIOError(t *testing.T) {
	ni := fakeNetworkInterfaces{err: errors.New("netlink unavailable")}

	if _, err := CollectIO(ni); err == nil {
		t.Fatal("CollectIO() error = nil, want error propagated from NetworkInterfaces")
	}
}
