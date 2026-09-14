package render_test

import (
	"bytes"
	"regexp"
	"strings"
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

func TestHTMLIsSelfContainedThemedOrderedAndEscaped(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t, "worked-example")
	rep.Observations = append(rep.Observations, report.Observation{
		ID: "O9", Text: "<script>alert(1)</script>", Origin: report.ObservationOriginAssistant,
		Evidence: []report.Evidence{}, CriterionIDs: []string{},
	})
	out := renderAs(t, "html", rep)
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
			out := renderAs(t, "html", loadFixture(t, testCase.name))
			// The structural checks run first: they name what is wrong,
			// where a golden mismatch only says that something is.
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
			assertGolden(t, testCase.name+".html", out)
		})
	}
}

func TestHTMLOutputIsCleanMarkup(t *testing.T) {
	t.Parallel()
	out := renderAs(t, "html", loadFixture(t, "extended-example"))
	if !bytes.HasSuffix(out, []byte("</html>\n")) {
		t.Error("a rendering must be a closed document ending in one newline")
	}
	for i, line := range bytes.Split(out, []byte("\n")) {
		if len(line) != len(bytes.TrimRight(line, " \t")) {
			t.Errorf("line %d ends in whitespace: %q", i+1, line)
		}
	}
}

func TestHTMLRenderingAnEmptyReportSaysSoRatherThanFailing(t *testing.T) {
	t.Parallel()
	out := renderAs(t, "html", report.Report{})
	for _, want := range []string{`id="status"`, "None recorded.", "Blocked by"} {
		if !bytes.Contains(out, []byte(want)) {
			t.Errorf("rendering of an empty report does not contain %q", want)
		}
	}
	if bytes.Contains(out, []byte(`id="learning"`)) || bytes.Contains(out, []byte(`id="dimensions"`)) {
		t.Error("an empty report has no learning or dimensions block to show")
	}
}

func TestHTMLKeepsTheShapeOfAnExcerpt(t *testing.T) {
	t.Parallel()
	const excerpt = "first line\n    indented line"
	rep := report.Report{
		Observations: []report.Observation{{
			ID: "O1", Text: excerpt,
			Evidence: []report.Evidence{{Kind: report.KindInspection, Observation: excerpt}},
		}},
		Criteria: []report.CriterionResult{{ID: "C1", Result: report.ResultPass, Reasoning: excerpt}},
	}
	out := renderAs(t, "html", rep)
	if !bytes.Contains(out, []byte("white-space: pre-wrap")) {
		t.Fatal("excerpt text must keep its newlines and indentation")
	}
	for _, want := range []string{
		`<p class="text">` + excerpt + `</p>`,
		`<span class="text">` + excerpt + `</span>`,
	} {
		if !bytes.Contains(out, []byte(want)) {
			t.Errorf("rendering does not carry %q on a pre-wrapped element", want)
		}
	}
	if got := bytes.Count(out, []byte(`<p class="text">`+excerpt+`</p>`)); got != 2 {
		t.Errorf("%d pre-wrapped paragraphs carry the excerpt, want 2 (observation text and reasoning)", got)
	}
}

func TestHTMLTablesScrollSidewaysSoTheBodyDoesNot(t *testing.T) {
	t.Parallel()
	out := renderAs(t, "html", loadFixture(t, "extended-example"))
	const wrapper = `<div class="scroll">`
	tables := bytes.Count(out, []byte("<table"))
	if tables == 0 {
		t.Fatal("the fixture renders no table to wrap")
	}
	if got := bytes.Count(out, []byte(wrapper)); got != tables {
		t.Errorf("%d scroll wrappers for %d tables, want one each", got, tables)
	}
	for i, before := range bytes.Split(out, []byte("<table"))[:tables] {
		if !bytes.HasSuffix(bytes.TrimRight(before, " \t\n"), []byte(wrapper)) {
			t.Errorf("table %d is not directly inside %s", i+1, wrapper)
		}
	}
	for _, want := range []string{"overflow-x: auto", "overflow-wrap: anywhere", "padding: 2rem 1rem 4rem"} {
		if !bytes.Contains(out, []byte(want)) {
			t.Errorf("the stylesheet does not carry %q, so the body can scroll sideways", want)
		}
	}
}

func TestBadgeUnknownRuleDoesNotRepeatTheBaseBadgeColors(t *testing.T) {
	t.Parallel()
	out := renderAs(t, "html", report.Report{})
	if bytes.Contains(out, []byte(".badge.unknown { background: var(--unknown-bg); color: var(--unknown-fg); }")) {
		t.Error(".badge.unknown repeats colors the base .badge rule already sets as its default")
	}
}

// customPropertyNames finds every "--name" custom property declared in a
// block of CSS text, by a plain regexp scan: good enough for a stylesheet we
// wrote ourselves, where a property is never assigned as someone else's value
// inside the block we are scanning.
var customPropertyNameRE = regexp.MustCompile(`--[a-z-]+`)

func customPropertyNames(t *testing.T, style, openLine string) map[string]bool {
	t.Helper()
	start := strings.Index(style, openLine)
	if start < 0 {
		t.Fatalf("style block does not contain %q", openLine)
	}
	end := strings.Index(style[start:], "\n}")
	if end < 0 {
		t.Fatalf("no closing brace found for the block opened by %q", openLine)
	}
	names := map[string]bool{}
	for _, name := range customPropertyNameRE.FindAllString(style[start:start+end], -1) {
		names[name] = true
	}
	return names
}

func TestDarkThemeRedefinesExactlyTheLightCustomProperties(t *testing.T) {
	t.Parallel()
	out := renderAs(t, "html", report.Report{})
	i, j := bytes.Index(out, []byte("<style>")), bytes.Index(out, []byte("</style>"))
	if i < 0 || j < 0 || j <= i {
		t.Fatal("rendering does not carry a <style> block")
	}
	style := string(out[i:j])

	// The light-mode block opens unindented, at "\n:root {"; the dark-mode
	// block nests its :root two spaces in, under the prefers-color-scheme
	// media query, so the same marker distinguishes the two.
	light := customPropertyNames(t, style, "\n:root {")
	dark := customPropertyNames(t, style, "\n  :root {")

	if len(light) == 0 {
		t.Fatal("no custom properties found under :root")
	}
	for name := range light {
		if !dark[name] {
			t.Errorf("%s is declared under :root but never redefined for dark mode", name)
		}
	}
	for name := range dark {
		if !light[name] {
			t.Errorf("%s is redefined for dark mode but never declared under :root", name)
		}
	}
}

func TestHTMLByFormat(t *testing.T) {
	t.Parallel()
	if got := mustHTMLRenderer(t).Format(); got != "html" {
		t.Errorf("Format() = %q, want %q", got, "html")
	}
}
