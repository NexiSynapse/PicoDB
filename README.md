<div align="center">

<svg width="860" height="220" viewBox="0 0 860 220" xmlns="http://www.w3.org/2000/svg" role="img" aria-label="PicoDB hero">
  <defs>
    <linearGradient id="bg" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#0f172a"/>
      <stop offset="50%" stop-color="#1e3a5f"/>
      <stop offset="100%" stop-color="#0ea5e9"/>
    </linearGradient>
    <linearGradient id="card" x1="0%" y1="0%" x2="100%" y2="0%">
      <stop offset="0%" stop-color="#38bdf8"/>
      <stop offset="100%" stop-color="#818cf8"/>
    </linearGradient>
    <filter id="glow">
      <feGaussianBlur stdDeviation="8" result="coloredBlur"/>
      <feMerge><feMergeNode in="coloredBlur"/><feMergeNode in="SourceGraphic"/></feMerge>
    </filter>
  </defs>
  <rect width="860" height="220" rx="18" fill="url(#bg)"/>
  <!-- grid -->
  <g opacity="0.08" stroke="#fff" stroke-width="0.8">
    <path d="M0 40 H860 M0 80 H860 M0 120 H860 M0 160 H860 M0 200 H860"/>
    <path d="M80 0 V220 M160 0 V220 M240 0 V220 M320 0 V220 M400 0 V220 M480 0 V220 M560 0 V220 M640 0 V220 M720 0 V220 M800 0 V220"/>
  </g>
  <g font-family="Segoe UI, Inter, sans-serif">
    <text x="36" y="58" font-size="13" font-weight="600" letter-spacing="3" fill="#7dd3fc">TRACK D — DATA &amp; STORAGE</text>
    <text x="36" y="108" font-size="54" font-weight="800" fill="#fff" letter-spacing="-1">PicoDB</text>
    <text x="36" y="138" font-size="16" fill="#cbd5e1" letter-spacing="0.3">Embedded crash-safe key-value store — Go stdlib only</text>
    <g transform="translate(36,158)">
      <rect width="118" height="28" rx="14" fill="#0ea5e9"/>
      <text x="59" y="19" text-anchor="middle" font-size="12" font-weight="700" fill="#fff">Go 1.22 • stdlib</text>
    </g>
    <g transform="translate(164,158)">
      <rect width="118" height="28" rx="14" fill="none" stroke="#38bdf8" stroke-width="1.4"/>
      <text x="59" y="19" text-anchor="middle" font-size="12" font-weight="600" fill="#7dd3fc">WAL • CRC32 • flock</text>
    </g>
    <g transform="translate(292,158)">
      <rect width="120" height="28" rx="14" fill="none" stroke="#818cf8" stroke-width="1.4"/>
      <text x="60" y="19" text-anchor="middle" font-size="12" font-weight="600" fill="#c4b5fd">Replay • Truncate</text>
    </g>
  </g>
  <!-- right card: WAL preview -->
  <g transform="translate(560,28)">
    <rect width="264" height="164" rx="14" fill="#ffffff" opacity="0.96"/>
    <text x="18" y="24" font-size="11" font-weight="700" letter-spacing="1.2" fill="#64748b">WAL FRAME</text>
    <g transform="translate(18,36)">
      <rect width="52" height="28" rx="6" fill="#0ea5e9"/><text x="26" y="18" text-anchor="middle" font-size="10" font-weight="700" fill="#fff">Len</text>
      <rect x="60" width="52" height="28" rx="6" fill="#8b5cf6"/><text x="86" y="18" text-anchor="middle" font-size="10" font-weight="700" fill="#fff">CRC</text>
      <rect x="120" width="36" height="28" rx="6" fill="#0f172a"/><text x="138" y="18" text-anchor="middle" font-size="10" font-weight="700" fill="#fff">Typ</text>
      <rect x="164" width="28" height="28" rx="6" fill="#334155"/><text x="178" y="18" text-anchor="middle" font-size="8" font-weight="700" fill="#fff">KLen</text>
      <rect x="200" width="28" height="28" rx="6" fill="#334155"/><text x="214" y="18" text-anchor="middle" font-size="8" font-weight="700" fill="#fff">VLen</text>
    </g>
    <g transform="translate(18,72)">
      <rect width="108" height="28" rx="6" fill="#e0f2fe" stroke="#7dd3fc"/><text x="54" y="18" text-anchor="middle" font-size="11" font-weight="600" fill="#0c4a6e">Key</text>
      <rect x="116" width="112" height="28" rx="6" fill="#ede9fe" stroke="#a78bfa"/><text x="172" y="18" text-anchor="middle" font-size="11" font-weight="600" fill="#4c1d95">Value</text>
    </g>
    <text x="18" y="122" font-size="10" fill="#64748b">RecordLen = 9 + KeyLen + ValLen</text>
    <text x="18" y="138" font-size="9.5" fill="#94a3b8">CRC = CRC32(Type‖KeyLen‖ValLen‖Key‖Value)</text>
    <rect x="18" y="148" width="228" height="6" rx="3" fill="#f1f5f9"/>
    <rect x="18" y="148" width="158" height="6" rx="3" fill="url(#card)"/>
  </g>
