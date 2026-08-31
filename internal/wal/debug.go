package wal

import (
	"os"
)

// crashEnv is the deterministic fault-injection switch.
// It is intentionally demo/QA-only and isolated here so it never leaks into
// normal storage logic.
//
//	README wording: "Fault injection exists only for demo/QA reproducibility
//	and is not required for ordinary database operation."
const crashEnv = "MICRODB_CRASH_AFTER_PREFIX"

// crashEnabled reports whether the process should simulate a torn write.
func crashEnabled() bool {
	return os.Getenv(crashEnv) == "1"
}

// maybeCrashAfterPrefix implements the v2 crash-injection sequence:
//
//	make previous records durable -> fsync
//	start new record -> write only 8-byte RecordLen|CRC prefix -> os.Exit(137)
//
// It is called from Writer.Append after the record has been encoded but before
// the full body is written. If injection is disabled it returns false and the
// caller performs a normal append. If enabled it never returns.
func maybeCrashAfterPrefix(f *os.File, prefix []byte) bool {
	if !crashEnabled() {
		return false
	}
	// Ensure prior records are durable (H5).
	_ = f.Sync()
	// Write only the 8-byte prefix to create a torn tail.
	// Ignore errors: the point is to leave a partial record then die.
	_, _ = f.Write(prefix)
	_ = f.Sync()
	os.Exit(137)
	return true // unreachable
}
