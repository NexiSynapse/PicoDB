//go:build !windows

package lock

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

var ErrLocked = errors.New("lock: database file is locked by another process")

type FileLock struct {
	file *os.File
}

// LockFile acquires an exclusive, non-blocking lock on the given file descriptor using flock.
func LockFile(f *os.File) (*FileLock, error) {
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, ErrLocked
		}
		return nil, fmt.Errorf("lock: flock acquire: %w", err)
	}

	return &FileLock{file: f}, nil
}

// Unlock releases the flock on the file descriptor.
func (l *FileLock) Unlock() error {
	if l.file == nil {
		return nil
	}
	err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	l.file = nil
	if err != nil {
		return fmt.Errorf("lock: flock release: %w", err)
	}
	return nil
}
