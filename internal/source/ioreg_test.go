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
