// Command lsmac prints hardware and system information about the Apple
// Silicon Mac it runs on.
package main

import (
	"fmt"
	"os"

	"fcuny.net/lsmac/internal/chips"
	"fcuny.net/lsmac/internal/collect"
	"fcuny.net/lsmac/internal/render"
	"fcuny.net/lsmac/internal/source"
)

func main() {
	cmd := source.RealCommand{}
	ok := false

	if cpu, err := collect.CollectCPU(cmd); err != nil {
		fmt.Fprintf(os.Stderr, "lsmac: cpu: %v\n", err)
	} else {
		specs := chips.FromBrandString(cpu.BrandName).Specs()
		if err := render.CPU(os.Stdout, cpu, specs); err != nil {
			fmt.Fprintf(os.Stderr, "lsmac: %v\n", err)
		}
		ok = true
	}

	if gpu, err := collect.CollectGPU(cmd); err != nil {
		fmt.Fprintf(os.Stderr, "lsmac: gpu: %v\n", err)
	} else {
		if err := render.GPU(os.Stdout, gpu); err != nil {
			fmt.Fprintf(os.Stderr, "lsmac: %v\n", err)
		}
		ok = true
	}

	if !ok {
		os.Exit(1)
	}
}
