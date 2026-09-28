package source

import "syscall"

// FilesystemStats reports the total and available space on the filesystem
// containing path.
type FilesystemStats interface {
	Stat(path string) (total, available uint64, err error)
}

// RealFilesystemStats reads the real filesystem via statfs(2).
type RealFilesystemStats struct{}

func (RealFilesystemStats) Stat(path string) (total, available uint64, err error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, 0, err
	}
	// Bavail (blocks available to non-superuser) rather than Bfree, so
	// UsedBytes = total - available matches what df reports.
	return uint64(st.Bsize) * st.Blocks, uint64(st.Bsize) * st.Bavail, nil
}
