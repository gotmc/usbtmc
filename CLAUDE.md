# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
# Format and vet
just check

# Run unit tests
just unit

# Lint with golangci-lint (v2 config in .golangci.yaml)
just lint

# HTML coverage report
just cover

# View documentation in browser via pkgsite
just docs

# Run a single test
go test -run TestFuncName ./...

# Tidy and verify modules
just tidy
```

Requires `libusb` C library installed on the system (used by both driver backends).

`just unit` runs `just check` (fmt + vet) automatically before tests. Tests run with `-short` and `-race` flags by default.

## Architecture

This is a Go library for communicating with USB Test & Measurement Class (USBTMC) devices (oscilloscopes, multimeters, function generators, etc.) via SCPI commands over USB.

**Driver abstraction pattern:** The package uses a registration-based driver model (similar to `database/sql`). The `driver/` package defines interfaces (`Driver`, `Context`, `USBDevice`), and two concrete backends exist:

- `driver/google/` — uses `github.com/google/gousb` (recommended, includes Keysight boot-mode handling)
- `driver/gotmc/` — uses `github.com/gotmc/libusb/v2`

Drivers self-register via `init()` calling `usbtmc.Register()`. Users select a driver with a blank import:
```go
import _ "github.com/gotmc/usbtmc/driver/google"
```

**Core flow:** `Context` (context.go) wraps a driver context and creates `Device` instances. `Device` (device.go) implements USBTMC protocol framing — it encodes/decodes the 12-byte USBTMC bulk headers (helpers.go) around user data for USB bulk transfers. Key device methods: `Write`, `Read` (ASCII with termChar), `BulkRead` (binary without termChar), `WriteString`, `Command` (SCPI with auto-newline), `Query` (write+read). `Write`, `Read`, `BulkRead`, and `WriteString` have context-free signatures to satisfy the `ivi.Instrument` interface, with `*Context` variants (e.g., `WriteContext`, `ReadContext`) that accept a `context.Context`.

**VISA addressing:** visa.go parses VISA resource strings (e.g., `USB0::0x0957::0x2818::0::INSTR`) to extract VID/PID for device lookup.

This package is designed to serve as an `Instrument` interface provider for the [gotmc/ivi](https://github.com/gotmc/ivi) and [gotmc/visa](https://github.com/gotmc/visa) packages.
