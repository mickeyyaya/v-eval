package render

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/mickeyyaya/v-eval/core/report"
)

// funcMap is the whole vocabulary a report template may call. Each function
// turns one report value into the words a reader sees; the template decides
// where those words go and computes nothing of its own.
func funcMap() template.FuncMap {
	return template.FuncMap{
		"badge":    badge,
		"joinIDs":  joinIDs,
		"supplied": supplied,
		"locator":  locator,
		"cell":     cell,
	}
}

// badge is the word a result is announced by.
func badge(result report.Result) string { return string(result) }

// joinIDs lists ids on one line, and says "none" for an empty list: a reader
// must be able to tell an empty list from a list the renderer dropped.
func joinIDs(ids []string) string {
	if len(ids) == 0 {
		return "none"
	}
	return strings.Join(ids, ", ")
}

// supplied marks evidence the candidate handed over rather than evidence the
// evaluator gathered, wherever that evidence is shown.
func supplied(evidence report.Evidence) string {
	if evidence.Origin == report.OriginCandidateSupplied {
		return " (supplied)"
	}
	return ""
}

// locator names where a piece of evidence came from, in the shortest form
// that still lets a reader go back to it. A locator that fills no complete
// group says so rather than rendering half of one.
func locator(l report.Locator) string {
	switch l.Shape() {
	case report.ShapeFile:
		return fmt.Sprintf("%s:%d-%d", l.File, l.LineStart, l.LineEnd)
	case report.ShapeCommand:
		return fmt.Sprintf("%s (exit %d)", l.Command, *l.ExitStatus)
	case report.ShapePassage:
		return fmt.Sprintf("%s from %s (%s)", l.Passage, l.SourceRef, l.SourceDateOrVersion)
	case report.ShapeNote:
		return l.Note
	default:
		return "(no locator)"
	}
}

// cellEscapes rewrites the two characters that end a Markdown table cell or
// row. A backslash-escaped pipe reads as a pipe outside a table too, so the
// same escaping is safe everywhere evidence text appears.
var cellEscapes = strings.NewReplacer("|", `\|`, "\r\n", "<br>", "\n", "<br>")

// cell makes report text safe inside a Markdown table cell.
func cell(text string) string { return cellEscapes.Replace(text) }
