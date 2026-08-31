package wal

import (
	"os"
)

const crashEnvVar = "PICODB_CRASH_AFTER_PREFIX"
const crashEnvVarLegacy = "MICRODB_CRASH_AFTER_PREFIX"

func shouldCrashAfterPrefix() bool {
	return os.Getenv(crashEnvVar) == "1" || os.Getenv(crashEnvVarLegacy) == "1"
}

func crashProcess() {
	os.Exit(137)
}
