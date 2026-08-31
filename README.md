# PicoDB

**PicoDB combines a Bitcask-style in-memory keydir with CRC-protected length-prefixed WAL records, deterministic batched syncing, OS-managed process locking, and self-healing tail truncation during recovery.**

PicoDB is an embedded crash-safe key-value store written in Go using only the standard library. Its core is an append-only write-ahead log combined with an in-memory hash index.

## Architecture

```text
                         ┌──────────────────────┐
                         │         CLI          │
                         │   put/get/del        │
                         └──────────┬───────────┘
                                    │
                                    ▼
                         ┌──────────────────────┐
                         │        Store         │
                         │  single lock owner   │
                         └────────┬──────┬──────┘
                                  │      │
                        ┌─────────┘      └──────────┐
                        ▼                            ▼
                ┌──────────────┐            ┌──────────────┐
                │    Index     │            │      WAL     │
                │ plain map    │            │ Reader/Writer│
                └──────────────┘            └──────┬───────┘
                                                   │
                                                   ▼
                                           ┌───────────────┐
                                           │ Database file │
                                           │ + flock       │
                                           └───────────────┘
```

## WAL Format

```text
0          4          8          9            13           17
┌──────────┬──────────┬────┬──────────────┬──────────────┬────────┬─────────┐
│RecordLen │  CRC32   │Type│   KeyLen     │    ValLen    │  Key   │  Value  │
│   BE     │   BE     │1 B │     BE       │      BE      │        │         │
└──────────┴──────────┴────┴──────────────┴──────────────┴────────┴─────────┘
    4 B         4 B      1 B      4 B             4 B
```
The `RecordLen` field is outside the CRC scope to prevent unbounded allocations during corruption checks. CRC protects the entire body.

## Recovery Algorithm

During database initialization, `Store.Open` replays the WAL.
- Valid records reconstruct the in-memory key directory.
- The process guarantees immediate termination at the first corrupt or truncated record (ErrCorruptTail).
- The recovery protocol enforces tail truncation, completely erasing the corrupt tail and generating a fresh and reliable file boundary capable of continuous operations without crashing the server.

## Locking Model

The system utilizes one single lock authority: `Store` holds an OS-level file lock (via `syscall.Flock` or `LockFileEx` depending on the platform) on the primary database file descriptor. There are no sidecar `.lock` files that could outlive process crashes.

## Durability Policy

Writes are batched automatically via a `SyncBatch` parameter (currently 100). The database also explicitly syncs upon closing, preventing any unwritten buffers from being lost in clean shutdown conditions.

## CLI Examples

```bash
picodb put demo.wal name keshav
picodb get demo.wal name
picodb del demo.wal name
```

## Crash Demo

The repository contains `scripts/crash_demo.sh` (and `.ps1`), which deterministically crashes the database during an active WAL append. On restart, the system safely recognizes the torn body, prints a recovery warning, cleanly truncates the broken bytes, recovers earlier writes, and allows subsequent writes without missing a beat.

## Complexity

- **GET:** O(1) average hash-map lookup
- **PUT:** O(record size) append WAL + O(1) average index update
- **DELETE:** O(record size) append tombstone + O(1) average index delete
- **REPLAY:** O(file size) sequential scan
- **CRC:** O(record size)

## Design Decisions / Trade-offs

- **Strong Guarantees:** CRC-protected records, explicit length validation, single OS-managed process lock, replay integrity, self-healing recovery truncation.
- **Deliberate Compromises:** Complete key index lives in memory (Bitcask-style limits database size to available RAM), WAL grows indefinitely without background compaction, batched fsync creates a bounded durability window, no multi-key transactions.

## Prior Art

We took established storage-engine ideas and implemented a deliberately minimal version using only the Go standard library.
- **Bitcask:** Append-only log + in-memory key directory mapping.
- **LevelDB / RocksDB:** Framed log records and integrity-aware storage.
- **Kafka:** Append-oriented log framing and recovery boundaries.
- **SQLite / Postgres:** Explicit durability semantics and crash recovery guarantees.

## Tests

Extensive automated unit and integration tests handle structural consistency, file bounding, EOF detection, missing key states, deterministic recovery testing, lock checking, and tombstone resolution. Run them via `make check`.

## Dependency Proof

Execute `GOPROXY=off GOTOOLCHAIN=local go list -m all`. The framework relies exclusively on native `go 1.22+` packages (see `STDLIB.md`).

## Future Work

- Compaction logic
- Snapshots / segment rotation
- Configurable sync policies
- Scan APIs
- Metrics