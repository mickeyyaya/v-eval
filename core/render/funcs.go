package render

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/mickeyyaya/v-eval/core/report"
)

// funcMap is the whole vocabulary a Markdown report template may call. Each
// function turns one report value into the words a reader sees; the template
// decides where those words go and computes nothing of its own.
//
// joinIDs is bound to the escaping form here and to the plain one in the HTML
// map: text/template writes whatever it is handed, so Markdown escaping is
// this map's job, while html/template escapes by the context a value lands in
// and needs none of it.
func funcMap() template.FuncMap {
	return template.FuncMap{
		"badge":         badge,
		"joinIDs":       joinIDsCell,
		"supplied":      supplied,
		"locator":       locator,
		"cell":          cell,
		"requirement":   requirement,
		"isJudgment":    isJudgment,
		"orNotRecorded": orNotRecorded,
		"provenance":    provenance,
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

// joinIDsCell is joinIDs for Markdown. An id is a string a report carries like
// any other -- a path, a criterion id, a constraint -- so a newline or a pipe
// inside one would end the line or the row it sits in, and the rest of it
// would read as the renderer's own words.
func joinIDsCell(ids []string) string { return cell(joinIDs(ids)) }

// isSupplied reports whether the candidate handed a piece of evidence over
// rather than the evaluator gathering it: the one origin a reader must be
// able to discount, wherever that evidence is shown.
func isSupplied(evidence report.Evidence) bool {
	return evidence.Origin == report.OriginCandidateSupplied
}

// supplied is the marker Markdown appends to supplied evidence, and nothing
// for the rest.
func supplied(evidence report.Evidence) string {
	if isSupplied(evidence) {
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

// requirement is what the contract asks of a criterion, by id. A result whose
// id the contract does not name says so: the row is still shown, because a
// result the contract has no criterion for is exactly what a reader must see.
// An optional criterion is marked, because its result never blocks the
// overall verdict and a reader weighing a failure must know that. A
// provisional criterion is marked too, because its verdict rests on a
// requirement nobody has confirmed yet. A criterion that is both carries both
// markers, in that order: whether it must hold, then whether anyone has
// confirmed what it asks.
func requirement(contract report.Contract, id string) string {
	criterion, ok := contract.CriterionByID(id)
	if !ok {
		return "(not in contract)"
	}
	out := criterion.Requirement
	if !criterion.Required {
		out += " (optional)"
	}
	if criterion.Provisional {
		out += " (provisional)"
	}
	return out
}

// isJudgment reports whether evidence is a judgment, which is shown with the
// rubric it applied and the model that applied it.
func isJudgment(evidence report.Evidence) bool { return evidence.Kind == report.KindJudgment }

// orNotRecorded names an absent value rather than leaving a blank a reader
// would read as an omission by the renderer. It says "(not recorded)" rather
// than "unknown", because a value nobody wrote down and a value recorded as
// the literal "unknown" are different states, and the second one is a
// statement the evaluation actually made.
func orNotRecorded(value string) string {
	if value == "" {
		return "(not recorded)"
	}
	return value
}

// cellEscapes rewrites the two characters that end a Markdown table cell or
// row. A backslash-escaped pipe reads as a pipe outside a table too, so the
// same escaping is safe everywhere evidence text appears.
var cellEscapes = strings.NewReplacer("|", `\|`, "\r\n", "<br>", "\n", "<br>")

// cell makes report text safe inside a Markdown table cell.
func cell(text string) string { return cellEscapes.Replace(text) }

// provenance names what produced a piece of evidence and when, e.g. "via
// pytest 7.4 at 2026-09-14T00:00:00Z". A tool nobody recorded says so
// explicitly rather than leaving a dangling "via" with nothing after it, and a
// tool with no recorded version is named without a trailing space. The
// timestamp is appended whenever one was recorded, whatever the tool is known
// to be: when a piece of evidence was captured is a fact of its own, and a
// reader comparing a citation against a later state of the artifact needs it.
func provenance(p report.EvidenceProvenance) string {
	out := "via (not recorded)"
	if p.Tool != "" {
		out = "via " + p.Tool
		if p.Version != "" {
			out += " " + p.Version
		}
	}
	if p.Timestamp != "" {
		out += " at " + p.Timestamp
	}
	return out
}
