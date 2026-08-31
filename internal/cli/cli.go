package cli

import (
	"errors"
	"fmt"
	"io"

	"picodb/internal/store"
)

const (
	ExitOK       = 0
	ExitError    = 1
	ExitUsage    = 2
	ExitNotFound = 3
)

// Stats is consumer-owned (no interface{} leak, Plan Â§24).
type Stats struct {
	Keys int
}

// Run is the CLI entry point. It never calls os.Exit; the caller decides.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: microdb <put|get|del|dump> <dbfile> [key] [value]")
		return ExitUsage
	}
	cmd := args[0]
	switch cmd {
	case "put":
		if len(args) != 4 {
			fmt.Fprintln(stderr, "usage: microdb put <dbfile> <key> <value>")
			return ExitUsage
		}
		db, key, val := args[1], args[2], args[3]
		s, err := store.Open(db)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return ExitError
		}
		defer s.Close()
		if err := s.Put([]byte(key), []byte(val)); err != nil {
			fmt.Fprintln(stderr, err)
			return ExitError
		}
		return ExitOK
	case "get":
		if len(args) != 3 {
			fmt.Fprintln(stderr, "usage: microdb get <dbfile> <key>")
			return ExitUsage
		}
		db, key := args[1], args[2]
		s, err := store.Open(db)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return ExitError
		}
		defer s.Close()
		val, err := s.Get([]byte(key))
		if err != nil {
			if errors.Is(err, store.ErrKeyNotFound) {
				fmt.Fprintln(stderr, "key not found")
				return ExitNotFound
			}
			fmt.Fprintln(stderr, err)
			return ExitError
		}
		fmt.Fprintln(stdout, string(val))
		return ExitOK
	case "del", "delete":
		if len(args) != 3 {
			fmt.Fprintln(stderr, "usage: microdb del <dbfile> <key>")
			return ExitUsage
		}
		db, key := args[1], args[2]
		s, err := store.Open(db)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return ExitError
		}
		defer s.Close()
		if err := s.Delete([]byte(key)); err != nil {
			if errors.Is(err, store.ErrKeyNotFound) {
				fmt.Fprintln(stderr, "key not found")
				return ExitNotFound
			}
			fmt.Fprintln(stderr, err)
			return ExitError
		}
		return ExitOK
	case "dump":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: microdb dump <dbfile>")
			return ExitUsage
		}
		// Optional command â€” cut first per ladder. Minimal stub.
		fmt.Fprintln(stderr, "dump: not yet implemented")
		return ExitError
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", cmd)
		return ExitUsage
	}
}

