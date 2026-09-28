// Command lsmac prints hardware and system information about the Apple
// Silicon Mac it runs on.
package main

import (
	"fmt"
	"os"

	"fcuny.net/lsmac/internal/chips"
	"fcuny.net/lsmac/internal/collect"
	"fcuny.net/lsmac/internal/models"
	"fcuny.net/lsmac/internal/render"
	"fcuny.net/lsmac/internal/source"
)

// showSerial gates Machine.Serial/HardwareUUID. Hardcoded off until the
// --show-serial flag is wired up.
const showSerial = false

func main() {
	cmd := source.NewCachingCommand(source.RealCommand{})
	fc := source.RealFileChecker{}
	ok := false

	if osInfo, err := collect.CollectOS(cmd, fc); err != nil {
		fmt.Fprintf(os.Stderr, "lsmac: os: %v\n", err)
	} else {
		if err := render.OS(os.Stdout, osInfo); err != nil {
			fmt.Fprintf(os.Stderr, "lsmac: %v\n", err)
		}
		ok = true
	}

	if machine, err := collect.CollectMachine(cmd, showSerial); err != nil {
		fmt.Fprintf(os.Stderr, "lsmac: machine: %v\n", err)
	} else {
		model, _ := models.Lookup(machine.ModelIdentifier)
		if err := render.Machine(os.Stdout, machine, model); err != nil {
			fmt.Fprintf(os.Stderr, "lsmac: %v\n", err)
		}
		ok = true
	}

	if fw, err := collect.CollectFirmware(cmd); err != nil {
		fmt.Fprintf(os.Stderr, "lsmac: firmware: %v\n", err)
	} else {
		if err := render.Firmware(os.Stdout, fw); err != nil {
			fmt.Fprintf(os.Stderr, "lsmac: %v\n", err)
		}
		ok = true
	}

	if chipID, err := collect.CollectChipID(cmd); err != nil {
		fmt.Fprintf(os.Stderr, "lsmac: chip: %v\n", err)
	} else {
		chip, found := chips.Lookup(chipID)
		if !found {
			chip = chips.Chip{ID: chipID}
		}
		if err := render.Chip(os.Stdout, chip); err != nil {
			fmt.Fprintf(os.Stderr, "lsmac: %v\n", err)
		}
		ok = true
	}

	if cpu, err := collect.CollectCPU(cmd); err != nil {
		fmt.Fprintf(os.Stderr, "lsmac: cpu: %v\n", err)
	} else {
		if err := render.CPU(os.Stdout, cpu); err != nil {
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
