package report

import "testing"

func TestLocatorShapes(t *testing.T) {
	t.Parallel()
	zero := 0
	cases := map[string]struct {
		loc  Locator
		want LocatorShape
	}{
		"file":            {Locator{File: "a.go", LineStart: 3, LineEnd: 9}, ShapeFile},
		"file bad range":  {Locator{File: "a.go", LineStart: 9, LineEnd: 3}, ShapeNone},
		"file no lines":   {Locator{File: "a.go"}, ShapeNone},
		"command":         {Locator{Command: "go test ./...", Cwd: "/w", ExitStatus: &zero, LogRef: "logs/1.txt"}, ShapeCommand},
		"command no exit": {Locator{Command: "go test", Cwd: "/w", LogRef: "l"}, ShapeNone},
		"passage":         {Locator{Passage: "Export requires a connection.", SourceRef: "guide.md", SourceDateOrVersion: "v3", AccessDate: "2026-09-14"}, ShapePassage},
		"passage no date": {Locator{Passage: "x", SourceRef: "guide.md"}, ShapeNone},
		"note":            {Locator{Note: "PR description"}, ShapeNote},
		"mixed":           {Locator{File: "a.go", LineStart: 1, LineEnd: 1, Note: "x"}, ShapeNone},
		"empty":           {Locator{}, ShapeNone},
	}
	for name, c := range cases {
		if got := c.loc.Shape(); got != c.want {
			t.Errorf("%s: Shape = %q, want %q", name, got, c.want)
		}
	}
}

func TestSupportsPassNeedsObservedQualifyingLocator(t *testing.T) {
	t.Parallel()
	file := Locator{File: "a.go", LineStart: 1, LineEnd: 1}
	if !(Evidence{Origin: OriginObserved, Locator: file}).SupportsPass() {
		t.Fatal("observed file evidence must support PASS")
	}
	if (Evidence{Origin: OriginCandidateSupplied, Locator: file}).SupportsPass() {
		t.Fatal("candidate-supplied evidence must never support PASS")
	}
	if (Evidence{Origin: OriginObserved, Locator: Locator{Note: "x"}}).SupportsPass() {
		t.Fatal("note locator must never support PASS")
	}
}
