package report

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/mickeyyaya/v-eval/schema"
)

// Violation is one rule broken at one place in a report.
type Violation struct {
	Path    string // where it was broken, e.g. "criteria[3].evidence"
	Rule    string // which rule, e.g. "evidence.pass_requires_observed_locator"
	Message string // what is wrong, in one sentence
}

// Rule ids are stable: a caller can act on a specific rule without reading
// the message, and a report's violations mean the same thing across versions.
const (
	RuleJSON                          = "json"
	RuleSchemaVersionSupported        = "schema_version.supported"
	RuleRequiredNonempty              = "required.nonempty"
	RuleEnumValid                     = "enum.valid"
	RuleTimeRFC3339                   = "time.rfc3339"
	RuleLocatorShape                  = "locator.shape"
	RuleCriteriaNonempty              = "criteria.nonempty"
	RuleRangeValid                    = "range.valid"
	RuleCriteriaContractLink          = "criteria.contract_link"
	RulePassRequiresObservedLocator   = "evidence.pass_requires_observed_locator"
	RuleCandidateSuppliedKind         = "evidence.candidate_supplied_kind"
	RuleJudgmentMetadata              = "evidence.judgment_metadata"
	RuleNotApplicableReasoning        = "criteria.not_applicable_reasoning"
	RuleErrorIsOperational            = "criteria.error_is_operational"
	RuleClaimsVerificationPresent     = "claims.verification_present"
	RuleForensicsCriterionLink        = "forensics.criterion_link"
	RuleForensicsConfirmedSeverity    = "forensics.confirmed_severity"
	RuleForensicsConfirmedImpliesFail = "forensics.confirmed_implies_fail"
	RuleDimensionsNoComposite         = "dimensions.no_composite"
	RuleCountsMatch                   = "counts.match"
	RuleStatusMatch                   = "status.match"
	RuleEvidenceDigest                = "provenance.evidence_digest"
	RuleReportID                      = "identity.report_id"
)

// rule checks one property of a decoded report and reports what it finds.
type rule func(rep Report) []Violation

// structuralRules hold for any report, whether or not its derived fields have
// been filled in.
func structuralRules() []rule {
	return []rule{
		ruleSchemaVersion, ruleRequiredNonempty, ruleEnumValid, ruleTimeRFC3339,
		ruleLocatorShape, ruleCriteriaNonempty, ruleRangeValid, ruleCriteriaContractLink,
		rulePassRequiresObservedLocator, ruleCandidateSuppliedKind, ruleJudgmentMetadata,
		ruleNotApplicableReasoning, ruleErrorIsOperational, ruleClaimsVerificationPresent,
		ruleForensicsCriterionLink, ruleForensicsConfirmedSeverity,
		ruleForensicsConfirmedImpliesFail, ruleDimensionsNoComposite,
	}
}

// derivedRules compare what a report states about itself against a fresh
// recomputation. They are the rules Aggregate satisfies, and so the rules a
// report on its way into Aggregate cannot yet be held to.
func derivedRules() []rule {
	return []rule{ruleCountsMatch, ruleStatusMatch, ruleEvidenceDigest, ruleReportID}
}

// Validate decodes raw and applies every rule. The error is reserved for
// internal failures: a report that is wrong is reported as violations.
func Validate(raw []byte) (Report, []Violation, error) {
	return runRules(raw, append(structuralRules(), derivedRules()...))
}

// ValidateForAggregate applies the structural rules only, so a report can be
// checked before Aggregate computes its counts, status, and digests.
func ValidateForAggregate(raw []byte) (Report, []Violation, error) {
	return runRules(raw, structuralRules())
}

// runRules decodes raw and collects every rule's violations in rule order. A
// report that does not decode yields the single json violation and nothing
// else: no later rule has a report to inspect.
func runRules(raw []byte, rules []rule) (Report, []Violation, error) {
	rep, err := Decode(raw)
	if err != nil {
		return Report{}, []Violation{{Path: "$", Rule: RuleJSON, Message: err.Error()}}, nil
	}
	var violations []Violation
	for _, check := range rules {
		violations = append(violations, check(rep)...)
	}
	return rep, violations, nil
}

