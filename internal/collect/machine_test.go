package collect

import (
	"os"
	"testing"
)

func platformExpertFixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("../../testdata/Mac14,2/ioreg-ioplatformexpertdevice.txt")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	return string(data)
}

func machineFixture(t *testing.T) cmdRouter {
	t.Helper()
	return cmdRouter{outputs: map[string]string{
		"/usr/sbin/sysctl -n hw.model":                   "Mac14,2",
		"/usr/sbin/ioreg -rc IOPlatformExpertDevice -d1": platformExpertFixture(t),
	}}
}

func TestCollectMachine(t *testing.T) {
	got, err := CollectMachine(machineFixture(t), false)
	if err != nil {
		t.Fatalf("CollectMachine() error = %v", err)
	}

	want := Machine{ModelIdentifier: "Mac14,2", SKU: "MN703LL/A"}
	if got != want {
		t.Errorf("CollectMachine() = %+v, want %+v", got, want)
	}
}

func TestCollectMachineSerialHiddenByDefault(t *testing.T) {
	got, err := CollectMachine(machineFixture(t), false)
	if err != nil {
		t.Fatalf("CollectMachine() error = %v", err)
	}
	if got.Serial != "" || got.HardwareUUID != "" {
		t.Errorf("Serial/HardwareUUID = %q/%q, want both empty when showSerial=false", got.Serial, got.HardwareUUID)
	}
}

func TestCollectMachineShowSerial(t *testing.T) {
	got, err := CollectMachine(machineFixture(t), true)
	if err != nil {
		t.Fatalf("CollectMachine() error = %v", err)
	}
	// The fixture has these redacted to placeholder values; just confirm
	// they were actually read rather than left empty.
	if got.Serial == "" || got.HardwareUUID == "" {
		t.Error("Serial/HardwareUUID empty, want populated when showSerial=true")
	}
}
