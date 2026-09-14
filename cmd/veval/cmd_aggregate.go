package main

import (
	"io"

	"github.com/mickeyyaya/v-eval/core/report"
)

// runAggregate recomputes a report's counts, status, and digests from the
// rest of it and writes the result.
//
// It validates with the structural rules only: the derived rules compare a
// report against its own recomputation, which is what this command is about
// to do, so holding an input to them would reject every report that needs
// aggregating.
func runAggregate(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := newFlags("aggregate", stderr)
	output := flags.String("o", "", "write to this file instead of standard output")
	operands, code := parseArgs(flags, args, 1)
	if code != exitOK {
		return code
	}
	rep, code, ok := loadReport(operands[0], stdin, stdout, stderr, report.ValidateForAggregate)
	if !ok {
		return code
	}
	data, err := report.Encode(report.Aggregate(rep))
	return writeEncoded(stdout, stderr, *output, data, err)
}