// ruleSchemaVersion requires the report to declare the schema this build reads.
func ruleSchemaVersion(rep Report) []Violation {
	if rep.Identity.SchemaVersion == schema.Version {
		return nil
	}
	return []Violation{{Path: "identity.schema_version", Rule: RuleSchemaVersionSupported,
		Message: fmt.Sprintf("schema version %q is not supported; this build reads %q",
			rep.Identity.SchemaVersion, schema.Version)}}
}

// requiredField is one string that must carry content, and where it lives.
type requiredField struct {
	path  string
	value string
}

// nonEmpty reports a violation for every field that is empty or all blanks.
func nonEmpty(fields []requiredField) []Violation {
	var violations []Violation
	for _, field := range fields {
		if strings.TrimSpace(field.value) == "" {
			violations = append(violations, Violation{Path: field.path, Rule: RuleRequiredNonempty,
				Message: "must not be empty"})
		}
	}
	return violations
}

// ruleRequiredNonempty checks the fields a report says nothing without.
// status.rule_applied is deliberately absent: status.match compares the whole
// derived status, which is stricter, and is the rule that owns those fields.
// Whether a list has any entries at all is criteria.nonempty's business, not
// this rule's: this one is about a field that is present and says nothing.
func ruleRequiredNonempty(rep Report) []Violation {
	violations := nonEmpty(headerFields(rep))
	violations = append(violations, nonEmpty(contractFields(rep.Contract))...)
	violations = append(violations, nonEmpty(elementFields(rep))...)
	for i, criterion := range rep.Contract.Criteria {
		if len(criterion.MethodsAllowed) == 0 {
			violations = append(violations, Violation{
				Path:    fmt.Sprintf("contract.criteria[%d].methods_allowed", i),
				Rule:    RuleRequiredNonempty,
				Message: "must allow at least one method"})
		}
	}
	for _, ref := range WalkEvidence(rep) {
		if strings.TrimSpace(ref.Evidence.Observation) == "" {
			violations = append(violations, Violation{Path: ref.Path + ".observation",
				Rule: RuleRequiredNonempty, Message: "must not be empty"})
		}
	}
	return violations
}

// headerFields are the report-level strings that must carry content.
func headerFields(rep Report) []requiredField {
	return []requiredField{
		{"identity.veval_version", rep.Identity.VevalVersion},
		{"identity.skill_revision", rep.Identity.SkillRevision},
		{"identity.artifact.kind", rep.Identity.Artifact.Kind},
		{"identity.artifact.revision", rep.Identity.Artifact.Revision},
		{"identity.task.requested_outcome", rep.Identity.Task.RequestedOutcome},
		{"contract.contract_id", rep.Contract.ContractID},
		{"contract.contract_version", rep.Contract.ContractVersion},
		{"routing.rationale", rep.Routing.Rationale},
	}
}

// contractFields are the strings each contract criterion must carry.
func contractFields(contract Contract) []requiredField {
	var fields []requiredField
	for i, criterion := range contract.Criteria {
		base := fmt.Sprintf("contract.criteria[%d]", i)
		fields = append(fields,
			requiredField{base + ".id", criterion.ID},
			requiredField{base + ".requirement", criterion.Requirement},
			requiredField{base + ".acceptance_rule", criterion.AcceptanceRule})
	}
	return fields
}

