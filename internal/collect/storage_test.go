package collect

import (
	"errors"
	"testing"
)

type fakeFilesystemStats struct {
	total, available uint64
	err              error
}

func (f fakeFilesystemStats) Stat(path string) (total, available uint64, err error) {
	return f.total, f.available, f.err
}

func storageFixture(fileVaultOutput string) cmdRouter {
	return cmdRouter{outputs: map[string]string{
		"/usr/sbin/ioreg -rc AppleANS3CGv2Controller -d1": "+-o AppleANS3CGv2Controller\n    {\n      \"Model Number\" = \"APPLE SSD AP1024Z\"\n    }\n",
		"/usr/bin/fdesetup status":                        fileVaultOutput,
	}}
}

func TestCollectStorage(t *testing.T) {
	fs := fakeFilesystemStats{total: 994662584320, available: 560475308032}

	got, err := CollectStorage(storageFixture("FileVault is On."), fs, true)
	if err != nil {
		t.Fatalf("CollectStorage() error = %v", err)
	}

	if got.Model != "APPLE SSD AP1024Z" {
		t.Errorf("Model = %q, want %q", got.Model, "APPLE SSD AP1024Z")
	}
	if got.TotalBytes != 994662584320 {
		t.Errorf("TotalBytes = %d, want %d", got.TotalBytes, 994662584320)
	}
	wantUsed := uint64(994662584320 - 560475308032)
	if got.UsedBytes != wantUsed {
		t.Errorf("UsedBytes = %d, want %d", got.UsedBytes, wantUsed)
	}
	if got.FileVaultOn == nil || !*got.FileVaultOn {
		t.Errorf("FileVaultOn = %v, want a pointer to true", got.FileVaultOn)
	}
}

func TestCollectStorageFileVaultOff(t *testing.T) {
	fs := fakeFilesystemStats{total: 1000, available: 500}

	got, err := CollectStorage(storageFixture("FileVault is Off."), fs, true)
	if err != nil {
		t.Fatalf("CollectStorage() error = %v", err)
	}
	if got.FileVaultOn == nil || *got.FileVaultOn {
		t.Errorf("FileVaultOn = %v, want a pointer to false", got.FileVaultOn)
	}
}

func TestCollectStorageSkipsFileVaultByDefault(t *testing.T) {
	fs := fakeFilesystemStats{total: 1000, available: 500}

	// No fdesetup fixture entry: if CollectStorage called it anyway, the
	// mock would error on the unrecognized command line.
	fixture := cmdRouter{outputs: map[string]string{
		"/usr/sbin/ioreg -rc AppleANS3CGv2Controller -d1": "+-o AppleANS3CGv2Controller\n    {\n      \"Model Number\" = \"APPLE SSD AP1024Z\"\n    }\n",
	}}

	got, err := CollectStorage(fixture, fs, false)
	if err != nil {
		t.Fatalf("CollectStorage() error = %v", err)
	}
	if got.FileVaultOn != nil {
		t.Errorf("FileVaultOn = %v, want nil when checkFileVault is false", got.FileVaultOn)
	}
}

func TestCollectStorageUnexpectedFileVaultOutput(t *testing.T) {
	fs := fakeFilesystemStats{total: 1000, available: 500}

	if _, err := CollectStorage(storageFixture("garbage"), fs, true); err == nil {
		t.Fatal("CollectStorage() error = nil, want error for unrecognized fdesetup output")
	}
}

func TestCollectStorageStatfsError(t *testing.T) {
	fs := fakeFilesystemStats{err: errors.New("statfs failed")}

	if _, err := CollectStorage(storageFixture("FileVault is On."), fs, true); err == nil {
		t.Fatal("CollectStorage() error = nil, want error propagated from statfs")
	}
}
