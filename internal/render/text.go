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

// Chip writes the chip identity line, plus process node and memory
// bandwidth when known. name falls back to the raw chip ID when the chip
// is not in the internal/chips table.
func Chip(w io.Writer, chip chips.Chip) error {
	name := chip.MarketingName
	if name == "" {
		name = chip.ID
	}
	if _, err := fmt.Fprintf(w, "Chip       %s (%s)\n", name, chip.ID); err != nil {
		return err
	}
	if chip.ProcessNode != "" {
		if _, err := fmt.Fprintf(w, "Process    %s\n", chip.ProcessNode); err != nil {
			return err
		}
	}
	if chip.MemoryBandwidthGBs > 0 {
		if _, err := fmt.Fprintf(w, "Bandwidth  %d GB/s\n", chip.MemoryBandwidthGBs); err != nil {
			return err
		}
	}
	return nil
}

// CPU writes the CPU core-count summary line.
func CPU(w io.Writer, cpu collect.CPU) error {
	_, err := fmt.Fprintf(w, "CPU        %d cores: %dP + %dE\n",
		cpu.TotalCores, cpu.PerformanceCores, cpu.EfficiencyCores)
	return err
}

// GPU writes the GPU summary line.
func GPU(w io.Writer, gpu collect.GPU) error {
	_, err := fmt.Fprintf(w, "GPU        %d cores\n", gpu.CoreCount)
	return err
}