// elementFields are the identifying strings each list element must carry.
func elementFields(rep Report) []requiredField {
	var fields []requiredField
	for i, result := range rep.Criteria {
		fields = append(fields, requiredField{fmt.Sprintf("criteria[%d].id", i), result.ID})
	}
	for i, observation := range rep.Observations {
		base := fmt.Sprintf("observations[%d]", i)
		fields = append(fields,
			requiredField{base + ".id", observation.ID},
			requiredField{base + ".text", observation.Text})
	}
	for i, claim := range rep.Claims {
		base := fmt.Sprintf("claims[%d]", i)
		fields = append(fields,
			requiredField{base + ".claim_id", claim.ClaimID},
			requiredField{base + ".text", claim.Text})
	}
	for i, finding := range rep.Forensics {
		fields = append(fields, requiredField{fmt.Sprintf("forensics[%d].finding_id", i), finding.FindingID})
	}
	for i, dimension := range rep.Dimensions {
		fields = append(fields, requiredField{
			fmt.Sprintf("dimensions[%d].definition_ref", i), dimension.DefinitionRef})
	}
	return fields
}

// enumField is one enum-typed field and where it lives.
type enumField struct {
	path  string
	value interface{ IsValid() bool }
}

// ruleEnumValid checks every enum-typed field against its defined values, and
// requires each declared conflict to state a resolution. status.overall is
// deliberately absent, for the reason given on ruleRequiredNonempty.
func ruleEnumValid(rep Report) []Violation {
	fields := contractEnums(rep.Contract)
	fields = append(fields, elementEnums(rep)...)
	fields = append(fields, evidenceEnums(rep)...)
	fields = append(fields, provenanceEnums(rep.Provenance)...)

	var violations []Violation
	for _, field := range fields {
		if !field.value.IsValid() {
			violations = append(violations, Violation{Path: field.path, Rule: RuleEnumValid,
				Message: fmt.Sprintf("%q is not a defined value", field.value)})
		}
	}
	for i, conflict := range rep.Contract.Conflicts {
		if strings.TrimSpace(conflict.Resolution) == "" {
			violations = append(violations, Violation{
				Path:    fmt.Sprintf("contract.conflicts[%d].resolution", i),
				Rule:    RuleEnumValid,
				Message: "must state how the conflict was resolved"})
		}
	}
	return violations
}

// contractEnums are the enum-typed fields of a contract.
func contractEnums(contract Contract) []enumField {
	fields := []enumField{{"contract.status", contract.Status}}
	for i, criterion := range contract.Criteria {
		for j, method := range criterion.MethodsAllowed {
			fields = append(fields, enumField{
				fmt.Sprintf("contract.criteria[%d].methods_allowed[%d]", i, j), method})
		}
	}
	return fields
}

// elementEnums are the enum-typed fields of a report's list elements.
func elementEnums(rep Report) []enumField {
	var fields []enumField
	for i, result := range rep.Criteria {
		base := fmt.Sprintf("criteria[%d]", i)
		fields = append(fields,
			enumField{base + ".result", result.Result},
			enumField{base + ".method_used", result.MethodUsed})
	}
	for i, observation := range rep.Observations {
		fields = append(fields, enumField{fmt.Sprintf("observations[%d].origin", i), observation.Origin})
	}
	for i, claim := range rep.Claims {
		fields = append(fields, enumField{fmt.Sprintf("claims[%d].status", i), claim.Status})
	}
	for i, finding := range rep.Forensics {
		base := fmt.Sprintf("forensics[%d]", i)
		fields = append(fields,
			enumField{base + ".severity", finding.Severity},
			enumField{base + ".disposition", finding.Disposition})
	}
	for i, dimension := range rep.Dimensions {
		fields = append(fields, enumField{fmt.Sprintf("dimensions[%d].authority_type", i), dimension.AuthorityType})
	}
	return fields
}

// evidenceEnums are the enum-typed fields of every evidence entry.
func evidenceEnums(rep Report) []enumField {
	var fields []enumField
	for _, ref := range WalkEvidence(rep) {
		fields = append(fields,
			enumField{ref.Path + ".kind", ref.Evidence.Kind},
			enumField{ref.Path + ".isolation", ref.Evidence.Isolation},
			enumField{ref.Path + ".origin", ref.Evidence.Origin})
	}
	return fields
}

