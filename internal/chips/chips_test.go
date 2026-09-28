package chips

import "testing"

func TestFromBrandString(t *testing.T) {
	tests := []struct {
		brand string
		want  Chip
	}{
		{"Apple M1", M1},
		{"Apple M1 Pro", M1Pro},
		{"Apple M1 Max", M1Max},
		{"Apple M1 Ultra", M1Ultra},
		{"Apple M2", M2},
		{"Apple M2 Pro", M2Pro},
		{"Apple M2 Max", M2Max},
		{"Apple M2 Ultra", M2Ultra},
		{"Apple M3", M3},
		{"Apple M3 Pro", M3Pro},
		{"Apple M3 Max", M3Max},
		{"Intel Core i7", Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.brand, func(t *testing.T) {
			if got := FromBrandString(tt.brand); got != tt.want {
				t.Errorf("FromBrandString(%q) = %v, want %v", tt.brand, got, tt.want)
			}
		})
	}
}

func TestSpecs(t *testing.T) {
	tests := []struct {
		chip Chip
		want Specs
	}{
		{M1, Specs{CPUBandwidthGBs: 70, GPUBandwidthGBs: 70}},
		{M1Pro, Specs{CPUBandwidthGBs: 200, GPUBandwidthGBs: 200}},
		{M1Max, Specs{CPUBandwidthGBs: 250, GPUBandwidthGBs: 400}},
		{M1Ultra, Specs{CPUBandwidthGBs: 500, GPUBandwidthGBs: 800}},
		{M2, Specs{CPUBandwidthGBs: 100, GPUBandwidthGBs: 100}},
		{M2Pro, Specs{}},
		{Unknown, Specs{}},
	}

	for _, tt := range tests {
		if got := tt.chip.Specs(); got != tt.want {
			t.Errorf("%v.Specs() = %+v, want %+v", tt.chip, got, tt.want)
		}
	}
}
