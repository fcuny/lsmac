package collect

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fcuny.net/lsmac/internal/source"
)

const rosettaRuntimePath = "/Library/Apple/usr/libexec/oah/libRosettaRuntime"

// versionNames maps a macOS major version number to its marketing name.
// Sourced from https://support.apple.com/en-us/109033. Apple jumped
// straight from 15 (Sequoia) to 26 (Tahoe) in 2025 to align the major
// version number with the release year; 16-25 were never used.
var versionNames = map[int]string{
	11: "Big Sur",
	12: "Monterey",
	13: "Ventura",
	14: "Sonoma",
	15: "Sequoia",
	26: "Tahoe",
	27: "Golden Gate",
}

// OS holds the facts collected for the OS section.
type OS struct {
	ProductName      string
	VersionName      string // empty if the major version isn't in versionNames
	ProductVersion   string
	Build            string
	DarwinVersion    string
	Uptime           time.Duration
	RosettaInstalled bool
}

var bootTimeSecRE = regexp.MustCompile(`sec = (\d+)`)

// CollectOS reads OS product/version/build info via sw_vers, the Darwin
// kernel release via sysctl, uptime via kern.boottime, and whether the
// Rosetta 2 runtime is installed by checking for its runtime file on disk.
func CollectOS(cmd source.SystemCommand, fc source.FileChecker) (OS, error) {
	swVers, err := source.Run(cmd, "/usr/bin/sw_vers")
	if err != nil {
		return OS{}, fmt.Errorf("sw_vers: %w", err)
	}
	fields := make(map[string]string)
	for line := range strings.SplitSeq(swVers, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		fields[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	productName, productVersion, build := fields["ProductName"], fields["ProductVersion"], fields["BuildVersion"]
	if productName == "" || productVersion == "" || build == "" {
		return OS{}, fmt.Errorf("sw_vers: missing ProductName/ProductVersion/BuildVersion in output: %q", swVers)
	}

	darwinVersion, err := source.Sysctl(cmd, "kern.osrelease")
	if err != nil {
		return OS{}, err
	}

	boottime, err := source.Sysctl(cmd, "kern.boottime")
	if err != nil {
		return OS{}, err
	}
	m := bootTimeSecRE.FindStringSubmatch(boottime[0])
	if m == nil {
		return OS{}, fmt.Errorf("kern.boottime: unexpected format: %q", boottime[0])
	}
	sec, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return OS{}, fmt.Errorf("kern.boottime: %w", err)
	}

	majorStr, _, _ := strings.Cut(productVersion, ".")
	major, _ := strconv.Atoi(majorStr)

	return OS{
		ProductName:      productName,
		VersionName:      versionNames[major],
		ProductVersion:   productVersion,
		Build:            build,
		DarwinVersion:    darwinVersion[0],
		Uptime:           time.Since(time.Unix(sec, 0)),
		RosettaInstalled: fc.Exists(rosettaRuntimePath),
	}, nil
}