// provenanceEnums are the enum-typed fields of a report's provenance.
func provenanceEnums(provenance Provenance) []enumField {
	var fields []enumField
	for i, isolation := range provenance.IsolationLevelsUsed {
		fields = append(fields, enumField{fmt.Sprintf("provenance.isolation_levels_used[%d]", i), isolation})
	}
	for i, command := range provenance.Commands {
		fields = append(fields, enumField{fmt.Sprintf("provenance.commands[%d].isolation", i), command.Isolation})
	}
	return fields
}

// ruleTimeRFC3339 requires the creation time to be machine-readable.
func ruleTimeRFC3339(rep Report) []Violation {
	if _, err := time.Parse(time.RFC3339, rep.Identity.CreatedAt); err == nil {
		return nil
	}
	return []Violation{{Path: "identity.created_at", Rule: RuleTimeRFC3339,
		Message: fmt.Sprintf("%q is not an RFC 3339 timestamp", rep.Identity.CreatedAt)}}
}

// ruleLocatorShape requires every locator to populate exactly one complete
// field group, so a reader can always go back to what the evaluator saw.
func ruleLocatorShape(rep Report) []Violation {
	var violations []Violation
	for _, ref := range WalkEvidence(rep) {
		if ref.Evidence.Locator.Shape() == ShapeNone {
			violations = append(violations, Violation{Path: ref.Path + ".locator", Rule: RuleLocatorShape,
				Message: "must fill exactly one complete group: file, command, passage, or note"})
		}
	}
	return violations
}

// ruleCriteriaNonempty requires a report to judge something. A contract with
// no criteria states no requirement to be judged against, and a report with no
// results reaches no verdict about the requirements it does state; either one
// would still tally and derive a status, which is how an empty evaluation
// comes to look like a finished one.
func ruleCriteriaNonempty(rep Report) []Violation {
	var violations []Violation
	if len(rep.Contract.Criteria) == 0 {
		violations = append(violations, Violation{Path: "contract.criteria", Rule: RuleCriteriaNonempty,
			Message: "a contract must state at least one criterion"})
	}
	if len(rep.Criteria) == 0 {
		violations = append(violations, Violation{Path: "criteria", Rule: RuleCriteriaNonempty,
			Message: "a report must reach a result for at least one criterion"})
	}
	return violations
}

// ruleRangeValid holds the two numbers whose range is part of what they mean:
// a similarity is a fraction of one, and a count of the inputs an evaluator
// received cannot be negative.
func ruleRangeValid(rep Report) []Violation {
	var violations []Violation
	for i, supplied := range rep.Routing.Supplied {
		if supplied.Count < 0 {
			violations = append(violations, Violation{
				Path:    fmt.Sprintf("routing.supplied[%d].count", i),
				Rule:    RuleRangeValid,
				Message: fmt.Sprintf("%d is not a count: it must be zero or more", supplied.Count)})
		}
	}
	if rep.Learning == nil {
		return violations
	}
	for i, precedent := range rep.Learning.PrecedentsRetrieved {
		if precedent.Similarity < 0 || precedent.Similarity > 1 {
			violations = append(violations, Violation{
				Path: fmt.Sprintf("learning.precedents_retrieved[%d].similarity", i),
				Rule: RuleRangeValid,
				Message: fmt.Sprintf("%v is not a similarity: it must be between 0 and 1 inclusive",
					precedent.Similarity)})
		}
	}
	return violations
}

// ruleCriteriaContractLink requires the results and the contract to name the
// same criteria: every result links to a contract criterion, and every
// contract criterion has exactly one result.
func ruleCriteriaContractLink(rep Report) []Violation {
	var violations []Violation
	seen := map[string]int{}
	for i, result := range rep.Criteria {
		base := fmt.Sprintf("criteria[%d]", i)
		if _, ok := rep.Contract.CriterionByID(result.ID); !ok {
			violations = append(violations, Violation{Path: base, Rule: RuleCriteriaContractLink,
				Message: fmt.Sprintf("result for %q is not a criterion of this contract", result.ID)})
		}
		seen[result.ID]++
		if seen[result.ID] > 1 {
			violations = append(violations, Violation{Path: base, Rule: RuleCriteriaContractLink,
				Message: fmt.Sprintf("criterion %q already has a result", result.ID)})
		}
	}
	for i, criterion := range rep.Contract.Criteria {
		if seen[criterion.ID] == 0 {
			violations = append(violations, Violation{
				Path:    fmt.Sprintf("contract.criteria[%d]", i),
				Rule:    RuleCriteriaContractLink,
				Message: fmt.Sprintf("criterion %q has no result", criterion.ID)})
		}
	}
	return violations
}

