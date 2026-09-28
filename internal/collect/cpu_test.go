package collect

import "testing"

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

func TestCollectCPU(t *testing.T) {
	tests := []struct {
		name    string
		fixture cmdRouter
		want    CPU
	}{
		{
			name: "symmetric M2",
			fixture: cmdRouter{outputs: map[string]string{
				"/usr/sbin/sysctl -n machdep.cpu.brand_string machdep.cpu.core_count hw.nperflevels":                          "Apple M2\n8\n2",
				"/usr/sbin/sysctl -n hw.perflevel0.name hw.perflevel0.logicalcpu hw.perflevel1.name hw.perflevel1.logicalcpu": "Performance\n4\nEfficiency\n4",
			}},
			want: CPU{
				BrandName:  "Apple M2",
				TotalCores: 8,
				Clusters: []CPUClusterCores{
					{Name: "Performance", Cores: 4},
					{Name: "Efficiency", Cores: 4},
				},
			},
		},
		{
			// Regression test for the P/E swap bug in the original socinfo
			// code: with a symmetric core count the swap is invisible, so
			// this fixture uses an asymmetric M3 Max configuration where a
			// swap would be caught.
			name: "asymmetric M3 Max",
			fixture: cmdRouter{outputs: map[string]string{
				"/usr/sbin/sysctl -n machdep.cpu.brand_string machdep.cpu.core_count hw.nperflevels":                          "Apple M3 Max\n16\n2",
				"/usr/sbin/sysctl -n hw.perflevel0.name hw.perflevel0.logicalcpu hw.perflevel1.name hw.perflevel1.logicalcpu": "Performance\n12\nEfficiency\n4",
			}},
			want: CPU{
				BrandName:  "Apple M3 Max",
				TotalCores: 16,
				Clusters: []CPUClusterCores{
					{Name: "Performance", Cores: 12},
					{Name: "Efficiency", Cores: 4},
				},
			},
		},
		{
			// Regression test for a real bug found by running lsmac on
			// GitHub's macOS Actions runners: a virtualized Apple Silicon
			// environment can report a single homogeneous tier
			// (hw.nperflevels=1) rather than the Performance/Efficiency
			// pair every real Mac has. CollectCPU used to hardcode
			// hw.perflevel0/hw.perflevel1 together and errored outright
			// here ("expected 4 lines, got 3").
			name: "single tier (virtualized)",
			fixture: cmdRouter{outputs: map[string]string{
				"/usr/sbin/sysctl -n machdep.cpu.brand_string machdep.cpu.core_count hw.nperflevels": "VMAPPLE2\n3\n1",
				"/usr/sbin/sysctl -n hw.perflevel0.name hw.perflevel0.logicalcpu":                    "Standard\n3",
			}},
			want: CPU{
				BrandName:  "VMAPPLE2",
				TotalCores: 3,
				Clusters: []CPUClusterCores{
					{Name: "Standard", Cores: 3},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CollectCPU(tt.fixture)
			if err != nil {
				t.Fatalf("CollectCPU() error = %v", err)
			}
			if got.BrandName != tt.want.BrandName || got.TotalCores != tt.want.TotalCores {
				t.Errorf("CollectCPU() = %+v, want %+v", got, tt.want)
			}
			if len(got.Clusters) != len(tt.want.Clusters) {
				t.Fatalf("Clusters = %+v, want %+v", got.Clusters, tt.want.Clusters)
			}
			for i := range tt.want.Clusters {
				if got.Clusters[i] != tt.want.Clusters[i] {
					t.Errorf("Clusters[%d] = %+v, want %+v", i, got.Clusters[i], tt.want.Clusters[i])
				}
			}
		})
	}
}

func TestCollectCPUError(t *testing.T) {
	if _, err := CollectCPU(mockCommand{output: "Apple M2\n"}); err == nil {
		t.Fatal("CollectCPU() error = nil, want error for insufficient output")
	}
}
