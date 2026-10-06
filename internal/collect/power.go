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
type Power struct {
	Percentage uint8 `json:"percentage"`
	Charging   bool  `json:"charging"`
	// PluggedIn is true whenever a power adapter is connected, regardless
	// of whether the battery is actively charging (e.g. already full).
	// Charging is only true while current is actually flowing into the
	// battery - a fully-charged, plugged-in Mac has PluggedIn=true,
	// Charging=false, which is a real state, not "discharging".
	PluggedIn bool `json:"pluggedIn"`
	// AdapterWatts is the connected adapter's rated wattage, from
	// AdapterDetails.Watts. Nil when no adapter is connected or that
	// sub-property isn't present.
	AdapterWatts  *uint16 `json:"adapterWatts,omitempty"`
	CycleCount    uint32  `json:"cycleCount"`
	HealthPercent *uint8  `json:"healthPercent,omitempty"` // see parseBatteryHealth
	ThermalState  string  `json:"thermalState,omitempty"`  // "nominal" or "elevated" - see parseThermalState; "" when not checked
	LowPowerMode  bool    `json:"lowPowerMode"`
}

var (
	fullChargeCapacityRE = regexp.MustCompile(`"FullChargeCapacity"=(\d+)`)
	// fccCompRE matches FccComp1 ("Full Charge Capacity, Compensated"),
	// the bq40z651 gauge's newer-firmware analog to FullChargeCapacity -
	// see the comment on parseBatteryHealth.
	fccCompRE        = regexp.MustCompile(`"FccComp1"=(\d+)`)
	designCapacityRE = regexp.MustCompile(`"DesignCapacity"=(\d+)`)
	adapterWattsRE   = regexp.MustCompile(`"Watts"=(\d+)`)
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
	pluggedIn, err := source.BoolProperty(props, "ExternalConnected")
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
		return Power{}, fmt.Errorf("BatteryData: %w", err)
	}

	var adapterWatts *uint16
	if pluggedIn {
		if adapterDetails, ok := props["AdapterDetails"]; ok {
			adapterWatts = parseAdapterWatts(adapterDetails)
		}
	}

	power := Power{
		Percentage:    uint8(percentage),
		Charging:      charging,
		PluggedIn:     pluggedIn,
		AdapterWatts:  adapterWatts,
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

// parseBatteryHealth extracts a full-charge capacity and DesignCapacity
// from AppleSmartBattery's nested BatteryData dict and returns their ratio
// as a percentage - the standard "battery health" computation (matches
// what tools like coconutBattery show), since AppleSmartBattery's own
// top-level MaxCapacity is pinned at 100 rather than reflecting
// degradation.
//
// Not every machine's battery-gauge firmware uses the same key for this.
// An M5 Pro MacBook Pro (bq40z651 gauge, newer firmware than the M2 this
// was first written against) was found by a user running a real release to
// report BatteryData with no FullChargeCapacity key at all. It does carry
// FccComp1 ("Full Charge Capacity, Compensated" - a standard TI bq-gauge
// name), which tracked the battery's charge state across two captures
// (8516 at 99% charge, 8579 - equal to DesignCapacity - at 100%/fully
// charged) and, at 100%, computed a health percentage that matched
// `system_profiler`'s own "Maximum Capacity: 100%" exactly. That's one
// data point at 100% health, not a confirmed formula across degraded
// batteries, but it's a real measured value moving with charge state, not
// a static echo of DesignCapacity - a meaningfully different case from
// Qmax and BatteryHealthMetric, whose semantics are still unverified and
// so still aren't used here.
//
// If neither key is present, health is unknown: returns (nil, nil), not an
// error, so the rest of the Power section still collects normally.
//
// The result is capped at 100. A new battery can report a full-charge
// capacity slightly above DesignCapacity (FccComp1=8651 against 8579 on the
// same M5 Pro), and System Information reports that as 100%.
//
// Errors name the missing key without echoing batteryData, since the dict
// carries the battery's serial number.
func parseBatteryHealth(batteryData string) (*uint8, error) {
	full := fullChargeCapacityRE.FindStringSubmatch(batteryData)
	if full == nil {
		full = fccCompRE.FindStringSubmatch(batteryData)
	}
	if full == nil {
		return nil, nil
	}
	design := designCapacityRE.FindStringSubmatch(batteryData)
	if design == nil {
		return nil, fmt.Errorf("DesignCapacity not found in BatteryData")
	}
	fullN, err := strconv.ParseFloat(full[1], 64)
	if err != nil {
		return nil, fmt.Errorf("full charge capacity: %w", err)
	}
	designN, err := strconv.ParseFloat(design[1], 64)
	if err != nil {
		return nil, fmt.Errorf("DesignCapacity: %w", err)
	}
	if designN == 0 {
		return nil, fmt.Errorf("DesignCapacity is 0")
	}
	health := uint8(min(fullN/designN*100, 100))
	return &health, nil
}

// parseAdapterWatts extracts Watts from a raw AdapterDetails dict value.
// Returns nil (not an error) when the sub-property isn't present, since
// the exact AdapterDetails shape isn't guaranteed across adapter types.
func parseAdapterWatts(adapterDetails string) *uint16 {
	m := adapterWattsRE.FindStringSubmatch(adapterDetails)
	if m == nil {
		return nil
	}
	n, err := strconv.ParseUint(m[1], 10, 16)
	if err != nil {
		return nil
	}
	watts := uint16(n)
	return &watts
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