// rulePassRequiresObservedLocator requires a PASS to rest on something the
// evaluator opened or ran, not on a claim that it did.
func rulePassRequiresObservedLocator(rep Report) []Violation {
	var violations []Violation
	for i, result := range rep.Criteria {
		if result.Result != ResultPass || slices.ContainsFunc(result.Evidence, Evidence.SupportsPass) {
			continue
		}
		violations = append(violations, Violation{
			Path:    fmt.Sprintf("criteria[%d].evidence", i),
			Rule:    RulePassRequiresObservedLocator,
			Message: "PASS needs observed evidence with a file, command, or passage locator"})
	}
	return violations
}

// ruleCandidateSuppliedKind keeps candidate-supplied evidence labelled as
// supplied, so it can never be counted as something the evaluator observed.
func ruleCandidateSuppliedKind(rep Report) []Violation {
	var violations []Violation
	for _, ref := range WalkEvidence(rep) {
		if ref.Evidence.Origin != OriginCandidateSupplied || ref.Evidence.Kind == KindSupplied {
			continue
		}
		violations = append(violations, Violation{Path: ref.Path + ".kind", Rule: RuleCandidateSuppliedKind,
			Message: fmt.Sprintf("candidate_supplied evidence must be of kind supplied, not %q", ref.Evidence.Kind)})
	}
	return violations
}

// ruleJudgmentMetadata requires a judgment to name the rubric it applied.
func ruleJudgmentMetadata(rep Report) []Violation {
	var violations []Violation
	for _, ref := range WalkEvidence(rep) {
		if ref.Evidence.Kind != KindJudgment || strings.TrimSpace(ref.Evidence.RubricVersion) != "" {
			continue
		}
		violations = append(violations, Violation{Path: ref.Path + ".rubric_version",
			Rule: RuleJudgmentMetadata, Message: "judgment evidence must name the rubric version it applied"})
	}
	return violations
}

// ruleNotApplicableReasoning requires a dismissed criterion to say why.
func ruleNotApplicableReasoning(rep Report) []Violation {
	var violations []Violation
	for i, result := range rep.Criteria {
		if result.Result != ResultNotApplicable || strings.TrimSpace(result.Reasoning) != "" {
			continue
		}
		violations = append(violations, Violation{
			Path:    fmt.Sprintf("criteria[%d].reasoning", i),
			Rule:    RuleNotApplicableReasoning,
			Message: "NOT_APPLICABLE must state why the criterion does not apply"})
	}
	return violations
}

// ruleErrorIsOperational requires each ERROR to rest on something that
// actually ran under that criterion: execution evidence of its own, or
// evidence pointing at a command that exited non-zero. The report's own list
// of commands is not enough -- a command run for another criterion says
// nothing about why this one could not be decided -- and an ERROR with
// neither is a judgment in disguise.
func ruleErrorIsOperational(rep Report) []Violation {
	var violations []Violation
	for i, result := range rep.Criteria {
		if result.Result != ResultError || slices.ContainsFunc(result.Evidence, Evidence.SupportsError) {
			continue
		}
		violations = append(violations, Violation{
			Path:    fmt.Sprintf("criteria[%d]", i),
			Rule:    RuleErrorIsOperational,
			Message: "ERROR needs its own execution evidence, or evidence of a command that exited non-zero"})
	}
	return violations
}

