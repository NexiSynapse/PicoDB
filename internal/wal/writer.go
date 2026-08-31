package wal

import (
	"fmt"
	"io"
	"os"
)

// SyncBatch controls deterministic durability: fsync every N appends inline.
// No timer, no goroutine, no time-dependent durability test.
const SyncBatch = 100

// Writer is an append-only WAL writer. It always appends at EOF; callers never
// supply an offset. Recovery owns truncation via TruncateTo.
type Writer struct {
	f     *os.File
	path  string
	count int // successful appends since open
}

// OpenWriter opens (or creates) the WAL file at path for appending.
// The file is opened with O_CREATE|O_RDWR so the same descriptor can be used
// for both reading during replay and appending after recovery. The descriptor
// is positioned at EOF.
func OpenWriter(path string) (*Writer, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("open wal: %w", err)
	}
	// Ensure we are at EOF for appends.
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("seek wal: %w", err)
	}
	return &Writer{f: f, path: path}, nil
}

// Append encodes rec and appends it atomically at EOF.
// Returns the start offset of the new record (useful for tests).
func (w *Writer) Append(rec Record) (int64, error) {
	// Input validation before encoding.
	if !ValidRecordType(rec.Type) {
		return 0, fmt.Errorf("wal: invalid record type %d: %w", rec.Type, ErrCorruptTail)
	}
	keyLen := len(rec.Key)
	valLen := len(rec.Value)
	if rec.Type == RecordDelete {
		valLen = 0
	}
	recordLen := 9 + keyLen + valLen
	if recordLen > MaxRecordSize {
		return 0, fmt.Errorf("wal: record too large %d > %d: %w", recordLen, MaxRecordSize, ErrCorruptTail)
	}

	encoded := Encode(rec)
	if len(encoded) < PrefixSize {
		return 0, fmt.Errorf("wal: encode failed: %w", ErrCorruptTail)
	}
	prefix := encoded[:PrefixSize]

	// Determine current EOF offset (append point).
	offset, err := w.f.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, fmt.Errorf("wal seek: %w", err)
	}

	// Deterministic crash injection: flush prior records, write only prefix, exit.
	// Isolated in debug.go so normal path has zero overhead when disabled.
	if maybeCrashAfterPrefix(w.f, prefix) {
		// unreachable - process exits inside maybeCrashAfterPrefix
		return offset, nil
	}

	// Normal path: write full record.
	if _, err := w.f.Write(encoded); err != nil {
		return 0, fmt.Errorf("wal write: %w", err)
	}
	w.count++
	// Deterministic SyncBatch: fsync every N appends inline.
	if w.count%SyncBatch == 0 {
		if err := w.f.Sync(); err != nil {
			return offset, fmt.Errorf("wal sync: %w", err)
		}
	}
	return offset, nil
}

// TruncateTo truncates the WAL file to exactly offset bytes, discarding the
// corrupt tail. It syncs the file after truncation so the repair is durable.
// offset must be <= current file size and aligned to a record boundary (caller
// ensures this via Reader.Offset).
func (w *Writer) TruncateTo(offset int64) error {
	if offset < 0 {
		return fmt.Errorf("wal truncate: negative offset %d", offset)
	}
	// Seek to offset to ensure file position is consistent after truncate.
	if err := w.f.Truncate(offset); err != nil {
		return fmt.Errorf("wal truncate: %w", err)
	}
	if _, err := w.f.Seek(offset, io.SeekStart); err != nil {
		return fmt.Errorf("wal seek after truncate: %w", err)
	}
	if err := w.f.Sync(); err != nil {
		return fmt.Errorf("wal sync after truncate: %w", err)
	}
	return nil
}

// Sync forces buffered data to stable storage.
func (w *Writer) Sync() error {
	if err := w.f.Sync(); err != nil {
		return fmt.Errorf("wal sync: %w", err)
	}
	return nil
}

// Close flushes pending writes, fsyncs, and closes the descriptor.
// Normal shutdown is fully durable.
func (w *Writer) Close() error {
	// No hidden bufio buffer in base path, but sync before close for durability.
	if err := w.f.Sync(); err != nil {
		_ = w.f.Close()
		return fmt.Errorf("wal sync on close: %w", err)
	}
	if err := w.f.Close(); err != nil {
		return fmt.Errorf("wal close: %w", err)
	}
	return nil
}

// File returns the underlying os.File for lock ownership (used by Store).
// The caller must not close it.
func (w *Writer) File() *os.File { return w.f }

// Path returns the WAL file path.
func (w *Writer) Path() string { return w.path }
