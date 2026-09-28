package collect

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"fcuny.net/lsmac/internal/source"
)

// ErrNoBattery means the machine has no battery (e.g. a Mac mini or Mac
// Studio). Callers should omit the Power section entirely rather than
// treat this as a failure - see CLAUDE.md's rule that a section which
// doesn't apply is left out, never printed as "N/A".
var ErrNoBattery = errors.New("no battery present")

// Power holds the facts collected for the Power section.
//
// ThermalState is empty and LowPowerMode is false unless includeThermal was
// passed to CollectPower: both come from `pmset -g therm` and `pmset -g`,
// two extra subprocess calls (~19ms combined) for facts that are "nominal"
// and "off" the overwhelming majority of the time - the default view skips
// them, matching Storage's FileVaultOn deferral for the same reason.
//
// Adapter wattage is not collected: this machine wasn't connected to power
// while developing this collector, so the relevant AppleSmartBattery
// fields (AdapterDetails/PowerDistribution) couldn't be verified against
// real data - left out rather than guessed at from field names alone.
type Power struct {
	Percentage    uint8  `json:"percentage"`
	Charging      bool   `json:"charging"`
	CycleCount    uint32 `json:"cycleCount"`
	HealthPercent uint8  `json:"healthPercent"`          // FullChargeCapacity / DesignCapacity, both from BatteryData
	ThermalState  string `json:"thermalState,omitempty"` // "nominal" or "elevated" - see parseThermalState; "" when not checked
	LowPowerMode  bool   `json:"lowPowerMode"`
}

var (
	fullChargeCapacityRE = regexp.MustCompile(`"FullChargeCapacity"=(\d+)`)
	designCapacityRE     = regexp.MustCompile(`"DesignCapacity"=(\d+)`)
)

const pmsetPath = "/usr/bin/pmset"

// CollectPower reads charge/cycle-count/charging state from
// AppleSmartBattery and battery health from its nested BatteryData dict.
// Thermal state (`pmset -g therm`) and Low Power Mode (`pmset -g`) are only
// read when includeThermal is true - see the comment on Power.
func CollectPower(cmd source.SystemCommand, includeThermal bool) (Power, error) {
	props, err := source.IORegProperties(cmd, "AppleSmartBattery")
	if err != nil {
		if errors.Is(err, source.ErrNotFound) {
			return Power{}, ErrNoBattery
		}
		return Power{}, err
	}

	percentage, err := source.IntProperty(props, "CurrentCapacity")
	if err != nil {
		return Power{}, err
	}
	charging, err := source.BoolProperty(props, "IsCharging")
	if err != nil {
		return Power{}, err
	}
	cycleCount, err := source.IntProperty(props, "CycleCount")
	if err != nil {
		return Power{}, err
	}

	batteryData, ok := props["BatteryData"]
	if !ok {
		return Power{}, fmt.Errorf("property %q not found", "BatteryData")
	}
	health, err := parseBatteryHealth(batteryData)
	if err != nil {
		return Power{}, err
	}

	power := Power{
		Percentage:    uint8(percentage),
		Charging:      charging,
		CycleCount:    uint32(cycleCount),
		HealthPercent: health,
	}

	if includeThermal {
		therm, err := source.Run(cmd, pmsetPath, "-g", "therm")
		if err != nil {
			return Power{}, fmt.Errorf("pmset -g therm: %w", err)
		}
		power.ThermalState = parseThermalState(therm)

		pmsetG, err := source.Run(cmd, pmsetPath, "-g")
		if err != nil {
			return Power{}, fmt.Errorf("pmset -g: %w", err)
		}
		lowPowerMode, err := parseLowPowerMode(pmsetG)
		if err != nil {
			return Power{}, err
		}
		power.LowPowerMode = lowPowerMode
	}

	return power, nil
}

// parseBatteryHealth extracts FullChargeCapacity and DesignCapacity from
// AppleSmartBattery's nested BatteryData dict and returns their ratio as a
// percentage - the standard "battery health" computation (matches what
// tools like coconutBattery show), since AppleSmartBattery's own top-level
// MaxCapacity is pinned at 100 rather than reflecting degradation on this
// machine's firmware version.
func parseBatteryHealth(batteryData string) (uint8, error) {
	full := fullChargeCapacityRE.FindStringSubmatch(batteryData)
	design := designCapacityRE.FindStringSubmatch(batteryData)
	if full == nil || design == nil {
		return 0, fmt.Errorf("BatteryData: FullChargeCapacity/DesignCapacity not found in %q", batteryData)
	}
	fullN, err := strconv.ParseFloat(full[1], 64)
	if err != nil {
		return 0, fmt.Errorf("BatteryData: FullChargeCapacity: %w", err)
	}
	designN, err := strconv.ParseFloat(design[1], 64)
	if err != nil || designN == 0 {
		return 0, fmt.Errorf("BatteryData: DesignCapacity: %w", err)
	}
	return uint8(fullN / designN * 100), nil
}

// parseThermalState reads `pmset -g therm`'s output. When macOS has never
// recorded a thermal or performance warning, pmset says so in plain text;
// that's the only case verified against a real machine, so anything else
// is reported generically as "elevated" rather than mapped to a specific
// named level that hasn't been observed.
func parseThermalState(therm string) string {
	if strings.Contains(therm, "No thermal warning level has been recorded") &&
		strings.Contains(therm, "No performance warning level has been recorded") {
		return "nominal"
	}
	return "elevated"
}

// parseLowPowerMode reads the "lowpowermode" line from `pmset -g`'s
// "Currently in use" settings block, e.g. " lowpowermode         0".
func parseLowPowerMode(pmsetG string) (bool, error) {
	for line := range strings.SplitSeq(pmsetG, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "lowpowermode" {
			return fields[1] != "0", nil
		}
	}
	return false, fmt.Errorf("pmset -g: lowpowermode not found in output")
}
