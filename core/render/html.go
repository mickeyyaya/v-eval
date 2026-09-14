package render

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/mickeyyaya/v-eval/core/report"
)

// htmlTemplate is parsed once, at package load, into its own template set:
// html/template escapes by the context each value lands in, so it cannot
// share a set, or a function map, with the Markdown renderer's text/template.
var htmlTemplate = template.Must(
	template.New("report.html.tmpl").Funcs(htmlFuncMap()).ParseFS(templates, "templates/report.html.tmpl"))

// htmlFuncMap is the whole vocabulary the HTML template may call. It repeats
// the shared functions rather than extending the Markdown set so that neither
// renderer's vocabulary can drift into the other's: cell, which escapes
// Markdown table syntax, has no meaning here, and the class functions have
// none there. joinIDs is the plain form for the same reason -- html/template
// escapes by the context a value lands in, so escaping it first would show a
// reader the escape rather than the id.
func htmlFuncMap() template.FuncMap {
	return template.FuncMap{
		"badge":         badge,
		"resultClass":   resultClass,
		"overallClass":  overallClass,
		"joinIDs":       joinIDs,
		"supplied":      supplied,
		"locator":       locator,
		"requirement":   requirement,
		"isJudgment":    isJudgment,
		"orNotRecorded": orNotRecorded,
		"provenance":    provenance,
	}
}

// resultClass is the class a criterion result is shown with, e.g. "badge
// pass". The stylesheet colours the verdict; the word beside it still says
// which verdict it is, so colour carries nothing on its own.
func resultClass(result report.Result) string {
	return "badge " + strings.ToLower(string(result))
}

// overallClass is the class the report-level verdict is shown with.
func overallClass(overall report.Overall) string {
	return "badge " + strings.ToLower(string(overall))
}

// htmlRenderer writes a report as one self-contained HTML document: no
// scripts, no external styles, fonts, or images, so it opens offline from a
// CI archive years later.
type htmlRenderer struct{}

// Format is the name this renderer is registered under.
func (htmlRenderer) Format() string { return "html" }

// Render writes rep as HTML.
func (htmlRenderer) Render(rep report.Report) ([]byte, error) {
	var out bytes.Buffer
	if err := htmlTemplate.Execute(&out, rep); err != nil {
		return nil, fmt.Errorf("render: html: %w", err)
	}
	return out.Bytes(), nil
}
