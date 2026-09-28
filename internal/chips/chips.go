// Package chips holds facts about Apple Silicon chips that the hardware
// itself does not report: marketing name, process node, memory bandwidth,
// Neural Engine core count, and media engine capabilities.
//
// The table is keyed on the chip ID (e.g. "T8112"), read from the device
// tree's platform-name property (see internal/collect.CollectChipID) rather
// than matched against the CPU brand string. Apple does not publish chip
// IDs itself; they are documented by community trackers such as
// theiphonewiki.com and theapplewiki.com, cross-checked against each other
// and, for recent chips, against Asahi Linux kernel device-tree patches and
// pre-release OS beta strings reported by outlets like MacRumors. Every
// other fact is sourced from an Apple newsroom article or an
// support.apple.com tech-specs page, cited on the entry that uses it.
//
// A field is left at its zero value when Apple has not published it, rather
// than estimated or inferred from a related chip. In particular:
//   - ProcessNode is empty for the M1 Ultra, M3 Ultra and M5 Ultra: each is
//     a multi-die package built from a Max die whose node Apple states
//     elsewhere, but Apple's Ultra-specific announcement does not restate
//     it, so it is not carried over here.
//   - MemoryBandwidthGBs is 0 for the base M1, which Apple never gave a
//     GB/s figure for.
//   - MemoryBandwidthGBs is also 0 for the M4 Max and M5 Max: Apple
//     publishes two figures for each (depending on GPU core count), the
//     chip ID does not distinguish which configuration is present, and
//     there is currently no collector that reads GPU core count to
//     disambiguate. The two published figures are noted in a comment on
//     the entry instead of guessed into the field. M3 Max does not have
//     this problem: its two configurations have distinct chip IDs
//     (T6031/T6034), so each gets its own entry with its own figure.
//   - MemoryBandwidthGBs is also 0 for the M6: Apple's Mac mini tech-specs
//     page lists 153GB/s for two of its three configurations and 170GB/s
//     for the third, tied to unified memory capacity rather than a
//     distinct chip ID, so again neither figure is picked over the other.
//   - Media.AV1Decode is only set true when Apple's tech-specs page
//     explicitly lists AV1 decode; false means "not confirmed", not
//     "confirmed absent".
//   - Media video decode/encode engine counts are 0 when Apple described
//     the media engine only in prose ("hardware-accelerated encode and
//     decode") without publishing a count.
package chips

import "strings"

// Media describes a chip's video engine capabilities.
type Media struct {
	VideoDecodeEngines uint8 `json:"videoDecodeEngines"`
	VideoEncodeEngines uint8 `json:"videoEncodeEngines"`
	ProResEngines      uint8 `json:"proResEngines"`
	AV1Decode          bool  `json:"av1Decode"`
}

// Chip holds the published facts for one Apple Silicon chip.
type Chip struct {
	ID                 string `json:"id"`
	MarketingName      string `json:"marketingName,omitempty"`
	ProcessNode        string `json:"processNode,omitempty"`
	MemoryBandwidthGBs uint32 `json:"memoryBandwidthGBs,omitempty"`
	NeuralEngineCores  uint16 `json:"neuralEngineCores,omitempty"`
	Media              Media  `json:"media"`
}

