package collect

import (
	"strconv"

	"fcuny.net/lsmac/internal/source"
)

// CPU holds the CPU facts collected for the SoC section.
type CPU struct {
	BrandName        string
	TotalCores       uint16
	PerformanceCores uint16
	EfficiencyCores  uint16
}

// CollectCPU reads CPU brand and core counts via sysctl.
//
// hw.perflevel0 is the performance cluster and hw.perflevel1 is the
// efficiency cluster; the original socinfo code swapped these when
// assigning them to PCoreCount/ECoreCount.
func CollectCPU(cmd source.SystemCommand) (CPU, error) {
	values, err := source.Sysctl(cmd,
		"machdep.cpu.brand_string",
		"machdep.cpu.core_count",
		"hw.perflevel0.logicalcpu",
		"hw.perflevel1.logicalcpu",
	)
	if err != nil {
		return CPU{}, err
	}

	total, err := strconv.ParseUint(values[1], 10, 16)
	if err != nil {
		return CPU{}, err
	}
	pCores, err := strconv.ParseUint(values[2], 10, 16)
	if err != nil {
		return CPU{}, err
	}
	eCores, err := strconv.ParseUint(values[3], 10, 16)
	if err != nil {
		return CPU{}, err
	}

	return CPU{
		BrandName:        values[0],
		TotalCores:       uint16(total),
		PerformanceCores: uint16(pCores),
		EfficiencyCores:  uint16(eCores),
	}, nil
}
