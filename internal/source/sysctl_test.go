package source

import (
	"errors"
	"testing"
)

type mockCommand struct {
	output string
	err    error
}

func (m mockCommand) Execute(binary string, args ...string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []byte(m.output), nil
}

func TestSysctl(t *testing.T) {
	cmd := mockCommand{output: "Apple M2\n8\n4\n4\n"}

	values, err := Sysctl(cmd,
		"machdep.cpu.brand_string",
		"machdep.cpu.core_count",
		"hw.perflevel0.logicalcpu",
		"hw.perflevel1.logicalcpu",
	)
	if err != nil {
		t.Fatalf("Sysctl() error = %v", err)
	}

	want := []string{"Apple M2", "8", "4", "4"}
	if len(values) != len(want) {
		t.Fatalf("Sysctl() = %v, want %v", values, want)
	}
	for i := range want {
		if values[i] != want[i] {
			t.Errorf("Sysctl()[%d] = %q, want %q", i, values[i], want[i])
		}
	}
}

func TestSysctlLineCountMismatch(t *testing.T) {
	cmd := mockCommand{output: "Apple M2\n"}

	_, err := Sysctl(cmd, "machdep.cpu.brand_string", "machdep.cpu.core_count")
	if err == nil {
		t.Fatal("Sysctl() error = nil, want error for line count mismatch")
	}
}

func TestSysctlCommandError(t *testing.T) {
	cmd := mockCommand{err: errors.New("command failed")}

	_, err := Sysctl(cmd, "machdep.cpu.brand_string")
	if err == nil {
		t.Fatal("Sysctl() error = nil, want error propagated from command")
	}
}
