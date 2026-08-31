package store

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"
)

func TestPutGetDelete(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "crud.wal")

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer s.Close()

	key := []byte("username")
	val := []byte("alice")

	// Get on empty database should return ErrKeyNotFound
	_, err = s.Get(key)
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}

	// Put
	if err := s.Put(key, val); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Get
	got, err := s.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !bytes.Equal(got, val) {
		t.Fatalf("expected value %q, got %q", val, got)
	}

	// Delete
	if err := s.Delete(key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Get after Delete
	_, err = s.Get(key)
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound after Delete, got %v", err)
	}
}

func TestOverwriteKeepsLatestValue(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "overwrite.wal")

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer s.Close()

	key := []byte("counter")

	for i := 1; i <= 5; i++ {
		val := []byte{byte(i)}
		if err := s.Put(key, val); err != nil {
			t.Fatalf("Put %d failed: %v", i, err)
		}
	}

	got, err := s.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !bytes.Equal(got, []byte{5}) {
		t.Fatalf("expected latest value [5], got %v", got)
	}
	if s.Len() != 1 {
		t.Fatalf("expected Len 1, got %d", s.Len())
	}
}

func TestDeleteMissingKey(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "del_missing.wal")

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer s.Close()

	err = s.Delete([]byte("non_existent_key"))
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestStoreReplayNormal(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "replay.wal")

	// Phase 1: Write entries and close
	s1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open phase 1 failed: %v", err)
	}

	if err := s1.Put([]byte("k1"), []byte("v1")); err != nil {
		t.Fatalf("Put k1 failed: %v", err)
	}
	if err := s1.Put([]byte("k2"), []byte("v2")); err != nil {
		t.Fatalf("Put k2 failed: %v", err)
	}
	if err := s1.Put([]byte("k3"), []byte("v3")); err != nil {
		t.Fatalf("Put k3 failed: %v", err)
	}
	if err := s1.Delete([]byte("k2")); err != nil {
		t.Fatalf("Delete k2 failed: %v", err)
	}
	if err := s1.Close(); err != nil {
		t.Fatalf("Close phase 1 failed: %v", err)
	}

	// Phase 2: Reopen and verify recovered state
	s2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open phase 2 failed: %v", err)
	}
	defer s2.Close()

	v1, err := s2.Get([]byte("k1"))
	if err != nil || !bytes.Equal(v1, []byte("v1")) {
		t.Fatalf("expected k1='v1', got %q, err=%v", v1, err)
	}

	_, err = s2.Get([]byte("k2"))
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected k2 to be deleted, got err=%v", err)
	}

	v3, err := s2.Get([]byte("k3"))
	if err != nil || !bytes.Equal(v3, []byte("v3")) {
		t.Fatalf("expected k3='v3', got %q, err=%v", v3, err)
	}

	if s2.Len() != 2 {
		t.Fatalf("expected Len 2, got %d", s2.Len())
	}
}

func TestStoreLen(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "len.wal")

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer s.Close()

	if s.Len() != 0 {
		t.Fatalf("expected empty store, got %d", s.Len())
	}

	_ = s.Put([]byte("a"), []byte("1"))
	_ = s.Put([]byte("b"), []byte("2"))
	_ = s.Put([]byte("c"), []byte("3"))

	if s.Len() != 3 {
		t.Fatalf("expected Len 3, got %d", s.Len())
	}

	_ = s.Delete([]byte("b"))
	if s.Len() != 2 {
		t.Fatalf("expected Len 2 after delete, got %d", s.Len())
	}
}
