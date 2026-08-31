package cli

import (
	"errors"
	"fmt"
	"io"

	"microdb/internal/store"
)

const (
	ExitOK       = 0
	ExitError    = 1
	ExitUsage    = 2
	ExitNotFound = 3
)

type Stats struct {
	Keys int
}

// Run executes a CLI command with the given arguments, writing output to stdout/stderr.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		printUsage(stderr)
		return ExitUsage
	}

	cmd := args[0]
	dbPath := args[1]

	switch cmd {
	case "put":
		if len(args) != 4 {
			fmt.Fprintln(stderr, "usage: microdb put <dbfile> <key> <value>")
			return ExitUsage
		}
		key := []byte(args[2])
		value := []byte(args[3])

		s, err := store.Open(dbPath)
		if err != nil {
			fmt.Fprintf(stderr, "error opening database: %v\n", err)
			return ExitError
		}
		defer s.Close()

		if err := s.Put(key, value); err != nil {
			fmt.Fprintf(stderr, "error putting key: %v\n", err)
			return ExitError
		}
		return ExitOK

	case "get":
		if len(args) != 3 {
			fmt.Fprintln(stderr, "usage: microdb get <dbfile> <key>")
			return ExitUsage
		}
		key := []byte(args[2])

		s, err := store.Open(dbPath)
		if err != nil {
			fmt.Fprintf(stderr, "error opening database: %v\n", err)
			return ExitError
		}
		defer s.Close()

		val, err := s.Get(key)
		if err != nil {
			if errors.Is(err, store.ErrKeyNotFound) {
				fmt.Fprintln(stderr, "key not found")
				return ExitNotFound
			}
			fmt.Fprintf(stderr, "error getting key: %v\n", err)
			return ExitError
		}

		fmt.Fprintln(stdout, string(val))
		return ExitOK

	case "del":
		if len(args) != 3 {
			fmt.Fprintln(stderr, "usage: microdb del <dbfile> <key>")
			return ExitUsage
		}
		key := []byte(args[2])

		s, err := store.Open(dbPath)
		if err != nil {
			fmt.Fprintf(stderr, "error opening database: %v\n", err)
			return ExitError
		}
		defer s.Close()

		if err := s.Delete(key); err != nil {
			if errors.Is(err, store.ErrKeyNotFound) {
				fmt.Fprintln(stderr, "key not found")
				return ExitNotFound
			}
			fmt.Fprintf(stderr, "error deleting key: %v\n", err)
			return ExitError
		}
		return ExitOK

	default:
		printUsage(stderr)
		return ExitUsage
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "MicroDB - Embedded Crash-Safe Key-Value Store")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  microdb put <dbfile> <key> <value>")
	fmt.Fprintln(w, "  microdb get <dbfile> <key>")
	fmt.Fprintln(w, "  microdb del <dbfile> <key>")
}
