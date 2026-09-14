package render_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"testing"

	"github.com/mickeyyaya/v-eval/core/render"
	"github.com/mickeyyaya/v-eval/core/report"
)

// update rewrites the golden files instead of comparing against them. The
// goldens are read by eye, so regenerating them is a deliberate act.
var update = flag.Bool("update", false, "rewrite golden files")

// loadFixture decodes one of the report package's worked fixtures by name.
func loadFixture(t *testing.T, name string) report.Report {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "report", "testdata", name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	rep, err := report.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	return rep
}

// assertGolden compares got against testdata/name, or rewrites it under -update.
func assertGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("missing golden %s; run: go test ./core/render -update", path)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("%s differs from golden; run go test ./core/render -update and review the diff", name)
	}
}

// reportSection is one section of a report and how each renderer names it:
// the heading a Markdown reader scans for, and the id the HTML section
// carries.
type reportSection struct {
	name    string
	heading string
	anchor  string
}

// reportSections is the one section order both renderers hold to, the
// conditional sections included. It lives here rather than beside either
// golden test so that a template which reorders a section fails both formats,
// and so that the order is stated once rather than kept in step by hand.
var reportSections = []reportSection{
	{"status", "## Status", `id="status"`},
	{"observations", "## Observations", `id="observations"`},
	{"claims", "## Claims", `id="claims"`},
	{"criteria", "## Criteria", `id="criteria"`},
	{"forensics", "## Forensics", `id="forensics"`},
	{"dimensions", "## Dimensions", `id="dimensions"`},
	{"counts", "## Counts", `id="counts"`},
	{"improvement", "## Improvement", `id="improvement"`},
	{"limitations", "## Limitations", `id="limitations"`},
	{"routing", "## Routing", `id="routing"`},
	{"provenance", "## Provenance", `id="provenance"`},
	{"learning", "## Learning", `id="learning"`},
}

// sectionNamesExcept names every section but the ones given, in table order.
// A fixture states what it lacks -- dimensions, learning -- rather than
// restating the whole order, so adding a section to the table adds it to
// every case that does not opt out.
func sectionNamesExcept(absent ...string) []string {
	names := make([]string, 0, len(reportSections))
	for _, section := range reportSections {
		if !slices.Contains(absent, section.name) {
			names = append(names, section.name)
		}
	}
	return names
}

// sectionMarkers splits the table, in table order, into the markers a report
// carrying the named sections must show and the markers it must not: a
// fixture without dimensions must be checked for their absence as strictly as
// for the presence of the rest.
func sectionMarkers(t *testing.T, format string, present []string) (want, unwanted []string) {
	t.Helper()
	for _, name := range present {
		if !slices.ContainsFunc(reportSections, func(s reportSection) bool { return s.name == name }) {
			t.Fatalf("section %q is not in the section table", name)
		}
	}
	marker, ok := sectionMarkerFunc(format)
	if !ok {
		t.Fatalf("no section markers for format %q", format)
	}
	for _, section := range reportSections {
		if slices.Contains(present, section.name) {
			want = append(want, marker(section))
			continue
		}
		unwanted = append(unwanted, marker(section))
	}
	return want, unwanted
}

// sectionMarkerFunc returns the function that names one section's marker in
// the given format, and whether the format is known.
func sectionMarkerFunc(format string) (marker func(reportSection) string, ok bool) {
	switch format {
	case "md":
		return func(s reportSection) string { return s.heading }, true
	case "html":
		return func(s reportSection) string { return s.anchor }, true
	default:
		return nil, false
	}
}

// assertOrder requires every needle to appear in out, in order.
func assertOrder(t *testing.T, out []byte, needles []string) {
	t.Helper()
	last := -1
	for _, needle := range needles {
		i := bytes.Index(out, []byte(needle))
		if i < 0 || i < last {
			t.Fatalf("section %q missing or out of order", needle)
		}
		last = i
	}
}

// mustRenderer returns the markdown renderer, or fails the test.
func mustRenderer(t *testing.T) render.Renderer {
	t.Helper()
	renderer, ok := render.ByFormat("md")
	if !ok {
		t.Fatal("md renderer missing")
	}
	return renderer
}

// renderAs renders one report in one format, or fails the test.
func renderAs(t *testing.T, format string, rep report.Report) []byte {
	t.Helper()
	renderer, ok := render.ByFormat(format)
	if !ok {
		t.Fatalf("%s renderer missing", format)
	}
	out, err := renderer.Render(rep)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// evidenceFixtures are the fixtures assertEveryEvidenceRendered checks: both
// carry evidence, so a renderer that drops a citation is caught without
// hand-picking which fixture would show it.
var evidenceFixtures = []string{"worked-example", "extended-example"}

// assertEveryEvidenceRendered renders each fixture in format and asserts that
// pattern matches exactly once per evidence entry the fixture carries: a
// renderer that silently drops a citation is a failure no golden alone can
// catch, because a golden only knows what was rendered last time.
func assertEveryEvidenceRendered(t *testing.T, format string, pattern *regexp.Regexp) {
	t.Helper()
	for _, name := range evidenceFixtures {
		t.Run(name, func(t *testing.T) {
			rep := loadFixture(t, name)
			out := renderAs(t, format, rep)
			want := len(report.WalkEvidence(rep))
			if want == 0 {
				t.Fatal("fixture carries no evidence to render")
			}
			if got := len(pattern.FindAll(out, -1)); got != want {
				t.Errorf("%d evidence entries rendered, want %d", got, want)
			}
		})
	}
}
