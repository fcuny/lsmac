// Package models maps a Mac's model identifier (the value
// `sysctl -n hw.model` reports) to its retail marketing name, for Apple
// Silicon Macs only.
//
// Apple Silicon spans two identifier eras, both still returned verbatim by
// hw.model on current macOS:
//   - 2020-2021 (M1 generation only): product-specific strings, e.g.
//     "MacBookAir10,1", "MacBookPro18,3", "Macmini9,1".
//   - 2022 onward (M2 generation and later): a flat, product-agnostic
//     "Mac<N>,<N>" scheme shared across every product line - there is no
//     way to tell MacBook from Mac Studio from the number alone, so the
//     table below must be keyed on the whole string.
//
// A single identifier frequently maps to more than one chip configuration
// (e.g. the 14-inch MacBook Pro chassis shared between M3 Pro and M3 Max in
// 2023): hw.model cannot disambiguate that by itself, so where Apple's own
// marketing name is disjunctive ("MacBook Pro (14-inch, M5 Pro or M5 Max)")
// the table stores it exactly that way rather than guessing which chip
// tier is present - internal/chips, keyed on the actual chip ID from the
// device tree, is what disambiguates the chip.
//
// Sourced from Apple's own "Identify your Mac model" support pages
// (support.apple.com), cross-checked against everymac.com's per-model
// identifier pages. A gap here (a model identifier not in the table) is
// left as a lookup miss rather than guessed.
package models

// Model holds the marketing name for one Mac model identifier.
type Model struct {
	Identifier    string
	MarketingName string
}

