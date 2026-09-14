package render_test

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/mickeyyaya/v-eval/core/render"
	"github.com/mickeyyaya/v-eval/core/report"
)

// goldenCase is one fixture and what its rendering must say.
type goldenCase struct {
	name     string   // fixture basename, under core/report/testdata and testdata
	sections []string // every section it carries, named as the section table names them
	contains []string // substrings that prove a rendering decision was made
}

// goldenCases covers both worked fixtures: one without dimensions or
// learning, one with both, so the conditional sections are exercised in each
// direction.
func goldenCases() []goldenCase {
	return []goldenCase{
		{
			name:     "worked-example",
			sections: sectionNamesExcept("dimensions", "learning"),
			contains: []string{
				"# v-eval report: code_change ",
				"Overall: FAIL",
				"required applicable criterion failed",
				"Blocked by: C1, C2, C3",
				"(supplied)",
				"examples/code-review/input.md:27-27",
				"Coverage: 4/5",
				"- Intended user: developer reviewing an AI-generated change",
				"- Brief: examples/code-review/input.md#intent",
				"- Host: fixture on any/any\n",
				"- Bundle digest: (none)",
				"| ID | Requirement | Result | Method |",
				"| C1 | Whitespace and case variants collapse into one address | FAIL |",
				"via assistant fixture",
			},
		},
		{
			name:     "extended-example",
			sections: sectionNamesExcept(),
			contains: []string{
				"# v-eval report: service_change ",
				"Overall: INCOMPLETE",
				"unresolved contract conflict",
				"Blocked by: none",
				"(supplied)",
				"go test ./... (exit 0)",
				"docker build --no-cache . (exit 127)",
				"isolation worktree",
				"from https://api.example.test/docs/rate-limits (2026-08-30)",
				"not measured",
				"- Host: fixture on any/any, model unknown",
				"- Bundle digest: sha256:1f0a5c7f2b3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7",
				"| E1 | The regression suite passes against this exact revision | PASS |",
				"via go 1.23.1",
				"via mtime-drift 0.3",
				", rubric integrity-rubric-0.2, model unknown",
			},
		},
	}
}

func TestMarkdownGoldens(t *testing.T) {
	for _, testCase := range goldenCases() {
		t.Run(testCase.name, func(t *testing.T) {
			out := renderAs(t, "md", loadFixture(t, testCase.name))
			// The structural checks run first: they name what is wrong,
			// where a golden mismatch only says that something is.
			want, unwanted := sectionMarkers(t, "md", testCase.sections)
			assertOrder(t, out, want)
			for _, heading := range unwanted {
				if bytes.Contains(out, []byte(heading)) {
					t.Errorf("section %q must not appear: the fixture has no such content", heading)
				}
			}
			for _, want := range testCase.contains {
				if !bytes.Contains(out, []byte(want)) {
					t.Errorf("rendering does not contain %q", want)
				}
			}
			assertGolden(t, testCase.name+".md", out)
		})
	}
}

// evidenceBadge is the prefix every rendered evidence line must carry, so a
// reader never meets a citation without knowing what it is and how isolated
// it ran.
var evidenceBadge = regexp.MustCompile(`\[kind [a-z_]+, isolation [a-z_]+\]`)

func TestEveryEvidenceLineShowsKindAndIsolation(t *testing.T) {
	assertEveryEvidenceRendered(t, "md", evidenceBadge)
}

func TestRenderingIsCleanMarkdown(t *testing.T) {
	for _, testCase := range goldenCases() {
		t.Run(testCase.name, func(t *testing.T) {
			out := renderAs(t, "md", loadFixture(t, testCase.name))
			if !bytes.HasSuffix(out, []byte("\n")) || bytes.HasSuffix(out, []byte("\n\n")) {
				t.Error("a rendering must end with exactly one newline")
			}
			for i, line := range bytes.Split(out, []byte("\n")) {
				if len(line) != len(bytes.TrimRight(line, " \t")) {
					t.Errorf("line %d ends in whitespace: %q", i+1, line)
				}
			}
		})
	}
}

func TestRenderingAnEmptyReportSaysSoRatherThanFailing(t *testing.T) {
	t.Parallel()
	renderer, ok := render.ByFormat("md")
	if !ok {
		t.Fatal("md renderer missing")
	}
	out, err := renderer.Render(report.Report{})
	if err != nil {
		t.Fatalf("an empty report must still render: %v", err)
	}
	for _, want := range []string{"## Status", "## Observations\n\nNone recorded.", "Blocked by: none"} {
		if !bytes.Contains(out, []byte(want)) {
			t.Errorf("rendering of an empty report does not contain %q", want)
		}
	}
	if bytes.Contains(out, []byte("## Learning")) {
		t.Error("an empty report has no learning block to show")
	}
}

