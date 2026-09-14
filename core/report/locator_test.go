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

// TestSupportsErrorSurvivesACommandLocatorWithNoExitStatus covers the one way
// the exit status could be read before it exists. A command group is complete
// only with an exit status, so Shape already rules this out and the answer is
// false either way; the guard is there so that a later change to what makes a
// command group complete cannot turn this into a panic in a validation rule.
func TestSupportsErrorSurvivesACommandLocatorWithNoExitStatus(t *testing.T) {
	t.Parallel()
	evidence := Evidence{
		Kind:    KindInspection,
		Locator: Locator{Command: "go test ./...", Cwd: "/w", LogRef: "logs/1.txt"},
	}
	if evidence.SupportsError() {
		t.Fatal("a command locator with no exit status shows no failed attempt")
	}
}

// TestSupportsErrorReadsExecutionAndNonZeroExits pins what an ERROR may rest
// on: evidence that ran, or evidence pointing at a command that came back
// non-zero. A command that exited zero shows an attempt that succeeded, which
// is not what an ERROR is about.
func TestSupportsErrorReadsExecutionAndNonZeroExits(t *testing.T) {
	t.Parallel()
	zero, failed := 0, 127
	command := func(exit *int) Locator {
		return Locator{Command: "docker build .", Cwd: "/w", ExitStatus: exit, LogRef: "logs/1.txt"}
	}
	cases := map[string]struct {
		evidence Evidence
		want     bool
	}{
		"execution evidence":      {Evidence{Kind: KindExecution, Locator: Locator{Note: "the runner died"}}, true},
		"a command that failed":   {Evidence{Kind: KindInspection, Locator: command(&failed)}, true},
		"a command that exited 0": {Evidence{Kind: KindInspection, Locator: command(&zero)}, false},
		"a note":                  {Evidence{Kind: KindSupplied, Locator: Locator{Note: "the author says so"}}, false},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := testCase.evidence.SupportsError(); got != testCase.want {
				t.Errorf("SupportsError() = %v, want %v", got, testCase.want)
			}
		})
	}
}
