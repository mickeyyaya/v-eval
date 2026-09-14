package report

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
