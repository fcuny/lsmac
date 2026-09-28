package source

import "strings"

// Run executes binary with args and returns its trimmed stdout. It's for
// commands whose output collectors parse themselves (e.g. csrutil status),
// as opposed to sysctl/ioreg which have dedicated helpers above.
func Run(cmd SystemCommand, binary string, args ...string) (string, error) {
	output, err := cmd.Execute(binary, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
