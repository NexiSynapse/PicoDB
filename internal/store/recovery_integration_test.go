//go:build integration

package store

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"microdb/internal/wal"
)

func TestReplayAfterSimulatedCrash_TornTail(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "crash_sim.wal")

	// Step 1: Open Store and write records A and B
	s1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("step 1: Open failed: %v", err)
	}

	if err := s1.Put([]byte("key_A"), []byte("val_A")); err != nil {
		t.Fatalf("put A failed: %v", err)
	}
	if err := s1.Put([]byte("key_B"), []byte("val_B")); err != nil {
		t.Fatalf("put B failed: %v", err)
	}
	if err := s1.Close(); err != nil {
		t.Fatalf("close s1 failed: %v", err)
	}

	// Record the file size after A and B are durably written
	infoAB, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	sizeAB := infoAB.Size()

	// Step 2: Simulate crash during append of record C by appending only 8-byte prefix
	recC := wal.Record{
		Type:  wal.RecordPut,
		Key:   []byte("key_C"),
		Value: []byte("val_C_interrupted"),
	}
	encodedC := wal.Encode(recC)
	prefixC := encodedC[:wal.PrefixSize]

	f, err := os.OpenFile(dbPath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		t.Fatalf("open for crash injection failed: %v", err)
	}
	if _, err := f.Write(prefixC); err != nil {
		t.Fatalf("write torn prefix failed: %v", err)
	}
	_ = f.Sync()
	_ = f.Close()

	// Step 3: Reopen database - recovery should detect corrupt tail, warn, and truncate
	s2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("step 3: Open after crash failed: %v", err)
	}

	// Verify A and B exist
	valA, err := s2.Get([]byte("key_A"))
	if err != nil || !bytes.Equal(valA, []byte("val_A")) {
		t.Fatalf("expected key_A='val_A', got %q, err=%v", valA, err)
	}

	valB, err := s2.Get([]byte("key_B"))
	if err != nil || !bytes.Equal(valB, []byte("val_B")) {
		t.Fatalf("expected key_B='val_B', got %q, err=%v", valB, err)
	}

	// Verify C is absent
	_, err = s2.Get([]byte("key_C"))
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected key_C to be absent, got err=%v", err)
	}

	// Verify truncation back to sizeAB
	infoAfterRecovery, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("stat after recovery failed: %v", err)
	}
	if infoAfterRecovery.Size() != sizeAB {
		t.Fatalf("expected recovered file size %d, got %d", sizeAB, infoAfterRecovery.Size())
	}

	// Step 4: Append new record D after recovery
	if err := s2.Put([]byte("key_D"), []byte("val_D")); err != nil {
		t.Fatalf("put D failed: %v", err)
	}
	if err := s2.Close(); err != nil {
		t.Fatalf("close s2 failed: %v", err)
	}

	// Step 5: Reopen again - verify A, B, and D exist intact
	s3, err := Open(dbPath)
	if err != nil {
		t.Fatalf("step 5: Open final failed: %v", err)
	}
	defer s3.Close()

	valA, err = s3.Get([]byte("key_A"))
	if err != nil || !bytes.Equal(valA, []byte("val_A")) {
		t.Fatalf("final check: expected key_A='val_A', got %q", valA)
	}
	valB, err = s3.Get([]byte("key_B"))
	if err != nil || !bytes.Equal(valB, []byte("val_B")) {
		t.Fatalf("final check: expected key_B='val_B', got %q", valB)
	}
	valD, err := s3.Get([]byte("key_D"))
	if err != nil || !bytes.Equal(valD, []byte("val_D")) {
		t.Fatalf("final check: expected key_D='val_D', got %q", valD)
	}
	_, err = s3.Get([]byte("key_C"))
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("final check: expected key_C to be absent, got %v", err)
	}

	if s3.Len() != 3 {
		t.Fatalf("expected store Len 3, got %d", s3.Len())
	}
}

func TestDeleteTombstoneSurvivesReplay(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "tombstone.wal")

	s1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	_ = s1.Put([]byte("k1"), []byte("v1"))
	_ = s1.Put([]byte("k2"), []byte("v2"))
	_ = s1.Delete([]byte("k1"))
	_ = s1.Close()

	s2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Reopen failed: %v", err)
	}
	defer s2.Close()

	_, err = s2.Get([]byte("k1"))
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected k1 to be deleted across replay, got err=%v", err)
	}

	val, err := s2.Get([]byte("k2"))
	if err != nil || !bytes.Equal(val, []byte("v2")) {
		t.Fatalf("expected k2='v2', got %q", val)
	}
}

func TestRecoveryTruncatesCorruptTail(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "corrupt_tail.wal")

	s1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	_ = s1.Put([]byte("x"), []byte("12345"))
	_ = s1.Close()

	infoBefore, _ := os.Stat(dbPath)
	validSize := infoBefore.Size()

	// Append corrupt trailing bytes
	f, _ := os.OpenFile(dbPath, os.O_WRONLY|os.O_APPEND, 0666)
	_, _ = f.Write([]byte{0xDE, 0xAD, 0xBE, 0xEF, 0x01, 0x02})
	_ = f.Close()

	// Reopen
	s2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open after corrupt tail failed: %v", err)
	}
	_ = s2.Close()

	infoAfter, _ := os.Stat(dbPath)
	if infoAfter.Size() != validSize {
		t.Fatalf("expected file truncated back to %d, got %d", validSize, infoAfter.Size())
	}
}

func TestLockRejectsSecondWriter(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "single_writer.wal")

	s1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("first Open failed: %v", err)
	}
	defer s1.Close()

	_, err = Open(dbPath)
	if err == nil {
		t.Fatalf("expected second Open to fail due to file lock")
	}
}
