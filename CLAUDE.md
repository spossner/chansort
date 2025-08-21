# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Chansort is a Go-based CLI tool for manipulating Samsung SCM satellite channel archives. It can read `.scm` files exported from Samsung TVs, parse satellite channel data, and reorder channels through both command-line and interactive TUI interfaces.

## Build and Development Commands

```bash
# Build the project
go mod tidy
go build -o chansort

# Run with different commands
./chansort dump --scm /path/to/channel_list.scm
./chansort edit --scm /path/to/channel_list.scm
./chansort swap --scm /path/to/channel_list.scm
```

## Architecture

### Core Components

- **main.go**: Entry point that delegates to cmd package
- **cmd/**: Cobra-based CLI commands
  - `root.go`: Main command definition with required `--scm` flag
  - `edit.go`: Interactive TUI channel browser using Bubble Tea
  - `swap.go`: Hardcoded example swapping ARD/ZDF channels
- **internal/scm/**: Samsung SCM archive handling
  - Parses 168-byte fixed records from `map-SateD` file
  - Non-destructive parsing preserves raw bytes for safe editing
  - Channel names stored at offset 0x24 as UTF-16BE
- **internal/tui/**: Bubble Tea TUI implementation
  - Multiple modes: Navigation, Search, Jump (numeric), Move
  - Real-time filtering and channel reordering
  - Toast notifications for user feedback
- **internal/utils/**: Utility functions (debug logging)

### Key Data Structures

- `scm.Channel`: Represents a satellite channel with ID, OrderId (LCN), Name, and Raw 168-byte data
- Channel records are kept as raw bytes to ensure non-destructive editing
- TUI model maintains filtered/unfiltered channel lists and cursor state

### Samsung SCM Format

- SCM files are ZIP archives containing Samsung TV channel data
- `map-SateD` contains satellite channels as 168-byte fixed records
- LCN (Logical Channel Number) stored as little-endian uint16 at bytes 0-1
- Channel names at offset 0x24 encoded as UTF-16BE zero-terminated
- Archive structure preserved when writing modified channel lists

## Dependencies

- **Cobra**: CLI framework
- **Bubble Tea**: Terminal UI framework with Lip Gloss for styling
- **Standard library**: zip, encoding utilities for Samsung format parsing