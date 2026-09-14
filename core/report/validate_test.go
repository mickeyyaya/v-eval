package report

import (
	"go/ast"
	"go/parser"
	"go/token"
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

// ruleCase is one rule id and a way to break it. mutate changes the decoded
// fixture; rewrite changes the encoded bytes, which is the only way to break
// the one rule that is about the bytes rather than about the report.
type ruleCase struct {
	rule    string
	mutate  func(rep *Report)
	rewrite func(raw []byte) []byte
}

// violationsFor validates the worked example with one case's break applied.
func violationsFor(t *testing.T, testCase ruleCase) []Violation {
	t.Helper()
	rep := loadFixture(t)
	if testCase.mutate != nil {
		testCase.mutate(&rep)
	}
	raw, err := Encode(rep)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if testCase.rewrite != nil {
		raw = testCase.rewrite(raw)
	}
	_, violations, err := Validate(raw)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	return violations
}

// ruleIDsInSource returns every rule id the package declares, read back out of
// validate.go itself. A rule added there without a firing case below must fail
// the test rather than pass unnoticed, and no hand-kept list can promise that.
func ruleIDsInSource(t *testing.T) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "validate.go", nil, 0)
	if err != nil {
		t.Fatalf("parse validate.go: %v", err)
	}
	var ids []string
	for _, decl := range file.Decls {
		general, ok := decl.(*ast.GenDecl)
		if !ok || general.Tok != token.CONST {
			continue
		}
		ids = append(ids, ruleIDsInConst(general)...)
	}
	if len(ids) == 0 {
		t.Fatal("no rule ids found in validate.go; this test can no longer find what it guards")
	}
	return ids
}

// ruleIDsInConst returns the string values of every Rule-prefixed constant in
// one const declaration.
func ruleIDsInConst(general *ast.GenDecl) []string {
	var ids []string
	for _, spec := range general.Specs {
		value, ok := spec.(*ast.ValueSpec)
		if !ok || len(value.Names) == 0 || len(value.Values) == 0 {
			continue
		}
		if !strings.HasPrefix(value.Names[0].Name, "Rule") {
			continue
		}
		if literal, ok := value.Values[0].(*ast.BasicLit); ok {
			ids = append(ids, strings.Trim(literal.Value, `"`))
		}
	}
	return ids
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

// TestValidateErrorIsPerCriterion pins the rule to the criterion it judges. A
// command the report ran for something else says nothing about why this
// criterion could not be decided, so an ERROR must rest on evidence of its
// own: execution evidence, or a locator naming a command that exited non-zero.
func TestValidateErrorIsPerCriterion(t *testing.T) {
	t.Parallel()
	exit := 127
	commandEvidence := Evidence{
		Kind:        KindInspection,
		Locator:     Locator{Command: "docker build .", Cwd: "/w", ExitStatus: &exit, LogRef: "logs/1.txt"},
		Observation: "the build could not run: docker is not installed",
		Isolation:   IsolationNone,
		Origin:      OriginObserved,
	}
	executionEvidence := commandEvidence
	executionEvidence.Kind = KindExecution
	executionEvidence.Locator = Locator{Note: "the runner died before it wrote a log"}

	cases := []struct {
		name     string
		evidence []Evidence
		commands []CommandRecord
		wantFire bool
	}{
		{"a command run elsewhere in the report is not this criterion's", nil,
			[]CommandRecord{{Command: "docker build .", Cwd: "/w", ExitStatus: exit,
				StartedAt: "2026-09-14T00:00:00Z", EndedAt: "2026-09-14T00:00:01Z",
				LogRef: "logs/1.txt", Isolation: IsolationNone}}, true},
		{"no evidence at all", nil, nil, true},
		{"a command locator that exited non-zero", []Evidence{commandEvidence}, nil, false},
		{"execution evidence of its own", []Evidence{executionEvidence}, nil, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rep := loadFixture(t)
			rep.Criteria[4].Result = ResultError
			rep.Criteria[4].Evidence = testCase.evidence
			rep.Provenance.Commands = testCase.commands
			raw, err := Encode(Aggregate(rep))
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			_, violations, err := Validate(raw)
			if err != nil {
				t.Fatalf("validate: %v", err)
			}
			if got := hasRule(violations, RuleErrorIsOperational); got != testCase.wantFire {
				t.Fatalf("rule fired = %v, want %v: %v", got, testCase.wantFire, violations)
			}
		})
	}
}

