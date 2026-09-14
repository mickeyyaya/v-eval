package render_test

import (
	"bytes"
	"testing"

	"github.com/mickeyyaya/v-eval/core/render"
	"github.com/mickeyyaya/v-eval/core/report"
)

// htmlSectionIDs are the section anchors every report carries, in the order
// it carries them. The two conditional sections, dimensions and learning, are
// asserted per fixture instead: each fixture must show the one it has and
// must not invent the one it lacks.
var htmlSectionIDs = []string{
	`id="status"`, `id="observations"`, `id="claims"`, `id="criteria"`,
	`id="forensics"`, `id="counts"`, `id="improvement"`, `id="limitations"`,
	`id="routing"`, `id="provenance"`,
}

// mustHTMLRenderer returns the html renderer, or fails the test.
func mustHTMLRenderer(t *testing.T) render.Renderer {
	t.Helper()
	renderer, ok := render.ByFormat("html")
	if !ok {
		t.Fatal("html renderer missing")
	}
	return renderer
}

// renderHTML renders one report as HTML, or fails the test.
func renderHTML(t *testing.T, rep report.Report) []byte {
	t.Helper()
	out, err := mustHTMLRenderer(t).Render(rep)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// assertOrder requires every needle to appear, in order.
func assertOrder(t *testing.T, out []byte, needles []string) {
	t.Helper()
	last := -1
	for _, needle := range needles {
		i := bytes.Index(out, []byte(needle))
		if i < 0 || i < last {
			t.Fatalf("section %s missing or out of order", needle)
		}
		last = i
	}
}

func TestHTMLIsSelfContainedThemedOrderedAndEscaped(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t, "worked-example")
	rep.Observations = append(rep.Observations, report.Observation{
		ID: "O9", Text: "<script>alert(1)</script>", Origin: report.ObservationOriginAssistant,
		Evidence: []report.Evidence{}, CriterionIDs: []string{},
	})
	out := renderHTML(t, rep)
	for _, forbidden := range []string{"<script", "<link ", "@import", "url(http", "https://fonts"} {
		if bytes.Contains(out, []byte(forbidden)) {
			t.Fatalf("external resource reference %q", forbidden)
		}
	}
	if !bytes.Contains(out, []byte("prefers-color-scheme: dark")) {
		t.Fatal("must carry a dark theme")
	}
	if !bytes.Contains(out, []byte("&lt;script&gt;alert")) {
		t.Fatal("observation text must be escaped")
	}
	assertOrder(t, out, htmlSectionIDs)
	if !bytes.HasPrefix(out, []byte("<!doctype html>")) {
		t.Fatal("must be a complete document")
	}
	for _, want := range []string{`<meta charset="utf-8">`, `<meta name="viewport"`, "<title>", "<style>", ":root {"} {
		if !bytes.Contains(out, []byte(want)) {
			t.Errorf("a self-contained document must contain %q", want)
		}
	}
}

// htmlGoldenCase is one fixture and what its HTML rendering must say.
type htmlGoldenCase struct {
	name     string   // fixture basename, under core/report/testdata and testdata
	sections []string // every section anchor it must carry, in order
	absent   []string // anchors and words it must not carry
	contains []string // substrings that prove a rendering decision was made
}

func htmlGoldenCases() []htmlGoldenCase {
	return []htmlGoldenCase{
		{
			name:     "worked-example",
			sections: htmlSectionIDs,
			absent:   []string{`id="dimensions"`, `id="learning"`},
			contains: []string{
				"<title>v-eval report: code_change ",
				`class="badge fail"`, `class="badge pass"`, `class="badge unknown"`,
				"Whitespace and case variants collapse into one address",
				"(supplied)",
				"examples/code-review/input.md:27-27",
				"4/5 applicable criteria assessed",
				"developer reviewing an AI-generated change",
				"examples/code-review/input.md#intent",
				"kind inspection", "isolation none", "via assistant fixture",
				"required applicable criterion failed",
			},
		},
		{
			name: "extended-example",
			sections: []string{
				`id="status"`, `id="observations"`, `id="claims"`, `id="criteria"`,
				`id="forensics"`, `id="dimensions"`, `id="counts"`, `id="improvement"`,
				`id="limitations"`, `id="routing"`, `id="provenance"`, `id="learning"`,
			},
			contains: []string{
				"<title>v-eval report: service_change ",
				`class="badge error"`, `class="badge not_applicable"`,
				"unresolved contract conflict",
				"go test ./... (exit 0)",
				"docker build --no-cache . (exit 127)",
				"isolation worktree",
				"from https://api.example.test/docs/rate-limits (2026-08-30)",
				"rubric integrity-rubric-0.2, model unknown",
				"A checkout or a formatter touched the file without changing its contents.",
				"suspicious", "disposition open",
				"sha256:1f0a5c7f2b3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7",
			},
		},
	}
}

