package collect

import (
	"strings"

	"fcuny.net/lsmac/internal/source"
)

// CollectChipID reads the SoC's internal chip ID (e.g. "T8112") from the
// device tree's platform-name property, exposed on IOPlatformExpertDevice.
// This is the key internal/chips looks facts up by, since Apple does not
// expose the marketing chip name or its specs directly.
func CollectChipID(cmd source.SystemCommand) (string, error) {
	props, err := source.IORegProperties(cmd, "IOPlatformExpertDevice")
	if err != nil {
		return "", err
	}

	name, err := source.DataProperty(props, "platform-name")
	if err != nil {
		return "", err
	}

	return strings.ToUpper(name), nil
}
