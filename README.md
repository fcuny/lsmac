# lsmac

`lsmac` prints hardware and system information about the Apple Silicon Mac it
runs on. Think `neofetch` without the logo, or `lscpu` / `lspci` for a whole
Mac.

This is an early scaffold. Only the SoC section (CPU and GPU core counts)
works so far.

## Build

    make build

## Usage

    lsmac

## Development

    make ci    # fmt-check, vet, test, build
    make test
    make fmt