// ruleClaimsVerificationPresent requires a settled claim to show its working.
func ruleClaimsVerificationPresent(rep Report) []Violation {
	var violations []Violation
	for i, claim := range rep.Claims {
		settled := claim.Status == ClaimStatusVerified || claim.Status == ClaimStatusContradicted
		if !settled || len(claim.Verification) > 0 {
			continue
		}
		violations = append(violations, Violation{
			Path:    fmt.Sprintf("claims[%d].verification", i),
			Rule:    RuleClaimsVerificationPresent,
			Message: fmt.Sprintf("a %q claim must cite the evidence that settled it", claim.Status)})
	}
	return violations
}

// ruleForensicsCriterionLink ties every finding to a criterion of the
// contract, so a detector hit can always be traced to what it bears on.
func ruleForensicsCriterionLink(rep Report) []Violation {
	var violations []Violation
	for i, finding := range rep.Forensics {
		if _, ok := rep.Contract.CriterionByID(finding.CriterionID); ok {
			continue
		}
		violations = append(violations, Violation{
			Path:    fmt.Sprintf("forensics[%d].criterion_id", i),
			Rule:    RuleForensicsCriterionLink,
			Message: fmt.Sprintf("%q is not a criterion of this contract", finding.CriterionID)})
	}
	return violations
}

// ruleForensicsConfirmedSeverity keeps a confirmed disposition backed by
// confirmed severity: a finding cannot be settled as real while its own
// severity still says it might not be. It is a separate rule id from the
// contract link because the two say different things about a finding and a
// caller acts on them differently.
func ruleForensicsConfirmedSeverity(rep Report) []Violation {
	var violations []Violation
	for i, finding := range rep.Forensics {
		if finding.Disposition != DispositionConfirmed || finding.Severity == SeverityConfirmed {
			continue
		}
		violations = append(violations, Violation{
			Path:    fmt.Sprintf("forensics[%d]", i),
			Rule:    RuleForensicsConfirmedSeverity,
			Message: fmt.Sprintf("a confirmed disposition needs confirmed severity, not %q", finding.Severity)})
	}
	return violations
}

// ruleForensicsConfirmedImpliesFail requires a confirmed finding against a
// required criterion to be reflected in that criterion's result, so a report
// cannot confirm a problem and pass the criterion it belongs to.
func ruleForensicsConfirmedImpliesFail(rep Report) []Violation {
	var violations []Violation
	for i, finding := range rep.Forensics {
		criterion, known := rep.Contract.CriterionByID(finding.CriterionID)
		if finding.Disposition != DispositionConfirmed || !known || !criterion.Required {
			continue
		}
		if result, ok := ResultByID(rep.Criteria, finding.CriterionID); ok && result.Result == ResultFail {
			continue
		}
		violations = append(violations, Violation{
			Path: fmt.Sprintf("forensics[%d]", i),
			Rule: RuleForensicsConfirmedImpliesFail,
			Message: fmt.Sprintf("confirmed finding against required criterion %q, whose result is not FAIL",
				finding.CriterionID)})
	}
	return violations
}

// ResultByID returns the result recorded for a criterion id, and whether
// there is one. A criterion with no result is a real state, so the caller is
// told rather than handed a zero value that reads as a verdict.
func ResultByID(results []CriterionResult, id string) (CriterionResult, bool) {
	for _, result := range results {
		if result.ID == id {
			return result, true
		}
	}
	return CriterionResult{}, false
}

// compositeMetrics are the metric names that hide several measurements behind
// one number, which a report must not do.
var compositeMetrics = map[string]bool{"composite": true, "overall": true, "total": true, "score": true}

// ruleDimensionsNoComposite rejects a dimension that reports a blended score
// instead of a measured quantity.
func ruleDimensionsNoComposite(rep Report) []Violation {
	var violations []Violation
	for i, dimension := range rep.Dimensions {
		if !compositeMetrics[strings.ToLower(strings.TrimSpace(dimension.Metric))] {
			continue
		}
		violations = append(violations, Violation{
			Path:    fmt.Sprintf("dimensions[%d].metric", i),
			Rule:    RuleDimensionsNoComposite,
			Message: fmt.Sprintf("%q is a composite, not a measured quantity", dimension.Metric)})
	}
	return violations
}

