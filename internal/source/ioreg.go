package source

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const ioregPath = "/usr/sbin/ioreg"

// ErrNotFound wraps the error IORegProperties/IORegNodeProperties return
// when nothing matched - e.g. no AppleSmartBattery on a Mac without a
// battery. Callers use errors.Is to tell "this section doesn't apply" apart
// from a real failure.
var ErrNotFound = errors.New("no matching object found")

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

// maxPropertyLineBytes raises bufio.Scanner's default 64KB line limit.
// Some device-tree properties are large binary blobs printed as a single
// hex-encoded line - e.g. the "chosen" node's IOProgressBackbuffer (a boot
// progress image) can run past 100KB - even though nothing we read cares
// about their content.
const maxPropertyLineBytes = 8 * 1024 * 1024

// scanObjectProperties reads ioreg's default (non-plist) object dump - an
// object header line, then a `{ ... }` block of scalar properties - and
// returns the properties of the first (only expected) object, keyed by
// property name. Values keep their raw ioreg formatting; use IntProperty /
// StringProperty / DataProperty to decode them.

func scanObjectProperties(output []byte, errContext string) (map[string]string, error) {
	props := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Buffer(make([]byte, 0, 64*1024), maxPropertyLineBytes)
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
		return nil, fmt.Errorf("%s: reading output: %w", errContext, err)
	}

	if !inObject {
		return nil, fmt.Errorf("%s: %w", errContext, ErrNotFound)
	}
	return props, nil
}

// IORegProperties runs `ioreg -rc <class> -d1` and returns the scalar
// properties of the first matching object, keyed by property name.
func IORegProperties(cmd SystemCommand, class string) (map[string]string, error) {
	output, err := cmd.Execute(ioregPath, "-rc", class, "-d1")
	if err != nil {
		return nil, fmt.Errorf("ioreg -rc %s: %w", class, err)
	}
	return scanObjectProperties(output, fmt.Sprintf("ioreg -rc %s", class))
}

// IORegNodeProperties runs `ioreg -p IODeviceTree -n <name> -d1 -r` and
// returns the scalar properties of the named device-tree node, keyed by
// property name. Use this for device-tree nodes that aren't IOKit classes,
// such as "chosen".
func IORegNodeProperties(cmd SystemCommand, name string) (map[string]string, error) {
	output, err := cmd.Execute(ioregPath, "-p", "IODeviceTree", "-n", name, "-d1", "-r")
	if err != nil {
		return nil, fmt.Errorf("ioreg -p IODeviceTree -n %s: %w", name, err)
	}
	return scanObjectProperties(output, fmt.Sprintf("ioreg -p IODeviceTree -n %s", name))
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

// BoolProperty decodes a raw ioreg boolean property value: the bare words
// `Yes` / `No` ioreg prints for OSBoolean values.
func BoolProperty(props map[string]string, key string) (bool, error) {
	raw, ok := props[key]
	if !ok {
		return false, fmt.Errorf("property %q not found", key)
	}
	switch raw {
	case "Yes":
		return true, nil
	case "No":
		return false, nil
	default:
		return false, fmt.Errorf("property %q: not a boolean: %q", key, raw)
	}
}

// StringProperty decodes a raw ioreg property value as a string, stripping
// the surrounding quotes ioreg prints for OSString/CFString values
// (`"foo"`) or, for a data property ioreg has printed as readable text
// instead of hex because its bytes look like a clean C string, both the
// angle brackets and the quotes (`<"foo">`) - e.g. the device tree's
// dram-type property.
func StringProperty(props map[string]string, key string) (string, error) {
	raw, ok := props[key]
	if !ok {
		return "", fmt.Errorf("property %q not found", key)
	}
	raw = strings.TrimSuffix(strings.TrimPrefix(raw, "<"), ">")
	return strings.Trim(raw, `"`), nil
}

// dataBytes decodes a raw ioreg "data" property value (printed as
// `<68657820...>`) to its underlying bytes.
func dataBytes(props map[string]string, key string) ([]byte, error) {
	raw, ok := props[key]
	if !ok {
		return nil, fmt.Errorf("property %q not found", key)
	}
	raw = strings.TrimSuffix(strings.TrimPrefix(raw, "<"), ">")

	decoded, err := hex.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("property %q: not hex data: %q", key, raw)
	}
	return decoded, nil
}

// DataProperty decodes a raw ioreg "data" property as a NUL-terminated
// ASCII string. Device tree properties such as platform-name are encoded
// this way.
func DataProperty(props map[string]string, key string) (string, error) {
	decoded, err := dataBytes(props, key)
	if err != nil {
		return "", err
	}
	if i := bytes.IndexByte(decoded, 0); i >= 0 {
		decoded = decoded[:i]
	}
	return string(decoded), nil
}

// DataPropertyUint decodes a raw ioreg "data" property as a little-endian
// unsigned integer, e.g. `<01000000>` -> 1. Device tree boolean/enum flags
// such as secure-boot are encoded this way. Property values longer than 8
// bytes are rejected.
func DataPropertyUint(props map[string]string, key string) (uint64, error) {
	decoded, err := dataBytes(props, key)
	if err != nil {
		return 0, err
	}
	if len(decoded) > 8 {
		return 0, fmt.Errorf("property %q: %d bytes too long for an integer", key, len(decoded))
	}

	var n uint64
	for i, b := range decoded {
		n |= uint64(b) << (8 * i)
	}
	return n, nil
}
