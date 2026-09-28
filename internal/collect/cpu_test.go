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
		name   string
		output string
		want   CPU
	}{
		{
			name:   "symmetric M2",
			output: "Apple M2\n8\n4\n4\n",
			want: CPU{
				BrandName:        "Apple M2",
				TotalCores:       8,
				PerformanceCores: 4,
				EfficiencyCores:  4,
			},
		},
		{
			// Regression test for the P/E swap bug in the original socinfo
			// code: with a symmetric core count the swap is invisible, so
			// this fixture uses an asymmetric M3 Max configuration where a
			// swap would be caught.
			name:   "asymmetric M3 Max",
			output: "Apple M3 Max\n16\n12\n4\n",
			want: CPU{
				BrandName:        "Apple M3 Max",
				TotalCores:       16,
				PerformanceCores: 12,
				EfficiencyCores:  4,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CollectCPU(mockCommand{output: tt.output})
			if err != nil {
				t.Fatalf("CollectCPU() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("CollectCPU() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestCollectCPUError(t *testing.T) {
	if _, err := CollectCPU(mockCommand{output: "Apple M2\n"}); err == nil {
		t.Fatal("CollectCPU() error = nil, want error for insufficient output")
	}
}
