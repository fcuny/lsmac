package collect

import "fcuny.net/lsmac/internal/source"

// Machine holds the facts collected for the Machine section. Serial and
// HardwareUUID are left empty unless showSerial is passed to CollectMachine,
// so screenshots of the default output don't leak them.
type Machine struct {
	ModelIdentifier string `json:"modelIdentifier"` // e.g. "Mac14,2"
	SKU             string `json:"sku"`             // e.g. "MN703LL/A": model-number + region-info
	Serial          string `json:"serial,omitempty"`
	HardwareUUID    string `json:"hardwareUUID,omitempty"`
}

// CollectMachine reads the model identifier via sysctl and the SKU,
// serial number, and hardware UUID from IOPlatformExpertDevice.
func CollectMachine(cmd source.SystemCommand, showSerial bool) (Machine, error) {
	modelIdentifier, err := source.Sysctl(cmd, "hw.model")
	if err != nil {
		return Machine{}, err
	}

	props, err := source.IORegProperties(cmd, "IOPlatformExpertDevice")
	if err != nil {
		return Machine{}, err
	}

	modelNumber, err := source.DataProperty(props, "model-number")
	if err != nil {
		return Machine{}, err
	}
	regionInfo, err := source.DataProperty(props, "region-info")
	if err != nil {
		return Machine{}, err
	}

	m := Machine{
		ModelIdentifier: modelIdentifier[0],
		SKU:             modelNumber + regionInfo,
	}

	if showSerial {
		m.Serial, err = source.StringProperty(props, "IOPlatformSerialNumber")
		if err != nil {
			return Machine{}, err
		}
		m.HardwareUUID, err = source.StringProperty(props, "IOPlatformUUID")
		if err != nil {
			return Machine{}, err
		}
	}

	return m, nil
}
