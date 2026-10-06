package collect

import (
	"errors"
	"strings"
	"testing"
)

const batteryDataFixture = `{"FullChargeCapacity"=4257,"NominalChargeCapacity"=4384,"FullyCharged"=0,"AvgTimeToEmpty"=1105,"RemainingCapacity"=3241,"AbsoluteCapacity"=0,"MaxCapacity"=100,"DesignCapacity"=4563,"CurrentCapacity"=80,"BatteryPower"=18446744073709549395,"TrueRemainingCapacity"=0}`

// batteryDataFixtureFccComp is the real BatteryData string from an M5 Pro
// MacBook Pro (Mac17,8) discharging at 99%, reported by a user running a
// real release: this firmware's fuel gauge has no FullChargeCapacity key
// at all, which crashed the whole Power section before that fix. It does
// have FccComp1=8516 alongside DesignCapacity=8579 - the fallback this
// test now exercises (8516/8579*100 = 99.26 -> 99).
const batteryDataFixtureFccComp = `{"Ra03"=32,"Ra10"=38,"CellWom"=(0,0),"RaTableRaw"=(<005900170017001f002a0020002500250025002500250025004300cb017d0000>,<005d001600170020002c002500270027002700270027002c004800dd019e0000>,<0067001a001800200030002100260026002600260026002b004900da019d0000>),"Qstart"=0,"AdapterPower"=,"TrueRemainingCapacity"=0,"DailyMinSoc"=54,"Ra04"=48,"CurrentSenseMonitorStatus"=0,"Ra11"=43,"CellVoltage"=(4380,4377,4377),"PackCurrentAccumulator"=13534855,"PassedCharge"=18446744073709550257,"Flags"=16777216,"PresentDOD"=(2,2,2),"Ra05"=33,"Ra12"=73,"MiscStatus"=200,"FccComp1"=8516,"ChemID"=29767,"iMaxAndSocSmoothTable"=<0000000000000000000000000000000000000000000000000000000000000000>,"FccComp2"=8516,"PackCurrentAccumulatorCount"=3412125,"DOD0"=(2736,2736,2736),"Dod0AtQualifiedQmax"=0,"Ra06"=38,"ResScale"=0,"Ra13"=218,"FilteredCurrent"=0,"WeightedRa"=(35,37,38),"RSS"=0,"CellCurrentAccumulatorCount"=0,"Serial"="F5DHUA000HW0000VD8","DataFlashWriteCount"=3131,"DailyMaxSoc"=99,"DateOfFirstUse"=0,"Ra07"=38,"Ra14"=413,"MaxCapacity"=100,"ChemicalWeightedRa"=0,"Ra00"=103,"BatteryHealthMetric"=0,"DesignCapacity"=8579,"Ra08"=38,"BatteryState"=<00000000000000000083e40002ca020000>,"BatteryRsenseOpenCount"=0,"AlgoChemID"=29767,"MfgData"=<000000000b000100171d00000433353136033030410341544c000e0000000000>,"ManufactureDate"=55186943915059,"ISS"=488,"Ra01"=26,"Soc1Voltage"=0,"QmaxDisqualificationReason"=0,"ChargeAccum"=0,"SimRate"=0,"IdealCRate"=,"Qmax"=(9270,9277,9259),"ITMiscStatus"=0,"StateOfCharge"=99,"Ra09"=38,"GaugeFlagRaw"=128,"CellCurrentAccumulator"=(0,0),"PMUConfigured"=4424,"SystemPower"=,"CycleCount"=11,"Voltage"=13134,"LifetimeData"={"Raw"=<008c780e00011d010000000000000000008d950f0000c3e44012000000000000002a000c11220da13366288125f4d1b32e99cc46d1b3cc4600c80000a1160000>,"UpdateTime"=1790626394,"ResistanceUpdatedDisabledCount"=0,"CycleCountLastQmax"=6,"TimeAtHighSoc"=<00000000f002000006000000000000000000000000000000000000000000000087000000030000000000000000000000000000000000000000000000ce000000850000001200000000000000000000000000000000000000130500001300000001000000000000000000000000000000>,"TemperatureSamples"=41238,"TotalOperatingTime"=2577,"MaximumDischargeCurrent"=18446744073709539763,"MinimumPackVoltage"=10369,"MaximumPackVoltage"=13158,"MaximumChargeCurrent"=9716,"AverageTemperature"=200,"MinimumTemperature"=12,"RDISCnt"=0,"MaximumTemperature"=42},"Ra02"=24}`

