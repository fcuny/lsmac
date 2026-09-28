package source

import "os"

// FileChecker checks whether a path exists on disk, e.g. to detect an
// optional install like the Rosetta 2 runtime.
type FileChecker interface {
	Exists(path string) bool
}

// RealFileChecker checks the real filesystem.
type RealFileChecker struct{}

func (RealFileChecker) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
