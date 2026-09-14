package report

import "testing"

func TestTallyWorkedExample(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	got := TallyCounts(rep.Contract, rep.Criteria)
	want := Counts{
		Required: Tally{Applicable: 5, Pass: 1, Fail: 3, Unknown: 1},
		Coverage: Coverage{Numerator: 4, Denominator: 5},
	}
	if got != want {
		t.Fatalf("Tally = %+v, want %+v", got, want)
	}
}

func TestTallySeparatesOptionalAndExcludesNotApplicable(t *testing.T) {
	t.Parallel()
	contract := Contract{Criteria: []Criterion{{ID: "R1", Required: true}, {ID: "R2", Required: true}, {ID: "O1"}}}
	results := []CriterionResult{{ID: "R1", Result: ResultPass}, {ID: "R2", Result: ResultNotApplicable}, {ID: "O1", Result: ResultError}}
	got := TallyCounts(contract, results)
	if got.Required != (Tally{Applicable: 1, Pass: 1, NotApplicable: 1}) || got.Optional != (Tally{Applicable: 1, Error: 1}) {
		t.Fatalf("got %+v", got)
	}
	if got.Coverage != (Coverage{Numerator: 1, Denominator: 2}) {
		t.Fatalf("coverage = %+v", got.Coverage)
	}
}

func TestCoverageUndefinedWithZeroApplicable(t *testing.T) {
	t.Parallel()
	contract := Contract{Criteria: []Criterion{{ID: "C1", Required: true}}}
	got := TallyCounts(contract, []CriterionResult{{ID: "C1", Result: ResultNotApplicable}})
	if !got.Coverage.Undefined || got.Coverage.Denominator != 0 || got.Required.NotApplicable != 1 {
		t.Fatalf("coverage = %+v", got.Coverage)
	}
}
