package collect

import "testing"

const vmStatFixture = `Mach Virtual Memory Statistics: (page size of 16384 bytes)
Pages free:                                    36374.
Pages active:                                 244389.
Pages inactive:                               242785.
Pages speculative:                               979.
Pages throttled:                                   0.
Pages wired down:                             137696.
Pages purgeable:                               13537.
"Translation faults":                      825690562.
Pages copy-on-write:                         8000928.
Pages zero filled:                         771665627.
Pages reactivated:                          82925680.
Pages purged:                               27845575.
File-backed pages:                            185693.
Anonymous pages:                              302460.
Pages stored in compressor:                   727536.
Pages occupied by compressor:                 336877.
Decompressions:                             53522082.
Compressions:                               81491067.
Pageins:                                    48268322.
Pageouts:                                     111908.
Swapins:                                           0.
Swapouts:                                          0.
`

func memoryFixture() cmdRouter {
	return cmdRouter{outputs: map[string]string{
		"/usr/sbin/sysctl -n hw.memsize hw.pagesize kern.memorystatus_vm_pressure_level": "17179869184\n16384\n1",
		"/usr/bin/vm_stat": vmStatFixture,
		"/usr/sbin/ioreg -p IODeviceTree -n chosen -d1 -r": `+-o chosen
    {
      "dram-type" = <"LPDDR5">
    }
`,
	}}
}

func TestCollectMemory(t *testing.T) {
	got, err := CollectMemory(memoryFixture())
	if err != nil {
		t.Fatalf("CollectMemory() error = %v", err)
	}

	want := Memory{
		TotalBytes:      17179869184,
		UsedBytes:       (244389 + 137696 + 336877) * 16384,
		WiredBytes:      137696 * 16384,
		CompressedBytes: 336877 * 16384,
		Type:            "LPDDR5",
		Pressure:        "normal",
	}
	if got != want {
		t.Errorf("CollectMemory() = %+v, want %+v", got, want)
	}
}

func TestCollectMemoryPressureLevels(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{"1", "normal"},
		{"2", "warning"},
		{"4", "critical"},
	}
	for _, tt := range tests {
		fixture := memoryFixture()
		fixture.outputs["/usr/sbin/sysctl -n hw.memsize hw.pagesize kern.memorystatus_vm_pressure_level"] =
			"17179869184\n16384\n" + tt.raw

		got, err := CollectMemory(fixture)
		if err != nil {
			t.Fatalf("CollectMemory() error = %v", err)
		}
		if got.Pressure != tt.want {
			t.Errorf("Pressure = %q, want %q", got.Pressure, tt.want)
		}
	}
}

func TestCollectMemoryMissingVMStatField(t *testing.T) {
	fixture := memoryFixture()
	fixture.outputs["/usr/bin/vm_stat"] = "Mach Virtual Memory Statistics: (page size of 16384 bytes)\n"

	if _, err := CollectMemory(fixture); err == nil {
		t.Fatal("CollectMemory() error = nil, want error for missing vm_stat fields")
	}
}

func TestCollectMemoryUnknownDRAMTypeLeftEmpty(t *testing.T) {
	fixture := memoryFixture()
	delete(fixture.outputs, "/usr/sbin/ioreg -p IODeviceTree -n chosen -d1 -r")

	got, err := CollectMemory(fixture)
	if err != nil {
		t.Fatalf("CollectMemory() error = %v", err)
	}
	if got.Type != "" {
		t.Errorf("Type = %q, want empty when the chosen node can't be read", got.Type)
	}
}
