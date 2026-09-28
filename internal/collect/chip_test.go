package collect

import (
	"os"
	"testing"
)

func TestCollectChipID(t *testing.T) {
	data, err := os.ReadFile("../../testdata/Mac14,2/ioreg-ioplatformexpertdevice.txt")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	got, err := CollectChipID(mockCommand{output: string(data)})
	if err != nil {
		t.Fatalf("CollectChipID() error = %v", err)
	}

	want := "T8112"
	if got != want {
		t.Errorf("CollectChipID() = %q, want %q", got, want)
	}
}

func TestCollectChipIDError(t *testing.T) {
	if _, err := CollectChipID(mockCommand{output: ""}); err == nil {
		t.Fatal("CollectChipID() error = nil, want error for empty output")
	}
}
