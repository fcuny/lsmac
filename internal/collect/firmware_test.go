package collect

import (
	"os"
	"testing"
)

func chosenFixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("../../testdata/Mac14,2/ioreg-chosen.txt")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	return string(data)
}

func firmwareFixture(t *testing.T, sipOutput string) cmdRouter {
	t.Helper()
	return cmdRouter{outputs: map[string]string{
		"/usr/sbin/ioreg -p IODeviceTree -n chosen -d1 -r": chosenFixture(t),
		"/usr/bin/csrutil status":                          sipOutput,
	}}
}

func TestCollectFirmware(t *testing.T) {
	got, err := CollectFirmware(firmwareFixture(t, "System Integrity Protection status: enabled."))
	if err != nil {
		t.Fatalf("CollectFirmware() error = %v", err)
	}

	want := Firmware{
		Version:    "mBoot-20457.1.29",
		SecureBoot: true,
		SIPEnabled: true,
	}
	if got != want {
		t.Errorf("CollectFirmware() = %+v, want %+v", got, want)
	}
}

func TestCollectFirmwareSIPDisabled(t *testing.T) {
	got, err := CollectFirmware(firmwareFixture(t, "System Integrity Protection status: disabled."))
	if err != nil {
		t.Fatalf("CollectFirmware() error = %v", err)
	}
	if got.SIPEnabled {
		t.Error("SIPEnabled = true, want false")
	}
}

func TestParseSIPStatusUnexpectedOutput(t *testing.T) {
	if _, err := parseSIPStatus("garbage"); err == nil {
		t.Fatal("parseSIPStatus() error = nil, want error for unrecognized output")
	}
}
