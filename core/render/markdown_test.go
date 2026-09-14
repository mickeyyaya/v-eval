package render_test

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/mickeyyaya/v-eval/core/render"
	"github.com/mickeyyaya/v-eval/core/report"
)

// commonSections are the headings every report shows, in the order it shows
// them. Dimensions and Learning are absent: both are conditional.
var commonSections = []string{
	"## Status", "## Observations", "## Claims", "## Criteria", "## Forensics",
	"## Counts", "## Improvement", "## Limitations", "## Routing", "## Provenance",
}

// goldenCase is one fixture and what its rendering must say.
type goldenCase struct {
	name     string   // fixture basename, under core/report/testdata and testdata
	sections []string // every heading it must carry, in order
	absent   []string // headings it must not carry
	contains []string // substrings that prove a rendering decision was made
}

// goldenCases covers both worked fixtures: one without dimensions or
// learning, one with both, so the conditional sections are exercised in each
// direction.
func goldenCases() []goldenCase {
	return []goldenCase{
		{
			name:     "worked-example",
			sections: commonSections,
			absent:   []string{"## Dimensions", "## Learning"},
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
			name: "extended-example",
			sections: []string{
				"## Status", "## Observations", "## Claims", "## Criteria", "## Forensics",
				"## Dimensions", "## Counts", "## Improvement", "## Limitations",
				"## Routing", "## Provenance", "## Learning",
			},
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
			out := renderFixture(t, testCase.name)
			// The structural checks run first: they name what is wrong,
			// where a golden mismatch only says that something is.
			assertOrder(t, out, testCase.sections)
			for _, heading := range testCase.absent {
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
	for _, testCase := range goldenCases() {
		t.Run(testCase.name, func(t *testing.T) {
			rep := loadFixture(t, testCase.name)
			out := renderFixture(t, testCase.name)
			want := len(report.WalkEvidence(rep))
			if want == 0 {
				t.Fatal("fixture carries no evidence to render")
			}
			if got := len(evidenceBadge.FindAll(out, -1)); got != want {
				t.Errorf("%d evidence lines carry a kind and an isolation, want %d", got, want)
			}
		})
	}
}

func TestRenderingIsCleanMarkdown(t *testing.T) {
	for _, testCase := range goldenCases() {
		t.Run(testCase.name, func(t *testing.T) {
			out := renderFixture(t, testCase.name)
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