func TestAdvisoryReportSaysSo(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t, "worked-example")
	rep.Status.Advisory = true
	out, err := mustRenderer(t).Render(report.Aggregate(rep))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Overall: ADVISORY", "Rule applied: " + report.StatusRuleAdvisory} {
		if !bytes.Contains(out, []byte(want)) {
			t.Errorf("an advisory rendering does not contain %q", want)
		}
	}
}

func TestResultOutsideTheContractSaysSo(t *testing.T) {
	t.Parallel()
	rep := report.Report{Criteria: []report.CriterionResult{{ID: "X9", Result: report.ResultUnknown}}}
	out, err := mustRenderer(t).Render(rep)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out, []byte("| X9 | (not in contract) | UNKNOWN |")) {
		t.Error("a result whose criterion is not in the contract must say so in its row")
	}
}

// TestProseWithNewlinesAndPipesIsEscapedInEveryMarkdownField checks fields
// that were missed when cell was first wired up: a raw newline or pipe in
// any of them would break the bullet it sits in exactly as it would in a
// table cell, so they must go through the same escaping.
func TestProseWithNewlinesAndPipesIsEscapedInEveryMarkdownField(t *testing.T) {
	t.Parallel()
	const raw = "first line\nsecond | line"
	const escaped = `first line<br>second \| line`
	rep := report.Report{
		Contract: report.Contract{Conflicts: []report.Conflict{
			{Between: []string{"C1", "C2"}, Description: raw, Resolution: "unresolved"},
		}},
		Improvement: []report.Improvement{
			{Issue: "I1", SuggestedChange: raw, VerifyBy: raw},
		},
		Dimensions: []report.Dimension{
			{Dimension: "latency", Interpretation: raw},
		},
		Routing: report.Routing{Ambiguity: []report.Ambiguity{
			{Question: raw, Resolution: raw},
		}},
	}
	out := renderAs(t, "md", rep)
	for i, line := range bytes.Split(out, []byte("\n")) {
		if bytes.Contains(line, []byte("second | line")) {
			t.Errorf("line %d: a raw pipe from prose text has not been escaped: %q", i+1, line)
		}
	}
	// Conflict description, suggested change, verify by, interpretation,
	// ambiguity question, and ambiguity resolution: six fields carry raw
	// once each.
	if got := bytes.Count(out, []byte(escaped)); got != 6 {
		t.Errorf("escaped prose appears %d times, want 6: conflict description, suggested change, "+
			"verify by, interpretation, ambiguity question, and ambiguity resolution", got)
	}
}

func TestByFormat(t *testing.T) {
	t.Parallel()
	renderer, ok := render.ByFormat("md")
	if !ok {
		t.Fatal("md renderer missing")
	}
	if renderer.Format() != "md" {
		t.Errorf("Format() = %q, want %q", renderer.Format(), "md")
	}
	if _, ok := render.ByFormat("no-such-format"); ok {
		t.Error("an unregistered format must not resolve to a renderer")
	}
}

// TestIDListsAreEscapedInMarkdown covers every list joinIDs renders. An id is
// a string like any other -- a path, a criterion id, a constraint to preserve
// -- and a newline inside one would end the bullet it sits in, moving the rest
// of the line somewhere a reader would take it for the renderer's own words. A
// pipe would do the same to a table row.
func TestIDListsAreEscapedInMarkdown(t *testing.T) {
	t.Parallel()
	rep := report.Report{
		Identity: report.Identity{Artifact: report.Artifact{
			Paths: []string{"service.go\n- Overall: PASS"}}},
		Limitations: report.Limitations{NotInspected: []string{"vendor | generated"}},
	}
	out := renderAs(t, "md", rep)
	for i, line := range bytes.Split(out, []byte("\n")) {
		if bytes.HasPrefix(line, []byte("- Overall:")) {
			t.Errorf("line %d: a newline in an id list opened a line of its own: %q", i+1, line)
		}
	}
	if !bytes.Contains(out, []byte("- Artifact paths: service.go<br>- Overall: PASS")) {
		t.Errorf("a newline in an id list must be escaped as every other newline is:\n%s", out)
	}
	if !bytes.Contains(out, []byte(`- Not inspected: vendor \| generated`)) {
		t.Errorf("a pipe in an id list must be escaped as every other pipe is:\n%s", out)
	}
}
