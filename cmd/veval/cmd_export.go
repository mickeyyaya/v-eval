package main

import (
	"io"
	"maps"
	"slices"

	"github.com/mickeyyaya/v-eval/core/export"
	"github.com/mickeyyaya/v-eval/core/report"
)

// exporters is every interchange format this build writes, keyed by the name
// the command takes. It is a table rather than a switch so that the accepted
// list an error message names is the set the command actually offers.
var exporters = map[string]func(report.Report) ([]byte, error){
	"sarif": toSARIF,
}

// exportFormats names every format export writes, sorted, so that a caller
// listing the choices reads the table rather than repeating it.
func exportFormats() []string {
	return slices.Sorted(maps.Keys(exporters))
}

// toSARIF maps a report onto a SARIF log and writes it in canonical form.
func toSARIF(rep report.Report) ([]byte, error) {
	log, err := export.ToSARIF(rep)
	if err != nil {
		return nil, err
	}
	return export.Marshal(log)
}

// runExport writes a report in an interchange format another tool reads. As
// with render, the report is held to every rule first: an export carries the
// report's own claims onward, so it may only carry sound ones.
func runExport(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := newFlags("export", stderr)
	output := flags.String("o", "", "write to this file instead of standard output")
	operands, code, ok := parseArgs(flags, args, 2, stdout)
	if !ok {
		return code
	}
	encode, ok := exporters[operands[0]]
	if !ok {
		return unknownFormat(stderr, operands[0], exportFormats())
	}
	rep, code, ok := loadReport(operands[1], stdin, stdout, stderr, report.Validate)
	if !ok {
		return code
	}
	data, err := encode(rep)
	return writeEncoded(stdout, stderr, *output, data, err)
}
