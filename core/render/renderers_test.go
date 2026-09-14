package render_test

import (
	"bytes"
	"slices"
	"testing"

	"github.com/mickeyyaya/v-eval/core/render"
	"github.com/mickeyyaya/v-eval/core/report"
)

// formats is every registered output format. A rule that must hold "in both
// renderers" is written once here and run over all of them, so a format
// cannot quietly opt out of it: the list is the registry's own, so a format
// added to the build is covered without a second edit here.
var formats = render.Formats()

// assertEveryFormat renders rep in every format and checks that each
// rendering carries every want and none of the unwanted strings.
func assertEveryFormat(t *testing.T, rep report.Report, want, unwanted []string) {
	t.Helper()
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			out := renderAs(t, format, rep)
			for _, needle := range want {
				if !bytes.Contains(out, []byte(needle)) {
					t.Errorf("rendering does not contain %q", needle)
				}
			}
			for _, needle := range unwanted {
				if bytes.Contains(out, []byte(needle)) {
					t.Errorf("rendering must not contain %q", needle)
				}
			}
		})
	}
}

func TestProvisionalCriterionIsMarkedInEveryFormat(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t, "worked-example")
	rep.Contract.Criteria[0].Provisional = true
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			out := renderAs(t, format, rep)
			want := "Whitespace and case variants collapse into one address (provisional)"
			if !bytes.Contains(out, []byte(want)) {
				t.Errorf("a provisional criterion is not marked in its row: want %q", want)
			}
			if got := bytes.Count(out, []byte("(provisional)")); got != 1 {
				t.Errorf("%d criteria carry the provisional marker, want 1", got)
			}
		})
	}
}

func TestOptionalCriterionIsMarkedInEveryFormat(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t, "extended-example")
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			out := renderAs(t, format, rep)
			// E3 is the one criterion the contract does not require.
			want := "Throughput stays at or above the stated target (optional)"
			if !bytes.Contains(out, []byte(want)) {
				t.Errorf("an optional criterion is not marked in its row: want %q", want)
			}
			if got := bytes.Count(out, []byte("(optional)")); got != 1 {
				t.Errorf("%d criteria carry the optional marker, want 1", got)
			}
		})
	}
}

// TestUnmarkedCriteriaCarryNoOptionalOrProvisionalMarker covers both markers
// the worked example's required, unmarked criteria must never carry: the two
// cases differ only in which marker they check for, so they run as one
// table over both.
func TestUnmarkedCriteriaCarryNoOptionalOrProvisionalMarker(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t, "worked-example")
	for _, marker := range []string{"(optional)", "(provisional)"} {
		t.Run(marker, func(t *testing.T) {
			assertEveryFormat(t, rep, nil, []string{marker})
		})
	}
}

func TestCoverageIsUndefinedWhenNothingIsApplicable(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t, "worked-example")
	for i := range rep.Criteria {
		rep.Criteria[i].Result = report.ResultNotApplicable
	}
	rep = report.Aggregate(rep)
	assertEveryFormat(t, rep,
		[]string{"undefined (0 applicable)", "Coverage is not a correctness score."},
		[]string{"applicable criteria assessed"})
}

func TestAMeasuredZeroIsAValueNotAMissingMeasurement(t *testing.T) {
	t.Parallel()
	zero := 0.0
	rep := report.Report{Dimensions: []report.Dimension{
		{Dimension: "latency", Unit: "ms", Value: &zero},
	}}
	assertEveryFormat(t, rep, []string{"0 ms"}, []string{"not measured"})
}

func TestAMissingDimensionValueSaysNotMeasured(t *testing.T) {
	t.Parallel()
	rep := report.Report{Dimensions: []report.Dimension{
		{Dimension: "latency", Unit: "ms", Value: nil},
	}}
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			out := renderAs(t, format, rep)
			if !bytes.Contains(out, []byte("not measured")) {
				t.Error(`rendering does not contain "not measured"`)
			}
			if bytes.Contains(out, []byte("0 ms")) {
				t.Error(`rendering must not contain "0 ms"`)
			}
			switch format {
			case "html":
				if bytes.Contains(out, []byte("<dd>missing</dd>")) {
					t.Error(`rendering must not contain "<dd>missing</dd>"`)
				}
			case "md":
				if !bytes.Contains(out, []byte("Value: not measured")) {
					t.Error(`rendering does not contain "Value: not measured"`)
				}
			}
		})
	}
}

func TestEveryEvidenceLineNamesAModelItHasAndOnlyJudgmentsNameARubric(t *testing.T) {
	t.Parallel()
	rep := report.Report{Criteria: []report.CriterionResult{{
		ID:     "C1",
		Result: report.ResultPass,
		Evidence: []report.Evidence{
			{Kind: report.KindInspection, Observation: "read the constant", Model: "claude-x"},
			{Kind: report.KindJudgment, Observation: "applied the integrity scale", RubricVersion: "r-1"},
		},
	}}}
	assertEveryFormat(t, rep, []string{"model claude-x", "rubric r-1", "model (not recorded)"}, nil)
	for _, format := range formats {
		out := renderAs(t, format, rep)
		if got := bytes.Count(out, []byte("rubric")); got != 1 {
			t.Errorf("%s names a rubric on %d evidence lines, want 1: only a judgment applies one", format, got)
		}
	}
}

func TestEvidenceWithNoToolRecordedNamesItRatherThanADanglingVia(t *testing.T) {
	t.Parallel()
	rep := report.Report{Criteria: []report.CriterionResult{{
		ID:     "C1",
		Result: report.ResultPass,
		Evidence: []report.Evidence{
			{Kind: report.KindInspection, Observation: "read the constant"},
		},
	}}}
	assertEveryFormat(t, rep, []string{"via (not recorded)"}, []string{"via  "})
}

func TestAnEmptyCriteriaListSaysNoneRecorded(t *testing.T) {
	t.Parallel()
	cases := map[string]struct{ want, unwanted string }{
		"md":   {"## Criteria\n\nNone recorded.", "| ID | Requirement |"},
		"html": {"<h2>Criteria</h2>\n<p class=\"none\">None recorded.</p>", "<th scope=\"col\">Requirement</th>"},
	}
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			testCase := cases[format]
			out := renderAs(t, format, report.Report{})
			if !bytes.Contains(out, []byte(testCase.want)) {
				t.Errorf("an empty criteria list does not say so: want %q", testCase.want)
			}
			if bytes.Contains(out, []byte(testCase.unwanted)) {
				t.Errorf("an empty criteria list must not render a header-only table: %q", testCase.unwanted)
			}
		})
	}
}

func TestFormatsListsEveryRegisteredRendererInOrder(t *testing.T) {
	t.Parallel()
	got := render.Formats()
	want := []string{"html", "md"}
	if !slices.Equal(got, want) {
		t.Fatalf("Formats() = %v, want %v", got, want)
	}
	for _, format := range got {
		renderer, ok := render.ByFormat(format)
		if !ok {
			t.Fatalf("Formats() names %q, which ByFormat does not resolve", format)
		}
		if renderer.Format() != format {
			t.Errorf("ByFormat(%q).Format() = %q: a name must round-trip", format, renderer.Format())
		}
	}
}
