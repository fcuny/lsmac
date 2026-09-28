# lsmac

`lsmac` prints hardware and system information about the Apple Silicon Mac it
runs on. Think `neofetch` without the logo, or `lscpu` / `lspci` for a whole
Mac: OS, machine, firmware, SoC, memory, storage, and power.

It targets Apple Silicon Macs only, reads everything through `sysctl` and
`ioreg`, and stays fast: the default view runs in under 100ms.

## Build

    make build

## Usage

    lsmac                      # compact view of every applicable section
    lsmac --section cpu        # detailed view of one section (repeatable)
    lsmac --json                # structured JSON instead of text
    lsmac --show-serial         # include serial number and hardware UUID

### Default view

```
$ lsmac
OS             macOS 27.0 Golden Gate (26A428)
Darwin         27.0.0
Uptime         13 days, 18 hours
Rosetta        not installed
Machine        MacBook Air (M2, 2022) (Mac14,2)
SKU            MN703LL/A
Firmware       mBoot-20457.1.29
Secure Boot    enabled
SIP            enabled
Chip           Apple M2 (T8112)
Process        5-nanometer (2nd generation)
CPU            8 cores: 4P + 4E
GPU            10 cores
Memory         11.0 GiB / 16 GiB
Type           LPDDR5
Bandwidth      100 GB/s
Pressure       normal
Disk           APPLE SSD AP1024Z
Capacity       404.4 GiB / 926.4 GiB
Battery        76%, discharging
Cycles         201
Health         92%
```

Every field label is padded to the same column, so values always line up
regardless of label length (down to "Low Power Mode", the longest one).

A section that doesn't apply to the machine it runs on (no battery on a Mac
mini, for instance) is left out, never printed as "N/A". A section whose
collector fails reports the error on stderr and the rest of the output still
prints.

Serial number, hardware UUID, FileVault status, thermal state, and Low
Power Mode are left out of the default view on purpose: the first two are
sensitive (screenshots of the default view shouldn't leak them; pass
`--show-serial` to include them), and the rest each cost tens of
milliseconds to check (`fdesetup status` and two `pmset` calls) for facts
that are "off"/"nominal" the overwhelming majority of the time. Pass
`--section storage` or `--section power` to get them.

### Detailed sections

`--section NAME` is repeatable and prints just the named section, in more
detail where lsmac has more to show. So far only `cpu` does:

```
$ lsmac --section cpu
CPU
  Chip              Apple M2 (T8112)
  Family            0xDA33D83D
  Page size         16 KiB

  Performance       4 cores  ·  1 cluster
    L1i / L1d       192 KiB / 128 KiB per core
    L2              16 MiB per cluster (4 cores per L2)

  Efficiency        4 cores  ·  1 cluster
    L1i / L1d       128 KiB / 64 KiB per core
    L2              4 MiB per cluster (4 cores per L2)

  Features          CRC32 FlagM FlagM2 FHM DotProd SHA3 RDM LSE SHA256 SHA512
                    SHA1 AES PMULL SB FRINTTS PACIMP LRCPC LRCPC2 FCMA JSCVT
                    PAuth PAuth2 FPAC DPB DPB2 BF16 I8MM ECV LSE2 CSV2 CSV3 DIT
                    AdvSIMD AdvSIMD_HPFPCvt FP16 SSBS BTI FP_SyncExceptions
```

Valid section names: `os`, `machine`, `firmware`, `chip`, `cpu`, `gpu`,
`memory`, `storage`, `power`, `io`.

### JSON

`--json` prints every applicable section as a single JSON object, one key
per section, and composes with `--section` and `--show-serial`:

```
$ lsmac --json --section gpu --section memory
{
  "gpu": {
    "coreCount": 10
  },
  "memory": {
    "totalBytes": 17179869184,
    "usedBytes": 11812012032,
    "wiredBytes": 2063745024,
    "compressedBytes": 5316214784,
    "type": "LPDDR5",
    "pressure": "normal",
    "bandwidthGBs": 100
  }
}
```

Unlike the text views, `--json` always fetches full detail (FileVault
status, thermal state, the full CPU breakdown) for whatever sections it
prints: it's an explicit request for complete structured data, not the
fast glance the default view is built for, so it isn't held to the same
speed budget.

## What's not there yet

A few fields from the original design aren't collected, each for a
concrete reason rather than being simply unfinished:

- **Display** (resolution, refresh rate): Apple Silicon exposes this only
  through deeply nested, undocumented GPU-driver structures. The reliable
  alternative, `system_profiler SPDisplaysDataType`, takes about 200ms.
- **Wi-Fi standard, Bluetooth version, Thunderbolt/USB4 port count**: same
  problem. `system_profiler SPAirPortDataType` alone took over 4 seconds in
  testing, and the ioreg equivalents (where they exist at all) don't map
  cleanly to a simple count or standard name.
- **CPU maximum frequency per cluster**: not exposed via `sysctl`. It exists
  as raw `pmgr` voltage/frequency tables in the device tree, but decoding
  them correctly needs a mapping from table index to cluster that's specific
  to each chip generation.
- **Adapter wattage while charging**: the relevant `AppleSmartBattery`
  fields were never observed populated during development (the machine
  wasn't plugged in), so nothing here has been verified against real data.
- **Secure Boot policy** (Full/Reduced/Permissive Security): the only
  reliable source, `bputil -d`, requires root, which lsmac never asks for.
  A plain secure-boot-enabled flag is collected instead.

## Development

    make ci    # fmt-check, vet, test, build
    make test
    make fmt

Collectors are tested against fixtures captured from real Macs, under
`testdata/<model-identifier>/`. Every collector goes through the
`SystemCommand` interface in `internal/source`, so tests replay captured
`sysctl`/`ioreg` output instead of touching the real machine.
