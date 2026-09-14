// Package render turns a report into a document a person reads. A renderer
// only presents: it never recomputes a count, a verdict, or a digest, so what
// a reader sees is what the report states.
package render

import "github.com/mickeyyaya/v-eval/core/report"

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
	"md": markdownRenderer{},
}

// ByFormat returns the renderer registered for a format name, and whether
// there is one. An unknown format is a caller's mistake, not an error state.
func ByFormat(format string) (Renderer, bool) {
	renderer, ok := renderers[format]
	return renderer, ok
}
