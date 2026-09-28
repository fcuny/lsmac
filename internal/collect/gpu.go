package collect

import "fcuny.net/lsmac/internal/source"

// GPU holds the GPU facts collected for the SoC section.
type GPU struct {
	CoreCount uint16 `json:"coreCount"`
}

// CollectGPU reads GPU core count from the AGXAccelerator IORegistry entry.
// This replaces the original socinfo code's use of
// `system_profiler SPDisplaysDataType`, which takes about a second to run.
func CollectGPU(cmd source.SystemCommand) (GPU, error) {
	props, err := source.IORegProperties(cmd, "AGXAccelerator")
	if err != nil {
		return GPU{}, err
	}

	cores, err := source.IntProperty(props, "gpu-core-count")
	if err != nil {
		return GPU{}, err
	}

	return GPU{CoreCount: uint16(cores)}, nil
}
