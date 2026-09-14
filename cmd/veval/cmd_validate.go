package main

import (
	"fmt"
	"io"

	"github.com/mickeyyaya/v-eval/core/report"
)

// runValidate applies every rule to one report and says which ones it breaks.
// It is the command the other three are gated by: nothing is rendered,
// exported, or aggregated from a report that does not hold together.
func runValidate(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := newFlags("validate", "<report.json|->", stderr)
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
