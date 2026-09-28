// Command lsmac prints hardware and system information about the Apple
// Silicon Mac it runs on.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"fcuny.net/lsmac/internal/chips"
	"fcuny.net/lsmac/internal/collect"
	"fcuny.net/lsmac/internal/models"
	"fcuny.net/lsmac/internal/render"
	"fcuny.net/lsmac/internal/source"
)

// sectionFlag collects repeated `--section NAME` occurrences.
type sectionFlag []string

func (s *sectionFlag) String() string { return strings.Join(*s, ",") }

func (s *sectionFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	var sections sectionFlag
	flag.Var(&sections, "section", "print the detailed view of one section (repeatable): "+strings.Join(validSectionNames, ", "))
	jsonOutput := flag.Bool("json", false, "print structured JSON instead of text")
	showSerial := flag.Bool("show-serial", false, "include serial number and hardware UUID (hidden by default)")
	flag.Parse()

	cmd := source.NewCachingCommand(source.RealCommand{})
	fc := source.RealFileChecker{}

	if *jsonOutput {
		runJSON(cmd, fc, sections, *showSerial)
		return
	}

	if len(sections) > 0 {
		ok := true
		for _, name := range sections {
			if err := runSection(cmd, fc, name, *showSerial); err != nil {
				fmt.Fprintf(os.Stderr, "lsmac: %s: %v\n", name, err)
				ok = false
			}
		}
		if !ok {
			os.Exit(1)
		}
		return
	}

	runDefault(cmd, fc, *showSerial)
}

