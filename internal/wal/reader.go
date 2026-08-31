package wal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

// Reader sequentially scans the WAL for replay.
// It follows the exact validation order from Plan.md §11:
//
//	Read 8-byte prefix
//	  -> clean EOF -> io.EOF
//	  -> short prefix -> ErrCorruptTail
//	Validate RecordLen <= MaxRecordSize -> ErrCorruptTail
//	Read exact body length -> short read -> ErrCorruptTail
//	Decode Type/KeyLen/ValLen -> validate RecordLen == 9+KeyLen+ValLen -> ErrCorruptTail
//	Validate record type -> ErrCorruptTail
//	Calculate CRC -> mismatch -> ErrCorruptTail
//	Return record
//
// No invalid length reaches an unbounded allocation.
type Reader struct {
	f   *os.File
	off int64 // file offset of next record to read (start of next prefix)
}

// OpenReader opens the WAL file at path for sequential reading from offset 0.
func OpenReader(path string) (*Reader, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Treat missing file as empty WAL; caller may create via Writer.
			// Create an empty file so Offset/Truncate work downstream, but keep reader empty.
			// Instead return a reader with no file and immediate EOF.
			// For simplicity, create file then open.
			nf, cerr := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0644)
			if cerr != nil {
				return nil, fmt.Errorf("open wal reader: %w", err)
			}
			return &Reader{f: nf, off: 0}, nil
		}
		return nil, fmt.Errorf("open wal reader: %w", err)
	}
	return &Reader{f: f, off: 0}, nil
}

// Offset returns the file offset of the next record to read.
// On success, it is positioned after the last fully-validated record.
// On ErrCorruptTail, it is the start offset of the corrupt/torn record, suitable
// for Writer.TruncateTo to discard the tail.
func (r *Reader) Offset() int64 { return r.off }

// Close closes the reader's file descriptor.
func (r *Reader) Close() error {
	if r.f == nil {
		return nil
	}
	return r.f.Close()
}

// Next returns the next valid record or an error.
// io.EOF indicates clean end-of-file (no partial record).
// ErrCorruptTail indicates the first corrupt/truncated tail record.
func (r *Reader) Next() (*Record, error) {
	cur := r.off
	var prefix [PrefixSize]byte
	// Ensure file offset matches logical offset (in case of prior Seek).
	if _, err := r.f.Seek(cur, io.SeekStart); err != nil {
		return nil, fmt.Errorf("wal seek: %w", err)
	}
	n, err := io.ReadFull(r.f, prefix[:])
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			// io.ReadFull returns ErrUnexpectedEOF on short read; map both to tail error
			// unless it was a clean EOF with 0 bytes.
			if n == 0 && err == io.EOF {
				return nil, io.EOF
			}
			return nil, ErrCorruptTail
		}
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, ErrCorruptTail
	}
	recordLen, wantCRC := DecodePrefix(prefix)
	if recordLen > MaxRecordSize {
		return nil, ErrCorruptTail
	}
	// Guard against impossible sizes before allocation.
	// We still need to read the body to detect torn writes, but we already know
	// recordLen is bounded so allocation is safe.
	body := make([]byte, recordLen)
	if _, err := io.ReadFull(r.f, body); err != nil {
		return nil, ErrCorruptTail
	}
	if len(body) < BodyHeaderSize {
		return nil, ErrCorruptTail
	}
	typ := RecordType(body[0])
	keyLen := binary.BigEndian.Uint32(body[1:5])
	valLen := binary.BigEndian.Uint32(body[5:9])

	// Validate structural consistency: RecordLen == 9 + KeyLen + ValLen
	if recordLen != 9+keyLen+valLen {
		return nil, ErrCorruptTail
	}
	// Validate type
	if !ValidRecordType(typ) {
		return nil, ErrCorruptTail
	}
	// Validate body length matches header
	expectedBody := int(9 + keyLen + valLen)
	if len(body) != expectedBody {
		return nil, ErrCorruptTail
	}
	if uint32(len(body[9:])) < keyLen+valLen {
		return nil, ErrCorruptTail
	}
	// CRC over body (Type||KeyLen||ValLen||Key||Value)
	gotCRC := crc32.ChecksumIEEE(body)
	if gotCRC != wantCRC {
		return nil, ErrCorruptTail
	}
	// Extract key/value
	key := make([]byte, keyLen)
	copy(key, body[9:9+keyLen])
	var val []byte
	if typ == RecordPut && valLen > 0 {
		val = make([]byte, valLen)
		copy(val, body[9+keyLen:9+keyLen+valLen])
	}
	// Advance logical offset only on success.
	r.off = cur + int64(PrefixSize) + int64(recordLen)
	rec := &Record{Type: typ, Key: key, Value: val}
	return rec, nil
}
