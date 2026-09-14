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
//
// Its usage line is the command table's own line for that command, so that
// "veval render -h" and "veval -h" answer with the same spelling instead of
// two that drift apart.
func newFlags(name string, stderr io.Writer) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintf(stderr, "usage: %s\n", usageLine(name))
		flags.PrintDefaults()
	}
	return flags
}

// parseArgs parses the flags in args and requires exactly want operands. It
// returns the operands, and the exit code to end on when it could not: the
// flag set has already written what went wrong.
//
// Only the flag tokens reach the flag set. The operands never do, so a path
// that begins with a dash is a path, and a value-taking flag left without a
// value is the last thing the set parses -- which is how it comes to say so.
func parseArgs(flags *flag.FlagSet, args []string, want int) ([]string, int) {
	flagArgs, operands := splitArgs(flags, args)
	if err := flags.Parse(flagArgs); err != nil {
		return nil, exitError
	}
	if len(operands) != want {
		fmt.Fprintf(flags.Output(), "error: %s takes %d argument(s), got %d\n", flags.Name(), want, len(operands))
		flags.Usage()
		return nil, exitError
	}
	return operands, exitOK
}

// terminator ends flag parsing: every token after it is an operand, however
// it is spelled.
const terminator = "--"

// splitArgs separates the flag tokens, each with its value, from the
// operands, so that "veval render report.json --format html" and "veval
// render --format html report.json" are the same command. Go's flag package
// stops at the first operand; a person typing a path first does not.
//
// Which flags take a following value is read from the command's own flag
// set, so the two can never disagree. Order within each group is kept, "-"
// is an operand -- the standard stream, not a flag -- and a value-taking
// flag with nothing after it stays where it is, with nothing after it, so
// that the flag set is the one to report the missing value.
func splitArgs(flags *flag.FlagSet, args []string) (flagArgs, operands []string) {
	flagArgs = make([]string, 0, len(args))
	operands = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == terminator {
			return flagArgs, append(operands, args[i+1:]...)
		}
		if !isFlag(args[i]) {
			operands = append(operands, args[i])
			continue
		}
		flagArgs = append(flagArgs, args[i])
		if takesValue(flags, args[i]) && i+1 < len(args) {
			i++
			flagArgs = append(flagArgs, args[i])
		}
	}
	return flagArgs, operands
}

// isFlag reports whether an argument is a flag rather than an operand.
func isFlag(arg string) bool {
	return len(arg) > 1 && strings.HasPrefix(arg, "-")
}

// takesValue reports whether a flag's value is the argument after it: the
// flag set defines it, it is not a boolean, and it is not the "-format=html"
// form that carries its own value. A flag the set does not define takes
// nothing, so an operand after it stays an operand and the flag set reports
// the undefined flag itself.
func takesValue(flags *flag.FlagSet, arg string) bool {
	name := strings.TrimLeft(arg, "-")
	if strings.Contains(name, "=") {
		return false
	}
	defined := flags.Lookup(name)
	if defined == nil {
		return false
	}
	asBool, ok := defined.Value.(interface{ IsBoolFlag() bool })
	return !ok || !asBool.IsBoolFlag()
}

// readInput reads the report at path, or standard input when path is "-".
func readInput(path string, stdin io.Reader) ([]byte, error) {
	if path != streamPath {
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

// writeOutput writes data to path, or to stdout when no path was named or
// the path is "-": the sentinel means the standard stream when a document
// goes out exactly as it does when a report comes in. The bytes go out
// exactly as they came in: what the core produced, line endings included, is
// what lands on disk on every operating system.
func writeOutput(path string, stdout io.Writer, data []byte) error {
	if path == "" || path == streamPath {
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
