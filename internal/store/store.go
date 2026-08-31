package store

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"picodb/internal/lock"
	"picodb/internal/wal"
)

var ErrKeyNotFound = errors.New("store: key not found")

// Store represents the embedded key-value store instance.
// Store is the single lock authority for both WAL mutations and index updates.
type Store struct {
	mu       sync.RWMutex
	idx      *index
	w        *wal.Writer
	path     string
	fileLock *lock.FileLock
}

// Open initializes the store, acquires the exclusive process file lock,
// replays existing records to populate the in-memory index, and performs
// self-healing tail truncation if corruption is detected.
func Open(path string) (*Store, error) {
	writer, err := wal.OpenWriter(path)
	if err != nil {
		return nil, fmt.Errorf("store: open wal writer: %w", err)
	}

	flock, err := lock.LockFile(writer.File())
	if err != nil {
		_ = writer.Close()
		return nil, fmt.Errorf("store: acquire file lock: %w", err)
	}

	idx := newIndex()
	reader := wal.NewReader(writer.File())

	var replayErr error
	for {
		rec, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if errors.Is(err, wal.ErrCorruptTail) {
			fmt.Fprintf(os.Stderr, "picodb: wal recovery: corrupt tail detected at offset %d, truncating\n", reader.Offset())
			break
		}
		if err != nil {
			replayErr = fmt.Errorf("store: replay scan: %w", err)
			break
		}

		if rec.Type == wal.RecordPut {
			idx.set(rec.Key, rec.Value)
		} else if rec.Type == wal.RecordDelete {
			idx.delete(rec.Key)
		}
	}

	validOffset := reader.Offset()
	_ = reader.Close()

	if replayErr != nil {
		_ = flock.Unlock()
		_ = writer.Close()
		return nil, replayErr
	}

	// Self-healing: truncate WAL to the last verified record offset
	if err := writer.TruncateTo(validOffset); err != nil {
		_ = flock.Unlock()
		_ = writer.Close()
		return nil, fmt.Errorf("store: truncate wal during recovery: %w", err)
	}

	return &Store{
		idx:      idx,
		w:        writer,
		path:     path,
		fileLock: flock,
	}, nil
}

// Put writes a key-value record to the WAL and updates the in-memory index.
// Invariant: WAL append occurs strictly before index update.
func (s *Store) Put(key, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rec := wal.Record{
		Type:  wal.RecordPut,
		Key:   key,
		Value: value,
	}

	if _, err := s.w.Append(rec); err != nil {
		return fmt.Errorf("store: put append: %w", err)
	}

	s.idx.set(key, value)
	return nil
}

// Get looks up a key from the in-memory index.
func (s *Store) Get(key []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.idx.get(key)
	if !ok {
		return nil, ErrKeyNotFound
	}
	return val, nil
}

// Delete removes a key by appending a tombstone record to the WAL and deleting from the index.
func (s *Store) Delete(key []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.idx.get(key); !ok {
		return ErrKeyNotFound
	}

	rec := wal.Record{
		Type:  wal.RecordDelete,
		Key:   key,
		Value: nil,
	}

	if _, err := s.w.Append(rec); err != nil {
		return fmt.Errorf("store: delete append: %w", err)
	}

	s.idx.delete(key)
	return nil
}

// Len returns the count of active keys in the database.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.idx.len()
}

// Close ensures all buffered WAL writes are fsynced, releases the process lock,
// and closes the underlying file descriptor.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var firstErr error

	if s.fileLock != nil {
		if err := s.fileLock.Unlock(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("store: release file lock: %w", err)
		}
		s.fileLock = nil
	}

	if s.w != nil {
		if err := s.w.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("store: close wal writer: %w", err)
		}
		s.w = nil
	}

	return firstErr
}
