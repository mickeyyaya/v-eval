package report

import (
	"os"
	"strings"
	"testing"
)

// hasRule reports whether any violation cites the given rule id.
func hasRule(vs []Violation, rule string) bool {
	for _, v := range vs {
		if v.Rule == rule {
			return true
		}
	}
	return false
}

// pathOf returns the path of the first violation citing the given rule id.
func pathOf(vs []Violation, rule string) string {
	for _, v := range vs {
		if v.Rule == rule {
			return v.Path
		}
	}
	return ""
}

// violationsFor validates the worked example with one mutation applied.
func violationsFor(t *testing.T, mutate func(rep *Report)) []Violation {
	t.Helper()
	rep := loadFixture(t)
	mutate(&rep)
	raw, err := Encode(rep)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	_, violations, err := Validate(raw)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	return violations
}

func TestValidateAcceptsFixture(t *testing.T) {
	t.Parallel()
	raw, _ := os.ReadFile("testdata/worked-example.json")
	_, violations, err := Validate(raw)
	if err != nil || len(violations) != 0 {
		t.Fatalf("err=%v violations=%v", err, violations)
	}
}

func TestValidateReportsMalformedJSONAsViolation(t *testing.T) {
	t.Parallel()
	_, violations, err := Validate([]byte("{not json"))
	if err != nil || len(violations) != 1 || violations[0].Rule != "json" || violations[0].Path != "$" {
		t.Fatalf("err=%v violations=%v", err, violations)
	}
}

func TestValidateRejectsPassWithoutObservedLocator(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Criteria[3].Evidence = []Evidence{{Kind: KindSupplied, Origin: OriginCandidateSupplied, Locator: Locator{Note: "author says so"}, Observation: "tests passed", Isolation: IsolationNone}}
	rep = Aggregate(rep)
	raw, _ := Encode(rep)
	_, violations, _ := Validate(raw)
	if !hasRule(violations, "evidence.pass_requires_observed_locator") {
		t.Fatalf("violations = %v", violations)
	}
	if got := pathOf(violations, "evidence.pass_requires_observed_locator"); got != "criteria[3].evidence" {
		t.Fatalf("path = %q, want %q", got, "criteria[3].evidence")
	}
}

func TestValidateCatchesInventedStatusAndBrokenLinks(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Status.Overall = OverallPass
	raw, _ := Encode(rep)
	_, violations, _ := Validate(raw)
	if !hasRule(violations, "status.match") || !hasRule(violations, "identity.report_id") {
		t.Fatalf("violations = %v", violations)
	}
	rep = loadFixture(t)
	rep.Criteria[0].ID = "C9"
	raw, _ = Encode(Aggregate(rep))
	_, violations, _ = Validate(raw)
	if !hasRule(violations, "criteria.contract_link") {
		t.Fatalf("violations = %v", violations)
	}
}

func TestValidateForAggregateSkipsDerivedFields(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Counts, rep.Status, rep.Identity.ReportID, rep.Provenance.EvidenceDigest = Counts{}, Status{}, "", ""
	raw, _ := Encode(rep)
	_, violations, err := ValidateForAggregate(raw)
	if err != nil || len(violations) != 0 {
		t.Fatalf("err=%v violations=%v", err, violations)
	}
}

func TestValidateErrorRequiresAttemptedCheck(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Criteria[4].Result = ResultError
	raw, _ := Encode(Aggregate(rep))
	_, violations, _ := Validate(raw)
	if !hasRule(violations, "criteria.error_is_operational") {
		t.Fatalf("violations = %v", violations)
	}
}

// TestValidateFiresEveryRule breaks one rule at a time, so each rule id is
// backed by a report that actually violates it.
func TestValidateFiresEveryRule(t *testing.T) {
	t.Parallel()
	cases := []struct {
		rule   string
		mutate func(rep *Report)
	}{
		{RuleSchemaVersionSupported, func(rep *Report) { rep.Identity.SchemaVersion = "9.9.9" }},
		{RuleRequiredNonempty, func(rep *Report) { rep.Routing.Rationale = "  " }},
		{RuleEnumValid, func(rep *Report) { rep.Criteria[0].Result = Result("MAYBE") }},
		{RuleTimeRFC3339, func(rep *Report) { rep.Identity.CreatedAt = "yesterday" }},
		{RuleLocatorShape, func(rep *Report) { rep.Criteria[0].Evidence[0].Locator = Locator{File: "a.go"} }},
		{RuleCriteriaContractLink, func(rep *Report) { rep.Criteria[0].ID = "C9" }},
		{RuleCandidateSuppliedKind, func(rep *Report) { rep.Criteria[4].Evidence[0].Kind = KindInspection }},
		{RuleJudgmentMetadata, func(rep *Report) { rep.Criteria[0].Evidence[0].Kind = KindJudgment }},
		{RuleNotApplicableReasoning, func(rep *Report) {
			rep.Criteria[0].Result, rep.Criteria[0].Reasoning = ResultNotApplicable, ""
		}},
		{RuleClaimsVerificationPresent, func(rep *Report) { rep.Claims[0].Status = ClaimStatusVerified }},
		{RuleForensicsCriterionLink, func(rep *Report) {
			rep.Forensics = append(rep.Forensics, Finding{FindingID: "F1", CriterionID: "C9",
				Severity: SeverityObserved, Disposition: DispositionOpen})
		}},
		{RuleForensicsConfirmedSeverity, func(rep *Report) {
			rep.Forensics = append(rep.Forensics, Finding{FindingID: "F1", CriterionID: "C1",
				Severity: SeveritySuspicious, Disposition: DispositionConfirmed})
		}},
		{RuleForensicsConfirmedImpliesFail, func(rep *Report) {
			rep.Forensics = append(rep.Forensics, Finding{FindingID: "F1", CriterionID: "C4",
				Severity: SeverityConfirmed, Disposition: DispositionConfirmed})
		}},
		{RuleDimensionsNoComposite, func(rep *Report) {
			rep.Dimensions = append(rep.Dimensions, Dimension{Dimension: "quality", Metric: "Overall",
				AuthorityType: AuthorityProjectRubric, DefinitionRef: "docs/rubric.md"})
		}},
		{RuleCountsMatch, func(rep *Report) { rep.Counts.Required.Pass = 99 }},
		{RuleStatusMatch, func(rep *Report) { rep.Status.RuleApplied = "because the author said so" }},
	}

	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			t.Parallel()
			if violations := violationsFor(t, tc.mutate); !hasRule(violations, tc.rule) {
				t.Fatalf("rule %s did not fire: %v", tc.rule, violations)
			}
		})
	}
}

