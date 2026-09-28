// Command lsmac prints hardware and system information about the Apple
// Silicon Mac it runs on.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"fcuny.net/lsmac/internal/chips"
	"fcuny.net/lsmac/internal/collect"
	"fcuny.net/lsmac/internal/models"
	"fcuny.net/lsmac/internal/render"
	"fcuny.net/lsmac/internal/source"
)

// showSerial gates Machine.Serial/HardwareUUID. Hardcoded off until the
// --show-serial flag is wired up.
const showSerial = false

// sectionFlag collects repeated `--section NAME` occurrences.
type sectionFlag []string

func (s *sectionFlag) String() string { return strings.Join(*s, ",") }

func (s *sectionFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	var sections sectionFlag
	flag.Var(&sections, "section", "print the detailed view of one section (repeatable): os, machine, firmware, chip, cpu, gpu")
	flag.Parse()

	cmd := source.NewCachingCommand(source.RealCommand{})
	fc := source.RealFileChecker{}

	if len(sections) > 0 {
		ok := true
		for _, name := range sections {
			if err := runSection(cmd, fc, name); err != nil {
				fmt.Fprintf(os.Stderr, "lsmac: %s: %v\n", name, err)
				ok = false
			}
		}
		if !ok {
			os.Exit(1)
		}
		return
	}

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

	if chip, err := collectChip(cmd); err != nil {
		fmt.Fprintf(os.Stderr, "lsmac: chip: %v\n", err)
	} else {
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

// collectChip reads the chip ID and looks it up in internal/chips, falling
// back to a bare Chip carrying just the ID when it's not in the table.
func collectChip(cmd source.SystemCommand) (chips.Chip, error) {
	chipID, err := collect.CollectChipID(cmd)
	if err != nil {
		return chips.Chip{}, err
	}
	chip, found := chips.Lookup(chipID)
	if !found {
		chip = chips.Chip{ID: chipID}
	}
	return chip, nil
}

// runSection prints the detailed view for a single named section. Most
// sections don't have a detailed view distinct from the compact one yet,
// so they just print their one compact line/block in isolation; "cpu" is
// the exception, with a genuinely more detailed view.
func runSection(cmd source.SystemCommand, fc source.FileChecker, name string) error {
	switch strings.ToLower(name) {
	case "os":
		osInfo, err := collect.CollectOS(cmd, fc)
		if err != nil {
			return err
		}
		return render.OS(os.Stdout, osInfo)
	case "machine":
		machine, err := collect.CollectMachine(cmd, showSerial)
		if err != nil {
			return err
		}
		model, _ := models.Lookup(machine.ModelIdentifier)
		return render.Machine(os.Stdout, machine, model)
	case "firmware":
		fw, err := collect.CollectFirmware(cmd)
		if err != nil {
			return err
		}
		return render.Firmware(os.Stdout, fw)
	case "chip":
		chip, err := collectChip(cmd)
		if err != nil {
			return err
		}
		return render.Chip(os.Stdout, chip)
	case "cpu":
		chip, err := collectChip(cmd)
		if err != nil {
			return err
		}
		detail, err := collect.CollectCPUDetail(cmd)
		if err != nil {
			return err
		}
		return render.CPUDetail(os.Stdout, chip, detail)
	case "gpu":
		gpu, err := collect.CollectGPU(cmd)
		if err != nil {
			return err
		}
		return render.GPU(os.Stdout, gpu)
	default:
		return fmt.Errorf("unknown section %q", name)
	}
}
