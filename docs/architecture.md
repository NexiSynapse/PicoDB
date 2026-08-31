# Architecture & Core Design

> One writer. One lock holder. No goroutines, no timers, no channels.

The store owns the lock. The WAL owns the disk. The CLI owns nothing — it just calls `Store` methods and prints results.

```mermaid
%%{init: {'theme':'base','themeVariables':{'primaryColor':'#fff','lineColor':'#64748b'}}}%%
flowchart TB
  subgraph CLI["<b>CLI</b> — cmd/picodb • internal/cli"]
    direction TB
    C1["picodb put / get / del"]
    C2["Exit codes: 0 OK • 1 I/O • 2 usage • 3 not found"]
  end

  subgraph Store["<b>Store</b> — internal/store — <i>only lock authority</i>"]
    direction TB
    S1["sync.RWMutex"]
    S2["index: map[string][]byte<br/><i>no internal mutex</i>"]
    S3["Replay on Open()"]
  end

  subgraph WAL["<b>WAL</b> — internal/wal — <i>append at EOF only</i>"]
    direction TB
    W1["Writer<br/>Append • TruncateTo • Close • SyncBatch=100"]
    W2["Reader<br/>Next() • Offset() — strict validation order"]
    W3["Record<br/>Encode • DecodePrefix • Checksum"]
    W4["debug.go<br/>deterministic crash hook"]
  end

  subgraph Disk["<b>Disk</b> — one file, kernel-managed lock"]
    D1[("demo.wal<br/>RecordLen | CRC32 | Type | KeyLen | ValLen | Key | Value")]
    D2["flock(LOCK_EX|LOCK_NB) on DB fd<br/>process dies → kernel releases"]
  end

  CLI --> Store
  Store --> S2
  Store --> WAL
  W1 --> D1
  W2 --> D1
  S3 -. replay .-> W2
  W1 -. truncate .-> D1
  Store -. lock/unlock .-> D2

  classDef cli fill:#e0f2fe,stroke:#0ea5e9,stroke-width:2px,color:#0c4a6e
  classDef store fill:#ede9fe,stroke:#8b5cf6,stroke-width:2px,color:#4c1d95
  classDef wal fill:#ecfdf5,stroke:#10b981,stroke-width:2px,color:#065f46
  classDef disk fill:#fffbeb,stroke:#f59e0b,stroke-width:2px,color:#92400e

  class CLI cli
  class Store store
  class WAL wal
  class Disk disk
```

**Composition root** `cmd/picodb/main.go:1` — the only place that wires things together:

```go
func main() { os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr)) }
```

**Who owns what:**

| Worker | Owns | File |
|---|---|---|
| **A** | WAL engine | `internal/wal/record.go`, `writer.go`, `reader.go`, `debug.go` |
| **B** | Store + Index | `internal/store/store.go`, `index.go` |
| **C** | Tests | `*_test.go`, `integration` |
| **D** | Docs/Release | `README.md`, `STDLIB.md`, `Makefile`, `go.mod` |
| **E** | Lock + crash harness | `internal/lock/*`, `scripts/crash_demo.sh` |

---

## WAL Format

Every record is length-prefixed, CRC-protected, and bounded. Corrupting a single byte anywhere in the body is caught; so is a record that claims to be larger than ~16 MB.

```mermaid
%%{init: {'theme':'base','themeVariables':{'primaryColor':'#fff','lineColor':'#0ea5e9'}}}%%
flowchart LR
  subgraph Header["Byte Offsets"]
    direction LR
    H0["0"] --- H4["4"] --- H8["8"] --- H9["9"] --- H13["13"] --- H17["17"] --- HEND["…"]
  end
  Header --> L["RecordLen<br/>BE uint32<br/>4 B"]
  L --> C["CRC32<br/>BE uint32<br/>4 B"]
  C --> T["Type<br/>1 B"]
  T --> KL["KeyLen<br/>BE u32<br/>4 B"]
  KL --> VL["ValLen<br/>BE u32<br/>4 B"]
  VL --> K["Key<br/>KeyLen B"]
  K --> V["Value<br/>ValLen B"]

  style L fill:#0ea5e9,stroke:#0284c7,color:#fff
  style C fill:#8b5cf6,stroke:#7c3aed,color:#fff
  style T fill:#0f172a,stroke:#020617,color:#fff
  style KL fill:#334155,stroke:#1e293b,color:#fff
  style VL fill:#334155,stroke:#1e293b,color:#fff
  style K fill:#e0f2fe,stroke:#7dd3fc,color:#0c4a6e
  style V fill:#ede9fe,stroke:#a78bfa,color:#4c1d95
  style Header fill:#f8fafc,stroke:#cbd5e1,color:#475569
```

