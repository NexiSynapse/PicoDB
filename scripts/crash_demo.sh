#!/usr/bin/env bash
set -euo pipefail

DB=demo.wal
BIN=./microdb

if [ ! -f "$BIN" ]; then
  echo "building microdb..."
  go build -o "$BIN" ./cmd/microdb
fi

rm -f "$DB"

echo "== Seed durable data =="
"$BIN" put "$DB" alpha one
"$BIN" put "$DB" beta two
echo "alpha/beta durable"

echo ""
echo "== Trigger deterministic crash (MICRODB_CRASH_AFTER_PREFIX=1) =="
export MICRODB_CRASH_AFTER_PREFIX=1
set +e
"$BIN" put "$DB" gamma THREE
STATUS=$?
set -e
echo "crashed process status: $STATUS (expected 137 or non-zero)"

unset MICRODB_CRASH_AFTER_PREFIX

echo ""
echo "== WAL tail after crash (hexdump) =="
if command -v xxd >/dev/null 2>&1; then
  xxd "$DB" | tail -5
elif command -v hexdump >/dev/null 2>&1; then
  hexdump -C "$DB" | tail -5
else
  od -An -tx1 "$DB" | tail -5
fi
echo "file size: $(wc -c < "$DB") bytes"

echo ""
echo "== Recovery: get surviving keys =="
echo -n "alpha: "
"$BIN" get "$DB" alpha
echo -n "beta: "
"$BIN" get "$DB" beta

echo -n "gamma should be missing: "
set +e
"$BIN" get "$DB" gamma
GAMMA_STATUS=$?
set -e
if [ "$GAMMA_STATUS" -eq 3 ]; then
  echo "gamma correctly missing (exit 3)"
else
  echo "UNEXPECTED gamma status $GAMMA_STATUS (want 3)"
  exit 1
fi

echo ""
echo "== Append after recovery (self-healing) =="
"$BIN" put "$DB" delta four
echo -n "delta: "
"$BIN" get "$DB" delta

echo ""
echo "== Reopen proof: A,B,D survive =="
"$BIN" get "$DB" alpha >/dev/null && echo "alpha OK"
"$BIN" get "$DB" beta  >/dev/null && echo "beta OK"
"$BIN" get "$DB" delta >/dev/null && echo "delta OK"

echo ""
echo "== Crash recovery demo PASSED =="
