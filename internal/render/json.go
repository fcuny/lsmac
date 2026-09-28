package render

import (
	"encoding/json"
	"io"

	"fcuny.net/lsmac/internal/collect"
	"fcuny.net/lsmac/internal/models"
)

// JSON writes sections (one entry per section name, e.g. from MachineJSON /
// CPUJSON / MemoryJSON or a plain collect.* struct) as indented JSON.
func JSON(w io.Writer, sections map[string]any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(sections)
}

// MachineJSON merges the collected Machine facts with the marketing name
// resolved from internal/models, which text rendering already does via its
// name/model parameters.
func MachineJSON(machine collect.Machine, model models.Model) any {
	type view struct {
		collect.Machine
		MarketingName string `json:"marketingName,omitempty"`
	}
	return view{Machine: machine, MarketingName: model.MarketingName}
}

// CPUJSON merges the compact CPU facts (brand, core counts) with the
// detailed per-cluster/family/features breakdown that's otherwise only
// available via `--section cpu`: JSON output favors completeness over the
// text view's fast/detailed split.
func CPUJSON(cpu collect.CPU, detail collect.CPUDetail) any {
	type view struct {
		BrandName        string               `json:"brandName"`
		TotalCores       uint16               `json:"totalCores"`
		PerformanceCores uint16               `json:"performanceCores"`
		EfficiencyCores  uint16               `json:"efficiencyCores"`
		Family           string               `json:"family"`
		PageSize         uint32               `json:"pageSize"`
		Clusters         []collect.CPUCluster `json:"clusters"`
		Features         []string             `json:"features"`
	}
	return view{
		BrandName:        cpu.BrandName,
		TotalCores:       cpu.TotalCores,
		PerformanceCores: cpu.PerformanceCores,
		EfficiencyCores:  cpu.EfficiencyCores,
		Family:           detail.Family,
		PageSize:         detail.PageSize,
		Clusters:         detail.Clusters,
		Features:         detail.Features,
	}
}

// MemoryJSON merges the collected Memory facts with bandwidth from the
// chip table, which text rendering already takes as a separate parameter.
func MemoryJSON(mem collect.Memory, bandwidthGBs uint32) any {
	type view struct {
		collect.Memory
		BandwidthGBs uint32 `json:"bandwidthGBs,omitempty"`
	}
	return view{Memory: mem, BandwidthGBs: bandwidthGBs}
}