func TestHTMLGoldens(t *testing.T) {
	for _, testCase := range htmlGoldenCases() {
		t.Run(testCase.name, func(t *testing.T) {
			out := renderHTML(t, loadFixture(t, testCase.name))
			assertGolden(t, testCase.name+".html", out)
			assertOrder(t, out, testCase.sections)
			for _, unwanted := range testCase.absent {
				if bytes.Contains(out, []byte(unwanted)) {
					t.Errorf("%q must not appear: the fixture has no such content", unwanted)
				}
			}
			for _, want := range testCase.contains {
				if !bytes.Contains(out, []byte(want)) {
					t.Errorf("rendering does not contain %q", want)
				}
			}
		})
	}
}

func TestHTMLShowsAMissingDimensionValueAsMissing(t *testing.T) {
	t.Parallel()
	out := renderHTML(t, loadFixture(t, "extended-example"))
	if !bytes.Contains(out, []byte("<dd>missing</dd>")) {
		t.Error("a dimension with no value must say the value is missing")
	}
	if bytes.Contains(out, []byte("0 req/s")) {
		t.Error("a missing dimension value must never be shown as zero")
	}
}

func TestHTMLShowsAMeasuredZeroAsZero(t *testing.T) {
	t.Parallel()
	zero := 0.0
	rep := report.Report{Dimensions: []report.Dimension{{Dimension: "latency", Unit: "ms", Value: &zero}}}
	out := renderHTML(t, rep)
	if !bytes.Contains(out, []byte("<dd>0 ms</dd>")) {
		t.Error("a measured zero is a measurement and must be shown as zero")
	}
	if bytes.Contains(out, []byte("<dd>missing</dd>")) {
		t.Error("a measured zero must not be reported as a missing value")
	}
}

func TestHTMLOutputIsCleanMarkup(t *testing.T) {
	t.Parallel()
	out := renderHTML(t, loadFixture(t, "extended-example"))
	if !bytes.HasSuffix(out, []byte("</html>\n")) {
		t.Error("a rendering must be a closed document ending in one newline")
	}
	for i, line := range bytes.Split(out, []byte("\n")) {
		if len(line) != len(bytes.TrimRight(line, " \t")) {
			t.Errorf("line %d ends in whitespace: %q", i+1, line)
		}
	}
}

func TestProvisionalCriterionIsMarkedInEveryFormat(t *testing.T) {
	t.Parallel()
	const marker = "(provisional)"
	for _, format := range []string{"md", "html"} {
		t.Run(format, func(t *testing.T) {
			t.Parallel()
			rep := loadFixture(t, "worked-example")
			rep.Contract.Criteria[0].Provisional = true
			renderer, ok := render.ByFormat(format)
			if !ok {
				t.Fatalf("%s renderer missing", format)
			}
			out, err := renderer.Render(rep)
			if err != nil {
				t.Fatal(err)
			}
			want := "Whitespace and case variants collapse into one address " + marker
			if !bytes.Contains(out, []byte(want)) {
				t.Errorf("a provisional criterion is not marked in its row: want %q", want)
			}
			if got := bytes.Count(out, []byte(marker)); got != 1 {
				t.Errorf("%d criteria carry the provisional marker, want 1", got)
			}
		})
	}
}

func TestUnmarkedCriteriaCarryNoProvisionalMarker(t *testing.T) {
	t.Parallel()
	for _, format := range []string{"md", "html"} {
		renderer, ok := render.ByFormat(format)
		if !ok {
			t.Fatalf("%s renderer missing", format)
		}
		out, err := renderer.Render(loadFixture(t, "worked-example"))
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(out, []byte("(provisional)")) {
			t.Errorf("%s marks a criterion the contract does not call provisional", format)
		}
	}
}

func TestHTMLRenderingAnEmptyReportSaysSoRatherThanFailing(t *testing.T) {
	t.Parallel()
	out := renderHTML(t, report.Report{})
	for _, want := range []string{`id="status"`, "None recorded.", "Blocked by"} {
		if !bytes.Contains(out, []byte(want)) {
			t.Errorf("rendering of an empty report does not contain %q", want)
		}
	}
	if bytes.Contains(out, []byte(`id="learning"`)) || bytes.Contains(out, []byte(`id="dimensions"`)) {
		t.Error("an empty report has no learning or dimensions block to show")
	}
}

func TestHTMLByFormat(t *testing.T) {
	t.Parallel()
	if got := mustHTMLRenderer(t).Format(); got != "html" {
		t.Errorf("Format() = %q, want %q", got, "html")
	}
}
