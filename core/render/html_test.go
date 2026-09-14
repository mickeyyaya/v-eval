package render_test

import (
	"bytes"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/mickeyyaya/v-eval/core/render"
	"github.com/mickeyyaya/v-eval/core/report"
)

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
	ordered, _ := sectionMarkers(t, "html", sectionNamesExcept("dimensions", "learning"))
	assertOrder(t, out, ordered)
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
	sections []string // every section it carries, named as the section table names them
	contains []string // substrings that prove a rendering decision was made
}

func htmlGoldenCases() []htmlGoldenCase {
	return []htmlGoldenCase{
		{
			name:     "worked-example",
			sections: sectionNamesExcept("dimensions", "learning"),
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
			name:     "extended-example",
			sections: sectionNamesExcept(),
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
			want, absent := sectionMarkers(t, "html", testCase.sections)
			assertOrder(t, out, want)
			for _, unwanted := range absent {
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

// htmlEvidenceEntry is the opening of one rendered evidence card: every
// card opens with a tag line that states what it is -- the kind tag carries
// its kind as a class, so the stylesheet can tint it -- and how isolated it
// ran, so counting the openings counts the entries.
var htmlEvidenceEntry = regexp.MustCompile(`<li><span class="tags"><span class="tag [a-z_]+">kind [a-z_]+</span> <span class="tag">isolation [a-z_]+</span>`)

// TestHTMLShowsEveryEvidenceRecord is the HTML twin of the Markdown guard:
// a renderer that silently drops a citation is the one failure a golden
// cannot name, because a golden only knows what was rendered last time.
func TestHTMLShowsEveryEvidenceRecord(t *testing.T) {
	assertEveryEvidenceRendered(t, "html", htmlEvidenceEntry)
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

// TestHTMLSeparatesEveryAmbiguity guards the one list in the routing block
// that the fixtures cannot guard: both carry a single ambiguity, so a missing
// separator would run two questions together only in reports no golden holds.
func TestHTMLSeparatesEveryAmbiguity(t *testing.T) {
	t.Parallel()
	rep := report.Report{Routing: report.Routing{Ambiguity: []report.Ambiguity{
		{Question: "Which suite counts?", Resolution: "the one the brief names"},
		{Question: "Which revision?", Resolution: "the one the bundle digest pins"},
	}}}
	out := renderAs(t, "html", rep)
	const want = "<dt>Ambiguity</dt><dd>Which suite counts? Resolved: the one the brief names; " +
		"Which revision? Resolved: the one the bundle digest pins</dd>"
	if !bytes.Contains(out, []byte(want)) {
		t.Errorf("ambiguities run together; want %q", want)
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

// navLinkRE matches one link in the section nav: the id it points at and the
// label a reader sees.
var navLinkRE = regexp.MustCompile(`<a href="#([a-z]+)">([^<]+)</a>`)

// renderedSectionRE matches the opening of one rendered section, by id.
var renderedSectionRE = regexp.MustCompile(`<section id="([a-z]+)">`)

// sectionLinks are the nav entries the section table predicts for a report
// carrying the named sections: the id is the section name -- every anchor in
// the table is `id="`+name+`"` -- the label the Markdown heading stripped of
// its marks, so the nav is held to the same table the section order is.
func sectionLinks(present []string) [][2]string {
	var links [][2]string
	for _, section := range reportSections {
		if slices.Contains(present, section.name) {
			links = append(links, [2]string{section.name, strings.TrimPrefix(section.heading, "## ")})
		}
	}
	return links
}

// between returns the text of out from the first open marker to the next
// close marker, or fails the test when either is missing.
func between(t *testing.T, out []byte, open, close string) string {
	t.Helper()
	i := bytes.Index(out, []byte(open))
	if i < 0 {
		t.Fatalf("rendering does not contain %q", open)
	}
	j := bytes.Index(out[i:], []byte(close))
	if j < 0 {
		t.Fatalf("%q is never closed by %q", open, close)
	}
	return string(out[i : i+j])
}

func TestHTMLNavLinksEveryRenderedSectionInOrder(t *testing.T) {
	t.Parallel()
	for _, testCase := range htmlGoldenCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			out := renderAs(t, "html", loadFixture(t, testCase.name))
			nav := between(t, out, `<nav class="toc" aria-label="Sections">`, "</nav>")
			var got [][2]string
			for _, m := range navLinkRE.FindAllStringSubmatch(nav, -1) {
				got = append(got, [2]string{m[1], m[2]})
			}
			if want := sectionLinks(testCase.sections); !slices.Equal(got, want) {
				t.Errorf("nav links = %v, want %v", got, want)
			}
			var rendered []string
			for _, m := range renderedSectionRE.FindAllSubmatch(out, -1) {
				rendered = append(rendered, string(m[1]))
			}
			var linked []string
			for _, link := range got {
				linked = append(linked, link[0])
			}
			if !slices.Equal(linked, rendered) {
				t.Errorf("nav links %v do not match the rendered sections %v", linked, rendered)
			}
		})
	}
}

func TestHTMLHeroShowsTheOverallVerdictBeforeTheFirstSection(t *testing.T) {
	t.Parallel()
	out := renderAs(t, "html", loadFixture(t, "worked-example"))
	hero := between(t, out, `class="hero"`, "</div>")
	if !strings.Contains(hero, `<span class="badge fail">FAIL</span>`) {
		t.Errorf("the hero does not carry the overall badge: %q", hero)
	}
	if !strings.Contains(hero, "Rule applied") {
		t.Errorf("the hero does not name the rule applied: %q", hero)
	}
	if i, j := bytes.Index(out, []byte(`class="hero"`)), bytes.Index(out, []byte("<section")); j < 0 || i > j {
		t.Errorf("the hero (at %d) must come before the first section (at %d)", i, j)
	}
}

func TestHTMLBlockedByRendersAsChipsOrNone(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"worked-example":   `<dt>Blocked by</dt><dd><span class="chip">C1</span> <span class="chip">C2</span> <span class="chip">C3</span></dd>`,
		"extended-example": `<dt>Blocked by</dt><dd>none</dd>`,
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			out := renderAs(t, "html", loadFixture(t, name))
			hero := between(t, out, `class="hero"`, "</div>")
			if !strings.Contains(hero, want) {
				t.Errorf("hero does not carry %q:\n%s", want, hero)
			}
		})
	}
}

func TestHTMLPrintStylesheetHidesTheNav(t *testing.T) {
	t.Parallel()
	out := renderAs(t, "html", report.Report{})
	print := between(t, out, "@media print {", "\n}")
	if !strings.Contains(print, "nav.toc { display: none; }") {
		t.Errorf("the print stylesheet does not hide the nav:\n%s", print)
	}
	for _, want := range []string{"overflow: visible", "break-inside: avoid", "print-color-adjust: exact", "max-width: none"} {
		if !strings.Contains(print, want) {
			t.Errorf("the print stylesheet does not carry %q", want)
		}
	}
}

func TestHTMLFooterNamesTheReport(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t, "worked-example")
	out := renderAs(t, "html", rep)
	footer := between(t, out, "<footer>", "</footer>")
	for _, want := range []string{
		"Report " + rep.Identity.ReportID,
		"schema " + rep.Identity.SchemaVersion,
		"v-eval " + rep.Identity.VevalVersion,
	} {
		if !strings.Contains(footer, want) {
			t.Errorf("footer does not carry %q:\n%s", want, footer)
		}
	}
}