// TestValidateForAggregateStillChecksStructure guards the split: dropping the
// four derived rules must not drop the structural ones with them.
func TestValidateForAggregateStillChecksStructure(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Counts, rep.Status = Counts{}, Status{}
	rep.Criteria[0].ID = "C9"
	raw, _ := Encode(rep)
	_, violations, err := ValidateForAggregate(raw)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if !hasRule(violations, RuleCriteriaContractLink) {
		t.Fatalf("violations = %v", violations)
	}
	if hasRule(violations, RuleCountsMatch) || hasRule(violations, RuleStatusMatch) {
		t.Fatalf("derived rules ran: %v", violations)
	}
}

// TestAggregateRecomputesCountsAndStatus covers the minimal Aggregate this
// task adds; Task 8 extends it with the two digests.
func TestAggregateRecomputesCountsAndStatus(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Counts, rep.Status = Counts{}, Status{}
	got := Aggregate(rep)
	if want := TallyCounts(rep.Contract, rep.Criteria); got.Counts != want {
		t.Fatalf("counts = %+v, want %+v", got.Counts, want)
	}
	if got.Status.Overall != OverallFail || got.Status.RuleApplied != StatusRuleFailed {
		t.Fatalf("status = %+v", got.Status)
	}
}

// TestStatusMismatchMessageNamesDifferingFieldsInWords guards the
// status.match message: it must name which fields of the stated status
// disagree with the recomputed one, in words, rather than dump both structs
// in Go syntax. Only overall and rule_applied are mutated, so blocked_by and
// advisory -- unchanged and matching -- must not appear in the message.
func TestStatusMismatchMessageNamesDifferingFieldsInWords(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Status.Overall = OverallPass
	rep.Status.RuleApplied = "made up"
	violations := ruleStatusMatch(rep)
	if len(violations) != 1 {
		t.Fatalf("violations = %v, want exactly one", violations)
	}
	want := `stated overall PASS, computed FAIL; stated rule "made up", computed "required applicable criterion failed"`
	if got := violations[0].Message; got != want {
		t.Errorf("message = %q, want %q", got, want)
	}
}

// TestCountsMismatchMessageNamesDifferingFieldsInWords guards the
// counts.match message: it must name the bucket and field that disagree, in
// words, rather than dump both structs in Go syntax. Only required.fail and
// coverage.numerator are mutated, so every other field -- unchanged and
// matching -- must not appear in the message.
func TestCountsMismatchMessageNamesDifferingFieldsInWords(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Counts.Required.Fail = 2
	rep.Counts.Coverage.Numerator = 5
	violations := ruleCountsMatch(rep)
	if len(violations) != 1 {
		t.Fatalf("violations = %v, want exactly one", violations)
	}
	want := "required.fail stated 2, computed 3; coverage.numerator stated 5, computed 4"
	if got := violations[0].Message; got != want {
		t.Errorf("message = %q, want %q", got, want)
	}
}

// TestWalkEvidenceIsOrderedAndUnique pins the exported walk, which both the
// validation rules and the evidence digest read: sections come in document
// order, and every entry has a path of its own, because a violation and a
// digest entry each name one.
func TestWalkEvidenceIsOrderedAndUnique(t *testing.T) {
	t.Parallel()
	_, rep := readExtendedFixture(t)
	refs := WalkEvidence(rep)
	if len(refs) == 0 {
		t.Fatal("the extended fixture cites no evidence")
	}

	sections := []string{"criteria[", "observations[", "claims[", "forensics[", "dimensions["}
	seen := map[string]bool{}
	at := 0
	for _, ref := range refs {
		for at < len(sections) && !strings.HasPrefix(ref.Path, sections[at]) {
			at++
		}
		if at == len(sections) {
			t.Fatalf("path %q is out of document order", ref.Path)
		}
		if seen[ref.Path] {
			t.Fatalf("path %q appears twice", ref.Path)
		}
		seen[ref.Path] = true
	}
	if at != len(sections)-1 {
		t.Fatalf("the walk reached section %d of %d; the fixture cites evidence in every one", at+1, len(sections))
	}
}

// TestResultByIDFindsWhatTheContractNames covers the exported lookup the
// forensics rules use: a known id returns its result, an unknown one does not
// return a zero value that could pass for a verdict.
func TestResultByIDFindsWhatTheContractNames(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	result, ok := ResultByID(rep.Criteria, "C4")
	if !ok || result.Result != ResultPass {
		t.Fatalf("C4 = %+v, %v", result, ok)
	}
	if _, ok := ResultByID(rep.Criteria, "C9"); ok {
		t.Fatal("C9 is not in the report")
	}
}