// ruleCountsMatch requires the stated counts to be the counts the results
// give. The message names only the bucket.field pairs that differ, in words,
// rather than dumping both structs: a report can be wrong about one tally
// field as often as about all fifteen.
func ruleCountsMatch(rep Report) []Violation {
	want := TallyCounts(rep.Contract, rep.Criteria)
	if rep.Counts == want {
		return nil
	}
	return []Violation{{Path: "counts", Rule: RuleCountsMatch,
		Message: strings.Join(countsFieldDiffs(rep.Counts, want), "; ")}}
}

// countsFieldDiffs names every bucket.field where stated counts differ from
// computed, e.g. "required.fail stated 2, computed 3".
func countsFieldDiffs(stated, computed Counts) []string {
	var diffs []string
	diffs = append(diffs, tallyFieldDiffs("required", stated.Required, computed.Required)...)
	diffs = append(diffs, tallyFieldDiffs("optional", stated.Optional, computed.Optional)...)
	diffs = append(diffs, coverageFieldDiffs(stated.Coverage, computed.Coverage)...)
	return diffs
}

// tallyFieldDiffs names every field where two tallies in the same bucket
// differ, prefixed with the bucket name so "required.fail" and
// "optional.fail" cannot be confused.
func tallyFieldDiffs(bucket string, stated, computed Tally) []string {
	fields := []struct {
		name             string
		stated, computed int
	}{
		{"applicable", stated.Applicable, computed.Applicable},
		{"pass", stated.Pass, computed.Pass},
		{"fail", stated.Fail, computed.Fail},
		{"unknown", stated.Unknown, computed.Unknown},
		{"error", stated.Error, computed.Error},
		{"not_applicable", stated.NotApplicable, computed.NotApplicable},
	}
	var diffs []string
	for _, field := range fields {
		if field.stated != field.computed {
			diffs = append(diffs, fmt.Sprintf("%s.%s stated %d, computed %d",
				bucket, field.name, field.stated, field.computed))
		}
	}
	return diffs
}

// coverageFieldDiffs names every field where two coverage values differ.
func coverageFieldDiffs(stated, computed Coverage) []string {
	var diffs []string
	if stated.Numerator != computed.Numerator {
		diffs = append(diffs, fmt.Sprintf("coverage.numerator stated %d, computed %d",
			stated.Numerator, computed.Numerator))
	}
	if stated.Denominator != computed.Denominator {
		diffs = append(diffs, fmt.Sprintf("coverage.denominator stated %d, computed %d",
			stated.Denominator, computed.Denominator))
	}
	if stated.Undefined != computed.Undefined {
		diffs = append(diffs, fmt.Sprintf("coverage.undefined stated %t, computed %t",
			stated.Undefined, computed.Undefined))
	}
	return diffs
}

// ruleStatusMatch requires the stated verdict to be the verdict the rules
// give. The message names only the fields that differ, in words: a status
// differs from its recomputation by one field as often as by all four.
func ruleStatusMatch(rep Report) []Violation {
	want := Derive(rep.Contract, rep.Criteria, rep.Status.Advisory)
	if reflect.DeepEqual(rep.Status, want) {
		return nil
	}
	return []Violation{{Path: "status", Rule: RuleStatusMatch,
		Message: strings.Join(statusFieldDiffs(rep.Status, want), "; ")}}
}

// statusFieldDiffs names every field where a stated status differs from its
// recomputation, e.g. `stated overall PASS, computed FAIL`. BlockedBy uses
// reflect.DeepEqual, matching the nil-vs-empty sensitivity of the equality
// check above it, so a difference this function finds is never invisible
// there and vice versa.
func statusFieldDiffs(stated, computed Status) []string {
	var diffs []string
	if stated.Overall != computed.Overall {
		diffs = append(diffs, fmt.Sprintf("stated overall %s, computed %s", stated.Overall, computed.Overall))
	}
	if stated.RuleApplied != computed.RuleApplied {
		diffs = append(diffs, fmt.Sprintf("stated rule %q, computed %q", stated.RuleApplied, computed.RuleApplied))
	}
	if !reflect.DeepEqual(stated.BlockedBy, computed.BlockedBy) {
		diffs = append(diffs, fmt.Sprintf("stated blocked_by %v, computed %v", stated.BlockedBy, computed.BlockedBy))
	}
	if stated.Advisory != computed.Advisory {
		diffs = append(diffs, fmt.Sprintf("stated advisory %t, computed %t", stated.Advisory, computed.Advisory))
	}
	return diffs
}