// TestHTMLHeadingsCarryTheCountOfWhatTheyList checks that a heading's count
// chip is the length of the list it heads -- a presentation of report data,
// not a number the renderer computes on its own -- and that an empty list
// carries none, because "None recorded." already says so.
func TestHTMLHeadingsCarryTheCountOfWhatTheyList(t *testing.T) {
	t.Parallel()
	for _, name := range evidenceFixtures {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rep := loadFixture(t, name)
			out := renderAs(t, "html", rep)
			counts := map[string]int{
				"Observations": len(rep.Observations),
				"Claims":       len(rep.Claims),
				"Criteria":     len(rep.Criteria),
				"Forensics":    len(rep.Forensics),
				"Improvement":  len(rep.Improvement),
			}
			for heading, n := range counts {
				want := fmt.Sprintf("<h2>%s <span class=\"count\">%d</span></h2>", heading, n)
				if n == 0 {
					want = "<h2>" + heading + "</h2>"
				}
				if !bytes.Contains(out, []byte(want)) {
					t.Errorf("rendering does not contain %q", want)
				}
			}
		})
	}
}

func TestHTMLForensicsCarryTheSeverityAsABadge(t *testing.T) {
	t.Parallel()
	out := renderAs(t, "html", loadFixture(t, "extended-example"))
	const want = `On E6: severity <span class="badge suspicious">suspicious</span>, disposition open.`
	if !bytes.Contains(out, []byte(want)) {
		t.Errorf("rendering does not contain %q", want)
	}
}
