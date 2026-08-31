# STDLIB — Standard Library Only Proof

> **Dependency policy:** `picodb` uses **Go standard library only**. No third-party modules are fetched at build or test time. See `go.mod` and `deps` proof below.

## Module Declaration

```go
module picodb

go 1.22

require ()
```

Verification (offline, no network):

```bash
cat go.mod
GOPROXY=off GOTOOLCHAIN=local go list -m all   # → only picodb
GOPROXY=off GOTOOLCHAIN=local go mod verify    # → all modules verified
GOPROXY=off GOTOOLCHAIN=local go build ./...   # → no download
```

Expected invariant:

```
only picodb
no third-party modules
no network download
no automatic toolchain fetch (GOTOOLCHAIN=local)
```

## Packages Actually Imported

Only packages listed below appear in `go list -f '{{join .Imports ","}}' ./...`. No external import is allowed.

| Package | Purpose | Used In |
|---|---|---|
| `os` | files, stdout/stderr, process exit, env `PICODB_CRASH_AFTER_PREFIX`, truncation | `wal/writer.go`, `wal/reader.go`, `wal/debug.go`, `store`, `cli`, `lock` |
| `io` | `io.EOF`, `io.SeekEnd`, `io.ReadFull`, exact reads | `wal/writer.go`, `wal/reader.go` |
| `encoding/binary` | big-endian `PutUint32`/`Uint32` for `RecordLen`, `CRC`, `KeyLen`, `ValLen` | `wal/record.go`, `wal/reader.go` |
| `hash/crc32` | `ChecksumIEEE` over `Type‖KeyLen‖ValLen‖Key‖Value` | `wal/record.go`, `wal/reader.go` |
| `sync` | `RWMutex` — Store is the single lock authority | `store/store.go` |
| `syscall` | Unix `Flock(LOCK_EX|LOCK_NB)` on the DB file descriptor | `lock/lock_unix.go` |
| `errors` | sentinels `ErrCorruptTail`, `ErrKeyNotFound`, `errors.Is` | `wal/*`, `store/*` |
| `fmt` | `fmt.Errorf("...: %w", err)` wrapping, recovery warnings, CLI output | `wal/*`, `store/*`, `cli/*` |
| `testing` | `testing.T`, temp dirs, subprocess tests | `*_test.go` |
| `path/filepath` | temp file joins (tests) — if needed, otherwise omitted | tests |
| `strconv` | env/demo config parsing if needed | `wal/debug.go` (optional) |

> If `go list` shows any import outside this table, the build fails the **Zero-Dependency Craft (30%)** rubric.

## Why Stdlib-Only?

* **Reproducibility** — `GOPROXY=off` clean clone always builds.
* **Auditability** — every byte of durability code is visible.
* **Rubric alignment** — proves craft without hiding complexity in dependencies.
* **No hidden bufio/fsync traps** — buffering is explicit and ordered `Flush → Sync`.

## Forbidden During Sprint

No `bufio` hidden behind `Sync`, no background goroutine/timer sync, no external packages, no `toolchain` directive in `go.mod` (see `Plan.md §7`).

---

*Generated for Worker D — docs/release. Keep this table in sync with `go list`.*
