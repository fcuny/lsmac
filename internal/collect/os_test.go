package collect

import (
	"fmt"
	"strings"
	"testing"
)

// cmdRouter dispatches Execute by the full command line (binary + args),
// for collectors that shell out to more than one distinct command.
type cmdRouter struct {
	outputs map[string]string
	err     error
}

func (r cmdRouter) Execute(binary string, args ...string) ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	key := strings.Join(append([]string{binary}, args...), " ")
	out, ok := r.outputs[key]
	if !ok {
		return nil, fmt.Errorf("no mock output for %q", key)
	}
	return []byte(out), nil
}

type fakeFileChecker struct{ exists bool }

func (f fakeFileChecker) Exists(string) bool { return f.exists }

func osFixture() cmdRouter {
	return cmdRouter{outputs: map[string]string{
		"/usr/bin/sw_vers":                   "ProductName:\t\tmacOS\nProductVersion:\t\t27.0\nBuildVersion:\t\t26A428\n",
		"/usr/sbin/sysctl -n kern.osrelease": "27.0.0",
		"/usr/sbin/sysctl -n kern.boottime":  "{ sec = 1789435651, usec = 462663 } Mon Sep 14 18:27:31 2026",
	}}
}

func TestCollectOS(t *testing.T) {
	got, err := CollectOS(osFixture(), fakeFileChecker{exists: false})
	if err != nil {
		t.Fatalf("CollectOS() error = %v", err)
	}

	if got.ProductName != "macOS" {
		t.Errorf("ProductName = %q, want %q", got.ProductName, "macOS")
	}
	if got.ProductVersion != "27.0" {
		t.Errorf("ProductVersion = %q, want %q", got.ProductVersion, "27.0")
	}
	if got.Build != "26A428" {
		t.Errorf("Build = %q, want %q", got.Build, "26A428")
	}
	if got.VersionName != "Golden Gate" {
		t.Errorf("VersionName = %q, want %q", got.VersionName, "Golden Gate")
	}
	if got.DarwinVersion != "27.0.0" {
		t.Errorf("DarwinVersion = %q, want %q", got.DarwinVersion, "27.0.0")
	}
	if got.Uptime <= 0 {
		t.Errorf("Uptime = %v, want positive", got.Uptime)
	}
	if got.UptimeSeconds != got.Uptime.Seconds() {
		t.Errorf("UptimeSeconds = %v, want %v (Uptime.Seconds())", got.UptimeSeconds, got.Uptime.Seconds())
	}
	if got.RosettaInstalled {
		t.Error("RosettaInstalled = true, want false")
	}
}

func TestCollectOSRosettaInstalled(t *testing.T) {
	got, err := CollectOS(osFixture(), fakeFileChecker{exists: true})
	if err != nil {
		t.Fatalf("CollectOS() error = %v", err)
	}
	if !got.RosettaInstalled {
		t.Error("RosettaInstalled = false, want true")
	}
}

func TestCollectOSUnknownVersionName(t *testing.T) {
	fixture := osFixture()
	fixture.outputs["/usr/bin/sw_vers"] = "ProductName:\t\tmacOS\nProductVersion:\t\t99.0\nBuildVersion:\t\t99A1\n"

	got, err := CollectOS(fixture, fakeFileChecker{exists: false})
	if err != nil {
		t.Fatalf("CollectOS() error = %v", err)
	}
	if got.VersionName != "" {
		t.Errorf("VersionName = %q, want empty for unknown major version", got.VersionName)
	}
}

func TestCollectOSMissingFields(t *testing.T) {
	fixture := osFixture()
	fixture.outputs["/usr/bin/sw_vers"] = "ProductName:\t\tmacOS\n"

	if _, err := CollectOS(fixture, fakeFileChecker{}); err == nil {
		t.Fatal("CollectOS() error = nil, want error for incomplete sw_vers output")
	}
}
