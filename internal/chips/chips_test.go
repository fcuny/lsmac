package chips

import "testing"

func TestLookup(t *testing.T) {
	tests := []struct {
		id   string
		want Chip
	}{
		{
			"T8112",
			Chip{
				ID:                 "T8112",
				MarketingName:      "Apple M2",
				ProcessNode:        "5-nanometer (2nd generation)",
				MemoryBandwidthGBs: 100,
				NeuralEngineCores:  16,
				Media:              Media{VideoDecodeEngines: 1, VideoEncodeEngines: 1, ProResEngines: 1},
			},
		},
		{
			// lowercase input, as read straight off the device tree.
			"t8103",
			Chip{
				ID:                "T8103",
				MarketingName:     "Apple M1",
				ProcessNode:       "5-nanometer",
				NeuralEngineCores: 16,
			},
		},
		{
			"T8152",
			Chip{
				ID:                "T8152",
				MarketingName:     "Apple M6",
				ProcessNode:       "2-nanometer",
				NeuralEngineCores: 32,
				Media:             Media{VideoDecodeEngines: 1, VideoEncodeEngines: 1, ProResEngines: 1, AV1Decode: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got, ok := Lookup(tt.id)
			if !ok {
				t.Fatalf("Lookup(%q) not found", tt.id)
			}
			if got != tt.want {
				t.Errorf("Lookup(%q) = %+v, want %+v", tt.id, got, tt.want)
			}
		})
	}
}

func TestLookupUnknown(t *testing.T) {
	if _, ok := Lookup("T9999"); ok {
		t.Fatal("Lookup(T9999) ok = true, want false for unknown chip ID")
	}
}

// M3 Max has two chip IDs for the same marketing name, distinguished by
// die configuration; each must carry its own bandwidth figure.
func TestLookupM3MaxDieVariants(t *testing.T) {
	full, ok := Lookup("T6031")
	if !ok {
		t.Fatal("Lookup(T6031) not found")
	}
	binned, ok := Lookup("T6034")
	if !ok {
		t.Fatal("Lookup(T6034) not found")
	}

	if full.MarketingName != "Apple M3 Max" || binned.MarketingName != "Apple M3 Max" {
		t.Errorf("marketing names = %q, %q, want both %q", full.MarketingName, binned.MarketingName, "Apple M3 Max")
	}
	if full.MemoryBandwidthGBs != 400 {
		t.Errorf("T6031 MemoryBandwidthGBs = %d, want 400", full.MemoryBandwidthGBs)
	}
	if binned.MemoryBandwidthGBs != 300 {
		t.Errorf("T6034 MemoryBandwidthGBs = %d, want 300", binned.MemoryBandwidthGBs)
	}
}

// Ambiguous-bandwidth chips (single chip ID, multiple SKUs at different
// bandwidths - by GPU core count for the Max chips, by memory capacity for
// the M6) must be left at 0 rather than guessed.
func TestLookupAmbiguousBandwidthLeftUnset(t *testing.T) {
	for _, id := range []string{"T6041", "T6051", "T8152"} {
		c, ok := Lookup(id)
		if !ok {
			t.Fatalf("Lookup(%q) not found", id)
		}
		if c.MemoryBandwidthGBs != 0 {
			t.Errorf("Lookup(%q).MemoryBandwidthGBs = %d, want 0 (ambiguous across SKUs)", id, c.MemoryBandwidthGBs)
		}
	}
}

// Ultra chips built from fused dies don't have an Apple-stated process node
// in their own announcement; it must be left empty rather than copied from
// the Max die it's built from.
func TestLookupUltraProcessNodeUnset(t *testing.T) {
	for _, id := range []string{"T6002", "T6032", "T6052"} {
		c, ok := Lookup(id)
		if !ok {
			t.Fatalf("Lookup(%q) not found", id)
		}
		if c.ProcessNode != "" {
			t.Errorf("Lookup(%q).ProcessNode = %q, want empty (not stated by Apple)", id, c.ProcessNode)
		}
	}
}
