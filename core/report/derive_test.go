package report

import (
	"reflect"
	"testing"
)

func TestDerive(t *testing.T) {
	t.Parallel()
	req := func(id string, provisional bool) Criterion {
		return Criterion{ID: id, Required: true, Provisional: provisional}
	}
	opt := func(id string) Criterion { return Criterion{ID: id} }
	res := func(id string, r Result) CriterionResult { return CriterionResult{ID: id, Result: r} }
	cases := []struct {
		name      string
		contract  Contract
		results   []CriterionResult
		advisory  bool
		want      Overall
		blockedBy []string
	}{
		{"fail dominates", Contract{Criteria: []Criterion{req("A", false), req("B", false)}}, []CriterionResult{res("A", ResultFail), res("B", ResultUnknown)}, false, OverallFail, []string{"A"}},
		{"unknown incomplete", Contract{Criteria: []Criterion{req("A", false)}}, []CriterionResult{res("A", ResultUnknown)}, false, OverallIncomplete, []string{"A"}},
		{"error incomplete", Contract{Criteria: []Criterion{req("A", false)}}, []CriterionResult{res("A", ResultError)}, false, OverallIncomplete, []string{"A"}},
		{"all pass", Contract{Criteria: []Criterion{req("A", false)}}, []CriterionResult{res("A", ResultPass)}, false, OverallPass, []string{}},
		{"provisional blocks", Contract{Criteria: []Criterion{req("A", true)}}, []CriterionResult{res("A", ResultPass)}, false, OverallIncomplete, []string{"A"}},
		{"unresolved conflict", Contract{Conflicts: []Conflict{{Resolution: "unresolved"}}, Criteria: []Criterion{req("A", false)}}, []CriterionResult{res("A", ResultPass)}, false, OverallIncomplete, []string{}},
		{"zero gating", Contract{Criteria: []Criterion{opt("A")}}, []CriterionResult{res("A", ResultPass)}, false, OverallIncomplete, []string{}},
		{"optional fail ignored", Contract{Criteria: []Criterion{req("A", false), opt("B")}}, []CriterionResult{res("A", ResultPass), res("B", ResultFail)}, false, OverallPass, []string{}},
		{"not applicable excluded", Contract{Criteria: []Criterion{req("A", false), req("B", false)}}, []CriterionResult{res("A", ResultPass), res("B", ResultNotApplicable)}, false, OverallPass, []string{}},
		{"advisory", Contract{Criteria: []Criterion{req("A", false)}}, []CriterionResult{res("A", ResultFail)}, true, OverallAdvisory, []string{"A"}},
		{"blocked_by sorted", Contract{Criteria: []Criterion{req("B", false), req("A", false)}}, []CriterionResult{res("B", ResultFail), res("A", ResultFail)}, false, OverallFail, []string{"A", "B"}},
	}
	for _, c := range cases {
		got := Derive(c.contract, c.results, c.advisory)
		if got.Overall != c.want || !reflect.DeepEqual(got.BlockedBy, c.blockedBy) || got.RuleApplied == "" || got.Advisory != c.advisory {
			t.Errorf("%s: got %+v, want %s blocked_by %v", c.name, got, c.want, c.blockedBy)
		}
	}
}
