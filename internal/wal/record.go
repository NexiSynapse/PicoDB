package wal

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
)

type RecordType uint8

const (
	RecordPut    RecordType = 1
	RecordDelete RecordType = 2
)

const (
	PrefixSize     = 8
	BodyHeaderSize = 9
	MaxRecordSize  = 16 << 20 // 16 MiB
)

type Record struct {
	Type  RecordType
	Key   []byte
	Value []byte
}

var ErrCorruptTail = errors.New("wal: corrupt or truncated tail record")

// Checksum calculates IEEE CRC32 over the given body.
func Checksum(body []byte) uint32 {
	return crc32.ChecksumIEEE(body)
}

// Encode serializes a Record into wire format:
// [RecordLen: 4B BE][CRC32: 4B BE][Type: 1B][KeyLen: 4B BE][ValLen: 4B BE][Key][Value]
func Encode(rec Record) []byte {
	keyLen := len(rec.Key)
	valLen := len(rec.Value)
	recordLen := uint32(BodyHeaderSize + keyLen + valLen)

	body := make([]byte, recordLen)
	body[0] = byte(rec.Type)
	binary.BigEndian.PutUint32(body[1:5], uint32(keyLen))
	binary.BigEndian.PutUint32(body[5:9], uint32(valLen))
	copy(body[9:9+keyLen], rec.Key)
	copy(body[9+keyLen:], rec.Value)

	crc := Checksum(body)

	buf := make([]byte, PrefixSize+recordLen)
	binary.BigEndian.PutUint32(buf[0:4], recordLen)
	binary.BigEndian.PutUint32(buf[4:8], crc)
	copy(buf[8:], body)

	return buf
}

// DecodePrefix unpacks the 8-byte prefix into RecordLen and CRC32.
func DecodePrefix(p [PrefixSize]byte) (recordLen uint32, crc uint32) {
	recordLen = binary.BigEndian.Uint32(p[0:4])
	crc = binary.BigEndian.Uint32(p[4:8])
	return recordLen, crc
}