// batteryDataFixtureFccCompFull is the same M5 Pro machine's BatteryData a
// short time later, plugged in and fully charged: FccComp1/FccComp2 moved
// to 8579, exactly matching DesignCapacity (100% health), which also
// matches `system_profiler`'s own "Maximum Capacity: 100%" for this
// machine at the time of capture.
const batteryDataFixtureFccCompFull = `{"Ra03"=32,"Ra10"=38,"CellWom"=(0,0),"RaTableRaw"=(<005900170017001f002a0020002500250025002500250025004300cb017d0000>,<005d001600170020002c002500270027002700270027002c004800dd019e0000>,<0067001a001800200030002100260026002600260026002b004900da019d0000>),"Qstart"=0,"AdapterPower"=,"TrueRemainingCapacity"=0,"DailyMinSoc"=54,"Ra04"=48,"CurrentSenseMonitorStatus"=0,"Ra11"=43,"CellVoltage"=(4349,4348,4347),"PackCurrentAccumulator"=13605696,"PassedCharge"=20,"Flags"=16777729,"PresentDOD"=(2,3,3),"Ra05"=33,"Ra12"=73,"MiscStatus"=8,"FccComp1"=8579,"ChemID"=29767,"iMaxAndSocSmoothTable"=<0000000000000000000000000000000000000000000000000000000000000000>,"FccComp2"=8579,"PackCurrentAccumulatorCount"=3412965,"DOD0"=(2736,2736,2736),"Dod0AtQualifiedQmax"=0,"Ra06"=38,"ResScale"=0,"Ra13"=218,"FilteredCurrent"=0,"WeightedRa"=(35,37,38),"RSS"=0,"CellCurrentAccumulatorCount"=0,"Serial"="F5DHUA000HW0000VD8","DataFlashWriteCount"=3131,"DailyMaxSoc"=100,"DateOfFirstUse"=0,"Ra07"=38,"Ra14"=413,"MaxCapacity"=100,"ChemicalWeightedRa"=0,"Ra00"=103,"BatteryHealthMetric"=0,"DesignCapacity"=8579,"Ra08"=38,"BatteryState"=<000000000000000000c3e4001209020000>,"BatteryRsenseOpenCount"=0,"AlgoChemID"=29767,"MfgData"=<000000000b000100171d00000433353136033030410341544c000e0000000000>,"ManufactureDate"=55186943915059,"ISS"=18446744073709551226,"Ra01"=26,"Soc1Voltage"=0,"QmaxDisqualificationReason"=0,"ChargeAccum"=0,"SimRate"=0,"IdealCRate"=,"Qmax"=(9270,9277,9259),"ITMiscStatus"=0,"StateOfCharge"=100,"Ra09"=38,"GaugeFlagRaw"=224,"CellCurrentAccumulator"=(0,0),"PMUConfigured"=0,"SystemPower"=,"CycleCount"=11,"Voltage"=13047,"LifetimeData"={"Raw"=<008c780e00011d010000000000000000008d950f0000c3e44012000000000000002a000c11220da13366288125f4d1b32e99cc46d1b3cc4600c80000a1160000>,"UpdateTime"=1790627234,"ResistanceUpdatedDisabledCount"=0,"CycleCountLastQmax"=6,"TimeAtHighSoc"=<00000000f002000006000000000000000000000000000000000000000000000087000000030000000000000000000000000000000000000000000000ce000000850000001200000000000000000000000000000000000000130500001300000001000000000000000000000000000000>,"TemperatureSamples"=41238,"TotalOperatingTime"=2577,"MaximumDischargeCurrent"=18446744073709539763,"MinimumPackVoltage"=10369,"MaximumPackVoltage"=13158,"MaximumChargeCurrent"=9716,"AverageTemperature"=200,"MinimumTemperature"=12,"RDISCnt"=0,"MaximumTemperature"=42},"Ra02"=24}`

// adapterDetailsFixture is the real AdapterDetails string from the same M5
// Pro capture (94W USB-C adapter), matching `system_profiler`'s own
// "Wattage (W): 94" exactly.
const adapterDetailsFixture = `{"IsWireless"=No,"AdapterID"=0,"AdapterVoltage"=20000,"FamilyCode"=18446744073172697098,"AdapterPowerTier"=2,"Watts"=94,"UsbHvcHvcIndex"=3,"Current"=4700,"PMUConfiguration"=4460,"UsbHvcMenu"=({"Index"=0,"MaxCurrent"=3000,"MaxVoltage"=5000},{"Index"=1,"MaxCurrent"=3000,"MaxVoltage"=9000},{"Index"=2,"MaxCurrent"=3000,"MaxVoltage"=15000},{"Index"=3,"MaxCurrent"=4700,"MaxVoltage"=20000})}`