func runDefault(cmd source.SystemCommand, fc source.FileChecker, showSerial bool) {
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

	chip, chipErr := collectChip(cmd)
	if chipErr != nil {
		fmt.Fprintf(os.Stderr, "lsmac: chip: %v\n", chipErr)
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

	if mem, err := collect.CollectMemory(cmd); err != nil {
		fmt.Fprintf(os.Stderr, "lsmac: memory: %v\n", err)
	} else {
		var bandwidth uint32
		if chipErr == nil {
			bandwidth = chip.MemoryBandwidthGBs
		}
		if err := render.Memory(os.Stdout, mem, bandwidth); err != nil {
			fmt.Fprintf(os.Stderr, "lsmac: %v\n", err)
		}
		ok = true
	}

	if storage, err := collect.CollectStorage(cmd, source.RealFilesystemStats{}, false); err != nil {
		fmt.Fprintf(os.Stderr, "lsmac: storage: %v\n", err)
	} else {
		if err := render.Storage(os.Stdout, storage); err != nil {
			fmt.Fprintf(os.Stderr, "lsmac: %v\n", err)
		}
		ok = true
	}

	if power, err := collect.CollectPower(cmd, false); err != nil {
		if !errors.Is(err, collect.ErrNoBattery) {
			fmt.Fprintf(os.Stderr, "lsmac: power: %v\n", err)
		}
	} else {
		if err := render.Power(os.Stdout, power); err != nil {
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
func runSection(cmd source.SystemCommand, fc source.FileChecker, name string, showSerial bool) error {
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
	case "memory":
		mem, err := collect.CollectMemory(cmd)
		if err != nil {
			return err
		}
		var bandwidth uint32
		if chip, err := collectChip(cmd); err == nil {
			bandwidth = chip.MemoryBandwidthGBs
		}
		return render.Memory(os.Stdout, mem, bandwidth)
	case "storage":
		storage, err := collect.CollectStorage(cmd, source.RealFilesystemStats{}, true)
		if err != nil {
			return err
		}
		return render.Storage(os.Stdout, storage)
	case "power":
		power, err := collect.CollectPower(cmd, true)
		if err != nil {
			return err
		}
		return render.Power(os.Stdout, power)
	case "io":
		ioInfo, err := collect.CollectIO(source.RealNetworkInterfaces{})
		if err != nil {
			return err
		}
		return render.IO(os.Stdout, ioInfo)
	default:
		return fmt.Errorf("unknown section %q", name)
	}
}

// validSectionNames are the names accepted by --section, in both text and
// JSON mode.
var validSectionNames = []string{
	"os", "machine", "firmware", "chip", "cpu", "gpu", "memory", "storage", "power", "io",
}

// runJSON collects the requested sections (all applicable ones if sections
// is empty) and prints them as a single JSON object, one key per section.
// Unlike the text paths, JSON mode always fetches full detail (FileVault
// status, thermal state, the detailed CPU breakdown): --json is an
// explicit request for complete structured data, not the fast-glance
// default view, so it isn't held to the same latency budget.
func runJSON(cmd source.SystemCommand, fc source.FileChecker, sections sectionFlag, showSerial bool) {
	all := len(sections) == 0
	want := func(name string) bool {
		if all {
			return true
		}
		for _, n := range sections {
			if strings.EqualFold(n, name) {
				return true
			}
		}
		return false
	}

	unknown := false
	for _, name := range sections {
		if !slices.ContainsFunc(validSectionNames, func(v string) bool { return strings.EqualFold(v, name) }) {
			fmt.Fprintf(os.Stderr, "lsmac: unknown section %q\n", name)
			unknown = true
		}
	}
	if unknown {
		os.Exit(1)
	}

	result := make(map[string]any)
	ok := false

	if want("os") {
		if v, err := collect.CollectOS(cmd, fc); err != nil {
			fmt.Fprintf(os.Stderr, "lsmac: os: %v\n", err)
		} else {
			result["os"], ok = v, true
		}
	}

	if want("machine") {
		if v, err := collect.CollectMachine(cmd, showSerial); err != nil {
			fmt.Fprintf(os.Stderr, "lsmac: machine: %v\n", err)
		} else {
			model, _ := models.Lookup(v.ModelIdentifier)
			result["machine"], ok = render.MachineJSON(v, model), true
		}
	}

	if want("firmware") {
		if v, err := collect.CollectFirmware(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "lsmac: firmware: %v\n", err)
		} else {
			result["firmware"], ok = v, true
		}
	}

	chip, chipErr := collectChip(cmd)
	if want("chip") {
		if chipErr != nil {
			fmt.Fprintf(os.Stderr, "lsmac: chip: %v\n", chipErr)
		} else {
			result["chip"], ok = chip, true
		}
	}

	if want("cpu") {
		cpu, cpuErr := collect.CollectCPU(cmd)
		detail, detailErr := collect.CollectCPUDetail(cmd)
		switch {
		case cpuErr != nil:
			fmt.Fprintf(os.Stderr, "lsmac: cpu: %v\n", cpuErr)
		case detailErr != nil:
			fmt.Fprintf(os.Stderr, "lsmac: cpu: %v\n", detailErr)
		default:
			result["cpu"], ok = render.CPUJSON(cpu, detail), true
		}
	}

	if want("gpu") {
		if v, err := collect.CollectGPU(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "lsmac: gpu: %v\n", err)
		} else {
			result["gpu"], ok = v, true
		}
	}

	if want("memory") {
		if v, err := collect.CollectMemory(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "lsmac: memory: %v\n", err)
		} else {
			var bandwidth uint32
			if chipErr == nil {
				bandwidth = chip.MemoryBandwidthGBs
			}
			result["memory"], ok = render.MemoryJSON(v, bandwidth), true
		}
	}

	if want("storage") {
		if v, err := collect.CollectStorage(cmd, source.RealFilesystemStats{}, true); err != nil {
			fmt.Fprintf(os.Stderr, "lsmac: storage: %v\n", err)
		} else {
			result["storage"], ok = v, true
		}
	}

	if want("power") {
		if v, err := collect.CollectPower(cmd, true); err != nil {
			if !errors.Is(err, collect.ErrNoBattery) {
				fmt.Fprintf(os.Stderr, "lsmac: power: %v\n", err)
			}
		} else {
			result["power"], ok = v, true
		}
	}

	if want("io") {
		if v, err := collect.CollectIO(source.RealNetworkInterfaces{}); err != nil {
			fmt.Fprintf(os.Stderr, "lsmac: io: %v\n", err)
		} else {
			result["io"], ok = v, true
		}
	}

	if err := render.JSON(os.Stdout, result); err != nil {
		fmt.Fprintf(os.Stderr, "lsmac: %v\n", err)
		os.Exit(1)
	}
	if !ok {
		os.Exit(1)
	}
}
