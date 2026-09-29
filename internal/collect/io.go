package collect

import (
	"net"

	"fcuny.net/lsmac/internal/source"
)

// NetworkInterface describes one network interface.
type NetworkInterface struct {
	Name         string   `json:"name"`
	HardwareAddr string   `json:"hardwareAddr,omitempty"` // MAC address, empty when the interface doesn't have one
	IsUp         bool     `json:"isUp"`
	IsLoopback   bool     `json:"isLoopback"`
	Addrs        []string `json:"addrs,omitempty"` // CIDR notation, e.g. "192.168.1.5/24"
}

// IO holds the facts collected for the I/O section.
//
// This is deliberately narrow for now: Thunderbolt/USB4 port count, Wi-Fi
// standard, Bluetooth version, and the connected Wi-Fi network name are not
// collected.
//
//   - Thunderbolt/USB4 port count, Wi-Fi standard, Bluetooth version: none
//     of them have a clean, fast, verifiable source - ioreg exposes them
//     (if at all) only through deeply nested, undocumented driver-internal
//     structures (e.g. IOThunderboltPort matches 16 internal PCIe topology
//     objects on a Mac with 2 physical ports, not a simple count), and the
//     reliable alternative, system_profiler, is far too slow for the
//     default path (SPAirPortDataType alone took over 4 seconds in
//     testing).
//   - The connected Wi-Fi SSID: since macOS Sonoma, ioreg/networksetup/
//     system_profiler all return the literal string "<SSID Redacted>"
//     unless the calling application has been granted Location Services
//     permission in System Settings - an unconditional macOS privacy gate
//     (TCC), not a privilege issue sudo would fix. The only workaround
//     found requires root or a setuid binary, which is off the table per
//     this project's no-sudo rule. Not worth collecting a value that would
//     show a placeholder for nearly everyone running it.
//
// All left as a follow-up.
type IO struct {
	Interfaces []NetworkInterface `json:"interfaces"`
}

// CollectIO lists network interfaces and their addresses via the standard
// library.
func CollectIO(ni source.NetworkInterfaces) (IO, error) {
	ifaces, err := ni.Interfaces()
	if err != nil {
		return IO{}, err
	}

	list := make([]NetworkInterface, len(ifaces))
	for i, iface := range ifaces {
		// Addrs is best-effort: an error here (e.g. a transient interface
		// that disappeared between Interfaces() and this call) just means
		// no addresses for this one interface, not a failed run.
		var addrStrs []string
		if addrs, err := ni.Addrs(iface); err == nil {
			addrStrs = make([]string, len(addrs))
			for j, a := range addrs {
				addrStrs[j] = a.String()
			}
		}

		list[i] = NetworkInterface{
			Name:         iface.Name,
			HardwareAddr: iface.HardwareAddr.String(),
			IsUp:         iface.Flags&net.FlagUp != 0,
			IsLoopback:   iface.Flags&net.FlagLoopback != 0,
			Addrs:        addrStrs,
		}
	}

	return IO{Interfaces: list}, nil
}
