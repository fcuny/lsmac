package source

import (
	"fmt"
	"strings"
)

const sysctlPath = "/usr/sbin/sysctl"

// Sysctl runs `sysctl -n <keys...>` and returns one trimmed value per key,
// in the order requested. It errors if the output does not have exactly
// len(keys) lines.
func Sysctl(cmd SystemCommand, keys ...string) ([]string, error) {
	args := append([]string{"-n"}, keys...)

	output, err := cmd.Execute(sysctlPath, args...)
	if err != nil {
		return nil, fmt.Errorf("sysctl %s: %w", strings.Join(keys, " "), err)
	}

	lines := strings.Split(strings.TrimRight(string(output), "\n"), "\n")
	if len(lines) != len(keys) {
		return nil, fmt.Errorf("sysctl %s: expected %d lines, got %d", strings.Join(keys, " "), len(keys), len(lines))
	}

	return lines, nil
}