func powerFixture(thermOutput, pmsetGOutput string) cmdRouter {
	battery := "+-o AppleSmartBattery\n    {\n" +
		`      "CurrentCapacity" = 80` + "\n" +
		`      "IsCharging" = No` + "\n" +
		`      "ExternalConnected" = No` + "\n" +
		`      "CycleCount" = 201` + "\n" +
		`      "BatteryData" = ` + batteryDataFixture + "\n" +
		"    }\n"

	return cmdRouter{outputs: map[string]string{
		"/usr/sbin/ioreg -rc AppleSmartBattery -d1": battery,
		"/usr/bin/pmset -g therm":                   thermOutput,
		"/usr/bin/pmset -g":                         pmsetGOutput,
	}}
}

const nominalTherm = "Note: No thermal warning level has been recorded\nNote: No performance warning level has been recorded\n"

const pmsetGLowPowerOff = " standby              1\n lowpowermode         0\n womp                 0\n"
const pmsetGLowPowerOn = " standby              1\n lowpowermode         1\n womp                 0\n"

func TestCollectPower(t *testing.T) {
	got, err := CollectPower(powerFixture(nominalTherm, pmsetGLowPowerOff), true)
	if err != nil {
		t.Fatalf("CollectPower() error = %v", err)
	}

	if got.Percentage != 80 || got.Charging || got.PluggedIn || got.CycleCount != 201 ||
		got.ThermalState != "nominal" || got.LowPowerMode {
		t.Errorf("CollectPower() = %+v, unexpected", got)
	}
	if got.HealthPercent == nil || *got.HealthPercent != 93 { // 4257/4563*100 = 93.29 -> 93
		t.Errorf("HealthPercent = %v, want a pointer to 93", got.HealthPercent)
	}
	if got.AdapterWatts != nil {
		t.Errorf("AdapterWatts = %v, want nil when not plugged in", *got.AdapterWatts)
	}
}

func TestCollectPowerSkipsThermalByDefault(t *testing.T) {
	// No pmset fixture entries: if CollectPower called them anyway, the
	// mock would error on the unrecognized command line.
	fixture := cmdRouter{outputs: map[string]string{
		"/usr/sbin/ioreg -rc AppleSmartBattery -d1": "+-o AppleSmartBattery\n    {\n" +
			`      "CurrentCapacity" = 80` + "\n" +
			`      "IsCharging" = No` + "\n" +
			`      "ExternalConnected" = No` + "\n" +
			`      "CycleCount" = 201` + "\n" +
			`      "BatteryData" = ` + batteryDataFixture + "\n" +
			"    }\n",
	}}

	got, err := CollectPower(fixture, false)
	if err != nil {
		t.Fatalf("CollectPower() error = %v", err)
	}
	if got.ThermalState != "" || got.LowPowerMode {
		t.Errorf("ThermalState/LowPowerMode = %q/%v, want empty/false when includeThermal is false", got.ThermalState, got.LowPowerMode)
	}
}

func TestCollectPowerLowPowerModeOn(t *testing.T) {
	got, err := CollectPower(powerFixture(nominalTherm, pmsetGLowPowerOn), true)
	if err != nil {
		t.Fatalf("CollectPower() error = %v", err)
	}
	if !got.LowPowerMode {
		t.Error("LowPowerMode = false, want true")
	}
}

func TestCollectPowerThermalElevated(t *testing.T) {
	got, err := CollectPower(powerFixture("CPU Power notify\nCPU_Speed_Limit = 40\n", pmsetGLowPowerOff), true)
	if err != nil {
		t.Fatalf("CollectPower() error = %v", err)
	}
	if got.ThermalState != "elevated" {
		t.Errorf("ThermalState = %q, want %q", got.ThermalState, "elevated")
	}
}

func TestCollectPowerNoBattery(t *testing.T) {
	fixture := cmdRouter{outputs: map[string]string{
		"/usr/sbin/ioreg -rc AppleSmartBattery -d1": "",
	}}

	_, err := CollectPower(fixture, true)
	if !errors.Is(err, ErrNoBattery) {
		t.Errorf("CollectPower() error = %v, want errors.Is(err, ErrNoBattery)", err)
	}
}

// Regression test for the M5 Pro's BatteryData schema (no
// FullChargeCapacity key): health must fall back to FccComp1 rather than
// come back unknown, since FccComp1 is a real measured value (verified to
// move with charge state - see TestCollectPowerPluggedInFullyCharged for
// the same machine at 100%) and not a static echo of DesignCapacity.
func TestCollectPowerFccCompFallback(t *testing.T) {
	battery := "+-o AppleSmartBattery\n    {\n" +
		`      "CurrentCapacity" = 99` + "\n" +
		`      "IsCharging" = No` + "\n" +
		`      "ExternalConnected" = No` + "\n" +
		`      "CycleCount" = 11` + "\n" +
		`      "BatteryData" = ` + batteryDataFixtureFccComp + "\n" +
		"    }\n"
	fixture := cmdRouter{outputs: map[string]string{
		"/usr/sbin/ioreg -rc AppleSmartBattery -d1": battery,
	}}

	got, err := CollectPower(fixture, false)
	if err != nil {
		t.Fatalf("CollectPower() error = %v, want no error for the FccComp1 schema", err)
	}
	if got.HealthPercent == nil || *got.HealthPercent != 99 { // 8516/8579*100 = 99.27 -> 99
		t.Errorf("HealthPercent = %v, want a pointer to 99", got.HealthPercent)
	}
	if got.Percentage != 99 || got.Charging || got.CycleCount != 11 {
		t.Errorf("CollectPower() = %+v, want percentage/charging/cycleCount still collected", got)
	}
}

