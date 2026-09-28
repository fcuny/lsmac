// Package render turns collected section data into output: plain text for
// the terminal, or JSON. Renderers never call out to the system themselves;
// they only format what internal/collect already gathered.
package render

import (
	"fmt"
	"io"
	"strings"
	"time"

	"fcuny.net/lsmac/internal/chips"
	"fcuny.net/lsmac/internal/collect"
	"fcuny.net/lsmac/internal/models"
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

// Machine writes the machine identity line, plus SKU, serial number, and
// hardware UUID when known. name falls back to the raw model identifier
// when it's not in the internal/models table. Serial and HardwareUUID are
// only printed when collect.Machine actually carries them (i.e. --show-serial
// was passed at collection time); this function does no gating of its own.
func Machine(w io.Writer, machine collect.Machine, model models.Model) error {
	name := model.MarketingName
	if name == "" {
		name = machine.ModelIdentifier
	}
	if _, err := fmt.Fprintf(w, "Machine    %s (%s)\n", name, machine.ModelIdentifier); err != nil {
		return err
	}
	if machine.SKU != "" {
		if _, err := fmt.Fprintf(w, "SKU        %s\n", machine.SKU); err != nil {
			return err
		}
	}
	if machine.Serial != "" {
		if _, err := fmt.Fprintf(w, "Serial     %s\n", machine.Serial); err != nil {
			return err
		}
	}
	if machine.HardwareUUID != "" {
		if _, err := fmt.Fprintf(w, "UUID       %s\n", machine.HardwareUUID); err != nil {
			return err
		}
	}
	return nil
}

// OS writes the OS identity, Darwin kernel version, uptime, and Rosetta
// installed status.
func OS(w io.Writer, osInfo collect.OS) error {
	label := fmt.Sprintf("%s %s", osInfo.ProductName, osInfo.ProductVersion)
	if osInfo.VersionName != "" {
		label += " " + osInfo.VersionName
	}
	if _, err := fmt.Fprintf(w, "OS         %s (%s)\n", label, osInfo.Build); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Darwin     %s\n", osInfo.DarwinVersion); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Uptime     %s\n", formatUptime(osInfo.Uptime)); err != nil {
		return err
	}
	rosetta := "not installed"
	if osInfo.RosettaInstalled {
		rosetta = "installed"
	}
	if _, err := fmt.Fprintf(w, "Rosetta    %s\n", rosetta); err != nil {
		return err
	}
	return nil
}

func formatUptime(d time.Duration) string {
	switch {
	case d >= 24*time.Hour:
		days := int(d.Hours()) / 24
		hours := int(d.Hours()) % 24
		return fmt.Sprintf("%d %s, %d %s", days, plural(days, "day"), hours, plural(hours, "hour"))
	case d >= time.Hour:
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%d %s, %d %s", hours, plural(hours, "hour"), minutes, plural(minutes, "minute"))
	default:
		minutes := int(d.Minutes())
		return fmt.Sprintf("%d %s", minutes, plural(minutes, "minute"))
	}
}

func plural(n int, unit string) string {
	if n == 1 {
		return unit
	}
	return unit + "s"
}

// Firmware writes the system firmware version, Secure Boot, and SIP status.
func Firmware(w io.Writer, fw collect.Firmware) error {
	if _, err := fmt.Fprintf(w, "Firmware   %s\n", fw.Version); err != nil {
		return err
	}
	secureBoot := "disabled"
	if fw.SecureBoot {
		secureBoot = "enabled"
	}
	if _, err := fmt.Fprintf(w, "Secure Boot  %s\n", secureBoot); err != nil {
		return err
	}
	sip := "disabled"
	if fw.SIPEnabled {
		sip = "enabled"
	}
	if _, err := fmt.Fprintf(w, "SIP        %s\n", sip); err != nil {
		return err
	}
	return nil
}

// CPUDetail writes the detailed `--section cpu` view: chip identity,
// family, page size, one block per performance-level cluster with its core
// counts and cache sizes, and the active ARM architecture features.
//
// Maximum frequency per cluster is not shown - see the comment on
// collect.CPUDetail for why.
func CPUDetail(w io.Writer, chip chips.Chip, detail collect.CPUDetail) error {
	name := chip.MarketingName
	if name == "" {
		name = chip.ID
	}
	if _, err := fmt.Fprintf(w, "CPU\n  Chip              %s (%s)\n  Family            %s\n  Page size         %s\n",
		name, chip.ID, detail.Family, formatBytes(detail.PageSize)); err != nil {
		return err
	}

	for _, c := range detail.Clusters {
		if _, err := fmt.Fprintf(w, "\n  %-17s %d %s  ·  %d %s\n",
			c.Name, c.PhysicalCores, plural(int(c.PhysicalCores), "core"),
			c.Clusters, plural(int(c.Clusters), "cluster")); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "    L1i / L1d       %s / %s per core\n",
			formatBytes(c.L1ICacheSize), formatBytes(c.L1DCacheSize)); err != nil {
			return err
		}
		l2 := fmt.Sprintf("%s per cluster", formatBytes(c.L2CacheSize))
		if c.CoresPerCluster > 0 {
			l2 += fmt.Sprintf(" (%d cores per L2)", c.CoresPerCluster)
		}
		if _, err := fmt.Fprintf(w, "    L2              %s\n", l2); err != nil {
			return err
		}
	}

	if len(detail.Features) > 0 {
		if _, err := fmt.Fprintf(w, "\n  Features          %s\n", wrapFeatures(detail.Features)); err != nil {
			return err
		}
	}
	return nil
}

// formatBytes renders a byte count as whole KiB or MiB. Every size lsmac
// deals with (page sizes, cache sizes) is an exact multiple of one or the
// other, so plain integer division never loses precision here.
func formatBytes(n uint32) string {
	const ki, mi = 1024, 1024 * 1024
	if n >= mi {
		return fmt.Sprintf("%d MiB", n/mi)
	}
	return fmt.Sprintf("%d KiB", n/ki)
}

// featuresWrapWidth roughly matches the indentation used for continuation
// lines under the "Features" label.
const featuresWrapWidth = 60

// wrapFeatures joins feature names with spaces, wrapping continuation
// lines to align under the "Features" label.
func wrapFeatures(features []string) string {
	const indent = "                    "
	var b strings.Builder
	lineLen := 0
	for i, f := range features {
		if i > 0 {
			if lineLen+1+len(f) > featuresWrapWidth {
				b.WriteString("\n")
				b.WriteString(indent)
				lineLen = 0
			} else {
				b.WriteString(" ")
				lineLen++
			}
		}
		b.WriteString(f)
		lineLen += len(f)
	}
	return b.String()
}
