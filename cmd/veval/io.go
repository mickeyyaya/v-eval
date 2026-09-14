package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// newFlags returns a flag set that writes everything it has to say to the
// stderr the caller passed in, and that decides nothing: a parse failure is
// returned, so one place -- the subcommand -- owns the exit code.
func newFlags(name, operands string, stderr io.Writer) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintf(stderr, "usage: veval %s [flags] %s\n", name, operands)
		flags.PrintDefaults()
	}
	return flags
}

// parseArgs parses args against flags and requires exactly want operands. It
// returns the operands, and the exit code to end on when it could not: the
// flag set has already written what went wrong.
func parseArgs(flags *flag.FlagSet, args []string, want int) ([]string, int) {
	if err := flags.Parse(reorderArgs(args)); err != nil {
		return nil, exitError
	}
	if flags.NArg() != want {
		fmt.Fprintf(flags.Output(), "error: %s takes %d argument(s), got %d\n", flags.Name(), want, flags.NArg())
		flags.Usage()
		return nil, exitError
	}
	return flags.Args(), exitOK
}

// valueFlags names the flags whose value is a separate argument, so that
// reorderArgs keeps a flag and its value together.
var valueFlags = map[string]bool{"o": true, "format": true}

// reorderArgs moves every flag, with its value, in front of the operands, so
// that "veval render report.json --format html" and "veval render --format
// html report.json" are the same command. Go's flag package stops at the
// first operand; a person typing a path first does not.
//
// Order within each group is kept, and "-" is an operand: it is the input
// that means standard input, not a flag.
func reorderArgs(args []string) []string {
	flags := make([]string, 0, len(args))
	operands := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if !isFlag(args[i]) {
			operands = append(operands, args[i])
			continue
		}
		flags = append(flags, args[i])
		if takesValue(args[i]) && i+1 < len(args) {
			i++
			flags = append(flags, args[i])
		}
	}
	return append(flags, operands...)
}

// isFlag reports whether an argument is a flag rather than an operand.
func isFlag(arg string) bool {
	return len(arg) > 1 && strings.HasPrefix(arg, "-")
}

// takesValue reports whether a flag's value is the argument after it. A
// "-format=html" carries its own value and takes nothing.
func takesValue(arg string) bool {
	name := strings.TrimLeft(arg, "-")
	if strings.Contains(name, "=") {
		return false
	}
	return valueFlags[name]
}

// readInput reads the report at path, or standard input when path is "-".
func readInput(path string, stdin io.Reader) ([]byte, error) {
	if path != stdinPath {
		raw, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return nil, fmt.Errorf("veval: read input: %w", err)
		}
		return raw, nil
	}
	if stdin == nil {
		return nil, errors.New("veval: read input: no standard input to read")
	}
	raw, err := io.ReadAll(stdin)
	if err != nil {
		return nil, fmt.Errorf("veval: read input: %w", err)
	}
	return raw, nil
}

// writeOutput writes data to path, or to stdout when path is empty. The bytes
// go out exactly as they came in: what the core produced, line endings
// included, is what lands on disk on every operating system.
func writeOutput(path string, stdout io.Writer, data []byte) error {
	if path == "" {
		if _, err := stdout.Write(data); err != nil {
			return fmt.Errorf("veval: write output: %w", err)
		}
		return nil
	}
	if err := os.WriteFile(filepath.Clean(path), data, 0o644); err != nil {
		return fmt.Errorf("veval: write output: %w", err)
	}
	return nil
}
