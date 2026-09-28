package collect

import "testing"

func cpuDetailFixture() cmdRouter {
	return cmdRouter{outputs: map[string]string{
		"/usr/sbin/sysctl -n hw.cpufamily hw.pagesize hw.nperflevels": "-634136515\n16384\n2",
		"/usr/sbin/sysctl -n hw.perflevel0.name hw.perflevel0.physicalcpu hw.perflevel0.cpusperl2 hw.perflevel0.l1icachesize hw.perflevel0.l1dcachesize hw.perflevel0.l2cachesize": "Performance\n4\n4\n196608\n131072\n16777216",
		"/usr/sbin/sysctl -n hw.perflevel1.name hw.perflevel1.physicalcpu hw.perflevel1.cpusperl2 hw.perflevel1.l1icachesize hw.perflevel1.l1dcachesize hw.perflevel1.l2cachesize": "Efficiency\n4\n4\n131072\n65536\n4194304",
		"/usr/sbin/sysctl hw.optional.arm": "hw.optional.arm.FEAT_FP8: 0\nhw.optional.arm.FEAT_CRC32: 1\nhw.optional.arm.FEAT_FlagM: 1\nhw.optional.arm.FEAT_BTI: 1\n",
	}}
}

func TestCollectCPUDetail(t *testing.T) {
	got, err := CollectCPUDetail(cpuDetailFixture())
	if err != nil {
		t.Fatalf("CollectCPUDetail() error = %v", err)
	}

	if got.Family != "0xDA33D83D" {
		t.Errorf("Family = %q, want %q", got.Family, "0xDA33D83D")
	}
	if got.PageSize != 16384 {
		t.Errorf("PageSize = %d, want 16384", got.PageSize)
	}
	if len(got.Clusters) != 2 {
		t.Fatalf("len(Clusters) = %d, want 2", len(got.Clusters))
	}

	perf := got.Clusters[0]
	want := CPUCluster{
		Name: "Performance", PhysicalCores: 4, Clusters: 1, CoresPerCluster: 4,
		L1ICacheSize: 196608, L1DCacheSize: 131072, L2CacheSize: 16777216,
	}
	if perf != want {
		t.Errorf("Clusters[0] = %+v, want %+v", perf, want)
	}

	eff := got.Clusters[1]
	if eff.Name != "Efficiency" || eff.Clusters != 1 || eff.L2CacheSize != 4194304 {
		t.Errorf("Clusters[1] = %+v, unexpected", eff)
	}

	wantFeatures := []string{"CRC32", "FlagM", "BTI"}
	if len(got.Features) != len(wantFeatures) {
		t.Fatalf("Features = %v, want %v", got.Features, wantFeatures)
	}
	for i, f := range wantFeatures {
		if got.Features[i] != f {
			t.Errorf("Features[%d] = %q, want %q", i, got.Features[i], f)
		}
	}
}

// A chip with more than two performance levels (e.g. an M6-style extra
// "super core" tier) must be handled without special-casing P/E.
func TestCollectCPUDetailThreeLevels(t *testing.T) {
	fixture := cmdRouter{outputs: map[string]string{
		"/usr/sbin/sysctl -n hw.cpufamily hw.pagesize hw.nperflevels": "1\n16384\n3",
		"/usr/sbin/sysctl -n hw.perflevel0.name hw.perflevel0.physicalcpu hw.perflevel0.cpusperl2 hw.perflevel0.l1icachesize hw.perflevel0.l1dcachesize hw.perflevel0.l2cachesize": "Super\n2\n2\n1\n1\n1",
		"/usr/sbin/sysctl -n hw.perflevel1.name hw.perflevel1.physicalcpu hw.perflevel1.cpusperl2 hw.perflevel1.l1icachesize hw.perflevel1.l1dcachesize hw.perflevel1.l2cachesize": "Performance\n4\n4\n1\n1\n1",
		"/usr/sbin/sysctl -n hw.perflevel2.name hw.perflevel2.physicalcpu hw.perflevel2.cpusperl2 hw.perflevel2.l1icachesize hw.perflevel2.l1dcachesize hw.perflevel2.l2cachesize": "Efficiency\n6\n6\n1\n1\n1",
		"/usr/sbin/sysctl hw.optional.arm": "",
	}}

	got, err := CollectCPUDetail(fixture)
	if err != nil {
		t.Fatalf("CollectCPUDetail() error = %v", err)
	}
	if len(got.Clusters) != 3 {
		t.Fatalf("len(Clusters) = %d, want 3", len(got.Clusters))
	}
	names := []string{got.Clusters[0].Name, got.Clusters[1].Name, got.Clusters[2].Name}
	want := []string{"Super", "Performance", "Efficiency"}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("Clusters[%d].Name = %q, want %q", i, names[i], want[i])
		}
	}
}