// TestValidateFiresEveryRule breaks one rule at a time, so each rule id is
// backed by a report that actually violates it. The ids come from validate.go
// itself, so the claim this test makes -- that every rule fires for some
// report -- is checked rather than asserted.
func TestValidateFiresEveryRule(t *testing.T) {
	t.Parallel()
	cases := []ruleCase{
		// json is about the bytes rather than the report, so it is the one
		// case that breaks the encoding instead of the value encoded.
		{RuleJSON, nil, func(raw []byte) []byte { return append(raw, []byte("{\"junk\": 1}\n")...) }},
		{RuleSchemaVersionSupported, func(rep *Report) { rep.Identity.SchemaVersion = "9.9.9" }, nil},
		{RuleRequiredNonempty, func(rep *Report) { rep.Routing.Rationale = "  " }, nil},
		{RuleEnumValid, func(rep *Report) { rep.Criteria[0].Result = Result("MAYBE") }, nil},
		{RuleTimeRFC3339, func(rep *Report) { rep.Identity.CreatedAt = "yesterday" }, nil},
		{RuleLocatorShape, func(rep *Report) { rep.Criteria[0].Evidence[0].Locator = Locator{File: "a.go"} }, nil},
		{RuleCriteriaContractLink, func(rep *Report) { rep.Criteria[0].ID = "C9" }, nil},
		{RuleCandidateSuppliedKind, func(rep *Report) { rep.Criteria[4].Evidence[0].Kind = KindInspection }, nil},
		{RuleJudgmentMetadata, func(rep *Report) { rep.Criteria[0].Evidence[0].Kind = KindJudgment }, nil},
		{RuleNotApplicableReasoning, func(rep *Report) {
			rep.Criteria[0].Result, rep.Criteria[0].Reasoning = ResultNotApplicable, ""
		}, nil},
		{RuleClaimsVerificationPresent, func(rep *Report) { rep.Claims[0].Status = ClaimStatusVerified }, nil},
		{RuleForensicsCriterionLink, func(rep *Report) {
			rep.Forensics = append(rep.Forensics, Finding{FindingID: "F1", CriterionID: "C9",
				Severity: SeverityObserved, Disposition: DispositionOpen})
		}, nil},
		{RuleForensicsConfirmedSeverity, func(rep *Report) {
			rep.Forensics = append(rep.Forensics, Finding{FindingID: "F1", CriterionID: "C1",
				Severity: SeveritySuspicious, Disposition: DispositionConfirmed})
		}, nil},
		{RuleForensicsConfirmedImpliesFail, func(rep *Report) {
			rep.Forensics = append(rep.Forensics, Finding{FindingID: "F1", CriterionID: "C4",
				Severity: SeverityConfirmed, Disposition: DispositionConfirmed})
		}, nil},
		{RuleDimensionsNoComposite, func(rep *Report) {
			rep.Dimensions = append(rep.Dimensions, Dimension{Dimension: "quality", Metric: "Overall",
				AuthorityType: AuthorityProjectRubric, DefinitionRef: "docs/rubric.md"})
		}, nil},
		{RuleCriteriaNonempty, func(rep *Report) { rep.Criteria = nil }, nil},
		{RuleRangeValid, func(rep *Report) { rep.Routing.Supplied[0].Count = -1 }, nil},
		{RulePassRequiresObservedLocator, func(rep *Report) {
			rep.Criteria[3].Evidence = []Evidence{{Kind: KindSupplied, Origin: OriginCandidateSupplied,
				Locator:     Locator{Note: "the author says the tests pass"},
				Observation: "tests passed", Isolation: IsolationNone}}
		}, nil},
		{RuleErrorIsOperational, func(rep *Report) { rep.Criteria[4].Result = ResultError }, nil},
		{RuleCountsMatch, func(rep *Report) { rep.Counts.Required.Pass = 99 }, nil},
		{RuleStatusMatch, func(rep *Report) { rep.Status.RuleApplied = "because the author said so" }, nil},
		// The fixture is already aggregated, so changing what a piece of
		// evidence says leaves the digest it states behind, as altering the
		// evidence after the fact would.
		{RuleEvidenceDigest, func(rep *Report) {
			rep.Criteria[0].Evidence[0].Observation = "something else was seen after all"
		}, nil},
		{RuleReportID, func(rep *Report) {
			rep.Identity.ReportID = "sha256:" + strings.Repeat("0", 64)
		}, nil},
	}

	covered := map[string]bool{}
	for _, testCase := range cases {
		covered[testCase.rule] = true
		t.Run(testCase.rule, func(t *testing.T) {
			t.Parallel()
			if violations := violationsFor(t, testCase); !hasRule(violations, testCase.rule) {
				t.Fatalf("rule %s did not fire: %v", testCase.rule, violations)
			}
		})
	}
	for _, id := range ruleIDsInSource(t) {
		if !covered[id] {
			t.Errorf("rule %q is declared but no case here breaks it", id)
		}
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

// validate encodes a mutated fixture and validates it, or fails the test.
func validate(t *testing.T, rep Report) []Violation {
	t.Helper()
	raw, err := Encode(Aggregate(rep))
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	_, violations, err := Validate(raw)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	return violations
}

// TestValidateRequiresSomethingToJudge covers the two lists a report says
// nothing without: a contract with no criteria states no requirement, and a
// report with no results reaches no verdict about the requirements it states.
func TestValidateRequiresSomethingToJudge(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		mutate func(rep *Report)
		path   string
	}{
		"a contract with no criteria": {func(rep *Report) { rep.Contract.Criteria = nil }, "contract.criteria"},
		"a report with no results":    {func(rep *Report) { rep.Criteria = nil }, "criteria"},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rep := loadFixture(t)
			testCase.mutate(&rep)
			violations := validate(t, rep)
			if got := pathOf(violations, RuleCriteriaNonempty); got != testCase.path {
				t.Fatalf("path = %q, want %q: %v", got, testCase.path, violations)
			}
		})
	}
}