**Formulas:**

```
RecordLen = 9 + KeyLen + ValLen          // 1 + 4 + 4 + Key + Value
CRC span  = Type || KeyLen || ValLen || Key || Value   // RecordLen is OUTSIDE CRC
Total on disk = 8 (prefix) + RecordLen
MaxRecordSize = 16 MiB   // validated BEFORE allocation
```

| Field | Size | Encoding | Notes |
|---|---:|---|---|
| `RecordLen` | 4 B | `binary.BigEndian.Uint32` | `9 + KeyLen + ValLen`, `<= MaxRecordSize` |
| `CRC32` | 4 B | `binary.BigEndian.Uint32` | `crc32.ChecksumIEEE(body)` |
| `Type` | 1 B | `uint8` | `1 = Put`, `2 = Delete` |
| `KeyLen` | 4 B | BE | |
| `ValLen` | 4 B | BE | `0` for `Delete` tombstone |
| `Key` | `KeyLen` | raw | |
| `Value` | `ValLen` | raw | omitted for `Delete` |

Example (`put demo.wal name keshav`): `KeyLen=4`, `ValLen=6`, `RecordLen=19`, `total=27 B`.

```mermaid
%%{init: {'theme':'base'}}%%
flowchart LR
  P0["0–31<br/>RecordLen (32)"] --> P1["32–63<br/>CRC32 (32)"]
  P1 --> P2["64–71<br/>Type (8)"]
  P2 --> P3["72–103<br/>KeyLen (32)"]
  P3 --> P4["104–135<br/>ValLen (32)"]
  P4 --> P5["136–<br/>Key (variable)"]
  P5 --> P6["…<br/>Value (variable)"]

  style P0 fill:#0ea5e9,stroke:#0284c7,color:#fff
  style P1 fill:#8b5cf6,stroke:#7c3aed,color:#fff
  style P2 fill:#0f172a,stroke:#020617,color:#fff
  style P3 fill:#334155,stroke:#1e293b,color:#fff
  style P4 fill:#334155,stroke:#1e293b,color:#fff
  style P5 fill:#e0f2fe,stroke:#7dd3fc,color:#0c4a6e
  style P6 fill:#ede9fe,stroke:#a78bfa,color:#4c1d95
```

---

## Recovery Algorithm

**Policy:** stop at the first corrupt or truncated record and throw away everything after it. For an append-only, single-writer log, everything that precedes the first bad byte is trustworthy — so the engine trusts it and forgets the rest. A record with a zero-length key counts as corruption too: it could never have been written by the store, so the replay treats it as a torn tail and truncates.

```mermaid
%%{init: {'theme':'base','themeVariables':{'lineColor':'#64748b'}}}%%
flowchart TD
  START(["Store.Open(path)"]) --> OW["OpenWriter(path)<br/>O_CREATE|O_RDWR, seek EOF"]
  OW --> LOCK["Acquire flock(LOCK_EX|LOCK_NB)<br/>on DB fd"]
  LOCK --> IDX["newIndex() — empty map"]
  IDX --> OR["OpenReader(path)<br/>off = 0"]
  OR --> LOOP{"Next()"}
  LOOP -- "io.EOF<br/>clean end" --> TRUNC_OK["TruncateTo(Offset())<br/>no-op — file already aligned"]
  LOOP -- "ErrCorruptTail<br/>torn prefix / body / CRC / length / empty key" --> WARN["log: corrupt tail at Offset()<br/>break"]
  WARN --> TRUNC["TruncateTo(Offset())<br/>discard tail, fsync"]
  TRUNC --> RET["return Store — ready to append"]
  TRUNC_OK --> RET
  LOOP -- "other I/O error" --> FAIL["fail Open()"]
  LOOP -- "RecordPut" --> PUT["idx.set(Key, Value)"] --> LOOP
  LOOP -- "RecordDelete" --> DEL["idx.delete(Key)"] --> LOOP

  style START fill:#0f172a,stroke:#020617,color:#fff
  style WARN fill:#fef2f2,stroke:#ef4444,color:#7f1d1d
  style TRUNC fill:#eff6ff,stroke:#3b82f6,color:#1e3a8a
  style TRUNC_OK fill:#ecfdf5,stroke:#10b981,color:#065f46
  style RET fill:#ecfdf5,stroke:#10b981,stroke-width:2px,color:#065f46
  style FAIL fill:#fef2f2,stroke:#ef4444,color:#7f1d1d
```

**The one-line proof of self-healing:**

```
open damaged DB → recover → truncate → put new key → close → reopen → new key exists
```

