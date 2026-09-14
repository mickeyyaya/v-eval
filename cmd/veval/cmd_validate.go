package main

import (
	"fmt"
	"io"

	"github.com/mickeyyaya/v-eval/core/report"
)

// runValidate applies every rule to one report and says which ones it breaks.
// It is that check on its own: render and export run the same report.Validate
// before they write anything, and aggregate runs the structural half of it,
// report.ValidateForAggregate, because the rest is what it is about to
// recompute.
func runValidate(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := newFlags("validate", stderr)
	operands, code := parseArgs(flags, args, 1)
	if code != exitOK {
		return code
	}
	rep, code, ok := loadReport(operands[0], stdin, stdout, stderr, report.Validate)
	if !ok {
		return code
	}
	fmt.Fprintf(stdout, "valid: %s (schema %s, %d criteria)\n",
		operands[0], rep.Identity.SchemaVersion, len(rep.Criteria))
	return exitOK
}
