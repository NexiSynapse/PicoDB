//go:build windows

package lock

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var ErrLocked = errors.New("lock: database file is locked by another process")

const (
	LOCKFILE_FAIL_IMMEDIATELY = 0x00000001
	LOCKFILE_EXCLUSIVE_LOCK   = 0x00000002
	ERROR_LOCK_VIOLATION      = 33
)

var (
	modkernel32      = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = modkernel32.NewProc("LockFileEx")
	procUnlockFileEx = modkernel32.NewProc("UnlockFileEx")
)

type FileLock struct {
	file *os.File
}

// LockFile acquires an exclusive, non-blocking lock on the Windows file handle.
func LockFile(f *os.File) (*FileLock, error) {
	var ol syscall.Overlapped
	r1, _, err := procLockFileEx.Call(
		f.Fd(),
		uintptr(LOCKFILE_EXCLUSIVE_LOCK|LOCKFILE_FAIL_IMMEDIATELY),
		0,
		1,
		0,
		uintptr(unsafe.Pointer(&ol)),
	)
	if r1 == 0 {
		var errno syscall.Errno
		if errors.As(err, &errno) && errno == ERROR_LOCK_VIOLATION {
			return nil, ErrLocked
		}
		if errors.Is(err, syscall.Errno(ERROR_LOCK_VIOLATION)) {
			return nil, ErrLocked
		}
		return nil, fmt.Errorf("lock: LockFileEx acquire: %w", err)
	}

	return &FileLock{file: f}, nil
}

// Unlock releases the lock on the Windows file handle.
func (l *FileLock) Unlock() error {
	if l.file == nil {
		return nil
	}
	var ol syscall.Overlapped
	r1, _, err := procUnlockFileEx.Call(
		l.file.Fd(),
		0,
		1,
		0,
		uintptr(unsafe.Pointer(&ol)),
	)
	l.file = nil
	if r1 == 0 {
		return fmt.Errorf("lock: UnlockFileEx release: %w", err)
	}
	return nil
}
