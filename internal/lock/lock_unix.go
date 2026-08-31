//go:build !windows

package lock

import (
	"fmt"
	"os"
	"syscall"
)

// FileLock holds an exclusive flock on the database file descriptor.
// The lock is released automatically when the process dies (kernel closes fd).
type FileLock struct {
	f *os.File
}

// TryLock acquires an exclusive non-blocking flock on f.
// It returns an error if another process already holds the lock.
func TryLock(f *os.File) (*FileLock, error) {
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return nil, fmt.Errorf("lock: database is locked by another process: %w", err)
	}
	return &FileLock{f: f}, nil
}

// Unlock releases the flock.
func (l *FileLock) Unlock() error {
	if l == nil || l.f == nil {
		return nil
	}
	if err := syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN); err != nil {
		return fmt.Errorf("unlock: %w", err)
	}
	return nil
}
