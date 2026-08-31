package wal

import (
	"fmt"
	"io"
	"os"
)

const SyncBatch = 100

type Writer struct {
	file             *os.File
	offset           int64
	appendsSinceSync int
}

// OpenWriter opens the WAL file at the specified path in append/read-write mode.
func OpenWriter(path string) (*Writer, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("wal: open writer file: %w", err)
	}

	offset, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("wal: seek to end: %w", err)
	}

	return &Writer{
		file:   f,
		offset: offset,
	}, nil
}

// File returns the underlying *os.File descriptor for locking and coordinated operations.
func (w *Writer) File() *os.File {
	return w.file
}

// Append writes a record to the end of the WAL.
func (w *Writer) Append(rec Record) (int64, error) {
	recordOffset := w.offset
	encoded := Encode(rec)

	if shouldCrashAfterPrefix() {
		// Deterministic crash injection: ensure prior records are durable,
		// write only the 8-byte prefix, sync, and exit immediately.
		_ = w.file.Sync()
		prefixOnly := encoded[:PrefixSize]
		_, _ = w.file.Write(prefixOnly)
		_ = w.file.Sync()
		crashProcess()
	}

	n, err := w.file.Write(encoded)
	if err != nil {
		return 0, fmt.Errorf("wal: append write: %w", err)
	}

	w.offset += int64(n)
	w.appendsSinceSync++

	if w.appendsSinceSync >= SyncBatch {
		if err := w.file.Sync(); err != nil {
			return 0, fmt.Errorf("wal: sync: %w", err)
		}
		w.appendsSinceSync = 0
	}

	return recordOffset, nil
}

// TruncateTo truncates the WAL file to the specified offset and seeks to it.
func (w *Writer) TruncateTo(offset int64) error {
	if err := w.file.Truncate(offset); err != nil {
		return fmt.Errorf("wal: truncate to %d: %w", offset, err)
	}
	if _, err := w.file.Seek(offset, io.SeekStart); err != nil {
		return fmt.Errorf("wal: seek to %d: %w", offset, err)
	}
	if err := w.file.Sync(); err != nil {
		return fmt.Errorf("wal: sync after truncate: %w", err)
	}
	w.offset = offset
	w.appendsSinceSync = 0
	return nil
}

// Sync flushes all pending file writes to disk.
func (w *Writer) Sync() error {
	w.appendsSinceSync = 0
	return w.file.Sync()
}

// Offset returns the current file offset (end of WAL).
func (w *Writer) Offset() int64 {
	return w.offset
}

// Close flushes, fsyncs, and closes the WAL file descriptor.
func (w *Writer) Close() error {
	if w.file == nil {
		return nil
	}
	syncErr := w.file.Sync()
	closeErr := w.file.Close()
	w.file = nil
	if syncErr != nil {
		return fmt.Errorf("wal: sync on close: %w", syncErr)
	}
	if closeErr != nil {
		return fmt.Errorf("wal: close writer: %w", closeErr)
	}
	return nil
}
