#!/usr/bin/env bash
set -euo pipefail

DB=demo.wal

rm -f "$DB"

echo "== Seed durable data =="

./picodb put "$DB" alpha one
./picodb put "$DB" beta two

echo "== Trigger deterministic crash =="

export PICODB_CRASH_AFTER_PREFIX=1

set +e
./picodb put "$DB" gamma THREE
STATUS=$?
set -e

echo "crashed process status: $STATUS"

unset PICODB_CRASH_AFTER_PREFIX

echo "== WAL tail after crash =="

xxd "$DB" | tail -5

echo "== Recovery =="

echo "alpha:"
./picodb get "$DB" alpha

echo "beta:"
./picodb get "$DB" beta

echo "gamma should be missing:"

set +e
./picodb get "$DB" gamma
GAMMA_STATUS=$?
set -e

test "$GAMMA_STATUS" -eq 3

echo "== Append after recovery =="

./picodb put "$DB" delta four
./picodb get "$DB" delta

echo "== Crash recovery demo passed =="
