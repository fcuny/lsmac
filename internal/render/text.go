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

// labelWidth is the column every top-level field label is padded to, so
// values line up regardless of label length. It matches "Low Power Mode",
// the longest label currently in use - computing this from hand-counted
// spaces per format string is exactly how misalignment bugs like it crept
// in before, so every field below goes through the field() helper instead.
const labelWidth = 14

// field writes one "Label   value" line, with label padded to labelWidth.
func field(w io.Writer, label, value string) error {
	_, err := fmt.Fprintf(w, "%-*s %s\n", labelWidth, label, value)
	return err
}

// Chip writes the chip identity line, plus process node when known. name
// falls back to the raw chip ID when the chip is not in the internal/chips
// table. Memory bandwidth is part of the Memory section, not here.
func Chip(w io.Writer, chip chips.Chip) error {
	name := chip.MarketingName
	if name == "" {
		name = chip.ID
	}
	if err := field(w, "Chip", fmt.Sprintf("%s (%s)", name, chip.ID)); err != nil {
		return err
	}
	if chip.ProcessNode != "" {
		if err := field(w, "Process", chip.ProcessNode); err != nil {
			return err
		}
	}
	return nil
}

// CPU writes the CPU core-count summary line.
func CPU(w io.Writer, cpu collect.CPU) error {
	return field(w, "CPU", fmt.Sprintf("%d cores: %dP + %dE", cpu.TotalCores, cpu.PerformanceCores, cpu.EfficiencyCores))
}

// GPU writes the GPU summary line.
func GPU(w io.Writer, gpu collect.GPU) error {
	return field(w, "GPU", fmt.Sprintf("%d cores", gpu.CoreCount))
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
	if err := field(w, "Machine", fmt.Sprintf("%s (%s)", name, machine.ModelIdentifier)); err != nil {
		return err
	}
	if machine.SKU != "" {
		if err := field(w, "SKU", machine.SKU); err != nil {
			return err
		}
	}
	if machine.Serial != "" {
		if err := field(w, "Serial", machine.Serial); err != nil {
			return err
		}
	}
	if machine.HardwareUUID != "" {
		if err := field(w, "UUID", machine.HardwareUUID); err != nil {
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
	if err := field(w, "OS", fmt.Sprintf("%s (%s)", label, osInfo.Build)); err != nil {
		return err
	}
	if err := field(w, "Darwin", osInfo.DarwinVersion); err != nil {
		return err
	}
	if err := field(w, "Uptime", formatUptime(osInfo.Uptime)); err != nil {
		return err
	}
	rosetta := "not installed"
	if osInfo.RosettaInstalled {
		rosetta = "installed"
	}
	return field(w, "Rosetta", rosetta)
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
	if err := field(w, "Firmware", fw.Version); err != nil {
		return err
	}
	secureBoot := "disabled"
	if fw.SecureBoot {
		secureBoot = "enabled"
	}
	if err := field(w, "Secure Boot", secureBoot); err != nil {
		return err
	}
	sip := "disabled"
	if fw.SIPEnabled {
		sip = "enabled"
	}
	return field(w, "SIP", sip)
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

// Memory writes used/total, RAM type, bandwidth (from the chip table), and
// memory pressure. Type and bandwidth lines are omitted when unknown.
func Memory(w io.Writer, mem collect.Memory, bandwidthGBs uint32) error {
	if err := field(w, "Memory", fmt.Sprintf("%s / %s", formatGiB(mem.UsedBytes), formatGiB(mem.TotalBytes))); err != nil {
		return err
	}
	if mem.Type != "" {
		if err := field(w, "Type", mem.Type); err != nil {
			return err
		}
	}
	if bandwidthGBs > 0 {
		if err := field(w, "Bandwidth", fmt.Sprintf("%d GB/s", bandwidthGBs)); err != nil {
			return err
		}
	}
	return field(w, "Pressure", mem.Pressure)
}

// formatGiB renders a byte count in GiB, dropping the decimal point when
// it's an exact multiple (which totals like hw.memsize always are).
func formatGiB(n uint64) string {
	const gib = 1 << 30
	if n%gib == 0 {
		return fmt.Sprintf("%d GiB", n/gib)
	}
	return fmt.Sprintf("%.1f GiB", float64(n)/gib)
}

// Storage writes the internal disk model, used/total capacity, and
// FileVault status.
func Storage(w io.Writer, storage collect.Storage) error {
	if err := field(w, "Disk", storage.Model); err != nil {
		return err
	}
	if err := field(w, "Capacity", fmt.Sprintf("%s / %s", formatGiB(storage.UsedBytes), formatGiB(storage.TotalBytes))); err != nil {
		return err
	}
	if storage.FileVaultOn != nil {
		fileVault := "off"
		if *storage.FileVaultOn {
			fileVault = "on"
		}
		return field(w, "FileVault", fileVault)
	}
	return nil
}

// Power writes battery charge/charging state, cycle count, health,
// thermal state, and Low Power Mode. Call only when collect.CollectPower
// didn't return collect.ErrNoBattery - a Mac without a battery has no
// Power section at all, never one printed with zeroed fields.
func Power(w io.Writer, power collect.Power) error {
	state := "discharging"
	if power.Charging {
		state = "charging"
	}
	if err := field(w, "Battery", fmt.Sprintf("%d%%, %s", power.Percentage, state)); err != nil {
		return err
	}
	if err := field(w, "Cycles", fmt.Sprintf("%d", power.CycleCount)); err != nil {
		return err
	}
	if err := field(w, "Health", fmt.Sprintf("%d%%", power.HealthPercent)); err != nil {
		return err
	}
	if power.ThermalState == "" {
		return nil
	}
	if err := field(w, "Thermal", power.ThermalState); err != nil {
		return err
	}
	lowPowerMode := "off"
	if power.LowPowerMode {
		lowPowerMode = "on"
	}
	return field(w, "Low Power Mode", lowPowerMode)
}

// IO writes one line per network interface. This is the only I/O fact
// currently collected - see the comment on collect.IO for what's deferred
// and why.
func IO(w io.Writer, ioInfo collect.IO) error {
	for _, iface := range ioInfo.Interfaces {
		state := "down"
		if iface.IsUp {
			state = "up"
		}
		if iface.HardwareAddr != "" {
			if _, err := fmt.Fprintf(w, "%-10s %-4s %s\n", iface.Name, state, iface.HardwareAddr); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintf(w, "%-10s %s\n", iface.Name, state); err != nil {
			return err
		}
	}
	return nil
}
