package collect

import (
	"os"
	"testing"
)

func TestCollectGPU(t *testing.T) {
	data, err := os.ReadFile("../../testdata/Mac14,2/ioreg-agxaccelerator.txt")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	got, err := CollectGPU(mockCommand{output: string(data)})
	if err != nil {
		t.Fatalf("CollectGPU() error = %v", err)
	}

	want := GPU{CoreCount: 10}
	if got != want {
		t.Errorf("CollectGPU() = %+v, want %+v", got, want)
	}
}

func TestCollectGPUError(t *testing.T) {
	if _, err := CollectGPU(mockCommand{output: ""}); err == nil {
		t.Fatal("CollectGPU() error = nil, want error for empty output")
	}
}
