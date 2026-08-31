# Standard Library Packages

MicroDB is built entirely using the Go standard library, with absolutely zero external third-party dependencies.

| Package | Purpose |
|---|---|
| `os` | files, stdout/stderr, process exit |
| `io` | EOF and exact byte reads |
| `encoding/binary` | binary encoding for length-prefixed headers (Big Endian) |
| `hash/crc32` | checksums for data integrity |
| `sync` | Store locking (single lock authority via RWMutex) |
| `syscall` | OS-managed file locking (flock on Unix, LockFileEx on Windows) |
| `unsafe` | specific to Windows LockFileEx struct sizing |
| `errors` | sentinel errors (e.g. `ErrCorruptTail`, `ErrKeyNotFound`) |
| `fmt` | formatted string output and CLI error reporting |
| `bytes` | slice comparisons in tests |
| `strings` | basic string manipulation in tests |
| `path/filepath` | temp directory path construction in tests |
| `testing` | unit and integration test suite framework |
