package collect

import (
	"net"

	"fcuny.net/lsmac/internal/source"
)

// NetworkInterface describes one network interface.
type NetworkInterface struct {
	Name         string `json:"name"`
	HardwareAddr string `json:"hardwareAddr,omitempty"` // MAC address, empty when the interface doesn't have one
	IsUp         bool   `json:"isUp"`
	IsLoopback   bool   `json:"isLoopback"`
}

// IO holds the facts collected for the I/O section.
//
// This is deliberately narrow for now: Thunderbolt/USB4 port count, Wi-Fi
// standard, and Bluetooth version are not collected. None of them have a
// clean, fast, verifiable source - ioreg exposes them (if at all) only
// through deeply nested, undocumented driver-internal structures (e.g.
// IOThunderboltPort matches 16 internal PCIe topology objects on a Mac
// with 2 physical ports, not a simple count), and the reliable alternative,
// system_profiler, is far too slow for the default path (SPAirPortDataType
// alone took over 4 seconds in testing). Left as a follow-up.
type IO struct {
	Interfaces []NetworkInterface `json:"interfaces"`
}

// CollectIO lists network interfaces via the standard library.
func CollectIO(ni source.NetworkInterfaces) (IO, error) {
	ifaces, err := ni.Interfaces()
	if err != nil {
		return IO{}, err
	}

	list := make([]NetworkInterface, len(ifaces))
	for i, iface := range ifaces {
		list[i] = NetworkInterface{
			Name:         iface.Name,
			HardwareAddr: iface.HardwareAddr.String(),
			IsUp:         iface.Flags&net.FlagUp != 0,
			IsLoopback:   iface.Flags&net.FlagLoopback != 0,
		}
	}

	return IO{Interfaces: list}, nil
}
