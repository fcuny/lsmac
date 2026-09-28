package collect

import (
	"fmt"
	"strconv"
	"strings"

	"fcuny.net/lsmac/internal/source"
)

// Memory holds the facts collected for the Memory section.
type Memory struct {
	TotalBytes uint64
	// UsedBytes approximates macOS's own "Memory Used" figure as
	// (active + wired + compressed) pages. Apple doesn't publish the exact
	// formula Activity Monitor uses; this is the same approximation
	// standard community tools use, not a guess at an unknown exact value.
	UsedBytes       uint64
	WiredBytes      uint64
	CompressedBytes uint64
	Type            string // e.g. "LPDDR5", from the device tree; empty if unknown
	Pressure        string // "normal", "warning", or "critical"
}

// CollectMemory reads total memory via sysctl, page-level usage via
// vm_stat, memory pressure via kern.memorystatus_vm_pressure_level, and
// RAM type from the device tree's dram-type property.
func CollectMemory(cmd source.SystemCommand) (Memory, error) {
	values, err := source.Sysctl(cmd, "hw.memsize", "hw.pagesize", "kern.memorystatus_vm_pressure_level")
	if err != nil {
		return Memory{}, err
	}
	total, err := strconv.ParseUint(values[0], 10, 64)
	if err != nil {
		return Memory{}, fmt.Errorf("hw.memsize: %w", err)
	}
	pageSize, err := strconv.ParseUint(values[1], 10, 64)
	if err != nil {
		return Memory{}, fmt.Errorf("hw.pagesize: %w", err)
	}
	pressure, err := parseMemoryPressureLevel(values[2])
	if err != nil {
		return Memory{}, err
	}

	vmStatOutput, err := source.Run(cmd, "/usr/bin/vm_stat")
	if err != nil {
		return Memory{}, fmt.Errorf("vm_stat: %w", err)
	}
	pages, err := parseVMStat(vmStatOutput)
	if err != nil {
		return Memory{}, err
	}

	active, err := pages.require("Pages active")
	if err != nil {
		return Memory{}, err
	}
	wired, err := pages.require("Pages wired down")
	if err != nil {
		return Memory{}, err
	}
	compressed, err := pages.require("Pages occupied by compressor")
	if err != nil {
		return Memory{}, err
	}

	var dramType string
	if props, err := source.IORegNodeProperties(cmd, "chosen"); err == nil {
		dramType, _ = source.StringProperty(props, "dram-type")
	}

	return Memory{
		TotalBytes:      total,
		UsedBytes:       (active + wired + compressed) * pageSize,
		WiredBytes:      wired * pageSize,
		CompressedBytes: compressed * pageSize,
		Type:            dramType,
		Pressure:        pressure,
	}, nil
}

// parseMemoryPressureLevel decodes kern.memorystatus_vm_pressure_level's
// enum: 1=normal, 2=warning, 4=critical.
func parseMemoryPressureLevel(raw string) (string, error) {
	switch raw {
	case "1":
		return "normal", nil
	case "2":
		return "warning", nil
	case "4":
		return "critical", nil
	default:
		return "", fmt.Errorf("kern.memorystatus_vm_pressure_level: unexpected value %q", raw)
	}
}

type vmStatPages map[string]uint64

func (p vmStatPages) require(key string) (uint64, error) {
	v, ok := p[key]
	if !ok {
		return 0, fmt.Errorf("vm_stat: missing %q", key)
	}
	return v, nil
}

// parseVMStat parses vm_stat's "Label:   N." lines into a page-count map.
func parseVMStat(output string) (vmStatPages, error) {
	pages := make(vmStatPages)
	for line := range strings.SplitSeq(output, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		value = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(value), "."))
		n, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			continue // non-numeric lines, e.g. the "Mach Virtual Memory Statistics" header
		}
		pages[strings.TrimSpace(key)] = n
	}
	return pages, nil
}
