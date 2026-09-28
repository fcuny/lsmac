// Package source wraps the external commands lsmac reads hardware and
// system data from (sysctl, ioreg, ...) behind a single interface, so
// collectors can be tested against captured fixtures instead of the real
// machine.
package source

import (
	"os/exec"
	"strings"
	"sync"
)

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

// CachingCommand wraps a SystemCommand and memoizes results by the exact
// command line, so independent collectors that happen to query the same
// slow object (e.g. two sections both reading IOPlatformExpertDevice) only
// pay for it once per run. This is safe because lsmac is a short-lived,
// single-shot process: results never need to be fresher than "as of when
// this run started".
type CachingCommand struct {
	inner SystemCommand
	mu    sync.Mutex
	cache map[string]cachedResult
}

type cachedResult struct {
	output []byte
	err    error
}

// NewCachingCommand wraps inner with a cache.
func NewCachingCommand(inner SystemCommand) *CachingCommand {
	return &CachingCommand{inner: inner, cache: make(map[string]cachedResult)}
}

func (c *CachingCommand) Execute(binary string, args ...string) ([]byte, error) {
	key := strings.Join(append([]string{binary}, args...), " ")

	c.mu.Lock()
	r, ok := c.cache[key]
	c.mu.Unlock()
	if ok {
		return r.output, r.err
	}

	output, err := c.inner.Execute(binary, args...)

	c.mu.Lock()
	c.cache[key] = cachedResult{output, err}
	c.mu.Unlock()
	return output, err
}
