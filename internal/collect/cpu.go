package collect

import (
	"fmt"
	"strconv"
	"strings"

	"fcuny.net/lsmac/internal/source"
)

// CPUClusterCores names one performance-level tier and its core count, for
// the compact CPU line. Every real Apple Silicon Mac shipped so far has
// exactly two - "Performance" and "Efficiency" - but this doesn't assume
// that: environments like GitHub's macOS Actions runners (a virtualized
// Apple Silicon environment) report a single homogeneous "Standard" tier
// via hw.nperflevels=1, discovered by running lsmac there in CI.
type CPUClusterCores struct {
	Name  string `json:"name"`
	Cores uint16 `json:"cores"`
}

// CPU holds the CPU facts collected for the SoC section.
type CPU struct {
	BrandName  string            `json:"brandName"`
	TotalCores uint16            `json:"totalCores"`
	Clusters   []CPUClusterCores `json:"clusters"`
}

// CollectCPU reads CPU brand, total core count, and each performance-level
// tier's name and core count via sysctl.
//
// It queries hw.nperflevels first, then all levels' name/logicalcpu keys in
// one further batched call, rather than assuming exactly two tiers
// (hw.perflevel0 = performance, hw.perflevel1 = efficiency): the original
// socinfo code assumed exactly that pair and swapped them when assigning to
// PCoreCount/ECoreCount, and assuming the pair exists at all breaks outright
// on a single-tier machine (see the CPUClusterCores doc comment). This still
// costs only two sysctl calls total, regardless of how many tiers exist.
func CollectCPU(cmd source.SystemCommand) (CPU, error) {
	head, err := source.Sysctl(cmd, "machdep.cpu.brand_string", "machdep.cpu.core_count", "hw.nperflevels")
	if err != nil {
		return CPU{}, err
	}

	total, err := strconv.ParseUint(head[1], 10, 16)
	if err != nil {
		return CPU{}, fmt.Errorf("machdep.cpu.core_count: %w", err)
	}
	nperflevels, err := strconv.Atoi(head[2])
	if err != nil {
		return CPU{}, fmt.Errorf("hw.nperflevels: %w", err)
	}

	keys := make([]string, 0, nperflevels*2)
	for i := range nperflevels {
		keys = append(keys, fmt.Sprintf("hw.perflevel%d.name", i), fmt.Sprintf("hw.perflevel%d.logicalcpu", i))
	}
	values, err := source.Sysctl(cmd, keys...)
	if err != nil {
		return CPU{}, err
	}

	clusters := make([]CPUClusterCores, nperflevels)
	for i := range nperflevels {
		cores, err := strconv.ParseUint(values[i*2+1], 10, 16)
		if err != nil {
			return CPU{}, fmt.Errorf("hw.perflevel%d.logicalcpu: %w", i, err)
		}
		clusters[i] = CPUClusterCores{Name: values[i*2], Cores: uint16(cores)}
	}

	return CPU{
		BrandName:  head[0],
		TotalCores: uint16(total),
		Clusters:   clusters,
	}, nil
}

// CPUCluster describes one performance-level cluster's core counts and
// cache sizes. Apple Silicon so far has two levels (Performance and
// Efficiency), but this loops over however many hw.nperflevels reports, so
// a chip with more tiers (e.g. the M6's extra "super" core) is handled
// without a code change - CollectCPUDetail doesn't hardcode P/E.
type CPUCluster struct {
	Name            string `json:"name"` // e.g. "Performance", "Efficiency" - from hw.perflevelN.name
	PhysicalCores   uint16 `json:"physicalCores"`
	Clusters        uint16 `json:"clusters"`        // L2-sharing groups: physicalcpu / cpusperl2
	CoresPerCluster uint16 `json:"coresPerCluster"` // hw.perflevelN.cpusperl2
	L1ICacheSize    uint32 `json:"l1iCacheSize"`    // bytes, per core
	L1DCacheSize    uint32 `json:"l1dCacheSize"`    // bytes, per core
	L2CacheSize     uint32 `json:"l2CacheSize"`     // bytes, per cluster (not summed across clusters)
}

// CPUDetail holds the facts for the detailed `--section cpu` view.
//
// Maximum frequency per cluster is deliberately not collected: Apple
// Silicon doesn't expose it via sysctl. It exists as raw pmgr
// "voltage-states" tables in the device tree, but decoding those correctly
// requires knowing which voltage-statesN property belongs to which
// cluster - a mapping that's specific to each chip generation and isn't
// derivable from the property names themselves. That's out of scope here.
type CPUDetail struct {
	Family   string       `json:"family"`   // e.g. "0xDA33D83D", from hw.cpufamily
	PageSize uint32       `json:"pageSize"` // bytes, from hw.pagesize
	Clusters []CPUCluster `json:"clusters"`
	Features []string     `json:"features"` // active hw.optional.arm.FEAT_* names, FEAT_ prefix stripped, in sysctl's own listing order
}

