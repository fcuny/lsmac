package render

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"fcuny.net/lsmac/internal/chips"
	"fcuny.net/lsmac/internal/collect"
	"fcuny.net/lsmac/internal/models"
)

func TestChip(t *testing.T) {
	var buf bytes.Buffer
	chip := chips.Chip{
		ID:            "T8112",
		MarketingName: "Apple M2",
		ProcessNode:   "5-nanometer (2nd generation)",
	}

	if err := Chip(&buf, chip); err != nil {
		t.Fatalf("Chip() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"Apple M2 (T8112)", "5-nanometer (2nd generation)"} {
		if !strings.Contains(out, want) {
			t.Errorf("Chip() output = %q, want it to contain %q", out, want)
		}
	}
	if strings.Contains(out, "GB/s") {
		t.Errorf("Chip() output = %q, want no bandwidth line (that's the Memory section's job now)", out)
	}
}

func TestChipUnknownFallsBackToID(t *testing.T) {
	var buf bytes.Buffer

	if err := Chip(&buf, chips.Chip{ID: "T9999"}); err != nil {
		t.Fatalf("Chip() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "T9999 (T9999)") {
		t.Errorf("Chip() output = %q, want it to fall back to the chip ID", out)
	}
	if strings.Contains(out, "GB/s") {
		t.Errorf("Chip() output = %q, want no bandwidth line when unknown", out)
	}
}

func TestCPU(t *testing.T) {
	var buf bytes.Buffer
	cpu := collect.CPU{BrandName: "Apple M2", TotalCores: 8, PerformanceCores: 4, EfficiencyCores: 4}

	if err := CPU(&buf, cpu); err != nil {
		t.Fatalf("CPU() error = %v", err)
	}

	if !strings.Contains(buf.String(), "8 cores: 4P + 4E") {
		t.Errorf("CPU() output = %q, want it to contain %q", buf.String(), "8 cores: 4P + 4E")
	}
}

func TestGPU(t *testing.T) {
	var buf bytes.Buffer

	if err := GPU(&buf, collect.GPU{CoreCount: 10}); err != nil {
		t.Fatalf("GPU() error = %v", err)
	}

	if !strings.Contains(buf.String(), "10 cores") {
		t.Errorf("GPU() output = %q, want it to contain %q", buf.String(), "10 cores")
	}
}

func TestMachine(t *testing.T) {
	var buf bytes.Buffer
	machine := collect.Machine{ModelIdentifier: "Mac14,2", SKU: "MN703LL/A"}
	model := models.Model{Identifier: "Mac14,2", MarketingName: "MacBook Air (M2, 2022)"}

	if err := Machine(&buf, machine, model); err != nil {
		t.Fatalf("Machine() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"MacBook Air (M2, 2022) (Mac14,2)", "MN703LL/A"} {
		if !strings.Contains(out, want) {
			t.Errorf("Machine() output = %q, want it to contain %q", out, want)
		}
	}
	if strings.Contains(out, "Serial") || strings.Contains(out, "UUID") {
		t.Errorf("Machine() output = %q, want no Serial/UUID lines when empty", out)
	}
}

func TestMachineUnknownFallsBackToIdentifier(t *testing.T) {
	var buf bytes.Buffer
	machine := collect.Machine{ModelIdentifier: "Mac99,99"}

	if err := Machine(&buf, machine, models.Model{}); err != nil {
		t.Fatalf("Machine() error = %v", err)
	}

	if !strings.Contains(buf.String(), "Mac99,99 (Mac99,99)") {
		t.Errorf("Machine() output = %q, want it to fall back to the model identifier", buf.String())
	}
}

func TestMachineShowsSerialWhenPresent(t *testing.T) {
	var buf bytes.Buffer
	machine := collect.Machine{ModelIdentifier: "Mac14,2", Serial: "Y9YH5WKX5V", HardwareUUID: "B2D5EAE2-0000-0000-0000-000000000000"}

	if err := Machine(&buf, machine, models.Model{}); err != nil {
		t.Fatalf("Machine() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"Y9YH5WKX5V", "B2D5EAE2-0000-0000-0000-000000000000"} {
		if !strings.Contains(out, want) {
			t.Errorf("Machine() output = %q, want it to contain %q", out, want)
		}
	}
}

func TestOS(t *testing.T) {
	var buf bytes.Buffer
	osInfo := collect.OS{
		ProductName:      "macOS",
		VersionName:      "Golden Gate",
		ProductVersion:   "27.0",
		Build:            "26A428",
		DarwinVersion:    "27.0.0",
		Uptime:           13*24*time.Hour + 15*time.Hour,
		RosettaInstalled: false,
	}

	if err := OS(&buf, osInfo); err != nil {
		t.Fatalf("OS() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"macOS 27.0 Golden Gate (26A428)", "27.0.0", "13 days, 15 hours", "not installed"} {
		if !strings.Contains(out, want) {
			t.Errorf("OS() output = %q, want it to contain %q", out, want)
		}
	}
}

func TestOSRosettaInstalled(t *testing.T) {
	var buf bytes.Buffer
	osInfo := collect.OS{ProductName: "macOS", ProductVersion: "27.0", Build: "26A428", RosettaInstalled: true}

	if err := OS(&buf, osInfo); err != nil {
		t.Fatalf("OS() error = %v", err)
	}
	if !strings.Contains(buf.String(), "Rosetta    installed") {
		t.Errorf("OS() output = %q, want %q", buf.String(), "Rosetta    installed")
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{13*24*time.Hour + 15*time.Hour, "13 days, 15 hours"},
		{24 * time.Hour, "1 day, 0 hours"},
		{90 * time.Minute, "1 hour, 30 minutes"},
		{5 * time.Minute, "5 minutes"},
		{1 * time.Minute, "1 minute"},
	}
	for _, tt := range tests {
		if got := formatUptime(tt.d); got != tt.want {
			t.Errorf("formatUptime(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestFirmware(t *testing.T) {
	var buf bytes.Buffer
	fw := collect.Firmware{Version: "mBoot-20457.1.29", SecureBoot: true, SIPEnabled: true}

	if err := Firmware(&buf, fw); err != nil {
		t.Fatalf("Firmware() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"mBoot-20457.1.29", "Secure Boot  enabled", "SIP        enabled"} {
		if !strings.Contains(out, want) {
			t.Errorf("Firmware() output = %q, want it to contain %q", out, want)
		}
	}
}

func TestFirmwareDisabled(t *testing.T) {
	var buf bytes.Buffer
	fw := collect.Firmware{Version: "mBoot-1.2.3"}

	if err := Firmware(&buf, fw); err != nil {
		t.Fatalf("Firmware() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"Secure Boot  disabled", "SIP        disabled"} {
		if !strings.Contains(out, want) {
			t.Errorf("Firmware() output = %q, want it to contain %q", out, want)
		}
	}
}

func TestCPUDetail(t *testing.T) {
	var buf bytes.Buffer
	chip := chips.Chip{ID: "T8112", MarketingName: "Apple M2"}
	detail := collect.CPUDetail{
		Family:   "0xDA33D83D",
		PageSize: 16384,
		Clusters: []collect.CPUCluster{
			{Name: "Performance", PhysicalCores: 4, Clusters: 1, CoresPerCluster: 4, L1ICacheSize: 196608, L1DCacheSize: 131072, L2CacheSize: 16777216},
			{Name: "Efficiency", PhysicalCores: 4, Clusters: 1, CoresPerCluster: 4, L1ICacheSize: 131072, L1DCacheSize: 65536, L2CacheSize: 4194304},
		},
		Features: []string{"CRC32", "FlagM", "BTI"},
	}

	if err := CPUDetail(&buf, chip, detail); err != nil {
		t.Fatalf("CPUDetail() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"Apple M2 (T8112)",
		"0xDA33D83D",
		"16 KiB",
		"Performance       4 cores  ·  1 cluster",
		"192 KiB / 128 KiB per core",
		"16 MiB per cluster (4 cores per L2)",
		"Efficiency        4 cores  ·  1 cluster",
		"CRC32 FlagM BTI",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("CPUDetail() output = %q, want it to contain %q", out, want)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		n    uint32
		want string
	}{
		{16384, "16 KiB"},
		{196608, "192 KiB"},
		{16777216, "16 MiB"},
		{4194304, "4 MiB"},
	}
	for _, tt := range tests {
		if got := formatBytes(tt.n); got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestWrapFeaturesWrapsLongLists(t *testing.T) {
	features := []string{
		"CRC32", "FlagM", "FlagM2", "FHM", "DotProd", "SHA3", "RDM", "LSE",
		"SHA256", "SHA512", "SHA1", "AES", "PMULL", "SB", "FRINTTS",
	}
	got := wrapFeatures(features)
	if !strings.Contains(got, "\n") {
		t.Errorf("wrapFeatures() = %q, want it to wrap onto more than one line", got)
	}
	// Every feature name must survive the wrap.
	for _, f := range features {
		if !strings.Contains(got, f) {
			t.Errorf("wrapFeatures() output missing %q", f)
		}
	}
}

func TestMemory(t *testing.T) {
	var buf bytes.Buffer
	mem := collect.Memory{
		TotalBytes: 16 << 30,
		UsedBytes:  14173412454, // ~13.2 GiB
		Type:       "LPDDR5",
		Pressure:   "normal",
	}

	if err := Memory(&buf, mem, 100); err != nil {
		t.Fatalf("Memory() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"13.2 GiB / 16 GiB", "Type       LPDDR5", "Bandwidth  100 GB/s", "Pressure   normal"} {
		if !strings.Contains(out, want) {
			t.Errorf("Memory() output = %q, want it to contain %q", out, want)
		}
	}
}

func TestMemoryOmitsUnknownFields(t *testing.T) {
	var buf bytes.Buffer
	mem := collect.Memory{TotalBytes: 16 << 30, UsedBytes: 8 << 30, Pressure: "normal"}

	if err := Memory(&buf, mem, 0); err != nil {
		t.Fatalf("Memory() error = %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "Type") || strings.Contains(out, "Bandwidth") {
		t.Errorf("Memory() output = %q, want no Type/Bandwidth lines when unknown", out)
	}
}

func TestFormatGiB(t *testing.T) {
	tests := []struct {
		n    uint64
		want string
	}{
		{16 << 30, "16 GiB"},
		{8 << 30, "8 GiB"},
	}
	for _, tt := range tests {
		if got := formatGiB(tt.n); got != tt.want {
			t.Errorf("formatGiB(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
	if got := formatGiB(13*(1<<30) + (1 << 29)); got != "13.5 GiB" {
		t.Errorf("formatGiB(13.5 GiB) = %q, want %q", got, "13.5 GiB")
	}
}

func TestStorage(t *testing.T) {
	var buf bytes.Buffer
	fileVaultOn := true
	storage := collect.Storage{
		Model:       "APPLE SSD AP1024Z",
		TotalBytes:  994662584320,
		UsedBytes:   434191556608,
		FileVaultOn: &fileVaultOn,
	}

	if err := Storage(&buf, storage); err != nil {
		t.Fatalf("Storage() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"APPLE SSD AP1024Z", "FileVault  on"} {
		if !strings.Contains(out, want) {
			t.Errorf("Storage() output = %q, want it to contain %q", out, want)
		}
	}
}

func TestStorageFileVaultOff(t *testing.T) {
	var buf bytes.Buffer
	fileVaultOff := false
	storage := collect.Storage{Model: "APPLE SSD AP1024Z", TotalBytes: 1000, UsedBytes: 500, FileVaultOn: &fileVaultOff}

	if err := Storage(&buf, storage); err != nil {
		t.Fatalf("Storage() error = %v", err)
	}
	if !strings.Contains(buf.String(), "FileVault  off") {
		t.Errorf("Storage() output = %q, want %q", buf.String(), "FileVault  off")
	}
}

func TestStorageFileVaultOmittedWhenNotChecked(t *testing.T) {
	var buf bytes.Buffer
	storage := collect.Storage{Model: "APPLE SSD AP1024Z", TotalBytes: 1000, UsedBytes: 500}

	if err := Storage(&buf, storage); err != nil {
		t.Fatalf("Storage() error = %v", err)
	}
	if strings.Contains(buf.String(), "FileVault") {
		t.Errorf("Storage() output = %q, want no FileVault line when FileVaultOn is nil", buf.String())
	}
}

func TestPower(t *testing.T) {
	var buf bytes.Buffer
	power := collect.Power{
		Percentage:    80,
		Charging:      false,
		CycleCount:    201,
		HealthPercent: 93,
		ThermalState:  "nominal",
		LowPowerMode:  false,
	}

	if err := Power(&buf, power); err != nil {
		t.Fatalf("Power() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"80%, discharging", "Cycles     201", "Health     93%", "Thermal    nominal", "Low Power Mode  off"} {
		if !strings.Contains(out, want) {
			t.Errorf("Power() output = %q, want it to contain %q", out, want)
		}
	}
}

func TestPowerChargingAndLowPowerMode(t *testing.T) {
	var buf bytes.Buffer
	power := collect.Power{Percentage: 50, Charging: true, ThermalState: "nominal", LowPowerMode: true}

	if err := Power(&buf, power); err != nil {
		t.Fatalf("Power() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"50%, charging", "Low Power Mode  on"} {
		if !strings.Contains(out, want) {
			t.Errorf("Power() output = %q, want it to contain %q", out, want)
		}
	}
}

func TestPowerOmitsThermalAndLowPowerModeWhenNotChecked(t *testing.T) {
	var buf bytes.Buffer
	power := collect.Power{Percentage: 50, Charging: true}

	if err := Power(&buf, power); err != nil {
		t.Fatalf("Power() error = %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "Thermal") || strings.Contains(out, "Low Power Mode") {
		t.Errorf("Power() output = %q, want no Thermal/Low Power Mode lines when ThermalState is empty", out)
	}
}

func TestIO(t *testing.T) {
	var buf bytes.Buffer
	ioInfo := collect.IO{Interfaces: []collect.NetworkInterface{
		{Name: "lo0", IsUp: true, IsLoopback: true},
		{Name: "en0", HardwareAddr: "c4:35:d9:89:5c:6c", IsUp: true},
		{Name: "en1", IsUp: false},
	}}

	if err := IO(&buf, ioInfo); err != nil {
		t.Fatalf("IO() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"lo0", "en0        up   c4:35:d9:89:5c:6c", "en1        down"} {
		if !strings.Contains(out, want) {
			t.Errorf("IO() output = %q, want it to contain %q", out, want)
		}
	}
}