// table is keyed by the exact hw.model string.
var table = map[string]Model{
	// MacBook Air. Source: https://support.apple.com/en-us/102869.
	// Every size/chip combination gets its own identifier - no sharing.
	"MacBookAir10,1": {"MacBookAir10,1", "MacBook Air (M1, 2020)"},
	"Mac14,2":        {"Mac14,2", "MacBook Air (M2, 2022)"},
	"Mac14,15":       {"Mac14,15", "MacBook Air (15-inch, M2, 2023)"},
	"Mac15,12":       {"Mac15,12", "MacBook Air (13-inch, M3, 2024)"},
	"Mac15,13":       {"Mac15,13", "MacBook Air (15-inch, M3, 2024)"},
	"Mac16,12":       {"Mac16,12", "MacBook Air (13-inch, M4, 2025)"},
	"Mac16,13":       {"Mac16,13", "MacBook Air (15-inch, M4, 2025)"},
	"Mac17,3":        {"Mac17,3", "MacBook Air (13-inch, M5)"},
	"Mac17,4":        {"Mac17,4", "MacBook Air (15-inch, M5)"},

	// MacBook Pro. Source: https://support.apple.com/en-us/108052, cross-
	// checked with everymac.com/systems/by-identifier/all-macbook-pro-model-identifiers.html.
	// The 13-inch M1 (2020) kept the pre-2022 product-specific identifier
	// format even though it's Apple Silicon.
	"MacBookPro17,1": {"MacBookPro17,1", "MacBook Pro (13-inch, M1, 2020)"},
	"MacBookPro18,3": {"MacBookPro18,3", "MacBook Pro (14-inch, 2021)"}, // M1 Pro
	"MacBookPro18,4": {"MacBookPro18,4", "MacBook Pro (14-inch, 2021)"}, // M1 Max
	"MacBookPro18,1": {"MacBookPro18,1", "MacBook Pro (16-inch, 2021)"}, // M1 Pro
	"MacBookPro18,2": {"MacBookPro18,2", "MacBook Pro (16-inch, 2021)"}, // M1 Max
	"Mac14,7":        {"Mac14,7", "MacBook Pro (13-inch, M2, 2022)"},
	"Mac14,9":        {"Mac14,9", "MacBook Pro (14-inch, 2023)"},      // M2 Pro
	"Mac14,5":        {"Mac14,5", "MacBook Pro (14-inch, 2023)"},      // M2 Max
	"Mac14,10":       {"Mac14,10", "MacBook Pro (16-inch, 2023)"},     // M2 Pro
	"Mac14,6":        {"Mac14,6", "MacBook Pro (16-inch, 2023)"},      // M2 Max
	"Mac15,3":        {"Mac15,3", "MacBook Pro (14-inch, Nov 2023)"},  // M3
	"Mac15,6":        {"Mac15,6", "MacBook Pro (14-inch, Nov 2023)"},  // M3 Pro
	"Mac15,8":        {"Mac15,8", "MacBook Pro (14-inch, Nov 2023)"},  // M3 Max (14-core CPU/30-core GPU)
	"Mac15,10":       {"Mac15,10", "MacBook Pro (14-inch, Nov 2023)"}, // M3 Max (16-core CPU/40-core GPU)
	"Mac15,7":        {"Mac15,7", "MacBook Pro (16-inch, Nov 2023)"},  // M3 Pro
	"Mac15,9":        {"Mac15,9", "MacBook Pro (16-inch, Nov 2023)"},  // M3 Max (14-core CPU/30-core GPU)
	"Mac15,11":       {"Mac15,11", "MacBook Pro (16-inch, Nov 2023)"}, // M3 Max (16-core CPU/40-core GPU)
	"Mac16,1":        {"Mac16,1", "MacBook Pro (14-inch, 2024)"},      // M4
	"Mac16,8":        {"Mac16,8", "MacBook Pro (14-inch, 2024)"},      // M4 Pro
	"Mac16,6":        {"Mac16,6", "MacBook Pro (14-inch, 2024)"},      // M4 Max
	"Mac16,7":        {"Mac16,7", "MacBook Pro (16-inch, 2024)"},      // M4 Pro
	"Mac16,5":        {"Mac16,5", "MacBook Pro (16-inch, 2024)"},      // M4 Max
	"Mac17,2":        {"Mac17,2", "MacBook Pro (14-inch, M5)"},
	"Mac17,9":        {"Mac17,9", "MacBook Pro (14-inch, M5 Pro or M5 Max)"}, // M5 Pro
	"Mac17,7":        {"Mac17,7", "MacBook Pro (14-inch, M5 Pro or M5 Max)"}, // M5 Max
	"Mac17,8":        {"Mac17,8", "MacBook Pro (16-inch, M5 Pro or M5 Max)"}, // M5 Pro
	"Mac17,6":        {"Mac17,6", "MacBook Pro (16-inch, M5 Pro or M5 Max)"}, // M5 Max

	// Mac mini. Source: https://support.apple.com/en-us/102852, cross-
	// checked with everymac.com. No plain (non-Pro) M5 Mac mini identifier
	// was found - the mini appears to have gone from M4 straight to M6,
	// with an M5 Pro variant alongside; not asserting a base M5 mini
	// exists.
	"Macmini9,1": {"Macmini9,1", "Mac mini (M1, 2020)"},
	"Mac14,3":    {"Mac14,3", "Mac mini (2023)"},  // M2
	"Mac14,12":   {"Mac14,12", "Mac mini (2023)"}, // M2 Pro
	"Mac16,10":   {"Mac16,10", "Mac mini (2024)"}, // M4
	"Mac16,11":   {"Mac16,11", "Mac mini (2024)"}, // M4 Pro
	"Mac17,16":   {"Mac17,16", "Mac mini (M5 Pro)"},
	"Mac18,5":    {"Mac18,5", "Mac mini (M6)"},

	// iMac. Source: https://support.apple.com/en-us/108054, cross-checked
	// with everymac.com/systems/by-identifier/all-imac-model-identifiers.html.
	// No M2 iMac was ever released (M1 -> M3 directly); no M5 iMac as of
	// 2026-09-28.
	"iMac21,1": {"iMac21,1", "iMac (24-inch, M1, 2021)"},        // 4-port, 8-core GPU
	"iMac21,2": {"iMac21,2", "iMac (24-inch, M1, 2021)"},        // 2-port, 7-core GPU
	"Mac15,4":  {"Mac15,4", "iMac (24-inch, 2023, Two ports)"},  // M3, 8-core GPU
	"Mac15,5":  {"Mac15,5", "iMac (24-inch, 2023, Four ports)"}, // M3, 10-core GPU
	"Mac16,2":  {"Mac16,2", "iMac (24-inch, 2024, Two ports)"},  // M4, 8-core GPU
	"Mac16,3":  {"Mac16,3", "iMac (24-inch, 2024, Four ports)"}, // M4, 10-core GPU

	// Mac Studio. Source: https://support.apple.com/en-us/102231, cross-
	// checked with everymac.com. Note M3 Ultra and M4 Max share the "Mac
	// Studio (2025)" marketing name but have distinct identifiers.
	"Mac13,1":  {"Mac13,1", "Mac Studio (2022)"},  // M1 Max
	"Mac13,2":  {"Mac13,2", "Mac Studio (2022)"},  // M1 Ultra
	"Mac14,13": {"Mac14,13", "Mac Studio (2023)"}, // M2 Max
	"Mac14,14": {"Mac14,14", "Mac Studio (2023)"}, // M2 Ultra
	"Mac15,14": {"Mac15,14", "Mac Studio (2025)"}, // M3 Ultra
	"Mac16,9":  {"Mac16,9", "Mac Studio (2025)"},  // M4 Max
	"Mac17,14": {"Mac17,14", "Mac Studio (M5 Max)"},
	"Mac17,15": {"Mac17,15", "Mac Studio (M5 Ultra)"},

	// Mac Pro. Source: https://support.apple.com/en-us/102887. The only
	// Apple Silicon Mac Pro to date; tower and rack forms share the same
	// identifier (form factor isn't distinguishable from hw.model alone).
	"Mac14,8": {"Mac14,8", "Mac Pro (2023)"},
}

// Lookup returns the marketing name for a model identifier (e.g.
// "Mac14,2"), exactly as reported by hw.model. ok is false when the
// identifier is not in the table.
func Lookup(identifier string) (Model, bool) {
	m, ok := table[identifier]
	return m, ok
}