// TestValidateHoldsNumbersToTheirRange covers the two numbers whose range is
// part of what they mean: a similarity is a fraction of one, and a count of
// the inputs an evaluator received cannot be negative.
func TestValidateHoldsNumbersToTheirRange(t *testing.T) {
	t.Parallel()
	precedents := func(similarity float64) func(rep *Report) {
		return func(rep *Report) {
			rep.Learning = &Learning{PrecedentsRetrieved: []PrecedentRef{
				{PrecedentID: "sha256:abc", Criterion: "C1", Similarity: similarity}}}
		}
	}
	cases := map[string]struct {
		mutate func(rep *Report)
		path   string
	}{
		"a negative count":        {func(rep *Report) { rep.Routing.Supplied[0].Count = -1 }, "routing.supplied[0].count"},
		"a similarity above one":  {precedents(1.5), "learning.precedents_retrieved[0].similarity"},
		"a similarity below zero": {precedents(-0.1), "learning.precedents_retrieved[0].similarity"},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rep := loadFixture(t)
			testCase.mutate(&rep)
			violations := validate(t, rep)
			if got := pathOf(violations, RuleRangeValid); got != testCase.path {
				t.Fatalf("path = %q, want %q: %v", got, testCase.path, violations)
			}
		})
	}
}

// TestValidateHoldsASimilarityAtTheBounds keeps the check inclusive: 0 and 1
// are similarities a report may state, and a rule that rejected them would
// reject the two most definite answers it can give.
func TestValidateHoldsASimilarityAtTheBounds(t *testing.T) {
	t.Parallel()
	for _, similarity := range []float64{0, 1} {
		rep := loadFixture(t)
		rep.Learning = &Learning{PrecedentsRetrieved: []PrecedentRef{
			{PrecedentID: "sha256:abc", Criterion: "C1", Similarity: similarity}}}
		if violations := validate(t, rep); hasRule(violations, RuleRangeValid) {
			t.Fatalf("similarity %v is in range: %v", similarity, violations)
		}
	}
}

// TestValidateRequiresADimensionToNameItsDefinition: a dimension states a
// measured quantity, and a reader cannot check one against a definition the
// report never points at.
func TestValidateRequiresADimensionToNameItsDefinition(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Dimensions = []Dimension{{Dimension: "latency", Metric: "p99", AuthorityType: AuthorityProjectRubric}}
	violations := validate(t, rep)
	if got := pathOf(violations, RuleRequiredNonempty); got != "dimensions[0].definition_ref" {
		t.Fatalf("path = %q, want %q: %v", got, "dimensions[0].definition_ref", violations)
	}
}
