package report

import (
	"bytes"
	"os"
	"testing"
)

// extendedFixture is the second worked report. Where worked-example.json is a
// plain inspection of a code change, this one carries the shapes that example
// has no room for: commands that ran, a cited passage, an optional criterion,
// a dismissed one, an operational error, an unmeasured dimension, an open
// forensic finding, an unresolved contract conflict, and a learning record.
const extendedFixture = "testdata/extended-example.json"

// readExtendedFixture returns the extended fixture's bytes and its report.
func readExtendedFixture(t *testing.T) ([]byte, Report) {
	t.Helper()
	raw, err := os.ReadFile(extendedFixture)
	if err != nil {
		t.Fatal(err)
	}
	rep, err := Decode(raw)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return raw, rep
}

func TestExtendedFixtureIsCanonicalAggregatedAndValid(t *testing.T) {
	t.Parallel()
	raw, rep := readExtendedFixture(t)

	encoded, err := Encode(rep)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, raw) {
		t.Fatal("fixture is not in canonical form; write Encode output back to testdata")
	}
	if aggregated, err := Encode(Aggregate(rep)); err != nil {
		t.Fatal(err)
	} else if !bytes.Equal(aggregated, raw) {
		t.Fatal("fixture is not aggregated; write Aggregate output back to testdata")
	}

	_, violations, err := Validate(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, violation := range violations {
		t.Errorf("violation at %s (%s): %s", violation.Path, violation.Rule, violation.Message)
	}
}

func TestExtendedFixtureIsIncompleteByUnresolvedConflict(t *testing.T) {
	t.Parallel()
	_, rep := readExtendedFixture(t)

	if rep.Status.Overall != OverallIncomplete {
		t.Errorf("overall = %q, want %q", rep.Status.Overall, OverallIncomplete)
	}
	if rep.Status.RuleApplied != StatusRuleConflict {
		t.Errorf("rule_applied = %q, want %q", rep.Status.RuleApplied, StatusRuleConflict)
	}
	unresolved := 0
	for _, conflict := range rep.Contract.Conflicts {
		if conflict.Resolution == "unresolved" {
			unresolved++
		}
	}
	if unresolved != 1 {
		t.Errorf("unresolved conflicts = %d, want 1", unresolved)
	}
}

func TestExtendedFixtureCoversWhatTheWorkedExampleLacks(t *testing.T) {
	t.Parallel()
	_, rep := readExtendedFixture(t)

	shapes := map[LocatorShape]int{}
	kinds := map[Kind]int{}
	for _, ref := range walkEvidence(rep) {
		shapes[ref.evidence.Locator.Shape()]++
		kinds[ref.evidence.Kind]++
	}
	for _, shape := range []LocatorShape{ShapeFile, ShapeCommand, ShapePassage, ShapeNote} {
		if shapes[shape] == 0 {
			t.Errorf("no %s locator in the fixture", shape)
		}
	}
	if kinds[KindExecution] == 0 {
		t.Error("no execution evidence in the fixture")
	}
	if !hasCommandAt(rep, IsolationWorktree) {
		t.Error("no provenance command ran under worktree isolation")
	}
	checkExtendedResults(t, rep)
	checkExtendedSections(t, rep)
}

// hasCommandAt reports whether some recorded command ran at an isolation level.
func hasCommandAt(rep Report, isolation Isolation) bool {
	for _, command := range rep.Provenance.Commands {
		if command.Isolation == isolation {
			return true
		}
	}
	return false
}

// checkExtendedResults requires the criterion results the worked example has
// no instance of: an optional criterion, a dismissal, and an error.
func checkExtendedResults(t *testing.T, rep Report) {
	t.Helper()
	optional := 0
	for _, criterion := range rep.Contract.Criteria {
		if !criterion.Required {
			optional++
		}
	}
	if optional != 1 {
		t.Errorf("optional criteria = %d, want 1", optional)
	}

	dismissed, errored := 0, 0
	for _, result := range rep.Criteria {
		switch result.Result {
		case ResultNotApplicable:
			dismissed++
			if result.Reasoning == "" {
				t.Errorf("criterion %s is NOT_APPLICABLE without reasoning", result.ID)
			}
		case ResultError:
			errored++
		}
	}
	if dismissed != 1 || errored != 1 {
		t.Errorf("not_applicable = %d, error = %d, want 1 and 1", dismissed, errored)
	}
}

// checkExtendedSections requires the sections the worked example leaves empty:
// an unmeasured dimension, an open finding on a required criterion, learning.
func checkExtendedSections(t *testing.T, rep Report) {
	t.Helper()
	unmeasured := 0
	for _, dimension := range rep.Dimensions {
		if dimension.Value == nil {
			unmeasured++
		}
	}
	if unmeasured != 1 {
		t.Errorf("dimensions with a null value = %d, want 1", unmeasured)
	}

	if len(rep.Forensics) != 1 {
		t.Fatalf("forensics = %d findings, want 1", len(rep.Forensics))
	}
	finding := rep.Forensics[0]
	if finding.Severity != SeveritySuspicious || finding.Disposition != DispositionOpen {
		t.Errorf("finding is %q/%q, want suspicious/open", finding.Severity, finding.Disposition)
	}
	if criterion, ok := rep.Contract.CriterionByID(finding.CriterionID); !ok || !criterion.Required {
		t.Errorf("finding %s is not on a required criterion", finding.FindingID)
	}
	if rep.Learning == nil || len(rep.Learning.PrecedentsRetrieved) == 0 {
		t.Error("fixture carries no learning block")
	}
}
