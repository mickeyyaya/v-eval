package render_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
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
