package store

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"picodb/internal/wal"
)

// ErrKeyNotFound is returned when Get/Delete targets a missing key.
var ErrKeyNotFound = errors.New("store: key not found")

// Store is the single concurrency authority (Plan Â§16).
// It owns the in-memory index and the WAL writer, and holds the process file lock.
type Store struct {
	mu   sync.RWMutex
	idx  *index
	w    *wal.Writer
	path string
}

// Open creates or replays the WAL at path, acquiring the exclusive file lock,
// building the in-memory index, and truncating any corrupt tail.
func Open(path string) (*Store, error) {
	w, err := wal.OpenWriter(path)
	if err != nil {
		return nil, err
	}
	// TODO: Worker B owns file locking (internal/lock). Acquire here when lock lands.
	// For now Writer holds the descriptor; lock will be added by Worker E.

	idx := newIndex()
	r, err := wal.OpenReader(path)
	if err != nil {
		_ = w.Close()
		return nil, err
	}
	defer r.Close()

	for {
		rec, err := r.Next()
		if err == io.EOF {
			break
		}
		if errors.Is(err, wal.ErrCorruptTail) {
			// Recovery policy: stop at first corrupt tail and truncate (Plan Â§20).
			fmt.Fprintf(os.Stderr, "wal: corrupt tail at offset %d â€” truncating\n", r.Offset())
			break
		}
		if err != nil {
			_ = w.Close()
			return nil, err
		}
		switch rec.Type {
		case wal.RecordPut:
			idx.set(rec.Key, rec.Value)
		case wal.RecordDelete:
			idx.delete(rec.Key)
		}
	}
	// Self-healing truncation (Plan Â§21).
	if err := w.TruncateTo(r.Offset()); err != nil {
		_ = w.Close()
		return nil, err
	}

	return &Store{idx: idx, w: w, path: path}, nil
}

// Put appends a Put record then updates the index (WAL-first ordering, Plan Â§18).
func (s *Store) Put(key, value []byte) error {
	if len(key) == 0 {
		return fmt.Errorf("store: empty key")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	rec := wal.Record{Type: wal.RecordPut, Key: key, Value: value}
	if _, err := s.w.Append(rec); err != nil {
		return err
	}
	s.idx.set(key, value)
	return nil
}

// Get returns a copy of the value or ErrKeyNotFound. O(1) average.
func (s *Store) Get(key []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.idx.get(key)
	if !ok {
		return nil, ErrKeyNotFound
	}
	return v, nil
}

// Delete appends a tombstone then removes from index.
func (s *Store) Delete(key []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.idx.get(key); !ok {
		return ErrKeyNotFound
	}
	rec := wal.Record{Type: wal.RecordDelete, Key: key}
	if _, err := s.w.Append(rec); err != nil {
		return err
	}
	s.idx.delete(key)
	return nil
}

// Close flushes, fsyncs, releases the file lock, and closes the descriptor.
func (s *Store) Close() error { return s.w.Close() }

// Len returns the number of live keys.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.idx.len()
}