// Regression test built from the real capture that motivated PluggedIn and
// AdapterWatts: a Mac plugged into a 94W adapter, fully charged (so
// IsCharging=No despite ExternalConnected=Yes - the case that used to
// render as "100%, discharging").
func TestCollectPowerPluggedInFullyCharged(t *testing.T) {
	battery := "+-o AppleSmartBattery\n    {\n" +
		`      "CurrentCapacity" = 100` + "\n" +
		`      "IsCharging" = No` + "\n" +
		`      "ExternalConnected" = Yes` + "\n" +
		`      "CycleCount" = 11` + "\n" +
		`      "AdapterDetails" = ` + adapterDetailsFixture + "\n" +
		`      "BatteryData" = ` + batteryDataFixtureFccCompFull + "\n" +
		"    }\n"
	fixture := cmdRouter{outputs: map[string]string{
		"/usr/sbin/ioreg -rc AppleSmartBattery -d1": battery,
	}}

	got, err := CollectPower(fixture, false)
	if err != nil {
		t.Fatalf("CollectPower() error = %v", err)
	}
	if !got.PluggedIn {
		t.Error("PluggedIn = false, want true")
	}
	if got.Charging {
		t.Error("Charging = true, want false (fully charged, not actively charging)")
	}
	if got.AdapterWatts == nil || *got.AdapterWatts != 94 {
		t.Errorf("AdapterWatts = %v, want a pointer to 94", got.AdapterWatts)
	}
	if got.HealthPercent == nil || *got.HealthPercent != 100 { // 8579/8579*100 = 100, matches system_profiler's "Maximum Capacity: 100%"
		t.Errorf("HealthPercent = %v, want a pointer to 100", got.HealthPercent)
	}
}

// When a battery-gauge schema exposes neither FullChargeCapacity nor
// FccComp1, health must still come back unknown rather than error.
func TestCollectPowerHealthTrulyUnknown(t *testing.T) {
	battery := "+-o AppleSmartBattery\n    {\n" +
		`      "CurrentCapacity" = 50` + "\n" +
		`      "IsCharging" = No` + "\n" +
		`      "ExternalConnected" = No` + "\n" +
		`      "CycleCount" = 5` + "\n" +
		`      "BatteryData" = {"DesignCapacity"=8579,"MaxCapacity"=100}` + "\n" +
		"    }\n"
	fixture := cmdRouter{outputs: map[string]string{
		"/usr/sbin/ioreg -rc AppleSmartBattery -d1": battery,
	}}

	got, err := CollectPower(fixture, false)
	if err != nil {
		t.Fatalf("CollectPower() error = %v, want no error when health data is simply absent", err)
	}
	if got.HealthPercent != nil {
		t.Errorf("HealthPercent = %v, want nil (unknown)", *got.HealthPercent)
	}
}

func TestParseBatteryHealth(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    int // -1 means unknown (nil)
		wantErr bool
	}{
		{"FullChargeCapacity", `{"FullChargeCapacity"=4257,"DesignCapacity"=4563}`, 93, false},
		// Real values from an M5 Pro: FccComp1 above DesignCapacity on a
		// new battery. System Information reports this as 100%.
		{"above design capped", `{"FccComp1"=8651,"DesignCapacity"=8579}`, 100, false},
		// Without the cap this would overflow uint8 and wrap.
		{"far above design capped", `{"FccComp1"=30000,"DesignCapacity"=8579}`, 100, false},
		{"no full charge key", `{"DesignCapacity"=8579}`, -1, false},
		{"zero design", `{"FccComp1"=8651,"DesignCapacity"=0}`, -1, true},
		{"no design", `{"FccComp1"=8651,"Serial"="F5DHUA000HW0000VD8"}`, -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBatteryHealth(tt.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseBatteryHealth() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && strings.Contains(err.Error(), "F5DHUA000HW0000VD8") {
				t.Errorf("error %q leaks the battery serial", err)
			}
			gotN := -1
			if got != nil {
				gotN = int(*got)
			}
			if gotN != tt.want {
				t.Errorf("parseBatteryHealth() = %d, want %d", gotN, tt.want)
			}
		})
	}
}
