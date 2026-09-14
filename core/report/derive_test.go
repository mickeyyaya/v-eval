package report

import (
	"reflect"
	"testing"
)

// req is a required contract criterion, provisional or not.
func req(id string, provisional bool) Criterion {
	return Criterion{ID: id, Required: true, Provisional: provisional}
}

// opt is an optional contract criterion.
func opt(id string) Criterion { return Criterion{ID: id} }

// res is the result reached for one criterion id.
func res(id string, r Result) CriterionResult { return CriterionResult{ID: id, Result: r} }

func TestDerive(t *testing.T) {
	t.Parallel()
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

// TestDeriveEdgeCases pins the three routings that are easy to get wrong: a
// provisional criterion that also failed, an advisory report with nothing
// failing, and a result the contract does not mention. Unlike TestDerive
// these assert the rule applied, because the point of each case is which
// rule claims the report, not only what the verdict is.
func TestDeriveEdgeCases(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		contract  Contract
		results   []CriterionResult
		advisory  bool
		want      Overall
		rule      string
		blockedBy []string
	}{
		{
			name:      "fail on provisional required goes to the provisional rule",
			contract:  Contract{Criteria: []Criterion{req("A", true)}},
			results:   []CriterionResult{res("A", ResultFail)},
			want:      OverallIncomplete,
			rule:      StatusRuleProvisional,
			blockedBy: []string{"A"},
		},
		{
			name:      "advisory with no fail blocks on nothing",
			contract:  Contract{Criteria: []Criterion{req("A", false)}},
			results:   []CriterionResult{res("A", ResultPass)},
			advisory:  true,
			want:      OverallAdvisory,
			rule:      StatusRuleAdvisory,
			blockedBy: []string{},
		},
		{
			name:      "result outside the contract does not gate",
			contract:  Contract{Criteria: []Criterion{req("R1", false)}},
			results:   []CriterionResult{res("R1", ResultPass), res("ZZ", ResultFail)},
			want:      OverallPass,
			rule:      StatusRuleAllPass,
			blockedBy: []string{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := Derive(c.contract, c.results, c.advisory)
			if got.Overall != c.want || got.RuleApplied != c.rule || !reflect.DeepEqual(got.BlockedBy, c.blockedBy) {
				t.Errorf("got %+v, want %s via %q blocked_by %v", got, c.want, c.rule, c.blockedBy)
			}
		})
	}
}
