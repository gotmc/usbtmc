# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
just check              # go fmt + go vet
just unit               # unit tests (runs check first, -race -short -cover)
just unit -run TestName # run a single test
just lint               # golangci-lint v2 with .golangci.yaml config
just cover              # HTML coverage report (also: just cover int, just cover e2e, just cover all)
just tidy               # go mod tidy + verify
just docs               # local pkgsite viewer
```

Requires `libusb` C library installed on the system (used by both driver backends).

Set `USBTMC_DEBUG=1` to enable debug logging to stderr (see debug.go).

## Architecture

This is a Go library for communicating with USB Test & Measurement Class (USBTMC) devices (oscilloscopes, multimeters, function generators, etc.) via SCPI commands over USB.

**Driver abstraction pattern:** The package uses a registration-based driver model (similar to `database/sql`). The `driver/` package defines interfaces (`Driver`, `Context`, `USBDevice`), and two concrete backends exist:

- `driver/google/` — uses `github.com/google/gousb` (recommended, includes Keysight boot-mode handling)
- `driver/gotmc/` — uses `github.com/gotmc/libusb/v2`

Drivers self-register via `init()` calling `usbtmc.Register()`. Users select a driver with a blank import:
```go
import _ "github.com/gotmc/usbtmc/driver/google"
```

**Core flow:** `Context` (context.go) wraps a driver context and creates `Device` instances. `Device` (device.go) implements USBTMC protocol framing — it encodes/decodes the 12-byte USBTMC bulk headers (helpers.go) around user data for USB bulk transfers. Key device methods: `Write`, `Read` (ASCII with termChar), `ReadRaw` (binary without termChar), `WriteString`, `Command` (SCPI with auto-newline), `Query` (write+read). `Device` satisfies the `ivi.Transport` interface via `Command`, `Query`, `ReadBinary`, `WriteBinary`, and `Close`. Convenience methods `Write`, `Read`, `ReadRaw`, and `WriteString` provide context-free signatures for simple usage.

**USBTMC protocol details:** Each USB transfer uses a 12-byte header (defined in helpers.go per USBTMC Spec 1.0 Tables 1-4). The `bTag` field increments between transfers (1-255) for sequencing. The IO buffer is 1MB (`ioBufferSize` in constants.go). Constants in constants.go map directly to the USBTMC and USB488 specifications — comments reference specific table numbers.

**VISA addressing:** visa.go parses VISA resource strings (e.g., `USB0::0x0957::0x2818::0::INSTR`) to extract VID/PID for device lookup.

This package is designed to serve as an `Instrument` interface provider for the [gotmc/ivi](https://github.com/gotmc/ivi) and [gotmc/visa](https://github.com/gotmc/visa) packages.
