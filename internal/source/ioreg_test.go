package source

import (
	"os"
	"testing"
)

func TestIORegPropertiesFixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/Mac14,2/ioreg-agxaccelerator.txt")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	cmd := mockCommand{output: string(data)}

	props, err := IORegProperties(cmd, "AGXAccelerator")
	if err != nil {
		t.Fatalf("IORegProperties() error = %v", err)
	}

	cores, err := IntProperty(props, "gpu-core-count")
	if err != nil {
		t.Fatalf("IntProperty(gpu-core-count) error = %v", err)
	}
	if cores != 10 {
		t.Errorf("gpu-core-count = %d, want 10", cores)
	}

	model, err := StringProperty(props, "model")
	if err != nil {
		t.Fatalf("StringProperty(model) error = %v", err)
	}
	if model != "Apple M2" {
		t.Errorf("model = %q, want %q", model, "Apple M2")
	}
}

func TestIORegPropertiesNoMatch(t *testing.T) {
	cmd := mockCommand{output: ""}

	_, err := IORegProperties(cmd, "AGXAccelerator")
	if err == nil {
		t.Fatal("IORegProperties() error = nil, want error for no matching object")
	}
}

func TestIntPropertyMissing(t *testing.T) {
	_, err := IntProperty(map[string]string{}, "gpu-core-count")
	if err == nil {
		t.Fatal("IntProperty() error = nil, want error for missing key")
	}
}

func TestIntPropertyNotAnInteger(t *testing.T) {
	_, err := IntProperty(map[string]string{"gpu-core-count": `"not a number"`}, "gpu-core-count")
	if err == nil {
		t.Fatal("IntProperty() error = nil, want error for non-integer value")
	}
}

func TestDataPropertyFixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/Mac14,2/ioreg-ioplatformexpertdevice.txt")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	cmd := mockCommand{output: string(data)}

	props, err := IORegProperties(cmd, "IOPlatformExpertDevice")
	if err != nil {
		t.Fatalf("IORegProperties() error = %v", err)
	}

	platformName, err := DataProperty(props, "platform-name")
	if err != nil {
		t.Fatalf("DataProperty(platform-name) error = %v", err)
	}
	if platformName != "t8112" {
		t.Errorf("platform-name = %q, want %q", platformName, "t8112")
	}
}

func TestDataPropertyMissing(t *testing.T) {
	_, err := DataProperty(map[string]string{}, "platform-name")
	if err == nil {
		t.Fatal("DataProperty() error = nil, want error for missing key")
	}
}

func TestDataPropertyNotHex(t *testing.T) {
	_, err := DataProperty(map[string]string{"platform-name": `"not hex data"`}, "platform-name")
	if err == nil {
		t.Fatal("DataProperty() error = nil, want error for non-hex value")
	}
}

func TestDataPropertyUint(t *testing.T) {
	got, err := DataPropertyUint(map[string]string{"secure-boot": "<01000000>"}, "secure-boot")
	if err != nil {
		t.Fatalf("DataPropertyUint() error = %v", err)
	}
	if got != 1 {
		t.Errorf("DataPropertyUint() = %d, want 1", got)
	}
}

func TestDataPropertyUintTooLong(t *testing.T) {
	_, err := DataPropertyUint(map[string]string{"x": "<0000000000000000000000000000000000000000>"}, "x")
	if err == nil {
		t.Fatal("DataPropertyUint() error = nil, want error for a value longer than 8 bytes")
	}
}

// IORegNodeProperties runs ioreg against a named device-tree node (e.g.
// "chosen") rather than an IOKit class. On a real Mac, the chosen node
// carries an IOProgressBackbuffer property (a boot progress image) that can
// run past 100KB on a single line - well beyond bufio.Scanner's default
// 64KB line limit - so this fixture specifically exercises that.
func TestIORegNodePropertiesFixtureWithLargeProperty(t *testing.T) {
	data, err := os.ReadFile("../../testdata/Mac14,2/ioreg-chosen.txt")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	cmd := mockCommand{output: string(data)}

	props, err := IORegNodeProperties(cmd, "chosen")
	if err != nil {
		t.Fatalf("IORegNodeProperties() error = %v", err)
	}

	version, err := DataProperty(props, "firmware-version")
	if err != nil {
		t.Fatalf("DataProperty(firmware-version) error = %v", err)
	}
	if version != "mBoot-20457.1.29" {
		t.Errorf("firmware-version = %q, want %q", version, "mBoot-20457.1.29")
	}

	secureBoot, err := DataPropertyUint(props, "secure-boot")
	if err != nil {
		t.Fatalf("DataPropertyUint(secure-boot) error = %v", err)
	}
	if secureBoot != 1 {
		t.Errorf("secure-boot = %d, want 1", secureBoot)
	}
}
