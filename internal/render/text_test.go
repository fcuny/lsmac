package render

import (
	"bytes"
	"strings"
	"testing"

	"fcuny.net/lsmac/internal/chips"
	"fcuny.net/lsmac/internal/collect"
)

func TestChip(t *testing.T) {
	var buf bytes.Buffer
	chip := chips.Chip{
		ID:                 "T8112",
		MarketingName:      "Apple M2",
		ProcessNode:        "5-nanometer (2nd generation)",
		MemoryBandwidthGBs: 100,
	}

	if err := Chip(&buf, chip); err != nil {
		t.Fatalf("Chip() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"Apple M2 (T8112)", "5-nanometer (2nd generation)", "100 GB/s"} {
		if !strings.Contains(out, want) {
			t.Errorf("Chip() output = %q, want it to contain %q", out, want)
		}
	}
}

func TestChipUnknownFallsBackToID(t *testing.T) {
	var buf bytes.Buffer

	if err := Chip(&buf, chips.Chip{ID: "T9999"}); err != nil {
		t.Fatalf("Chip() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "T9999 (T9999)") {
		t.Errorf("Chip() output = %q, want it to fall back to the chip ID", out)
	}
	if strings.Contains(out, "GB/s") {
		t.Errorf("Chip() output = %q, want no bandwidth line when unknown", out)
	}
}

func TestCPU(t *testing.T) {
	var buf bytes.Buffer
	cpu := collect.CPU{BrandName: "Apple M2", TotalCores: 8, PerformanceCores: 4, EfficiencyCores: 4}

	if err := CPU(&buf, cpu); err != nil {
		t.Fatalf("CPU() error = %v", err)
	}

	if !strings.Contains(buf.String(), "8 cores: 4P + 4E") {
		t.Errorf("CPU() output = %q, want it to contain %q", buf.String(), "8 cores: 4P + 4E")
	}
}

func TestGPU(t *testing.T) {
	var buf bytes.Buffer

	if err := GPU(&buf, collect.GPU{CoreCount: 10}); err != nil {
		t.Fatalf("GPU() error = %v", err)
	}

	if !strings.Contains(buf.String(), "10 cores") {
		t.Errorf("GPU() output = %q, want it to contain %q", buf.String(), "10 cores")
	}
}
