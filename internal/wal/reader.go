package wal

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type Reader struct {
	file      *os.File
	offset    int64
	ownsClose bool
}

// OpenReader opens an existing WAL file for sequential replay.
func OpenReader(path string) (*Reader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("wal: open reader file: %w", err)
	}

	return &Reader{
		file:      f,
		offset:    0,
		ownsClose: true,
	}, nil
}

// NewReader creates a Reader from an existing *os.File descriptor.
func NewReader(f *os.File) *Reader {
	_, _ = f.Seek(0, io.SeekStart)
	return &Reader{
		file:      f,
		offset:    0,
		ownsClose: false,
	}
}

// Next reads and validates the next valid record in the WAL.
// Adheres strictly to the validation sequence in Section 11 of the architecture plan.
func (r *Reader) Next() (*Record, error) {
	var prefixBuf [PrefixSize]byte
	n, err := io.ReadFull(r.file, prefixBuf[:])
	if err != nil {
		if err == io.EOF || (err == io.ErrUnexpectedEOF && n == 0) {
			return nil, io.EOF
		}
		// Short prefix or read error at tail
		return nil, ErrCorruptTail
	}

	recordLen, expectedCRC := DecodePrefix(prefixBuf)

	// Length bounds check before allocating body
	if recordLen < BodyHeaderSize || recordLen > MaxRecordSize {
		return nil, ErrCorruptTail
	}

	body := make([]byte, recordLen)
	if _, err := io.ReadFull(r.file, body); err != nil {
		// Short body read at tail
		return nil, ErrCorruptTail
	}

	recType := RecordType(body[0])
	keyLen := binary.BigEndian.Uint32(body[1:5])
	valLen := binary.BigEndian.Uint32(body[5:9])

	// Structural length consistency check: recordLen == 9 + KeyLen + ValLen
	if uint64(recordLen) != uint64(BodyHeaderSize)+uint64(keyLen)+uint64(valLen) {
		return nil, ErrCorruptTail
	}

	// Validate record type
	if recType != RecordPut && recType != RecordDelete {
		return nil, ErrCorruptTail
	}

	// CRC32 checksum integrity check
	if Checksum(body) != expectedCRC {
		return nil, ErrCorruptTail
	}

	key := make([]byte, keyLen)
	copy(key, body[9:9+keyLen])

	val := make([]byte, valLen)
	copy(val, body[9+keyLen:])

	r.offset += int64(PrefixSize + recordLen)

	return &Record{
		Type:  recType,
		Key:   key,
		Value: val,
	}, nil
}

// Offset returns the offset of the last successfully validated record boundary.
func (r *Reader) Offset() int64 {
	return r.offset
}

// Close closes the underlying reader file descriptor if owned.
func (r *Reader) Close() error {
	if r.file == nil {
		return nil
	}
	var err error
	if r.ownsClose {
		err = r.file.Close()
	}
	r.file = nil
	return err
}
