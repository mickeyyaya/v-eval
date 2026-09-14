// Package report defines the v-eval report: its types, validation, and aggregation.
package report

// Result is the verdict of a single check.
type Result string

const (
	ResultPass          Result = "PASS"
	ResultFail          Result = "FAIL"
	ResultUnknown       Result = "UNKNOWN"
	ResultError         Result = "ERROR"
	ResultNotApplicable Result = "NOT_APPLICABLE"
)

// IsValid reports whether r is one of the defined Result values.
func (r Result) IsValid() bool {
	switch r {
	case ResultPass, ResultFail, ResultUnknown, ResultError, ResultNotApplicable:
		return true
	}
	return false
}

// Overall is the verdict for the report as a whole.
type Overall string

const (
	OverallPass       Overall = "PASS"
	OverallFail       Overall = "FAIL"
	OverallIncomplete Overall = "INCOMPLETE"
	OverallAdvisory   Overall = "ADVISORY"
)

// IsValid reports whether o is one of the defined Overall values.
func (o Overall) IsValid() bool {
	switch o {
	case OverallPass, OverallFail, OverallIncomplete, OverallAdvisory:
		return true
	}
	return false
}

// Method is how a check was carried out.
type Method string

const (
	MethodExecution          Method = "execution"
	MethodDeterministicCheck Method = "deterministic_check"
	MethodStaticInspection   Method = "static_inspection"
	MethodSourceVerification Method = "source_verification"
	MethodRubricJudgment     Method = "rubric_judgment"
	MethodHumanJudgment      Method = "human_judgment"
)

// IsValid reports whether m is one of the defined Method values.
func (m Method) IsValid() bool {
	switch m {
	case MethodExecution, MethodDeterministicCheck, MethodStaticInspection,
		MethodSourceVerification, MethodRubricJudgment, MethodHumanJudgment:
		return true
	}
	return false
}

// Kind classifies the nature of a check.
type Kind string

const (
	KindExecution  Kind = "execution"
	KindInspection Kind = "inspection"
	KindSupplied   Kind = "supplied"
	KindJudgment   Kind = "judgment"
)

// IsValid reports whether k is one of the defined Kind values.
func (k Kind) IsValid() bool {
	switch k {
	case KindExecution, KindInspection, KindSupplied, KindJudgment:
		return true
	}
	return false
}

// Origin describes where a piece of evidence came from.
type Origin string

const (
	OriginObserved          Origin = "observed"
	OriginCandidateSupplied Origin = "candidate_supplied"
	OriginRetrieved         Origin = "retrieved"
)

// IsValid reports whether o is one of the defined Origin values.
func (o Origin) IsValid() bool {
	switch o {
	case OriginObserved, OriginCandidateSupplied, OriginRetrieved:
		return true
	}
	return false
}

// Isolation is the sandboxing level a check ran under.
type Isolation string

const (
	IsolationNone          Isolation = "none"
	IsolationWorktree      Isolation = "worktree"
	IsolationContainer     Isolation = "container"
	IsolationRemoteSandbox Isolation = "remote_sandbox"
)

// IsValid reports whether i is one of the defined Isolation values.
func (i Isolation) IsValid() bool {
	switch i {
	case IsolationNone, IsolationWorktree, IsolationContainer, IsolationRemoteSandbox:
		return true
	}
	return false
}

// ContractStatus describes how a contract term was established.
type ContractStatus string

const (
	ContractStatusUserSpecified ContractStatus = "user_specified"
	ContractStatusProvisional   ContractStatus = "provisional"
	ContractStatusApproved      ContractStatus = "approved"
)

// IsValid reports whether c is one of the defined ContractStatus values.
func (c ContractStatus) IsValid() bool {
	switch c {
	case ContractStatusUserSpecified, ContractStatusProvisional, ContractStatusApproved:
		return true
	}
	return false
}

// ClaimStatus is the verification state of a claim.
type ClaimStatus string

const (
	ClaimStatusVerified     ClaimStatus = "verified"
	ClaimStatusContradicted ClaimStatus = "contradicted"
	ClaimStatusUnverified   ClaimStatus = "unverified"
	ClaimStatusNotCheckable ClaimStatus = "not_checkable"
)

// IsValid reports whether c is one of the defined ClaimStatus values.
func (c ClaimStatus) IsValid() bool {
	switch c {
	case ClaimStatusVerified, ClaimStatusContradicted, ClaimStatusUnverified, ClaimStatusNotCheckable:
		return true
	}
	return false
}

// Severity is how serious an observation is.
type Severity string

const (
	SeverityObserved   Severity = "observed"
	SeveritySuspicious Severity = "suspicious"
	SeverityConfirmed  Severity = "confirmed"
)

// IsValid reports whether s is one of the defined Severity values.
func (s Severity) IsValid() bool {
	switch s {
	case SeverityObserved, SeveritySuspicious, SeverityConfirmed:
		return true
	}
	return false
}

// Disposition is the resolution state of an observation.
type Disposition string

const (
	DispositionOpen      Disposition = "open"
	DispositionExplained Disposition = "explained"
	DispositionConfirmed Disposition = "confirmed"
)

// IsValid reports whether d is one of the defined Disposition values.
func (d Disposition) IsValid() bool {
	switch d {
	case DispositionOpen, DispositionExplained, DispositionConfirmed:
		return true
	}
	return false
}

// Authority is the source of legitimacy for a rubric or measure.
type Authority string

const (
	AuthorityFormalStandard     Authority = "formal_standard"
	AuthorityEstablishedMeasure Authority = "established_measure"
	AuthorityVendorRating       Authority = "vendor_rating"
	AuthorityProjectRubric      Authority = "project_rubric"
)

// IsValid reports whether a is one of the defined Authority values.
func (a Authority) IsValid() bool {
	switch a {
	case AuthorityFormalStandard, AuthorityEstablishedMeasure, AuthorityVendorRating, AuthorityProjectRubric:
		return true
	}
	return false
}

// ObservationOrigin identifies who or what produced an observation.
type ObservationOrigin string

const (
	ObservationOriginAssistant ObservationOrigin = "assistant"
	ObservationOriginAdapter   ObservationOrigin = "adapter"
	ObservationOriginDetector  ObservationOrigin = "detector"
)

// IsValid reports whether o is one of the defined ObservationOrigin values.
func (o ObservationOrigin) IsValid() bool {
	switch o {
	case ObservationOriginAssistant, ObservationOriginAdapter, ObservationOriginDetector:
		return true
	}
	return false
}