</svg>

**Kill it mid-write. Reopen it. Your data is still there. Zero dependencies. Zero goroutines. Just the write-ahead log.**

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat&logo=go&logoColor=white)](#dependency-proof)
[![stdlib only](https://img.shields.io/badge/stdlib-only-0ea5e9?style=flat)](#stdlibmd)
[![WAL](https://img.shields.io/badge/WAL-append--only-8b5cf6?style=flat)](#3-wal-format)
[![CRC32](https://img.shields.io/badge/CRC32-per--record-10b981?style=flat)](#4-recovery-algorithm)
[![flock](https://img.shields.io/badge/lock-flock%20on%20DB%20fd-f59e0b?style=flat)](#5-locking-model)

[Architecture](#2-architecture) • [WAL Format](#3-wal-format) • [Recovery](#4-recovery-algorithm) • [Crash Demo](#8-crash-demo) • [CLI](#7-cli) • [Tests](#12-tests)

</div>

---

## Table of Contents

1. [Description](#1-Description)
2. [Architecture](#2-architecture)
3. [WAL Format](#3-wal-format)
4. [Recovery Algorithm](#4-recovery-algorithm)
5. [Locking Model](#5-locking-model)
6. [Durability Policy](#6-durability-policy)
7. [CLI](#7-cli)
8. [Crash Demo](#8-crash-demo)
9. [Complexity](#9-complexity)
10. [Design Decisions & Trade-offs](#10-design-decisions--trade-offs)
11. [Prior Art](#11-prior-art)
12. [Tests](#12-tests)
13. [Dependency Proof](#13-dependency-proof)
14. [Future Work](#14-future-work)
15. [Rubric Mapping](#15-rubric-mapping)

---

## 1. Description

**PicoDB** is a crash-safe embedded key-value store built entirely on the Go standard library. No external packages, no background threads, no hidden complexity.

What you get:

* an **append-only WAL** with per-record CRC32 integrity,
* an **in-memory key directory** (Bitcask-style hash map),
* **replay-on-open** with automatic truncated-tail repair,
* and **kernel `flock`** on the database file itself — no stale lockfiles.

> Process dies mid-append? The next `Open()` detects the torn record, truncates back to the last clean offset, and the store is immediately writable again. No manual intervention. No data loss for committed records.

```mermaid
%%{init: {'theme':'base','themeVariables':{'primaryColor':'#e0f2fe','primaryTextColor':'#0c4a6e','lineColor':'#0ea5e9','tertiaryColor':'#f0f9ff'}}}%%
flowchart LR
  A["🌱 Seed<br/>put alpha/beta<br/><i>durable</i>"] --> B["⚡ Crash<br/>PICODB_CRASH_AFTER_PREFIX=1<br/><i>prefix only</i>"]
  B --> C["🔍 Reopen<br/>Reader detects<br/><i>ErrCorruptTail</i>"]
  C --> D["✂️ Truncate<br/>to last good offset"]
  D --> E["✅ Recovered<br/>alpha + beta survive<br/>gamma gone"]
  E --> F["➕ Append<br/>delta after repair<br/><i>proves self-healing</i>"]

  style A fill:#ecfdf5,stroke:#10b981,stroke-width:2px
  style B fill:#fef2f2,stroke:#ef4444,stroke-width:2px
  style C fill:#fffbeb,stroke:#f59e0b,stroke-width:2px
  style D fill:#eff6ff,stroke:#3b82f6,stroke-width:2px
  style E fill:#ecfdf5,stroke:#10b981,stroke-width:2px
  style F fill:#f5f3ff,stroke:#8b5cf6,stroke-width:2px
```

---

## 2. Architecture

One writer. One lock holder. No goroutines, no timers, no channels.

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

## 3. WAL Format

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

## 4. Recovery Algorithm

**Policy:** stop at the first corrupt or truncated record and throw away everything after it. For an append-only, single-writer log, everything that precedes the first bad byte is trustworthy — so the engine trusts it and forgets the rest.

```mermaid
%%{init: {'theme':'base','themeVariables':{'lineColor':'#64748b'}}}%%
flowchart TD
  START(["Store.Open(path)"]) --> OW["OpenWriter(path)<br/>O_CREATE|O_RDWR, seek EOF"]
  OW --> LOCK["Acquire flock(LOCK_EX|LOCK_NB)<br/>on DB fd"]
  LOCK --> IDX["newIndex() — empty map"]
  IDX --> OR["OpenReader(path)<br/>off = 0"]
  OR --> LOOP{"Next()"}
  LOOP -- "io.EOF<br/>clean end" --> TRUNC_OK["TruncateTo(Offset())<br/>no-op — file already aligned"]
  LOOP -- "ErrCorruptTail<br/>torn prefix / body / CRC / length" --> WARN["log: corrupt tail at Offset()<br/>break"]
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

## 5. Locking Model

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

## 6. Durability Policy

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

---

## 7. CLI

Small surface, explicit exit codes, and no `interface{}` leaking out of the public API. Tooling that scripts PicoDB never has to guess whether a command worked.

```go
// internal/cli/cli.go
const ( ExitOK=0; ExitError=1; ExitUsage=2; ExitNotFound=3 )
type Stats struct{ Keys int }
func Run(args []string, stdout, stderr io.Writer) int
```

| Command | Example | Exit |
|---|---|---:|
| `put` | `picodb put demo.wal name keshav` | `0` |
| `get` | `picodb get demo.wal name` → `keshav` on stdout | `0` / `3` if missing |
| `del` | `picodb del demo.wal name` | `0` / `3` if missing |
| `dump` | `picodb dump demo.wal` | optional — cut first |

```bash
./picodb put demo.wal alpha one
./picodb get demo.wal alpha      # → one
./picodb del demo.wal alpha
./picodb get demo.wal alpha; echo $?  # → 3
```

`get` prints **only the value** to stdout; everything else is diagnostic and goes to stderr.

---

## 8. Crash Demo

The centerpiece — and it's deterministic. No sleeps, no timing luck, no flaky tests. A fault-injection hook forces the process to exit mid-append, then we prove everything still works on the next open.

```mermaid
sequenceDiagram
  autonumber
  participant SH as crash_demo.sh
  participant DB as demo.wal
  participant W as Writer.Append
  participant K as Kernel

  SH->>DB: put alpha one — fsync (durable)
  SH->>DB: put beta two — fsync (durable)
  Note over SH,W: PICODB_CRASH_AFTER_PREFIX=1
  SH->>W: put gamma THREE
  W->>K: Sync() — flush alpha/beta
  W->>DB: write 8B prefix only (RecordLen|CRC)
  W->>K: Sync() prefix
  W->>W: os.Exit(137) — torn tail created
  Note over DB: alpha ✅ beta ✅ gamma = 8B prefix only ❌
  SH->>W: unset PICODB_CRASH_AFTER_PREFIX
  SH->>DB: get alpha → one
  SH->>DB: get beta → two
  SH->>DB: get gamma → exit 3 (not found, truncated)
  SH->>DB: put delta four → truncates tail, appends
  SH->>DB: get delta → four — self-healing proven
```

**Run it:**

```bash
go build -o picodb ./cmd/picodb
./scripts/crash_demo.sh
# or
make demo-crash
```

**What a reviewer sees when they run it:**

```
== Seed durable data ==
== Trigger deterministic crash ==
crashed process status: 137
== WAL tail after crash ==
00000130: 0000 0013 8f2a c4e1 ...   # 8-byte prefix, no body
== Recovery ==
alpha: one
beta: two
gamma should be missing: exit 3
== Append after recovery ==
delta: four
== Crash recovery demo passed ==
```

> The fault-injection hook exists purely so the demo and tests are reproducible. Normal operation never touches it (`internal/wal/debug.go`).

---

## 9. Complexity

No invented sophistication — just honest, stated bounds.

| Operation | Design | Complexity |
|---|---:|---|
| `Get` | hash-map lookup | **O(1)** avg |
| `Put` | WAL append + hash update | **O(record size)** + O(1) avg |
| `Delete` | WAL append (tombstone) + hash delete | **O(record size)** + O(1) avg |
| `Replay` | sequential log scan | **O(file size)** |
| `CRC` | `crc32.ChecksumIEEE` over body | **O(record size)** |

> We publish complexity bounds, not fake benchmark numbers. If a benchmark ever lands in this repo, it will carry real, measured output.

---

## 10. Design Decisions & Trade-offs

### Strong Guarantees

* CRC-protected records + `RecordLen == 9+KeyLen+ValLen` consistency check
* `MaxRecordSize = 16 MiB` — validated **before** allocation (prevents OOM on corrupt length)
* Single writer, single lock authority (`Store` only)
* Replay + deterministic truncation + `SyncBatch` + `Close` fsync

### Deliberate Compromises

| Compromise | Why |
|---|---|
| Full key index in memory | Simplicity; Bitcask-style — fits the 3-hour hackathon |
| WAL grows without compaction | Compaction is forbidden during sprint (`Plan §4`) |
| `SyncBatch=100` creates a bounded durability window | Explicit trade-off (Redis-style); `Close` is always durable |
| No replication / transactions / MVCC | Out of scope — would dilute rubric points |

### Rejected Ideas

React, REST, ML, Docker, K8s, JWT, cloud deploy — the research mentions them, but they **do not improve the storage engine** and violate `stdlib-only` + time constraints (`Plan §2.5`).

---

## 11. Prior Art

We're not claiming an invention — PicoDB is a deliberately lean synthesis of ideas that serious storage engines have used for decades.

```mermaid
mindmap
  root((PicoDB<br/>stdlib synthesis))
    Bitcask
      append-only log
      in-memory keydir
    LevelDB/RocksDB
      framed records
      integrity-aware storage
    Kafka
      append log framing
      recovery boundaries
    Redis
      explicit fsync frequency
      durability trade-off
    SQLite/Postgres
      crash recovery
      durability semantics
```

> "We took established storage-engine ideas and built the smallest reasonable version of them using nothing but the Go standard library." — `Plan §43`

---

## 12. Tests

A testing pyramid scaled to a three-hour project: one crash-demo up top, integration in the middle, and a foundation of fast unit tests.

```mermaid
%%{init: {'theme':'base'}}%%
flowchart TB
  C["☁️ Crash Demo / Smoke<br/>scripts/crash_demo.sh<br/>deterministic torn tail"] --> I["🔗 Integration — //go:build integration<br/>replay • truncation • append-after-recovery • flock"]
  I --> U["🧪 Unit — go test ./...<br/>WAL framing • CRC • lengths • Store CRUD • CLI exits"]

  style C fill:#f5f3ff,stroke:#8b5cf6,stroke-width:2px
  style I fill:#eff6ff,stroke:#3b82f6,stroke-width:2px
  style U fill:#ecfdf5,stroke:#10b981,stroke-width:2px
```

**Unit matrix** `Plan §28`:

| # | Test | Purpose |
|---:|---|---|
| 1 | `TestEncodeDecodeRoundTrip` | framing |
| 2 | `TestChecksumDetectsCorruption` | CRC |
| 3 | `TestRecordLengthConsistency` | `RecordLen == 9+K+V` |
| 4 | `TestRecordTooLargeRejected` | `MaxRecordSize` |
| 5 | `TestReaderCleanEOF` | normal EOF |
| 6 | `TestReaderTornPrefix` | partial prefix |
| 7 | `TestReaderTornBody` | partial body |
| 8 | `TestReaderUnknownRecordType` | invalid type |
| 9 | `TestPutGetDelete` | Store CRUD |
| 10 | `TestOverwriteKeepsLatestValue` | last write wins |
| 11 | `TestDeleteMissingKey` | sentinel |
| 12 | `TestCLIExitCodes` | `0/1/2/3` |

**Integration matrix** `Plan §29` (`go test -tags=integration ./...` — may use real files, reopen, truncation, subprocesses):

| Test | Purpose |
|---|---|
| `TestReplayAfterSimulatedCrash_TornTail` | centerpiece recovery |
| `TestDeleteTombstoneSurvivesReplay` | delete persists |
| `TestRecoveryTruncatesCorruptTail` | repair occurs |
| `TestReopenAfterRecoveryCanAppend` | WAL writable after repair |
| `TestLockRejectsSecondWriter` | single owner |
| `TestLockReleasedAfterProcessDeath` | no stale lock |

**Run:**

```bash
go test ./...
go test -tags=integration ./...
./scripts/crash_demo.sh
make check   # fmt-check + vet + test + integration
```

---

## 13. Dependency Proof

**The invariant a reviewer can verify in seconds:** a `GOPROXY=off` build with an empty `require ()` block.

```bash
cat go.mod
# module picodb
# go 1.22
# require ()

GOPROXY=off GOTOOLCHAIN=local go list -m all
# → picodb only

GOPROXY=off GOTOOLCHAIN=local go mod verify
# → all modules verified

GOPROXY=off GOTOOLCHAIN=local go build -o picodb ./cmd/picodb
# → no download, no toolchain fetch

# Clean-clone proof
git clone <public-repo> /tmp/picodb-clean
cd /tmp/picodb-clean
GOPROXY=off GOTOOLCHAIN=local go build -o picodb ./cmd/picodb
```

`make deps` runs all four checks in one shot, and `STDLIB.md` lists every single package the project imports.

---

## 14. Future Work

These are the directions we'd take PicoDB next — listed here explicitly so the scope is unambiguous and nothing shown was silently skipped:

```mermaid
flowchart LR
  A["compaction"] --> B["snapshots"]
  B --> C["segment rotation"]
  C --> D["configurable sync policies"]
  D --> E["scan / prefix APIs"]
  E --> F["metrics & observability"]

  style A fill:#f1f5f9,stroke:#94a3b8
  style F fill:#e0f2fe,stroke:#0ea5e9
```

* compaction (garbage-collect tombstones)
* snapshots
* segment rotation
* configurable `SyncBatch` / `fsync` policies
* scan / range APIs
* metrics

Forbidden during sprint: compaction, snapshots, segment rotation, replication, networking, HTTP, multi-key transactions, MVCC, encryption, compression, external packages, background durability goroutines (`Plan §4`).

---

## 15. Rubric Mapping

| Criterion | Weight | Evidence |
|---|---:|---|
| **Functionality & Usefulness** | 35% | `put/get/del`, persistence, replay, `scripts/crash_demo.sh` |
| **Zero-Dependency Craft** | 30% | `require ()`, `GOPROXY=off` proof, `STDLIB.md` |
| **Code Quality & Idiom** | 25% | `gofmt`, `go vet`, `errors.Is`, `Store`-only lock, `SyncBatch`, tests |
| **Innovation** | 10% | deterministic `PICODB_CRASH_AFTER_PREFIX` injection, self-healing truncate, prior-art synthesis |

**Priority rule:** crash safety beats optional features. If something has to go, it's `dump` — never the crash recovery (`Plan §5`).

---

## Repository Layout

```
picodb/
├── go.mod
├── Makefile
├── README.md
├── STDLIB.md
├── cmd/picodb/main.go
├── internal/
│   ├── wal/
│   │   ├── record.go     # Encode/DecodePrefix/Checksum — Worker A
│   │   ├── writer.go     # Append/TruncateTo/Close/SyncBatch — Worker A
│   │   ├── reader.go     # strict validation order — Worker A
│   │   └── debug.go      # deterministic crash hook — Worker A
│   ├── store/
│   │   ├── store.go      # Open/replay/Put/Get/Delete — Worker B
│   │   └── index.go      # plain map, no mutex — Worker B
│   ├── cli/
│   │   └── cli.go        # Run() with Exit codes — Worker B
│   └── lock/
│       ├── lock_unix.go     # flock on DB fd — Worker E
│       └── lock_windows.go  # compile-safe fallback
└── scripts/
    └── crash_demo.sh     # deterministic demo — Worker E
```

---

## Quick Start

```bash
# 1. Build (no network)
GOPROXY=off GOTOOLCHAIN=local go build -o picodb ./cmd/picodb

# 2. Basic CLI
./picodb put demo.wal name keshav
./picodb get demo.wal name        # → keshav
./picodb del demo.wal name
./picodb get demo.wal name; echo $?  # → 3

# 3. Crash recovery (centerpiece)
./scripts/crash_demo.sh

# 4. Quality gates
gofmt -l .                          # must be empty
go vet ./...
go test ./...
go test -tags=integration ./...
make check
```

---

<div align="center">

*Built in a 180-minute hackathon — every minute spent on reliability, not ceremony.*

`WAL` • `CRC32` • `MaxRecordSize` • `Replay` • `Truncate` • `flock` • `SyncBatch` • `Self-healing`

</div>
