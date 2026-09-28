package render

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"fcuny.net/lsmac/internal/collect"
	"fcuny.net/lsmac/internal/models"
)

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	sections := map[string]any{
		"gpu": collect.GPU{CoreCount: 10},
	}

	if err := JSON(&buf, sections); err != nil {
		t.Fatalf("JSON() error = %v", err)
	}

	var decoded map[string]collect.GPU
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("decoding JSON() output: %v", err)
	}
	if decoded["gpu"].CoreCount != 10 {
		t.Errorf("decoded gpu.coreCount = %d, want 10", decoded["gpu"].CoreCount)
	}
	if !strings.Contains(buf.String(), "\n  ") {
		t.Errorf("JSON() output = %q, want indented output", buf.String())
	}
}

func TestMachineJSON(t *testing.T) {
	machine := collect.Machine{ModelIdentifier: "Mac14,2", SKU: "MN703LL/A"}
	model := models.Model{MarketingName: "MacBook Air (M2, 2022)"}

	data, err := json.Marshal(MachineJSON(machine, model))
	if err != nil {
		t.Fatalf("json.Marshal error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if decoded["modelIdentifier"] != "Mac14,2" {
		t.Errorf("modelIdentifier = %v, want %q", decoded["modelIdentifier"], "Mac14,2")
	}
	if decoded["marketingName"] != "MacBook Air (M2, 2022)" {
		t.Errorf("marketingName = %v, want %q", decoded["marketingName"], "MacBook Air (M2, 2022)")
	}
}

func TestCPUJSON(t *testing.T) {
	cpu := collect.CPU{BrandName: "Apple M2", TotalCores: 8, PerformanceCores: 4, EfficiencyCores: 4}
	detail := collect.CPUDetail{Family: "0xDA33D83D", PageSize: 16384, Features: []string{"BTI"}}

	data, err := json.Marshal(CPUJSON(cpu, detail))
	if err != nil {
		t.Fatalf("json.Marshal error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if decoded["brandName"] != "Apple M2" {
		t.Errorf("brandName = %v, want %q", decoded["brandName"], "Apple M2")
	}
	if decoded["family"] != "0xDA33D83D" {
		t.Errorf("family = %v, want %q", decoded["family"], "0xDA33D83D")
	}
}

func TestMemoryJSON(t *testing.T) {
	mem := collect.Memory{TotalBytes: 16 << 30, UsedBytes: 8 << 30, Pressure: "normal"}

	data, err := json.Marshal(MemoryJSON(mem, 100))
	if err != nil {
		t.Fatalf("json.Marshal error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if decoded["bandwidthGBs"] != float64(100) {
		t.Errorf("bandwidthGBs = %v, want 100", decoded["bandwidthGBs"])
	}
	if decoded["totalBytes"] != float64(16<<30) {
		t.Errorf("totalBytes = %v, want %d", decoded["totalBytes"], 16<<30)
	}
}
