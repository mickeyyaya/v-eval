package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/mickeyyaya/v-eval/core/render"
	"github.com/mickeyyaya/v-eval/core/report"
)

// Exit codes are part of the command's contract: a script branches on them
// without reading a word of the output.
const (
	exitOK         = 0 // the report is sound
	exitViolations = 1 // the report breaks at least one rule
	exitError      = 2 // the command could not run: wrong usage, or an operational failure
)

// stdinPath is the input path that means standard input.
const stdinPath = "-"

// subcommand is one verb the command line offers. The handler takes its
// streams rather than reaching for the process's own, so a test drives it
// exactly as a shell does.
type subcommand struct {
	Name    string
	Summary string
	Run     func(args []string, stdin io.Reader, stdout, stderr io.Writer) int
}

// commands is the whole command line in one table: the dispatch, the usage
// text, and the order both are read in. A new verb is one entry here.
var commands = []subcommand{
	{"validate", "check a report against every rule", runValidate},
	{"aggregate", "recompute counts, status, and digests", runAggregate},
	{"render", "write a report as " + strings.Join(render.Formats(), " or "), runRender},
	{"export", "write a report as " + strings.Join(exportFormats(), " or "), runExport},
}

// run dispatches one invocation. No subcommand and an explicit request for
// help are the same case: the caller does not yet know what to ask for, so
// the usage goes to standard error and the exit code says nothing was done.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprint(stderr, usage())
		return exitError
	}
	for _, cmd := range commands {
		if cmd.Name == args[0] {
			return cmd.Run(args[1:], stdin, stdout, stderr)
		}
	}
	fmt.Fprintf(stderr, "error: unknown command %q\n\n%s", args[0], usage())
	return exitError
}

// usage describes the command line, reading the command table and the format
// registry rather than repeating either.
func usage() string {
	var b strings.Builder
	b.WriteString("usage: veval <command> [flags] <report.json|->\n\ncommands:\n")
	for _, cmd := range commands {
		fmt.Fprintf(&b, "  %-9s  %s\n", cmd.Name, cmd.Summary)
	}
	fmt.Fprintf(&b, "\nflags:\n  -o <path>       write to this file instead of standard output\n"+
		"  -format <name>  what render writes: %s\n", strings.Join(render.Formats(), ", "))
	fmt.Fprintf(&b, "\nan input of %q reads the report from standard input.\n", stdinPath)
	b.WriteString("exit codes: 0 the report is sound, 1 it breaks a rule, 2 the command could not run.\n")
	return b.String()
}

// loadReport reads one report and puts it through validate. It returns the
// report and true when the caller may go on; otherwise it has already said
// what is wrong and returns the exit code to end on.
//
// Violations go to standard output: they are the command's answer about the
// report, not a failure of the command.
func loadReport(path string, stdin io.Reader, stdout, stderr io.Writer,
	validate func([]byte) (report.Report, []report.Violation, error)) (report.Report, int, bool) {
	raw, err := readInput(path, stdin)
	if err != nil {
		return report.Report{}, fail(stderr, err), false
	}
	rep, violations, err := validate(raw)
	if err != nil {
		return report.Report{}, fail(stderr, err), false
	}
	if len(violations) > 0 {
		for _, violation := range violations {
			fmt.Fprintf(stdout, "%s: %s: %s\n", violation.Path, violation.Rule, violation.Message)
		}
		return report.Report{}, exitViolations, false
	}
	return rep, exitOK, true
}

// fail reports an operational failure: the command could not run at all.
func fail(stderr io.Writer, err error) int {
	fmt.Fprintf(stderr, "error: %v\n", err)
	return exitError
}

// unknownFormat reports a format this build does not write, naming the ones
// it does. The list comes from the registry, so it cannot go stale.
func unknownFormat(stderr io.Writer, format string, accepted []string) int {
	fmt.Fprintf(stderr, "error: unknown format %q (accepted: %s)\n", format, strings.Join(accepted, ", "))
	return exitError
}

// writeEncoded finishes a command that has already encoded a report: it
// reports err if the encoding failed, otherwise writes data to output (or
// standard output) and reports any failure to do that. Aggregate, render,
// and export all end this way, once the encoding is specific to each.
func writeEncoded(stdout, stderr io.Writer, output string, data []byte, err error) int {
	if err != nil {
		return fail(stderr, err)
	}
	if err := writeOutput(output, stdout, data); err != nil {
		return fail(stderr, err)
	}
	return exitOK
}
