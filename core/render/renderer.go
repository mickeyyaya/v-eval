// Package render turns a report into a document a person reads. A renderer
// only presents: it never recomputes a count, a verdict, or a digest, so what
// a reader sees is what the report states.
package render

import (
	"embed"
	"maps"
	"slices"

	"github.com/mickeyyaya/v-eval/core/report"
)

// templates holds every report template, compiled into the binary so that a
// rendering never reads a file at run time and a broken template is a
// build-time defect rather than a failure a caller could hit.
//
//go:embed templates
var templates embed.FS

// Renderer turns a report into the bytes of one output format.
type Renderer interface {
	// Format is the short name the renderer is selected by, e.g. "md".
	Format() string
	// Render writes rep in this format.
	Render(rep report.Report) ([]byte, error)
}

// renderers is the registry ByFormat looks in, keyed by format name. It is a
// package-level var rather than a registration call so that the set of
// formats this build offers can be read in one place.
var renderers = map[string]Renderer{
	"md":   markdownRenderer{},
	"html": htmlRenderer{},
}

// ByFormat returns the renderer registered for a format name, and whether
// there is one. An unknown format is a caller's mistake, not an error state.
func ByFormat(format string) (Renderer, bool) {
	renderer, ok := renderers[format]
	return renderer, ok
}

// Formats names every format this build renders, sorted, so that a caller
// listing the choices -- a CLI usage line, a test that must cover them all --
// reads the registry rather than repeating it.
func Formats() []string {
	return slices.Sorted(maps.Keys(renderers))
}
