// Package render turns collected section data into output: plain text for
// the terminal, or JSON. Renderers never call out to the system themselves;
// they only format what internal/collect already gathered.
package render

import (
	"fmt"
	"io"

	"fcuny.net/lsmac/internal/chips"
	"fcuny.net/lsmac/internal/collect"
)

// CPU writes the CPU summary line, plus its memory bandwidth when known.
func CPU(w io.Writer, cpu collect.CPU, specs chips.Specs) error {
	if _, err := fmt.Fprintf(w, "Chip       %s\n", cpu.BrandName); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "CPU        %d cores: %dP + %dE\n",
		cpu.TotalCores, cpu.PerformanceCores, cpu.EfficiencyCores); err != nil {
		return err
	}
	if specs.CPUBandwidthGBs > 0 {
		if _, err := fmt.Fprintf(w, "Bandwidth  %d GB/s\n", specs.CPUBandwidthGBs); err != nil {
			return err
		}
	}
	return nil
}

// GPU writes the GPU summary line.
func GPU(w io.Writer, gpu collect.GPU) error {
	_, err := fmt.Fprintf(w, "GPU        %d cores\n", gpu.CoreCount)
	return err
}
