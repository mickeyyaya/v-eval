package report

import "sort"

// Rule strings identify which status-derivation rule produced a Status, so
// the outcome is auditable rather than a bare enum value.
const (
	RuleAdvisory       = "advisory: acceptance not requested; per-criterion results stand"
	RuleFailed         = "required applicable criterion failed"
	RuleProvisional    = "provisional required criterion prevents acceptance"
	RuleConflict       = "unresolved contract conflict"
	RuleUnknownOrError = "required applicable criterion unknown or error"
	RuleNoGating       = "no required applicable criteria"
	RuleAllPass        = "all required applicable criteria pass"
)

// Derive computes the report-level Status from a contract and its criterion
// results, applying seven rules in order. The first rule that hits decides
// the outcome; each rule's own function reports which ids triggered it.
func Derive(contract Contract, results []CriterionResult, advisory bool) Status {
	g := gatingSet(contract, results)

	if ids, hit := ruleAdvisory(g, advisory); hit {
		return newStatus(OverallAdvisory, RuleAdvisory, ids, advisory)
	}
	if ids, hit := ruleFailed(g); hit {
		return newStatus(OverallFail, RuleFailed, ids, advisory)
	}
	if ids, hit := ruleProvisional(contract, results); hit {
		return newStatus(OverallIncomplete, RuleProvisional, ids, advisory)
	}
	if ids, hit := ruleConflict(contract); hit {
		return newStatus(OverallIncomplete, RuleConflict, ids, advisory)
	}
	if ids, hit := ruleUnknownOrError(g); hit {
		return newStatus(OverallIncomplete, RuleUnknownOrError, ids, advisory)
	}
	if ids, hit := ruleNoGating(g); hit {
		return newStatus(OverallIncomplete, RuleNoGating, ids, advisory)
	}
	return newStatus(OverallPass, RuleAllPass, nil, advisory)
}

// newStatus builds a Status, ensuring BlockedBy is always non-nil so
// canonical JSON emits [] rather than null.
func newStatus(overall Overall, rule string, ids []string, advisory bool) Status {
	if ids == nil {
		ids = []string{}
	}
	return Status{Overall: overall, RuleApplied: rule, BlockedBy: ids, Advisory: advisory}
}

// gatingSet is the subset of results whose contract criterion is required,
// not provisional, and whose result is not NOT_APPLICABLE. A result whose id
// is not in the contract is never in the gating set.
func gatingSet(contract Contract, results []CriterionResult) []CriterionResult {
	var g []CriterionResult
	for _, result := range results {
		criterion, ok := contract.CriterionByID(result.ID)
		if !ok || !criterion.Required || criterion.Provisional || result.Result == ResultNotApplicable {
			continue
		}
		g = append(g, result)
	}
	return g
}

// idsMatching returns the sorted ids of gating-set results whose Result
// satisfies match.
func idsMatching(g []CriterionResult, match func(Result) bool) []string {
	var ids []string
	for _, result := range g {
		if match(result.Result) {
			ids = append(ids, result.ID)
		}
	}
	sort.Strings(ids)
	return ids
}

// failedIDs returns the sorted ids of FAIL results in the gating set.
func failedIDs(g []CriterionResult) []string {
	return idsMatching(g, func(r Result) bool { return r == ResultFail })
}

// ruleAdvisory hits whenever advisory is requested, regardless of results.
func ruleAdvisory(g []CriterionResult, advisory bool) (ids []string, hit bool) {
	if !advisory {
		return nil, false
	}
	return failedIDs(g), true
}

// ruleFailed hits when any gating result is FAIL.
func ruleFailed(g []CriterionResult) (ids []string, hit bool) {
	ids = failedIDs(g)
	return ids, len(ids) > 0
}

// ruleProvisional hits when a required, applicable criterion is provisional,
// regardless of whether it is in the gating set.
func ruleProvisional(contract Contract, results []CriterionResult) (ids []string, hit bool) {
	for _, result := range results {
		criterion, ok := contract.CriterionByID(result.ID)
		if !ok || !criterion.Required || !criterion.Provisional || result.Result == ResultNotApplicable {
			continue
		}
		ids = append(ids, result.ID)
	}
	sort.Strings(ids)
	return ids, len(ids) > 0
}

// ruleConflict hits when the contract has an unresolved conflict.
func ruleConflict(contract Contract) (ids []string, hit bool) {
	for _, conflict := range contract.Conflicts {
		if conflict.Resolution == "unresolved" {
			return nil, true
		}
	}
	return nil, false
}

// ruleUnknownOrError hits when any gating result is UNKNOWN or ERROR.
func ruleUnknownOrError(g []CriterionResult) (ids []string, hit bool) {
	ids = idsMatching(g, func(r Result) bool { return r == ResultUnknown || r == ResultError })
	return ids, len(ids) > 0
}

// ruleNoGating hits when the gating set is empty.
func ruleNoGating(g []CriterionResult) (ids []string, hit bool) {
	return nil, len(g) == 0
}

// TallyCounts computes per-bucket result counts and assessment coverage for
// a contract's criteria against the results reached for them. A result whose
// id is not in the contract counts as optional; validation reports the
// broken link separately, so TallyCounts must not panic on it.
func TallyCounts(contract Contract, results []CriterionResult) Counts {
	var counts Counts
	for _, result := range results {
		bucket := &counts.Optional
		if criterion, ok := contract.CriterionByID(result.ID); ok && criterion.Required {
			bucket = &counts.Required
		}
		bump(bucket, result.Result)
	}

	numerator := counts.Required.Pass + counts.Required.Fail + counts.Optional.Pass + counts.Optional.Fail
	denominator := counts.Required.Applicable + counts.Optional.Applicable
	if denominator == 0 {
		counts.Coverage = Coverage{Undefined: true}
	} else {
		counts.Coverage = Coverage{Numerator: numerator, Denominator: denominator}
	}

	return counts
}

// bump increments the tally bucket for one result: NOT_APPLICABLE is counted
// only in its own field, every other result also counts as applicable.
func bump(t *Tally, result Result) {
	if result == ResultNotApplicable {
		t.NotApplicable++
		return
	}
	t.Applicable++
	switch result {
	case ResultPass:
		t.Pass++
	case ResultFail:
		t.Fail++
	case ResultUnknown:
		t.Unknown++
	case ResultError:
		t.Error++
	}
}