```mermaid
sequenceDiagram
  participant C as Client
  participant S as Store.Open
  participant R as Reader
  participant W as Writer
  participant D as demo.wal

  C->>S: Open("demo.wal")
  S->>R: OpenReader — off=0
  loop Replay
    R->>D: Read 8B prefix
    alt clean record
      R->>R: validate len/CRC/type
      R-->>S: RecordPut/Del
      S->>S: idx.set / delete
    else torn tail (H5 crash)
      D-->>R: short read / CRC mismatch
      R-->>S: ErrCorruptTail
      S->>S: break, warn
    end
  end
  S->>W: TruncateTo(Offset) — fsync
  W->>D: truncate tail, now writable
  C->>S: Put(delta, four)
  S->>W: Append(delta) — at new EOF
  S->>S: Close — fsync + unlock
  C->>S: Reopen — replay again
  S->>R: Next() sees alpha/beta/delta, no gamma
  R-->>C: recovered
```

---

## Locking Model

The single most important fix in this design: **we never create `demo.wal.lock`.** A sidecar lockfile is a lying artifact — it survives `kill -9` and blocks restarts forever. Instead, we lock the database file descriptor *itself*, so the kernel releases the lock the moment the process dies.

```mermaid
%%{init: {'theme':'base'}}%%
flowchart LR
  subgraph Good["✅ v2 — Lock the DB fd"]
    direction TB
    G1["Open demo.wal<br/>O_RDWR"] --> G2["syscall.Flock(fd, LOCK_EX|LOCK_NB)"]
    G2 --> G3["hold until Store.Close()<br/>flock + close"]
    G3 --> G4["process dies<br/>↓<br/>kernel closes fd<br/>↓<br/>lock released automatically"]
  end

  subgraph Bad["❌ v1 — Sidecar (forbidden)"]
    direction TB
    B1["create demo.wal.lock"] --> B2["kill -9"]
    B2 --> B3[".lock file remains<br/>restart blocked<br/>stale state"]
  end

  Good -. prevents .-> Bad

  style Good fill:#ecfdf5,stroke:#10b981,stroke-width:2px
  style Bad fill:#fef2f2,stroke:#ef4444,stroke-width:2px
  style G4 fill:#e0f2fe,stroke:#0ea5e9
  style B3 fill:#fef2f2,stroke:#ef4444
```

```mermaid
flowchart TB
  S["Store"] -- "ONE lock authority<br/>acquires on Open()<br/>holds until Close()" --> F[("DB fd<br/>flock")]
  I["Index<br/>map[string][]byte<br/>NO mutex"] -. "no lock" .-> S
  L1["lock_unix.go<br/>//go:build !windows<br/>syscall.Flock"] -. implements .-> F
  L2["lock_windows.go<br/>//go:build windows<br/>compile-safe fallback"] -. fallback .-> F

  style S fill:#ede9fe,stroke:#8b5cf6,stroke-width:2px
  style F fill:#fffbeb,stroke:#f59e0b,stroke-width:2px
  style I fill:#f1f5f9,stroke:#94a3b8
```

***Platform:*** `lock_unix.go` (`!windows`) uses `syscall.Flock` for a hard guarantee; `lock_windows.go` is a build-safe no-op so the project still compiles there. The demo host is Unix, so the invariant is real.

---

## Durability Policy

No background flusher, no timer ticks. Just a deterministic batch: fsync every 100 appends, and always on `Close()`.

```mermaid
%%{init: {'theme':'base'}}%%
gantt
  title SyncBatch = 100 — deterministic fsync every 100 appends + always on Close()
  dateFormat X
  axisFormat %L
  section Appends
  write 1..99 (no sync) :0, 99
  append 100 → write + fsync :100, 101
  write 101..199 (no sync) :101, 200
  append 200 → write + fsync :200, 201
  Close() → fsync :201, 202
```

```mermaid
flowchart LR
  A["Put()"] --> B["Store.Lock()"]
  B --> C["WAL Append<br/>at EOF"]
  C --> D{"count % 100 == 0 ?"}
  D -- yes --> E["fsync"]
  D -- no --> F["write only"]
  E --> G["idx.set / delete"]
  F --> G
  G --> H["Unlock()"]
  H --> I["Close() always fsyncs<br/>flush → fsync → unlock → close"]

  style E fill:#e0f2fe,stroke:#0ea5e9,stroke-width:2px
  style I fill:#ecfdf5,stroke:#10b981,stroke-width:2px
```

**The ordering rule that matters** — a `Sync` is pointless if the write hasn't reached the OS buffer first:

```
write → Flush → Sync    ✅
write → Sync → Flush    ❌  (data never reaches disk)
```

The base WAL path skips `bufio.Writer` entirely and writes straight to the `os.File`, so this ordering falls out for free.
