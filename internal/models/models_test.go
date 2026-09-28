package models

import "testing"

func TestLookup(t *testing.T) {
	tests := []struct {
		identifier string
		want       string
	}{
		{"Mac14,2", "MacBook Air (M2, 2022)"},
		{"MacBookAir10,1", "MacBook Air (M1, 2020)"},
		{"Macmini9,1", "Mac mini (M1, 2020)"},
		{"Mac18,5", "Mac mini (M6)"},
	}

	for _, tt := range tests {
		t.Run(tt.identifier, func(t *testing.T) {
			got, ok := Lookup(tt.identifier)
			if !ok {
				t.Fatalf("Lookup(%q) not found", tt.identifier)
			}
			if got.MarketingName != tt.want {
				t.Errorf("Lookup(%q).MarketingName = %q, want %q", tt.identifier, got.MarketingName, tt.want)
			}
		})
	}
}

func TestLookupUnknown(t *testing.T) {
	if _, ok := Lookup("Mac99,99"); ok {
		t.Fatal("Lookup(Mac99,99) ok = true, want false for unknown identifier")
	}
}

// Distinct chip tiers sharing one chassis (and one model identifier) must
// resolve to Apple's own disjunctive marketing name rather than picking one.
func TestLookupSharedIdentifierIsDisjunctive(t *testing.T) {
	max, ok := Lookup("Mac17,7")
	if !ok {
		t.Fatal("Lookup(Mac17,7) not found")
	}
	pro, ok := Lookup("Mac17,9")
	if !ok {
		t.Fatal("Lookup(Mac17,9) not found")
	}
	if max.MarketingName != pro.MarketingName {
		t.Errorf("shared 14-inch M5 Pro/Max identifiers resolved to different names: %q vs %q", max.MarketingName, pro.MarketingName)
	}
	if max.MarketingName != "MacBook Pro (14-inch, M5 Pro or M5 Max)" {
		t.Errorf("MarketingName = %q, want the disjunctive Apple name", max.MarketingName)
	}
}

// The Mac Pro tower and rack forms share Mac14,8 with no way to tell them
// apart from hw.model alone; both directions of the lookup return the same
// entry rather than one overwriting the other silently.
func TestLookupMacProSharedIdentifier(t *testing.T) {
	got, ok := Lookup("Mac14,8")
	if !ok {
		t.Fatal("Lookup(Mac14,8) not found")
	}
	if got.MarketingName != "Mac Pro (2023)" {
		t.Errorf("MarketingName = %q, want %q", got.MarketingName, "Mac Pro (2023)")
	}
}
