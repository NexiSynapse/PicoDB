# Crash Demo

The centerpiece — and it's deterministic. No sleeps, no timing luck, no flaky tests. A fault-injection hook forces the process to exit mid-append, then we prove everything still works on the next open.

## How it works

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

## Run it

```bash
go build -o picodb ./cmd/picodb
./scripts/crash_demo.sh
# or
make demo-crash
```

## Output

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

> The fault-injection hook exists purely so the demo and tests are reproducible. Normal operation never touches it (`internal/wal/debug.go`). For backward compatibility it still honors the old `MICRODB_CRASH_AFTER_PREFIX` name as well as `PICODB_CRASH_AFTER_PREFIX`.