// ruleEvidenceDigest requires the stated evidence digest to be the digest of
// the evidence the report carries.
func ruleEvidenceDigest(rep Report) []Violation {
	want := EvidenceDigest(rep)
	if rep.Provenance.EvidenceDigest == want {
		return nil
	}
	return []Violation{{Path: "provenance.evidence_digest", Rule: RuleEvidenceDigest,
		Message: fmt.Sprintf("stated %q, recomputed %q", rep.Provenance.EvidenceDigest, want)}}
}

// ruleReportID requires the stated report id to be the digest of the report.
func ruleReportID(rep Report) []Violation {
	want := ReportID(rep)
	if rep.Identity.ReportID == want {
		return nil
	}
	return []Violation{{Path: "identity.report_id", Rule: RuleReportID,
		Message: fmt.Sprintf("stated %q, recomputed %q", rep.Identity.ReportID, want)}}
}

// EvidenceRef is one evidence entry together with the path it lives at, e.g.
// "criteria[3].evidence[0]" — the path a violation about it cites.
type EvidenceRef struct {
	Path     string
	Evidence Evidence
}

// WalkEvidence lists every evidence entry in a report, in document order:
// criteria, observations, claims, forensics, dimensions. Every rule that
// looks at evidence reads this walk, so a rule cannot quietly disagree with
// another about where evidence lives or what it is called.
func WalkEvidence(rep Report) []EvidenceRef {
	var refs []EvidenceRef
	for _, section := range evidenceSections(rep) {
		for i, evidence := range section.Evidence {
			refs = append(refs, EvidenceRef{
				Path:     fmt.Sprintf("%s[%d]", section.Path, i),
				Evidence: evidence,
			})
		}
	}
	return refs
}

// evidenceSection is one evidence array, the report section that owns it, and
// the path it lives at. It is the single statement of where evidence sits:
// WalkEvidence flattens it for the rules, and EvidenceDigest hashes it whole,
// so the digest and the violation paths can never drift apart. The json tags
// are part of the digest's input and so part of its definition.
type evidenceSection struct {
	Section  string     `json:"section"`
	Path     string     `json:"path"`
	Evidence []Evidence `json:"evidence"`
}

// evidenceSections lists a report's evidence arrays in document order.
func evidenceSections(rep Report) []evidenceSection {
	sections := make([]evidenceSection, 0,
		len(rep.Criteria)+len(rep.Observations)+len(rep.Claims)+len(rep.Forensics)+len(rep.Dimensions))
	for i, result := range rep.Criteria {
		sections = append(sections, evidenceSection{"criteria", fmt.Sprintf("criteria[%d].evidence", i), result.Evidence})
	}
	for i, observation := range rep.Observations {
		sections = append(sections, evidenceSection{"observations", fmt.Sprintf("observations[%d].evidence", i), observation.Evidence})
	}
	for i, claim := range rep.Claims {
		sections = append(sections, evidenceSection{"claims", fmt.Sprintf("claims[%d].verification", i), claim.Verification})
	}
	for i, finding := range rep.Forensics {
		sections = append(sections, evidenceSection{"forensics", fmt.Sprintf("forensics[%d].evidence", i), finding.Evidence})
	}
	for i, dimension := range rep.Dimensions {
		sections = append(sections, evidenceSection{"dimensions", fmt.Sprintf("dimensions[%d].evidence", i), dimension.Evidence})
	}
	return sections
}
