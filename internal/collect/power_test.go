package collect

import (
	"errors"
	"testing"
)

const batteryDataFixture = `{"FullChargeCapacity"=4257,"NominalChargeCapacity"=4384,"FullyCharged"=0,"AvgTimeToEmpty"=1105,"RemainingCapacity"=3241,"AbsoluteCapacity"=0,"MaxCapacity"=100,"DesignCapacity"=4563,"CurrentCapacity"=80,"BatteryPower"=18446744073709549395,"TrueRemainingCapacity"=0}`

func powerFixture(thermOutput, pmsetGOutput string) cmdRouter {
	battery := "+-o AppleSmartBattery\n    {\n" +
		`      "CurrentCapacity" = 80` + "\n" +
		`      "IsCharging" = No` + "\n" +
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

	want := Power{
		Percentage:    80,
		Charging:      false,
		CycleCount:    201,
		HealthPercent: 93, // 4257/4563*100 = 93.29 -> 93
		ThermalState:  "nominal",
		LowPowerMode:  false,
	}
	if got != want {
		t.Errorf("CollectPower() = %+v, want %+v", got, want)
	}
}

func TestCollectPowerSkipsThermalByDefault(t *testing.T) {
	// No pmset fixture entries: if CollectPower called them anyway, the
	// mock would error on the unrecognized command line.
	fixture := cmdRouter{outputs: map[string]string{
		"/usr/sbin/ioreg -rc AppleSmartBattery -d1": "+-o AppleSmartBattery\n    {\n" +
			`      "CurrentCapacity" = 80` + "\n" +
			`      "IsCharging" = No` + "\n" +
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
