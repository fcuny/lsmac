package render

import (
	"bytes"
	"strings"
	"testing"

	"fcuny.net/lsmac/internal/chips"
	"fcuny.net/lsmac/internal/collect"
)

func TestCPU(t *testing.T) {
	var buf bytes.Buffer
	cpu := collect.CPU{BrandName: "Apple M2", TotalCores: 8, PerformanceCores: 4, EfficiencyCores: 4}
	specs := chips.Specs{CPUBandwidthGBs: 100}

	if err := CPU(&buf, cpu, specs); err != nil {
		t.Fatalf("CPU() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"Apple M2", "8 cores: 4P + 4E", "100 GB/s"} {
		if !strings.Contains(out, want) {
			t.Errorf("CPU() output = %q, want it to contain %q", out, want)
		}
	}
}

func TestCPUUnknownBandwidthOmitted(t *testing.T) {
	var buf bytes.Buffer
	cpu := collect.CPU{BrandName: "Apple M2 Pro", TotalCores: 10, PerformanceCores: 6, EfficiencyCores: 4}

	if err := CPU(&buf, cpu, chips.Specs{}); err != nil {
		t.Fatalf("CPU() error = %v", err)
	}

	if strings.Contains(buf.String(), "GB/s") {
		t.Errorf("CPU() output = %q, want no bandwidth line when unknown", buf.String())
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
