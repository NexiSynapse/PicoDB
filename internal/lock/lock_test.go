package lock

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLockRejectsSecondAcquisition(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lock_test.db")

	f1, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		t.Fatalf("open f1 failed: %v", err)
	}
	defer f1.Close()

	l1, err := LockFile(f1)
	if err != nil {
		t.Fatalf("acquire l1 failed: %v", err)
	}
	defer l1.Unlock()

	f2, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		t.Fatalf("open f2 failed: %v", err)
	}
	defer f2.Close()

	_, err = LockFile(f2)
	if !errors.Is(err, ErrLocked) {
		t.Fatalf("expected ErrLocked on second lock attempt, got: %v", err)
	}
}

func TestUnlockAllowsReacquisition(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "reacquire_test.db")

	f1, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		t.Fatalf("open f1 failed: %v", err)
	}
	defer f1.Close()

	l1, err := LockFile(f1)
	if err != nil {
		t.Fatalf("acquire l1 failed: %v", err)
	}

	if err := l1.Unlock(); err != nil {
		t.Fatalf("unlock l1 failed: %v", err)
	}

	f2, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		t.Fatalf("open f2 failed: %v", err)
	}
	defer f2.Close()

	l2, err := LockFile(f2)
	if err != nil {
		t.Fatalf("acquire l2 failed after l1 unlocked: %v", err)
	}
	_ = l2.Unlock()
}
