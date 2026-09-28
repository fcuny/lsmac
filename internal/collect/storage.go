package collect

import (
	"fmt"
	"strings"

	"fcuny.net/lsmac/internal/source"
)

// Storage holds the facts collected for the Storage section. FileVaultOn is
// nil unless checkFileVault was passed to CollectStorage: `fdesetup status`
// is the only source for it and costs ~50ms on its own, about a third of
// lsmac's entire default-view time budget, so the default view skips it
// and only `--section storage` pays for it.
type Storage struct {
	Model       string // internal SSD model, e.g. "APPLE SSD AP1024Z"
	TotalBytes  uint64
	UsedBytes   uint64
	FileVaultOn *bool
}

const fdesetupPath = "/usr/bin/fdesetup"

// CollectStorage reads the internal SSD model via ioreg and capacity/usage
// of the root filesystem via statfs. FileVault status via fdesetup is only
// read when checkFileVault is true - see the comment on Storage.
func CollectStorage(cmd source.SystemCommand, fs source.FilesystemStats, checkFileVault bool) (Storage, error) {
	props, err := source.IORegProperties(cmd, "AppleANS3CGv2Controller")
	if err != nil {
		return Storage{}, err
	}
	model, err := source.StringProperty(props, "Model Number")
	if err != nil {
		return Storage{}, err
	}

	total, available, err := fs.Stat("/")
	if err != nil {
		return Storage{}, fmt.Errorf("statfs /: %w", err)
	}

	storage := Storage{
		Model:      model,
		TotalBytes: total,
		UsedBytes:  total - available,
	}

	if checkFileVault {
		fileVault, err := source.Run(cmd, fdesetupPath, "status")
		if err != nil {
			return Storage{}, fmt.Errorf("fdesetup status: %w", err)
		}
		fileVaultOn, err := parseFileVaultStatus(fileVault)
		if err != nil {
			return Storage{}, err
		}
		storage.FileVaultOn = &fileVaultOn
	}

	return storage, nil
}

func parseFileVaultStatus(output string) (bool, error) {
	lower := strings.ToLower(output)
	switch {
	case strings.Contains(lower, "filevault is on"):
		return true, nil
	case strings.Contains(lower, "filevault is off"):
		return false, nil
	default:
		return false, fmt.Errorf("fdesetup status: unexpected output: %q", output)
	}
}
