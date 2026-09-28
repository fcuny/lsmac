# lsmac

`lsmac` is a command-line tool that prints hardware and system information about
the Mac it runs on. Think `neofetch` without the logo, or `lscpu` / `lspci` for a
whole Mac: OS, machine, firmware, SoC, GPU, memory, storage, display, power and
I/O. It targets Apple Silicon Macs only.

lsmac started from a small existing SoC-info tool (called `socinfo` in its own
README). That program only covers the SoC: brand string, core counts, GPU core
count, and a hardcoded TDP / bandwidth table. `lsmac` extracts the idea and
widens the scope.

## Conventions

- Go, module path `fcuny.net/lsmac`. Use the Go version from `~/workspace/cpuids/go.mod`.
- Use a `Makefile` for tasks, never a `justfile`. Targets: `build`, `test`,
  `test-cov`, `vet`, `fmt`, `fmt-check`, `ci` (runs fmt-check, vet, test, build),
  `clean`, with `.PHONY` declared. Otherwise mirror `~/workspace/cpuids`: BSD
  3-Clause `LICENSE`, `.github/workflows/ci.yml` with gofmt, vet, `go test -race`
  and build. CI must run on `macos-latest` (arm64), since the collectors need Darwin.
- Use the `gh` CLI for anything on GitHub.
- Work in feature branches with a PR per change; let CI run on the PR before
  merging. Don't push straight to `main`.
- Prefer the standard library. `golang.org/x/sys/unix` is fine for `sysctl`.
  Ask before adding any other dependency.
- Writing style for README, docs and commit messages: active voice, no em dash
  or en dash in sentences, and no "it's not X, it's Y" constructions.

## Architecture

Keep collection and rendering separate.

```
cmd/lsmac/          main: flags, pick sections, call the renderer
internal/collect/   one file per section; each returns a plain struct
internal/source/    thin wrappers over the data sources (sysctl, ioreg, ...)
internal/chips/     static table of chip facts the hardware does not report
internal/render/    text and JSON renderers
testdata/<machine>/ captured raw output per Mac, used as fixtures
```

- Every collector reads through an interface (the `SystemCommand` idea from the
  old code) so tests can replay fixtures. Keep that pattern.
- A section that does not apply (no battery on a Mac mini, no built-in display
  on a Mac Studio) is omitted from the output, never printed as "N/A".
- One failing collector must not fail the whole run. Print what works, report
  the error for that section on stderr.
- Output must be fast. Target well under 100 ms. Do not call `system_profiler`
  on the default path; it takes about a second per data type. Use `sysctl` and
  `ioreg -a` (plist output) first. Consider direct IOKit through cgo only if
  `ioreg` parsing turns out too slow or too fragile, and discuss it first.

## CLI

- `lsmac`: default compact view (see mockup below)
- `lsmac --section cpu` (repeatable): detailed view of one or more sections
- `lsmac --json`: same data as structured JSON, one object per section
- `lsmac --show-serial`: serial number and hardware UUID are hidden unless this
  flag is set, so screenshots don't leak them

## What to collect

The "source" column lists candidates. Verify each one on the real machine
before relying on it, and record the actual key you used.

| Section  | Fields | Candidate source |
|----------|--------|------------------|
| OS       | product name, version, build, Darwin version, uptime, Rosetta installed | `kern.osproductversion`, `kern.osversion`, `kern.osrelease`, `kern.boottime`, `sw_vers` |
| Machine  | marketing name, model identifier, model number / SKU, serial, hardware UUID | `hw.model`, `ioreg` `IOPlatformExpertDevice`, device tree root |
| Firmware | system firmware (iBoot) version, OS loader version, Secure Boot policy, SIP | `ioreg` device tree, `csrutil status` |
| SoC/CPU  | chip name and ID (e.g. T6031), P/E cores per cluster, max frequency per cluster, L1/L2 caches, page size, ARM feature flags | `machdep.cpu.brand_string`, `hw.perflevelN.*`, `hw.cpufamily`, `hw.pagesize`, `hw.optional.arm.FEAT_*`, `pmgr` voltage states in device tree |
| GPU      | core count, family, Metal support | `ioreg -arc AGXAccelerator` (`gpu-core-count`) |
| Neural / media | Neural Engine cores, video encode / ProRes engines | chip table |
| Memory   | total, type (LPDDR5/5X), used/wired/compressed, pressure, bandwidth | `hw.memsize`, `host_statistics64` / `vm_stat`, chip table for bandwidth |
| Storage  | internal SSD model, capacity, APFS usage, FileVault | `ioreg`, `statfs`, `fdesetup status` |
| Display  | panel name, native and scaled resolution, refresh rate, external displays | `ioreg` / CoreGraphics |
| Power    | battery %, charging, adapter watts, cycle count, health, thermal state, Low Power Mode | `ioreg -arc AppleSmartBattery`, `pmset` |
| I/O      | Thunderbolt/USB4 ports, Wi-Fi standard, Bluetooth version, network interfaces | `ioreg`, `net.Interfaces` |

