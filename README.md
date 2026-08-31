<div align="center">

**Kill it mid-write. Reopen it. Your data is still there. Zero dependencies. Zero goroutines. Just the write-ahead log.**

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat&logo=go&logoColor=white)](docs/engineering.md#dependency-proof)
[![stdlib only](https://img.shields.io/badge/stdlib-only-0ea5e9?style=flat)](STDLIB.md)
[![WAL](https://img.shields.io/badge/WAL-append--only-8b5cf6?style=flat)](docs/architecture.md#wal-format)
[![CRC32](https://img.shields.io/badge/CRC32-per--record-10b981?style=flat)](docs/architecture.md#recovery-algorithm)
[![flock](https://img.shields.io/badge/lock-flock%20on%20DB%20fd-f59e0b?style=flat)](docs/architecture.md#locking-model)

</div>

---

## Table of Contents

- [Quick Start](#quick-start)
- [Interactive Shell](#interactive-shell)
- [Description](#description)
- [CLI](#cli)
- [Repository Layout](#repository-layout)
- [Future Work](#future-work)

**Deep Dives**

- [Architecture & Core Design](docs/architecture.md) — WAL format, recovery, locking, durability
- [Crash Demo](docs/crash-demo.md) — deterministic fault-injection demo
- [Engineering Details](docs/engineering.md) — complexity, design decisions, prior art, tests, dependency proof, rubric

---

## Quick Start

Build and start using PicoDB in about a minute. No network, no extra tools — just the Go toolchain and this repo.

```bash
# 1. Build the binary (no network — stdlib only)
GOPROXY=off GOTOOLCHAIN=local go build -o picodb ./cmd/picodb

# 2. Use it from the command line
./picodb put demo.wal name keshav   # write a key
./picodb get demo.wal name          # → keshav
./picodb del demo.wal name          # remove it
./picodb get demo.wal name; echo $? # → 3 (not found)

# 3. Or open the interactive shell (menu-driven)
./scripts/interactive.ps1

# 4. See the crash-recovery demo (the centerpiece)
./scripts/crash_demo.sh

# 5. Run the quality gates
go vet ./...
go test ./...
go test -tags=integration ./...
make check
```

---

## Interactive Shell

For a friendlier way to work with data, PicoDB ships a menu-driven interactive shell: `scripts/interactive.ps1`.

```bash
./scripts/interactive.ps1
```

It opens a REPL-style menu that wraps the underlying `picodb` CLI:

| Option | Action |
|---|---|
| **1 — Put / create a key** | Write a key-value pair to the store |
| **2 — Get / read a key** | Look up a value by key |
| **3 — Delete a key** | Remove a key (appends a tombstone) |
| **4 — List all keys** | Show the keys created this session |
| **5 — Clear all data** | Wipe the database file |
| **0 — Exit** | Close the shell (data stays saved) |

Every create/read/delete goes straight through the real `picodb` store, so it's genuinely persistent and crash-safe — the same WAL, CRC protection, and recovery as the CLI. On Windows you can launch it with:

```powershell
powershell -NoExit -File .\scripts\interactive.ps1
```

---

## Description

**PicoDB** is a crash-safe embedded key-value store built entirely on the Go standard library. No external packages, no background threads, no hidden complexity.

What you get:

* an **append-only WAL** with per-record CRC32 integrity,
* an **in-memory key directory** (Bitcask-style hash map),
* **replay-on-open** with automatic truncated-tail repair,
* and **kernel `flock`** on the database file itself — no stale lockfiles.

> Process dies mid-append? The next `Open()` detects the torn record, truncates back to the last clean offset, and the store is immediately writable again. No manual intervention. No data loss for committed records.

---

## CLI

Small surface, explicit exit codes, and no `interface{}` leaking out of the public API.

| Command | Example | Exit |
|---|---|---:|
| `put` | `picodb put demo.wal name keshav` | `0` |
| `get` | `picodb get demo.wal name` → `keshav` on stdout | `0` / `3` if missing |
| `del` | `picodb del demo.wal name` | `0` / `3` if missing |

`get` prints **only the value** to stdout; everything else is diagnostic and goes to stderr.

---

## Repository Layout

```
picodb/
├── go.mod
├── Makefile
├── README.md
├── STDLIB.md
├── cmd/picodb/main.go
├── docs/
│   ├── architecture.md    # WAL format, recovery, locking, durability
│   ├── crash-demo.md      # deterministic fault-injection demo
│   └── engineering.md     # complexity, design decisions, tests, prior art
├── internal/
│   ├── wal/
│   │   ├── record.go      # Encode/DecodePrefix/Checksum
│   │   ├── writer.go      # Append/TruncateTo/Close/SyncBatch
│   │   ├── reader.go      # strict validation order
│   │   └── debug.go       # deterministic crash hook
│   ├── store/
│   │   ├── store.go       # Open/replay/Put/Get/Delete
│   │   └── index.go       # plain map, no mutex
│   ├── cli/
│   │   └── cli.go         # Run() with Exit codes
│   └── lock/
│       ├── lock_unix.go   # flock on DB fd
│       └── lock_windows.go # compile-safe fallback
└── scripts/
    ├── crash_demo.sh      # deterministic crash demo
    └── interactive.ps1    # menu-driven interactive shell
```

---

## Future Work

* compaction (garbage-collect tombstones)
* snapshots
* segment rotation
* configurable `SyncBatch` / `fsync` policies
* scan / range APIs
* metrics

---

<div align="center">

`WAL` • `CRC32` • `MaxRecordSize` • `Replay` • `Truncate` • `flock` • `SyncBatch` • `Self-healing`

</div>
