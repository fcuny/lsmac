// Package source wraps the external commands lsmac reads hardware and
// system data from (sysctl, ioreg, ...) behind a single interface, so
// collectors can be tested against captured fixtures instead of the real
// machine.
package source

import "os/exec"

// SystemCommand runs an external binary and returns its stdout. Collectors
// depend on this interface, never on os/exec directly, so tests can replay
// fixture output.
type SystemCommand interface {
	Execute(binary string, args ...string) ([]byte, error)
}

// RealCommand runs commands through the OS.
type RealCommand struct{}

func (RealCommand) Execute(binary string, args ...string) ([]byte, error) {
	return exec.Command(binary, args...).Output()
}
