package main

import (
	"io"
	"strings"

	"github.com/mickeyyaya/v-eval/core/render"
	"github.com/mickeyyaya/v-eval/core/report"
)

// defaultFormat is what render writes when the caller names no format.
const defaultFormat = "md"

// runRender writes a report as a document a person reads.
//
// The renderers themselves present whatever they are given; this command is
// where a report is held to the rules first, so that nobody is handed a
// document built from a report that contradicts itself.
func runRender(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := newFlags("render", stderr)
	format := flags.String("format", defaultFormat, "output format: "+strings.Join(render.Formats(), ", "))
	output := flags.String("o", "", "write to this file instead of standard output")
	operands, code := parseArgs(flags, args, 1)
	if code != exitOK {
		return code
	}
	renderer, ok := render.ByFormat(*format)
	if !ok {
		return unknownFormat(stderr, *format, render.Formats())
	}
	rep, code, ok := loadReport(operands[0], stdin, stdout, stderr, report.Validate)
	if !ok {
		return code
	}
	data, err := renderer.Render(rep)
	return writeEncoded(stdout, stderr, *output, data, err)
}