### Chip table (`internal/chips`)

For facts the hardware does not expose: marketing name, process node, memory
bandwidth, Neural Engine cores, media engines.

- Key it on the chip ID from the device tree (e.g. `T8103`, `T6031`) or on
  `hw.cpufamily`. Do not match on the brand string with `strings.Contains`;
  the old code did that and the `case` order mattered.
- Cover M1 through the current generation, all variants (base, Pro, Max, Ultra).
- Every value must come from a published source (Apple tech specs pages).
  Add a comment with the source. Leave a field empty instead of guessing.
- Drop TDP. Apple does not publish it; the old table held estimates.
- `~/workspace/cpuids` (`fcuny.net/cpuids`) resolves ARM implementer + part
  (Apple is implementer `0x61`) to core microarchitecture names. macOS does not
  expose MIDR directly, so check whether the part can be derived (device tree
  `compatible` strings, `hw.cpufamily`) before wiring it in. Treat this as an
  optional follow-up.

## Known bugs in the original code

1. **P and E cores are swapped.** `getCPUInfo` returns `(brand, total, pCores, eCores)`
   (`hw.perflevel0` is performance, `hw.perflevel1` is efficiency), but the caller
   assigns them to `(..., eCoreCount, pCoreCount, ...)`. The tests use an M2 with
   4P + 4E, so they don't catch it. Add a fixture with asymmetric counts.
2. The spec table has zero bandwidth for M2 Pro and M2 Max, no values for the M3
   family, and nothing for M4 and later.
3. GPU core count comes from `system_profiler SPDisplaysDataType`, which is slow.

## Target output (default view)

Mocked values. Use this as the layout reference, not as data.

```
fcuny@lamb  ·  MacBook Pro (14-inch, Nov 2023)
────────────────────────────────────────────────────────────
OS         macOS 15.6 Sequoia (24G84)  ·  Darwin 24.6.0
Uptime     6 days, 3 hours
Model      Mac15,8  ·  MRX53LL/A
Firmware   iBoot 11881.140.96  ·  Secure Boot: Full  ·  SIP: enabled

Chip       Apple M3 Max (T6031)  ·  3 nm
CPU        16 cores: 12P @ 4.05 GHz + 4E @ 2.75 GHz
Cache      P: 192K L1i / 128K L1d, 16M L2  ·  E: 128K / 64K, 4M L2
GPU        40 cores  ·  Metal 3
Neural     16-core Neural Engine
Media      2 video encode, 2 ProRes engines
Memory     22.4 GiB / 64 GiB  ·  LPDDR5  ·  400 GB/s  ·  pressure: normal

Disk       APPLE SSD AP1024Z  ·  612 GiB / 994 GiB  ·  FileVault on
Display    Liquid Retina XDR 3024x1964 (1512x982 @2x)  ·  120 Hz
Battery    87%, charging via 96W  ·  212 cycles  ·  health 94%
Thermal    nominal  ·  Low Power Mode off
Ports      3x Thunderbolt 4  ·  Wi-Fi 6E  ·  Bluetooth 5.3
```

Detailed view, `lsmac --section cpu`:

```
CPU
  Chip              Apple M3 Max (T6031)
  Family            0x2876F5B5
  Page size         16 KiB

  Performance       12 cores  ·  2 clusters  ·  up to 4.05 GHz
    L1i / L1d       192 KiB / 128 KiB per core
    L2              16 MiB per cluster (6 cores per L2)

  Efficiency        4 cores  ·  1 cluster  ·  up to 2.75 GHz
    L1i / L1d       128 KiB / 64 KiB per core
    L2              4 MiB per cluster

  Features          FP16 BF16 I8MM DotProd SHA3 SHA512 AES PMULL
                    LSE LSE2 FlagM2 FRINTTS SB SSBS BTI DPB2
```

## Testing

- Capture raw output from this Mac into `testdata/<model-identifier>/`
  (`sysctl -a`, the relevant `ioreg -a` dumps). Remove the serial number and
  UUID from fixtures before committing.
- Every collector has table-driven tests against the fixtures.
- After each section, run `lsmac` on this machine and compare the values with
  System Information (`system_profiler` is fine for verification, just not in
  the tool itself). Show the comparison before calling a section done.
- Fixtures from other Macs (M1, M2 Pro, M4...) are welcome later; design the
  tests so adding a machine means adding a folder.

## Plan

Work in this order, one commit (or PR) per step. Stop after each step so the
output can be reviewed.

1. Scaffold: `go.mod`, `Makefile`, `LICENSE`, CI, `README.md`, the directory
   layout above. Port the old code into it with the P/E fix, and switch the GPU
   core count to `ioreg`.
2. Chip table keyed on chip ID, M1 to current, with sources. Drop TDP.
3. OS, Machine and Firmware sections.
4. Full CPU section (clusters, frequencies, caches, features) and `--section`.
5. Memory, Storage, Display, Power, I/O.
6. `--json`, `--show-serial`, and the README with real output.