// CollectCPUDetail reads the per-cluster core/cache breakdown, CPU family,
// page size, and active ARM architecture feature flags.
func CollectCPUDetail(cmd source.SystemCommand) (CPUDetail, error) {
	head, err := source.Sysctl(cmd, "hw.cpufamily", "hw.pagesize", "hw.nperflevels")
	if err != nil {
		return CPUDetail{}, err
	}

	family, err := strconv.ParseInt(head[0], 10, 64)
	if err != nil {
		return CPUDetail{}, fmt.Errorf("hw.cpufamily: %w", err)
	}
	pageSize, err := strconv.ParseUint(head[1], 10, 32)
	if err != nil {
		return CPUDetail{}, fmt.Errorf("hw.pagesize: %w", err)
	}
	nperflevels, err := strconv.Atoi(head[2])
	if err != nil {
		return CPUDetail{}, fmt.Errorf("hw.nperflevels: %w", err)
	}

	clusters := make([]CPUCluster, nperflevels)
	for i := range nperflevels {
		prefix := fmt.Sprintf("hw.perflevel%d.", i)
		values, err := source.Sysctl(cmd,
			prefix+"name",
			prefix+"physicalcpu",
			prefix+"cpusperl2",
			prefix+"l1icachesize",
			prefix+"l1dcachesize",
			prefix+"l2cachesize",
		)
		if err != nil {
			return CPUDetail{}, err
		}

		physicalCores, err := strconv.ParseUint(values[1], 10, 16)
		if err != nil {
			return CPUDetail{}, fmt.Errorf("%sphysicalcpu: %w", prefix, err)
		}
		coresPerCluster, err := strconv.ParseUint(values[2], 10, 16)
		if err != nil {
			return CPUDetail{}, fmt.Errorf("%scpusperl2: %w", prefix, err)
		}
		l1i, err := strconv.ParseUint(values[3], 10, 32)
		if err != nil {
			return CPUDetail{}, fmt.Errorf("%sl1icachesize: %w", prefix, err)
		}
		l1d, err := strconv.ParseUint(values[4], 10, 32)
		if err != nil {
			return CPUDetail{}, fmt.Errorf("%sl1dcachesize: %w", prefix, err)
		}
		l2, err := strconv.ParseUint(values[5], 10, 32)
		if err != nil {
			return CPUDetail{}, fmt.Errorf("%sl2cachesize: %w", prefix, err)
		}

		var clusterCount uint16
		if coresPerCluster > 0 {
			clusterCount = uint16(physicalCores / coresPerCluster)
		}

		clusters[i] = CPUCluster{
			Name:            values[0],
			PhysicalCores:   uint16(physicalCores),
			Clusters:        clusterCount,
			CoresPerCluster: uint16(coresPerCluster),
			L1ICacheSize:    uint32(l1i),
			L1DCacheSize:    uint32(l1d),
			L2CacheSize:     uint32(l2),
		}
	}

	features, err := collectARMFeatures(cmd)
	if err != nil {
		return CPUDetail{}, err
	}

	return CPUDetail{
		Family:   fmt.Sprintf("0x%08X", uint32(family)),
		PageSize: uint32(pageSize),
		Clusters: clusters,
		Features: features,
	}, nil
}

const armFeaturesPrefix = "hw.optional.arm."

// collectARMFeatures reads the hw.optional.arm.FEAT_* namespace directly
// (rather than the full sysctl -a dump) and returns the names of the
// features that are actually active (value 1), FEAT_ prefix stripped, in
// the order sysctl lists them.
func collectARMFeatures(cmd source.SystemCommand) ([]string, error) {
	output, err := source.Run(cmd, "/usr/sbin/sysctl", "hw.optional.arm")
	if err != nil {
		return nil, fmt.Errorf("sysctl hw.optional.arm: %w", err)
	}

	var features []string
	for line := range strings.SplitSeq(output, "\n") {
		key, value, ok := strings.Cut(line, ": ")
		if !ok || value != "1" {
			continue
		}
		name := strings.TrimPrefix(key, armFeaturesPrefix)
		name = strings.TrimPrefix(name, "FEAT_")
		features = append(features, name)
	}
	return features, nil
}
