package collect

import (
	"fmt"
	"strings"

	"fcuny.net/lsmac/internal/source"
)

// Firmware holds the facts collected for the Firmware section.
type Firmware struct {
	Version    string `json:"version"` // e.g. "mBoot-20457.1.29", as the device tree names it
	SecureBoot bool   `json:"secureBoot"`
	SIPEnabled bool   `json:"sipEnabled"`
}

const csrutilPath = "/usr/bin/csrutil"

// CollectFirmware reads the system firmware version and Secure Boot flag
// from the device tree's "chosen" node, and SIP status via csrutil.
//
// Secure Boot's finer-grained policy (Full/Reduced/Permissive Security) is
// not collected: the only reliable source is `bputil -d`, which requires
// root, and the device tree's own security-mode properties
// (effective-security-mode-ap/-sep) have no documented mapping to those
// policy names. The plain secure-boot flag is unambiguous and unprivileged.
func CollectFirmware(cmd source.SystemCommand) (Firmware, error) {
	props, err := source.IORegNodeProperties(cmd, "chosen")
	if err != nil {
		return Firmware{}, err
	}

	version, err := source.DataProperty(props, "firmware-version")
	if err != nil {
		return Firmware{}, err
	}

	secureBoot, err := source.DataPropertyUint(props, "secure-boot")
	if err != nil {
		return Firmware{}, err
	}

	sip, err := source.Run(cmd, csrutilPath, "status")
	if err != nil {
		return Firmware{}, fmt.Errorf("csrutil status: %w", err)
	}
	sipEnabled, err := parseSIPStatus(sip)
	if err != nil {
		return Firmware{}, err
	}

	return Firmware{
		Version:    version,
		SecureBoot: secureBoot != 0,
		SIPEnabled: sipEnabled,
	}, nil
}

func parseSIPStatus(output string) (bool, error) {
	lower := strings.ToLower(output)
	switch {
	case strings.Contains(lower, "disabled"):
		return false, nil
	case strings.Contains(lower, "enabled"):
		return true, nil
	default:
		return false, fmt.Errorf("csrutil status: unexpected output: %q", output)
	}
}
