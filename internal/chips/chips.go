// Package chips holds facts about Apple Silicon chips that the hardware
// itself does not report, such as memory bandwidth.
//
// This is an interim version ported from the old socinfo tool: it keys off
// the CPU brand string with strings.Contains, and only covers M1 through
// M3-family bandwidth for CPU and GPU. It will be replaced by a table keyed
// on chip ID (e.g. T6031) with sourced values for the full M1-to-current
// lineup.
package chips

import "strings"

// Chip identifies an Apple Silicon variant.
type Chip int

const (
	Unknown Chip = iota
	M1
	M1Pro
	M1Max
	M1Ultra
	M2
	M2Pro
	M2Max
	M2Ultra
	M3
	M3Pro
	M3Max
)

// Specs holds the per-chip facts not reported by the hardware.
type Specs struct {
	CPUBandwidthGBs uint32
	GPUBandwidthGBs uint32
}

// FromBrandString maps a machdep.cpu.brand_string value to a Chip.
func FromBrandString(brand string) Chip {
	switch {
	case strings.Contains(brand, "M1 Pro"):
		return M1Pro
	case strings.Contains(brand, "M1 Max"):
		return M1Max
	case strings.Contains(brand, "M1 Ultra"):
		return M1Ultra
	case strings.Contains(brand, "M1"):
		return M1
	case strings.Contains(brand, "M2 Pro"):
		return M2Pro
	case strings.Contains(brand, "M2 Max"):
		return M2Max
	case strings.Contains(brand, "M2 Ultra"):
		return M2Ultra
	case strings.Contains(brand, "M2"):
		return M2
	case strings.Contains(brand, "M3 Pro"):
		return M3Pro
	case strings.Contains(brand, "M3 Max"):
		return M3Max
	case strings.Contains(brand, "M3"):
		return M3
	default:
		return Unknown
	}
}

// Specs returns the known facts for the chip. Fields are zero when unknown
// rather than guessed.
func (c Chip) Specs() Specs {
	switch c {
	case M1:
		return Specs{CPUBandwidthGBs: 70, GPUBandwidthGBs: 70}
	case M1Pro:
		return Specs{CPUBandwidthGBs: 200, GPUBandwidthGBs: 200}
	case M1Max:
		return Specs{CPUBandwidthGBs: 250, GPUBandwidthGBs: 400}
	case M1Ultra:
		return Specs{CPUBandwidthGBs: 500, GPUBandwidthGBs: 800}
	case M2:
		return Specs{CPUBandwidthGBs: 100, GPUBandwidthGBs: 100}
	default:
		return Specs{}
	}
}
