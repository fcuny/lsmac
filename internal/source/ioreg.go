package source

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

const ioregPath = "/usr/sbin/ioreg"

// propertyLine matches a single `"key" = value` line from ioreg's default
// (non-plist) property dump, e.g. `"gpu-core-count" = 10`.
var propertyLine = func(line string) (key, value string, ok bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, `"`) {
		return "", "", false
	}
	rest := line[1:]
	end := strings.Index(rest, `"`)
	if end < 0 {
		return "", "", false
	}
	key = rest[:end]
	rest = strings.TrimSpace(rest[end+1:])
	rest = strings.TrimPrefix(rest, "=")
	return key, strings.TrimSpace(rest), true
}

// IORegProperties runs `ioreg -rc <class> -d1` and returns the scalar
// properties of the first matching object, keyed by property name. Values
// keep their raw ioreg formatting (quoted strings still quoted); use
// IntProperty / StringProperty to decode them.
func IORegProperties(cmd SystemCommand, class string) (map[string]string, error) {
	output, err := cmd.Execute(ioregPath, "-rc", class, "-d1")
	if err != nil {
		return nil, fmt.Errorf("ioreg -rc %s: %w", class, err)
	}

	props := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	inObject := false
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "{":
			if inObject {
				return props, nil // nested object: stop, we only read scalars
			}
			inObject = true
		case trimmed == "}":
			if inObject {
				return props, nil
			}
		case inObject:
			if key, value, ok := propertyLine(line); ok {
				props[key] = value
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ioreg -rc %s: reading output: %w", class, err)
	}

	if !inObject {
		return nil, fmt.Errorf("ioreg -rc %s: no matching object found", class)
	}
	return props, nil
}

// IntProperty decodes a raw ioreg property value as an integer.
func IntProperty(props map[string]string, key string) (int64, error) {
	raw, ok := props[key]
	if !ok {
		return 0, fmt.Errorf("property %q not found", key)
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("property %q: not an integer: %q", key, raw)
	}
	return n, nil
}

// StringProperty decodes a raw ioreg property value as a string, stripping
// the surrounding quotes ioreg prints for string values.
func StringProperty(props map[string]string, key string) (string, error) {
	raw, ok := props[key]
	if !ok {
		return "", fmt.Errorf("property %q not found", key)
	}
	return strings.Trim(raw, `"`), nil
}

// DataProperty decodes a raw ioreg "data" property (printed as
// `<68657820...>`) as a NUL-terminated ASCII string. Device tree properties
// such as platform-name are encoded this way.
func DataProperty(props map[string]string, key string) (string, error) {
	raw, ok := props[key]
	if !ok {
		return "", fmt.Errorf("property %q not found", key)
	}
	raw = strings.TrimSuffix(strings.TrimPrefix(raw, "<"), ">")

	decoded, err := hex.DecodeString(raw)
	if err != nil {
		return "", fmt.Errorf("property %q: not hex data: %q", key, raw)
	}

	if i := bytes.IndexByte(decoded, 0); i >= 0 {
		decoded = decoded[:i]
	}
	return string(decoded), nil
}
