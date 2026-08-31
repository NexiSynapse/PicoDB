package wal

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
)

// RecordType distinguishes Put vs Delete tombstone.
type RecordType uint8

const (
	RecordPut    RecordType = 1
	RecordDelete RecordType = 2
)

const (
	// PrefixSize is the fixed 8-byte header: 4B RecordLen + 4B CRC32.
	PrefixSize = 8
	// BodyHeaderSize is Type(1) + KeyLen(4) + ValLen(4).
	BodyHeaderSize = 9
	// MaxRecordSize caps the body length to prevent giant allocations on corrupt length.
	MaxRecordSize = 16 << 20 // 16 MiB
)

// Record is the logical WAL entry.
type Record struct {
	Type  RecordType
	Key   []byte
	Value []byte
}

// ErrCorruptTail signals the first corrupt or truncated record at the log tail.
// Recovery must stop at this point and truncate everything after it.
var ErrCorruptTail = errors.New("wal: corrupt or truncated tail record")

// Checksum computes CRC32-IEEE over the record body (Type||KeyLen||ValLen||Key||Value).
func Checksum(body []byte) uint32 {
	return crc32.ChecksumIEEE(body)
}

// DecodePrefix splits the 8-byte prefix into record length and CRC.
func DecodePrefix(p [PrefixSize]byte) (recordLen uint32, crc uint32) {
	recordLen = binary.BigEndian.Uint32(p[0:4])
	crc = binary.BigEndian.Uint32(p[4:8])
	return
}

// Encode serializes a Record into the wire format:
//
//	[0:4]  RecordLen BE  ( = 9 + KeyLen + ValLen )
//	[4:8]  CRC32 BE     (over Type||KeyLen||ValLen||Key||Value)
//	[8]    Type        (1 byte)
//	[9:13] KeyLen BE
//	[13:17] ValLen BE
//	[17:17+KeyLen] Key
//	[17+KeyLen:] Value
//
// Encode does not validate sizes against MaxRecordSize; callers should validate
// if the record originates from untrusted input. For Put/Delete the caller
// should ensure record size is bounded.
func Encode(rec Record) []byte {
	keyLen := len(rec.Key)
	valLen := len(rec.Value)
	// Delete tombstones carry no value on disk; enforce zero ValLen for canonical form.
	if rec.Type == RecordDelete {
		valLen = 0
	}
	recordLen := 9 + keyLen + valLen
	// Total buffer = prefix (8) + body (recordLen)
	buf := make([]byte, PrefixSize+recordLen)
	body := buf[PrefixSize:]

	body[0] = byte(rec.Type)
	binary.BigEndian.PutUint32(body[1:5], uint32(keyLen))
	binary.BigEndian.PutUint32(body[5:9], uint32(valLen))
	copy(body[9:9+keyLen], rec.Key)
	if rec.Type == RecordPut && valLen > 0 {
		copy(body[9+keyLen:], rec.Value)
	}
	crc := Checksum(body)
	binary.BigEndian.PutUint32(buf[0:4], uint32(recordLen))
	binary.BigEndian.PutUint32(buf[4:8], crc)
	return buf
}

// ValidRecordType reports whether t is a known RecordType.
func ValidRecordType(t RecordType) bool {
	return t == RecordPut || t == RecordDelete
}

// ValidateSizes checks length consistency and bounds before allocation.
// Returns nil if RecordLen is consistent with KeyLen/ValLen and within MaxRecordSize.
func ValidateSizes(recordLen uint32, keyLen uint32, valLen uint32) error {
	if recordLen > MaxRecordSize {
		return ErrCorruptTail
	}
	if recordLen != 9+keyLen+valLen {
		return ErrCorruptTail
	}
	return nil
}
