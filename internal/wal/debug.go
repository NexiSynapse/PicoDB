package wal

import (
	"os"
)

const crashEnvVar = "MICRODB_CRASH_AFTER_PREFIX"

func shouldCrashAfterPrefix() bool {
	return os.Getenv(crashEnvVar) == "1"
}

func crashProcess() {
	os.Exit(137)
}