// table is keyed by chip ID, upper-cased.
var table = map[string]Chip{
	// M1 generation.
	// Chip ID T8103: community source https://theiphonewiki.com/wiki/T8103.
	"T8103": {
		ID:            "T8103",
		MarketingName: "Apple M1",
		// https://www.apple.com/newsroom/2020/11/apple-unleashes-m1/
		ProcessNode: "5-nanometer",
		// MemoryBandwidthGBs left at 0: Apple never published a GB/s figure
		// for the base M1 (checked the newsroom article and the M1 Mac
		// mini/MacBook Air tech-specs pages).
		// https://www.apple.com/newsroom/2020/11/apple-unleashes-m1/
		NeuralEngineCores: 16,
		// Apple's M1 material describes the media engine only in prose
		// ("low-power, highly efficient media encode and decode engines"),
		// no counts and no ProRes silicon (that starts at M1 Pro).
		Media: Media{},
	},
	// Chip ID T6000: community source https://theiphonewiki.com/wiki/T6000.
	"T6000": {
		ID:            "T6000",
		MarketingName: "Apple M1 Pro",
		// https://www.apple.com/newsroom/2021/10/introducing-m1-pro-and-m1-max-the-most-powerful-chips-apple-has-ever-built/
		ProcessNode:        "5-nanometer",
		MemoryBandwidthGBs: 200,
		NeuralEngineCores:  16,
		// https://support.apple.com/en-us/111902 (MacBook Pro 14-inch 2021 tech specs)
		Media: Media{VideoDecodeEngines: 1, VideoEncodeEngines: 1, ProResEngines: 1},
	},
	// Chip ID T6001: community source https://theiphonewiki.com/wiki/T6001.
	"T6001": {
		ID:            "T6001",
		MarketingName: "Apple M1 Max",
		// https://www.apple.com/newsroom/2021/10/introducing-m1-pro-and-m1-max-the-most-powerful-chips-apple-has-ever-built/
		ProcessNode:        "5-nanometer",
		MemoryBandwidthGBs: 400,
		NeuralEngineCores:  16,
		// https://support.apple.com/en-us/111900 (Mac Studio 2022 tech specs)
		Media: Media{VideoDecodeEngines: 1, VideoEncodeEngines: 2, ProResEngines: 2},
	},
	// Chip ID T6002: community source https://theiphonewiki.com/wiki/T6002.
	"T6002": {
		ID:            "T6002",
		MarketingName: "Apple M1 Ultra",
		// ProcessNode not stated in the M1 Ultra announcement (an
		// UltraFusion package of two M1 Max dies); see package doc.
		// https://www.apple.com/newsroom/2022/03/apple-unveils-m1-ultra-the-worlds-most-powerful-chip-for-a-personal-computer/
		MemoryBandwidthGBs: 800,
		NeuralEngineCores:  32,
		// https://support.apple.com/en-us/111900 (Mac Studio 2022 tech specs)
		Media: Media{VideoDecodeEngines: 2, VideoEncodeEngines: 4, ProResEngines: 4},
	},

	// M2 generation.
	// Chip ID T8112: community source https://theapplewiki.com/wiki/CHIP.
	"T8112": {
		ID:            "T8112",
		MarketingName: "Apple M2",
		// https://www.apple.com/newsroom/2022/06/apple-unveils-m2-with-breakthrough-performance-and-capabilities/
		ProcessNode:        "5-nanometer (2nd generation)",
		MemoryBandwidthGBs: 100,
		NeuralEngineCores:  16,
		// https://support.apple.com/en-us/111867 (MacBook Air M2 tech specs); no AV1.
		Media: Media{VideoDecodeEngines: 1, VideoEncodeEngines: 1, ProResEngines: 1},
	},
	// Chip ID T6020: community source https://theapplewiki.com/wiki/CHIP.
	"T6020": {
		ID:            "T6020",
		MarketingName: "Apple M2 Pro",
		// https://www.apple.com/newsroom/2023/01/apple-unveils-m2-pro-and-m2-max-next-generation-chips-for-next-level-workflows/
		ProcessNode:        "5-nanometer (2nd generation)",
		MemoryBandwidthGBs: 200,
		NeuralEngineCores:  16,
		// Same source describes the media engine only in prose, no counts published.
		Media: Media{},
	},
	// Chip ID T6021: community source https://theapplewiki.com/wiki/CHIP.
	"T6021": {
		ID:            "T6021",
		MarketingName: "Apple M2 Max",
		// https://www.apple.com/newsroom/2023/01/apple-unveils-m2-pro-and-m2-max-next-generation-chips-for-next-level-workflows/
		ProcessNode:        "5-nanometer (2nd generation)",
		MemoryBandwidthGBs: 400,
		NeuralEngineCores:  16,
		// Same source: "two video encode engines and two ProRes engines";
		// decode engine count not stated.
		Media: Media{VideoEncodeEngines: 2, ProResEngines: 2},
	},
	// Chip ID T6022: community source https://theapplewiki.com/wiki/CHIP
	// (pattern-consistent with T6020/T6021, not independently confirmed).
	"T6022": {
		ID:            "T6022",
		MarketingName: "Apple M2 Ultra",
		// https://www.apple.com/newsroom/2023/06/apple-introduces-m2-ultra/
		ProcessNode:        "5-nanometer (2nd generation)",
		MemoryBandwidthGBs: 800,
		NeuralEngineCores:  32,
		// https://support.apple.com/en-us/111835 (Mac Studio 2023 tech specs)
		Media: Media{VideoDecodeEngines: 2, VideoEncodeEngines: 4, ProResEngines: 4},
	},

	// M3 generation.
	// Chip ID T8122: community source https://en.wikipedia.org/wiki/Apple_M3.
	"T8122": {
		ID:            "T8122",
		MarketingName: "Apple M3",
		// https://www.apple.com/newsroom/2023/10/apple-unveils-m3-m3-pro-and-m3-max-the-most-advanced-chips-for-a-personal-computer/
		ProcessNode:        "3-nanometer",
		MemoryBandwidthGBs: 100,
		NeuralEngineCores:  16,
		// https://support.apple.com/en-us/117735 (MacBook Pro 14-inch M3 tech specs)
		Media: Media{VideoDecodeEngines: 1, VideoEncodeEngines: 1, ProResEngines: 1, AV1Decode: true},
	},
	// Chip ID T6030: community source https://theapplewiki.com/wiki/T6030.
	"T6030": {
		ID:            "T6030",
		MarketingName: "Apple M3 Pro",
		ProcessNode:   "3-nanometer",
		// https://support.apple.com/en-us/117736 (MacBook Pro 14-inch M3 Pro/M3 Max tech specs)
		MemoryBandwidthGBs: 150,
		NeuralEngineCores:  16,
		Media:              Media{VideoDecodeEngines: 1, VideoEncodeEngines: 1, ProResEngines: 1, AV1Decode: true},
	},
	// Chip ID T6031: full die (16-core CPU / 40-core GPU) of the M3 Max.
	// Community source https://theapplewiki.com/wiki/T6030 family pages /
	// Asahi Linux "Initial Apple M3 Pro, Max and Ultra SoC support" patches.
	"T6031": {
		ID:            "T6031",
		MarketingName: "Apple M3 Max",
		ProcessNode:   "3-nanometer",
		// https://support.apple.com/en-us/117736: 400GB/s for the 16-core
		// CPU / 40-core GPU configuration.
		MemoryBandwidthGBs: 400,
		NeuralEngineCores:  16,
		Media:              Media{VideoDecodeEngines: 1, VideoEncodeEngines: 2, ProResEngines: 2, AV1Decode: true},
	},
	// Chip ID T6034: binned die (14-core CPU / 30-core GPU) of the M3 Max.
	// Same community source as T6031.
	"T6034": {
		ID:            "T6034",
		MarketingName: "Apple M3 Max",
		ProcessNode:   "3-nanometer",
		// https://support.apple.com/en-us/117736: 300GB/s for the 14-core
		// CPU / 30-core GPU configuration.
		MemoryBandwidthGBs: 300,
		NeuralEngineCores:  16,
		Media:              Media{VideoDecodeEngines: 1, VideoEncodeEngines: 2, ProResEngines: 2, AV1Decode: true},
	},
	// Chip ID T6032: community source https://theapplewiki.com/wiki/T6030
	// family / Asahi Linux patch series, cross-checked against lowendmac.com.
	"T6032": {
		ID:            "T6032",
		MarketingName: "Apple M3 Ultra",
		// ProcessNode not stated in the M3 Ultra announcement (built from
		// two fused M3 Max dies); see package doc.
		// https://www.apple.com/newsroom/2025/03/apple-reveals-m3-ultra-taking-apple-silicon-to-a-new-extreme/
		// "over 800GB/s" is Apple's own wording; stored as a floor, not a
		// precise figure.
		MemoryBandwidthGBs: 800,
		NeuralEngineCores:  32,
		// Same source: "four ProRes encode and decode engines"; separate
		// video decode/encode engine counts not itemized for M3 Ultra.
		Media: Media{ProResEngines: 4},
	},

	// M4 generation.
	// Chip ID T8132: community source https://theapplewiki.com/wiki/CHIP,
	// cross-checked against Asahi/kernel devicetree patches ("T8132/T604x").
	"T8132": {
		ID:            "T8132",
		MarketingName: "Apple M4",
		// https://www.apple.com/newsroom/2024/10/apple-introduces-m4-pro-and-m4-max/
		// (applies to the whole M4 family)
		ProcessNode: "3-nanometer (2nd generation)",
		// https://support.apple.com/en-us/121552 (MacBook Pro 14-inch M4 tech specs)
		MemoryBandwidthGBs: 120,
		NeuralEngineCores:  16,
		Media:              Media{VideoDecodeEngines: 1, VideoEncodeEngines: 1, ProResEngines: 1, AV1Decode: true},
	},
	// Chip ID T6040: community source https://theapplewiki.com/wiki/CHIP.
	"T6040": {
		ID:            "T6040",
		MarketingName: "Apple M4 Pro",
		// https://www.apple.com/newsroom/2024/10/apple-introduces-m4-pro-and-m4-max/
		ProcessNode: "3-nanometer (2nd generation)",
		// https://support.apple.com/en-us/121553 (MacBook Pro 14-inch M4 Pro/M4 Max tech specs)
		MemoryBandwidthGBs: 273,
		NeuralEngineCores:  16,
		Media:              Media{VideoDecodeEngines: 1, VideoEncodeEngines: 1, ProResEngines: 1, AV1Decode: true},
	},
	// Chip ID T6041: community source https://theapplewiki.com/wiki/CHIP.
	"T6041": {
		ID:            "T6041",
		MarketingName: "Apple M4 Max",
		// https://www.apple.com/newsroom/2024/10/apple-introduces-m4-pro-and-m4-max/
		ProcessNode: "3-nanometer (2nd generation)",
		// MemoryBandwidthGBs left at 0: Apple publishes two figures for
		// this chip ID depending on GPU core count -
		// https://support.apple.com/en-us/121553: 410GB/s (14-core
		// CPU/32-core GPU) or 546GB/s (16-core CPU/40-core GPU). There is
		// no collector yet that reads GPU core count to disambiguate.
		NeuralEngineCores: 16,
		Media:             Media{VideoDecodeEngines: 1, VideoEncodeEngines: 2, ProResEngines: 2, AV1Decode: true},
	},

	// M5 generation.
	// Chip ID T8142: community source https://theapplewiki.com/wiki/CHIP,
	// cross-checked against "asahi-m5" T8142 SoC docs.
	"T8142": {
		ID:            "T8142",
		MarketingName: "Apple M5",
		// https://www.apple.com/newsroom/2025/10/apple-unleashes-m5-the-next-big-leap-in-ai-performance-for-apple-silicon/
		ProcessNode: "3-nanometer (3rd generation)",
		// https://support.apple.com/en-us/125405 (MacBook Pro 14-inch M5 tech specs)
		MemoryBandwidthGBs: 153,
		NeuralEngineCores:  16,
		Media:              Media{VideoDecodeEngines: 1, VideoEncodeEngines: 1, ProResEngines: 1, AV1Decode: true},
	},
	// Chip ID T6050: community source https://theapplewiki.com/wiki/M5_Pro;
	// pre-launch beta string "T6050 / H17S" reported by MacRumors.
	"T6050": {
		ID:            "T6050",
		MarketingName: "Apple M5 Pro",
		// Apple's own description: two 3rd-generation 3-nanometer dies
		// joined with Apple's new "Fusion Architecture" packaging.
		// https://www.apple.com/newsroom/2026/03/apple-debuts-m5-pro-and-m5-max-to-supercharge-the-most-demanding-pro-workflows/
		ProcessNode: "3-nanometer (3rd generation), dual-die Fusion Architecture",
		// https://support.apple.com/en-us/126318 (MacBook Pro 14-inch M5 Pro/M5 Max tech specs)
		MemoryBandwidthGBs: 307,
		NeuralEngineCores:  16,
		Media:              Media{VideoDecodeEngines: 1, VideoEncodeEngines: 1, ProResEngines: 1, AV1Decode: true},
	},
	// Chip ID T6051: community source https://theapplewiki.com/wiki/M5_Max;
	// pre-launch beta string "T6051 / H17C".
	"T6051": {
		ID:            "T6051",
		MarketingName: "Apple M5 Max",
		// https://www.apple.com/newsroom/2026/03/apple-debuts-m5-pro-and-m5-max-to-supercharge-the-most-demanding-pro-workflows/
		ProcessNode: "3-nanometer (3rd generation), dual-die Fusion Architecture",
		// MemoryBandwidthGBs left at 0: Apple publishes two figures for
		// this chip ID depending on GPU core count -
		// https://support.apple.com/en-us/126318 and
		// https://support.apple.com/en-us/128107: 460GB/s (32-core GPU) or
		// 614GB/s (40-core GPU). No collector yet disambiguates by GPU core
		// count.
		NeuralEngineCores: 16,
		Media:             Media{VideoDecodeEngines: 1, VideoEncodeEngines: 2, ProResEngines: 2, AV1Decode: true},
	},
	// Chip ID T6052: pre-launch beta string "T6052 / H17D" reported by
	// MacRumors/TechSpot/Macworld; tracked by theapplewiki.com under Mac
	// Studio (M5 Ultra).
	"T6052": {
		ID:            "T6052",
		MarketingName: "Apple M5 Ultra",
		// ProcessNode not stated in the M5 Ultra announcement (a
		// Fusion/UltraFusion dual-die package); see package doc.
		// https://www.apple.com/newsroom/2026/08/apple-introduces-new-mac-studio-with-m5-max-and-m5-ultra/
		MemoryBandwidthGBs: 1200, // Apple: "1.2TB/s"
		NeuralEngineCores:  32,
		// https://support.apple.com/en-us/128107 (Mac Studio M5 Max/Ultra tech specs)
		Media: Media{VideoDecodeEngines: 2, VideoEncodeEngines: 4, ProResEngines: 4, AV1Decode: true},
	},

	// M6 generation. Announced 2026-08-25, released 2026-09-22 in the Mac
	// mini; no Pro/Max/Ultra variant as of this table's writing (2026-09-28)
	// -- Apple has said M6 Pro/Max are being skipped in favor of a later M7
	// generation.
	// Chip ID T8152: community source https://theapplewiki.com/wiki/M6.
	"T8152": {
		ID:            "T8152",
		MarketingName: "Apple M6",
		// https://www.apple.com/newsroom/2026/08/apple-introduces-m6-and-m5-ultra-for-a-big-leap-in-performance-and-ai-compute/
		// "built using cutting-edge 2 nm process technology" - first Apple
		// silicon chip on TSMC's 2nm process.
		ProcessNode: "2-nanometer",
		// MemoryBandwidthGBs left at 0: Apple's Mac mini tech-specs page
		// (https://www.apple.com/mac-mini/specs/) lists 153GB/s for two of
		// the three M6 configurations and 170GB/s for the third, tied to
		// unified memory capacity, not a distinct chip ID.
		//
		// NeuralEngineCores: Apple's own wording is "Dual 16-core Neural
		// Engine" (same source), not a plain single core count like every
		// earlier generation. Read literally as two 16-core Neural Engine
		// units, consistent with Apple's "up to 2x the peak compute of
		// previous generations" framing for the same feature.
		NeuralEngineCores: 32,
		// https://www.apple.com/mac-mini/specs/: "Hardware-accelerated
		// H.264, HEVC, ProRes, and ProRes RAW", one video decode engine,
		// one video encode engine, one ProRes encode/decode engine, AV1
		// decode.
		Media: Media{VideoDecodeEngines: 1, VideoEncodeEngines: 1, ProResEngines: 1, AV1Decode: true},
	},
}

// Lookup returns the known facts for a chip ID (e.g. "T8112", case
// insensitive). ok is false when the chip ID is not in the table.
func Lookup(id string) (Chip, bool) {
	c, ok := table[strings.ToUpper(id)]
	return c, ok
}
