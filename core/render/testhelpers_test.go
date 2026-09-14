package render_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
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
	for _, section := range reportSections {
		var marker string
		switch format {
		case "md":
			marker = section.heading
		case "html":
			marker = section.anchor
		default:
			t.Fatalf("no section markers for format %q", format)
		}
		if slices.Contains(present, section.name) {
			want = append(want, marker)
			continue
		}
		unwanted = append(unwanted, marker)
	}
	return want, unwanted
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
