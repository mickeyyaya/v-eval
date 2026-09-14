package render

import (
	"testing"

	"github.com/mickeyyaya/v-eval/core/report"
)

func TestLocatorRendersEachShape(t *testing.T) {
	t.Parallel()
	exit := 127
	cases := []struct {
		name    string
		locator report.Locator
		want    string
	}{
		{"file", report.Locator{File: "svc.go", LineStart: 4, LineEnd: 9}, "svc.go:4-9"},
		{"command", report.Locator{Command: "go test ./...", Cwd: "/w", ExitStatus: &exit, LogRef: "l.log"},
			"go test ./... (exit 127)"},
		{"passage", report.Locator{Passage: "at most 100", SourceRef: "https://x.test", SourceDateOrVersion: "2026-08-30", AccessDate: "2026-09-14"},
			"at most 100 from https://x.test (2026-08-30)"},
		{"note", report.Locator{Note: "PR description, paragraph 2"}, "PR description, paragraph 2"},
		{"incomplete", report.Locator{File: "svc.go"}, "(no locator)"},
		{"mixed", report.Locator{File: "svc.go", LineStart: 1, LineEnd: 1, Note: "and a note"}, "(no locator)"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := locator(testCase.locator); got != testCase.want {
				t.Errorf("locator() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestSuppliedMarksOnlyCandidateSuppliedEvidence(t *testing.T) {
	t.Parallel()
	for _, origin := range []report.Origin{report.OriginObserved, report.OriginRetrieved} {
		if got := supplied(report.Evidence{Origin: origin}); got != "" {
			t.Errorf("supplied(%s) = %q, want empty", origin, got)
		}
	}
	if got := supplied(report.Evidence{Origin: report.OriginCandidateSupplied}); got != " (supplied)" {
		t.Errorf("supplied(candidate_supplied) = %q, want %q", got, " (supplied)")
	}
}

func TestJoinIDsNamesTheEmptyList(t *testing.T) {
	t.Parallel()
	if got := joinIDs(nil); got != "none" {
		t.Errorf("joinIDs(nil) = %q, want %q", got, "none")
	}
	if got := joinIDs([]string{"C1", "C2"}); got != "C1, C2" {
		t.Errorf("joinIDs() = %q, want %q", got, "C1, C2")
	}
}

func TestBadgeIsTheResultWord(t *testing.T) {
	t.Parallel()
	if got := badge(report.ResultNotApplicable); got != "NOT_APPLICABLE" {
		t.Errorf("badge() = %q, want %q", got, "NOT_APPLICABLE")
	}
}

func TestCellKeepsTableRowsIntact(t *testing.T) {
	t.Parallel()
	if got := cell("a | b\nc"); got != `a \| b<br>c` {
		t.Errorf("cell() = %q, want %q", got, `a \| b<br>c`)
	}
}

func TestRequirementFallsBackWhenTheContractDoesNotNameTheID(t *testing.T) {
	t.Parallel()
	contract := report.Contract{Criteria: []report.Criterion{
		{ID: "C1", Requirement: "Addresses collapse", Required: true},
	}}
	if got := requirement(contract, "C1"); got != "Addresses collapse" {
		t.Errorf("requirement() = %q, want %q", got, "Addresses collapse")
	}
	if got := requirement(contract, "C9"); got != "(not in contract)" {
		t.Errorf("requirement() = %q, want %q", got, "(not in contract)")
	}
}

func TestRequirementMarksWhatTheContractAsksOfACriterion(t *testing.T) {
	t.Parallel()
	contract := report.Contract{Criteria: []report.Criterion{
		{ID: "C1", Requirement: "Addresses collapse", Required: true, Provisional: true},
		{ID: "C2", Requirement: "Throughput holds"},
		{ID: "C3", Requirement: "Latency holds", Provisional: true},
	}}
	cases := map[string]string{
		"C1": "Addresses collapse (provisional)",
		"C2": "Throughput holds (optional)",
		// Both markers, in the order the contract is read in: what it asks
		// for, whether it must hold, and whether anyone has confirmed it.
		"C3": "Latency holds (optional) (provisional)",
	}
	for id, want := range cases {
		t.Run(id, func(t *testing.T) {
			t.Parallel()
			if got := requirement(contract, id); got != want {
				t.Errorf("requirement(%s) = %q, want %q", id, got, want)
			}
		})
	}
}

func TestProvenanceNamesAnUnrecordedToolRatherThanADanglingVia(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		p    report.EvidenceProvenance
		want string
	}{
		{"no tool", report.EvidenceProvenance{}, "via (not recorded)"},
		{"tool only", report.EvidenceProvenance{Tool: "pytest"}, "via pytest"},
		{"tool and version", report.EvidenceProvenance{Tool: "pytest", Version: "7.4"}, "via pytest 7.4"},
		{"tool, version, and timestamp",
			report.EvidenceProvenance{Tool: "pytest", Version: "7.4", Timestamp: "2026-09-14T00:00:00Z"},
			"via pytest 7.4 at 2026-09-14T00:00:00Z"},
		{"tool and timestamp, no version",
			report.EvidenceProvenance{Tool: "pytest", Timestamp: "2026-09-14T00:00:00Z"},
			"via pytest at 2026-09-14T00:00:00Z"},
		{"timestamp only",
			report.EvidenceProvenance{Timestamp: "2026-09-14T00:00:00Z"},
			"via (not recorded) at 2026-09-14T00:00:00Z"},
		{"revision is not shown: it names a state, not an act",
			report.EvidenceProvenance{Tool: "pytest", Version: "7.4", Revision: "abc123"},
			"via pytest 7.4"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := provenance(testCase.p); got != testCase.want {
				t.Errorf("provenance(%+v) = %q, want %q", testCase.p, got, testCase.want)
			}
		})
	}
}

func TestIsJudgmentAndOrNotRecorded(t *testing.T) {
	t.Parallel()
	if !isJudgment(report.Evidence{Kind: report.KindJudgment}) {
		t.Error("judgment evidence must be recognised")
	}
	if isJudgment(report.Evidence{Kind: report.KindInspection}) {
		t.Error("inspection evidence is not a judgment")
	}
	if got := orNotRecorded(""); got != "(not recorded)" {
		t.Errorf("orNotRecorded(\"\") = %q, want %q", got, "(not recorded)")
	}
	if got := orNotRecorded("unknown"); got != "unknown" {
		t.Errorf("orNotRecorded(%q) = %q, want %q: a recorded unknown is not an absent value", "unknown", got, "unknown")
	}
	if got := orNotRecorded("gpt-9"); got != "gpt-9" {
		t.Errorf("orNotRecorded() = %q, want %q", got, "gpt-9")
	}
}
