#!/usr/bin/env bash
set -euo pipefail

DB=demo.wal

rm -f "$DB"

echo "== Seed durable data =="

./microdb put "$DB" alpha one
./microdb put "$DB" beta two

echo "== Trigger deterministic crash =="

export MICRODB_CRASH_AFTER_PREFIX=1

set +e
./microdb put "$DB" gamma THREE
STATUS=$?
set -e

echo "crashed process status: $STATUS"

unset MICRODB_CRASH_AFTER_PREFIX

echo "== WAL tail after crash =="

xxd "$DB" | tail -5

echo "== Recovery =="

echo "alpha:"
./microdb get "$DB" alpha

echo "beta:"
./microdb get "$DB" beta

echo "gamma should be missing:"

set +e
./microdb get "$DB" gamma
GAMMA_STATUS=$?
set -e

test "$GAMMA_STATUS" -eq 3

echo "== Append after recovery =="

./microdb put "$DB" delta four
./microdb get "$DB" delta

echo "== Crash recovery demo passed =="
