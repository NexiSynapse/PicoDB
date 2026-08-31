package wal

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	rec := Record{
		Type:  RecordPut,
		Key:   []byte("test_key"),
		Value: []byte("test_value_payload"),
	}

	encoded := Encode(rec)

	dir := t.TempDir()
	path := filepath.Join(dir, "test.wal")

	if err := os.WriteFile(path, encoded, 0666); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	reader, err := OpenReader(path)
	if err != nil {
		t.Fatalf("OpenReader failed: %v", err)
	}
	defer reader.Close()

	readRec, err := reader.Next()
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}

	if readRec.Type != rec.Type {
		t.Errorf("expected type %v, got %v", rec.Type, readRec.Type)
	}
	if !bytes.Equal(readRec.Key, rec.Key) {
		t.Errorf("expected key %q, got %q", rec.Key, readRec.Key)
	}
	if !bytes.Equal(readRec.Value, rec.Value) {
		t.Errorf("expected value %q, got %q", rec.Value, readRec.Value)
	}

	// Should reach clean EOF
	_, err = reader.Next()
	if !errors.Is(err, io.EOF) {
		t.Errorf("expected EOF, got %v", err)
	}
}

func TestChecksumDetectsCorruption(t *testing.T) {
	rec := Record{
		Type:  RecordPut,
		Key:   []byte("safe_key"),
		Value: []byte("safe_value"),
	}

	encoded := Encode(rec)
	// Corrupt a byte in the payload body (after 8-byte prefix)
	encoded[len(encoded)-1] ^= 0xFF

	dir := t.TempDir()
	path := filepath.Join(dir, "corrupt_crc.wal")

	if err := os.WriteFile(path, encoded, 0666); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	reader, err := OpenReader(path)
	if err != nil {
		t.Fatalf("OpenReader failed: %v", err)
	}
	defer reader.Close()

	_, err = reader.Next()
	if !errors.Is(err, ErrCorruptTail) {
		t.Fatalf("expected ErrCorruptTail on CRC mismatch, got %v", err)
	}
}

func TestRecordLengthConsistency(t *testing.T) {
	rec := Record{
		Type:  RecordPut,
		Key:   []byte("k"),
		Value: []byte("v"),
	}

	encoded := Encode(rec)
	// Modify recordLen in prefix to not match 9 + keyLen + valLen
	binary.BigEndian.PutUint32(encoded[0:4], 50)

	dir := t.TempDir()
	path := filepath.Join(dir, "len_mismatch.wal")

	if err := os.WriteFile(path, encoded, 0666); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	reader, err := OpenReader(path)
	if err != nil {
		t.Fatalf("OpenReader failed: %v", err)
	}
	defer reader.Close()

	_, err = reader.Next()
	if !errors.Is(err, ErrCorruptTail) {
		t.Fatalf("expected ErrCorruptTail on length mismatch, got %v", err)
	}
}

func TestRecordTooLargeRejected(t *testing.T) {
	var prefixBuf [8]byte
	// Set RecordLen > MaxRecordSize (e.g. 32 MiB)
	binary.BigEndian.PutUint32(prefixBuf[0:4], MaxRecordSize+1024)
	binary.BigEndian.PutUint32(prefixBuf[4:8], 0x12345678)

	dir := t.TempDir()
	path := filepath.Join(dir, "oversized.wal")

	if err := os.WriteFile(path, prefixBuf[:], 0666); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	reader, err := OpenReader(path)
	if err != nil {
		t.Fatalf("OpenReader failed: %v", err)
	}
	defer reader.Close()

	_, err = reader.Next()
	if !errors.Is(err, ErrCorruptTail) {
		t.Fatalf("expected ErrCorruptTail on oversized length, got %v", err)
	}
}

func TestReaderCleanEOF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.wal")

	if err := os.WriteFile(path, []byte{}, 0666); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	reader, err := OpenReader(path)
	if err != nil {
		t.Fatalf("OpenReader failed: %v", err)
	}
	defer reader.Close()

	_, err = reader.Next()
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF, got %v", err)
	}
}

func TestReaderTornPrefix(t *testing.T) {
	// Write only 4 bytes (incomplete 8-byte prefix)
	torn := []byte{0x00, 0x00, 0x00, 0x20}

	dir := t.TempDir()
	path := filepath.Join(dir, "torn_prefix.wal")

	if err := os.WriteFile(path, torn, 0666); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	reader, err := OpenReader(path)
	if err != nil {
		t.Fatalf("OpenReader failed: %v", err)
	}
	defer reader.Close()

	_, err = reader.Next()
	if !errors.Is(err, ErrCorruptTail) {
		t.Fatalf("expected ErrCorruptTail on torn prefix, got %v", err)
	}
}

func TestReaderTornBody(t *testing.T) {
	rec := Record{
		Type:  RecordPut,
		Key:   []byte("alpha"),
		Value: []byte("beta"),
	}

	encoded := Encode(rec)
	// Truncate encoded bytes so body is incomplete
	torn := encoded[:len(encoded)-3]

	dir := t.TempDir()
	path := filepath.Join(dir, "torn_body.wal")

	if err := os.WriteFile(path, torn, 0666); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	reader, err := OpenReader(path)
	if err != nil {
		t.Fatalf("OpenReader failed: %v", err)
	}
	defer reader.Close()

	_, err = reader.Next()
	if !errors.Is(err, ErrCorruptTail) {
		t.Fatalf("expected ErrCorruptTail on torn body, got %v", err)
	}
}

func TestReaderUnknownRecordType(t *testing.T) {
	rec := Record{
		Type:  RecordType(99), // Invalid record type
		Key:   []byte("key"),
		Value: []byte("val"),
	}

	encoded := Encode(rec)

	dir := t.TempDir()
	path := filepath.Join(dir, "unknown_type.wal")

	if err := os.WriteFile(path, encoded, 0666); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	reader, err := OpenReader(path)
	if err != nil {
		t.Fatalf("OpenReader failed: %v", err)
	}
	defer reader.Close()

	_, err = reader.Next()
	if !errors.Is(err, ErrCorruptTail) {
		t.Fatalf("expected ErrCorruptTail on unknown type, got %v", err)
	}
}
