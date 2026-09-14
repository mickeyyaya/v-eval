package render

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/mickeyyaya/v-eval/core/report"
)

// markdownTemplate is parsed once, at package load. The template is compiled
// into the binary, so a broken template is a build-time defect rather than a
// failure a caller could hit at render time.
var markdownTemplate = template.Must(
	template.New("report.md.tmpl").Funcs(funcMap()).ParseFS(templates, "templates/report.md.tmpl"))

// markdownRenderer writes a report as Markdown, in a fixed section order.
type markdownRenderer struct{}

// Format is the name this renderer is registered under.
func (markdownRenderer) Format() string { return "md" }

// Render writes rep as Markdown.
func (markdownRenderer) Render(rep report.Report) ([]byte, error) {
	var out bytes.Buffer
	if err := markdownTemplate.Execute(&out, rep); err != nil {
		return nil, fmt.Errorf("render: markdown: %w", err)
	}
	return out.Bytes(), nil
}
