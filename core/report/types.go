package report

// Report is one complete evaluation: what was evaluated, against which
// contract, with what evidence, and what verdict follows.
type Report struct {
	Identity     Identity          `json:"identity"`
	Contract     Contract          `json:"contract"`
	Routing      Routing           `json:"routing"`
	Observations []Observation     `json:"observations"`
	Claims       []Claim           `json:"claims"`
	Criteria     []CriterionResult `json:"criteria"`
	Forensics    []Finding         `json:"forensics"`
	Dimensions   []Dimension       `json:"dimensions,omitempty"`
	Counts       Counts            `json:"counts"`
	Status       Status            `json:"status"`
	Improvement  []Improvement     `json:"improvement"`
	Limitations  Limitations       `json:"limitations"`
	Provenance   Provenance        `json:"provenance"`
	Learning     *Learning         `json:"learning,omitempty"`
}

// Identity names the report, the artifact it judged, and the task it served.
type Identity struct {
	ReportID      string   `json:"report_id"`
	SchemaVersion string   `json:"schema_version"`
	VevalVersion  string   `json:"veval_version"`
	SkillRevision string   `json:"skill_revision"`
	Artifact      Artifact `json:"artifact"`
	Task          Task     `json:"task"`
	CreatedAt     string   `json:"created_at"`
	Host          Host     `json:"host"`
}

// Artifact identifies the exact thing evaluated.
type Artifact struct {
	Kind         string   `json:"kind"`
	Revision     string   `json:"revision"`
	Paths        []string `json:"paths"`
	BundleDigest string   `json:"bundle_digest"`
}

// Task records what the artifact was meant to achieve, and for whom.
type Task struct {
	RequestedOutcome string `json:"requested_outcome"`
	IntendedUser     string `json:"intended_user"`
	BriefRef         string `json:"brief_ref"`
}

// Host describes where the evaluation ran.
type Host struct {
	CLI   string `json:"cli"`
	Model string `json:"model,omitempty"`
	OS    string `json:"os"`
	Arch  string `json:"arch"`
}

// Contract is the set of criteria the artifact was judged against.
type Contract struct {
	ContractID      string         `json:"contract_id"`
	ContractVersion string         `json:"contract_version"`
	Status          ContractStatus `json:"status"`
	Conflicts       []Conflict     `json:"conflicts"`
	Criteria        []Criterion    `json:"criteria"`
}

// CriterionByID returns the contract criterion with the given id.
func (c Contract) CriterionByID(id string) (Criterion, bool) {
	for _, criterion := range c.Criteria {
		if criterion.ID == id {
			return criterion, true
		}
	}
	return Criterion{}, false
}

// Conflict records criteria that cannot all hold at once.
type Conflict struct {
	Between     []string `json:"between"`
	Description string   `json:"description"`
	Resolution  string   `json:"resolution"` // literal "unresolved" blocks acceptance
}

// Criterion is a single rule the artifact must satisfy.
type Criterion struct {
	ID               string   `json:"id"`
	Requirement      string   `json:"requirement"`
	SourceRef        string   `json:"source_ref"`
	Required         bool     `json:"required"`
	Applicability    string   `json:"applicability"`
	MethodsAllowed   []Method `json:"methods_allowed"`
	AcceptanceRule   string   `json:"acceptance_rule"`
	ExpectedEvidence string   `json:"expected_evidence"`
	Provisional      bool     `json:"provisional"`
}

// Routing records which inputs were supplied and how they were dispatched.
type Routing struct {
	Supplied        []SuppliedSource `json:"supplied"`
	Profiles        []ProfileRef     `json:"profiles"`
	AdaptersRun     []AdapterRun     `json:"adapters_run"`
	AdaptersSkipped []AdapterSkipped `json:"adapters_skipped"`
	Ambiguity       []Ambiguity      `json:"ambiguity"`
	Rationale       string           `json:"rationale"`
}

// SuppliedSource counts one class of input the evaluator received.
type SuppliedSource struct {
	SourceType string `json:"source_type"`
	Count      int    `json:"count"`
	Origin     string `json:"origin"`
}

// ProfileRef names an evaluation profile and its version.
type ProfileRef struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// AdapterRun records an adapter that ran and what it consumed.
type AdapterRun struct {
	Name       string   `json:"name"`
	Version    string   `json:"version"`
	InputsUsed []string `json:"inputs_used"`
}

// AdapterSkipped records an adapter that could not run, and why.
type AdapterSkipped struct {
	Name         string `json:"name"`
	MissingInput string `json:"missing_input"`
}

// Ambiguity records an open question and how it was settled.
type Ambiguity struct {
	Question   string `json:"question"`
	Resolution string `json:"resolution"`
}

// Observation is something noticed about the artifact, with its support.
type Observation struct {
	ID           string            `json:"id"`
	Text         string            `json:"text"`
	Evidence     []Evidence        `json:"evidence"`
	CriterionIDs []string          `json:"criterion_ids"`
	Origin       ObservationOrigin `json:"origin"`
}

// Claim is an assertion made by the candidate, and its verification state.
type Claim struct {
	ClaimID      string      `json:"claim_id"`
	Text         string      `json:"text"`
	Location     string      `json:"location"`
	Status       ClaimStatus `json:"status"`
	Verification []Evidence  `json:"verification"`
	Notes        string      `json:"notes"`
}

// CriterionResult is the verdict reached for one criterion.
type CriterionResult struct {
	ID              string     `json:"id"`
	Result          Result     `json:"result"`
	MethodUsed      Method     `json:"method_used"`
	Evidence        []Evidence `json:"evidence"`
	Reasoning       string     `json:"reasoning"`
	NextAction      string     `json:"next_action"`
	SharedCauseWith []string   `json:"shared_cause_with"`
	DimensionRefs   []string   `json:"dimension_refs"`
}

// Finding is a detector hit recorded for forensic review.
type Finding struct {
	FindingID         string      `json:"finding_id"`
	Detector          string      `json:"detector"`
	DetectorVersion   string      `json:"detector_version"`
	CriterionID       string      `json:"criterion_id"`
	Severity          Severity    `json:"severity"`
	Evidence          []Evidence  `json:"evidence"`
	BenignAlternative string      `json:"benign_alternative"`
	Disposition       Disposition `json:"disposition"`
}

// Dimension is one measured quantity with its metric and threshold.
type Dimension struct {
	Dimension       string     `json:"dimension"`
	Metric          string     `json:"metric"`
	MetricVersion   string     `json:"metric_version"`
	AuthorityType   Authority  `json:"authority_type"`
	DefinitionRef   string     `json:"definition_ref"`
	Unit            string     `json:"unit"`
	Range           string     `json:"range"`
	Direction       string     `json:"direction"`
	Value           *float64   `json:"value"` // null when missing; never zero for missing
	Threshold       string     `json:"threshold"`
	ThresholdSource string     `json:"threshold_source"`
	Tool            string     `json:"tool"`
	ToolVersion     string     `json:"tool_version"`
	Workload        string     `json:"workload"`
	Evidence        []Evidence `json:"evidence"`
	Interpretation  string     `json:"interpretation"`
}

// Counts tallies criterion results and assessment coverage.
type Counts struct {
	Required Tally    `json:"required"`
	Optional Tally    `json:"optional"`
	Coverage Coverage `json:"coverage"`
}

// Tally counts criteria by result within one requiredness class.
type Tally struct {
	Applicable    int `json:"applicable"`
	Pass          int `json:"pass"`
	Fail          int `json:"fail"`
	Unknown       int `json:"unknown"`
	Error         int `json:"error"`
	NotApplicable int `json:"not_applicable"`
}

// Coverage is the fraction of applicable criteria actually assessed.
type Coverage struct {
	Numerator   int  `json:"numerator"`
	Denominator int  `json:"denominator"`
	Undefined   bool `json:"undefined"`
}

// Status is the report-level verdict and the rule that produced it.
type Status struct {
	Overall     Overall  `json:"overall"`
	RuleApplied string   `json:"rule_applied"`
	BlockedBy   []string `json:"blocked_by"`
	Advisory    bool     `json:"advisory"`
}

// Improvement is one actionable change, tied to the criteria it would resolve.
type Improvement struct {
	Issue                 string   `json:"issue"`
	Locations             []string `json:"locations"`
	SuggestedChange       string   `json:"suggested_change"`
	ConstraintsToPreserve []string `json:"constraints_to_preserve"`
	VerifyBy              string   `json:"verify_by"`
	LinkedCriteria        []string `json:"linked_criteria"`
}

// Limitations states what the evaluation did not cover.
type Limitations struct {
	NotInspected    []string `json:"not_inspected"`
	NotExecuted     []string `json:"not_executed"`
	UnknownMetadata []string `json:"unknown_metadata"`
	Assumptions     []string `json:"assumptions"`
}

// Provenance records the tools, commands, and environment behind the report.
type Provenance struct {
	Tools               []ToolRef       `json:"tools"`
	Commands            []CommandRecord `json:"commands"`
	Environment         Environment     `json:"environment"`
	IsolationLevelsUsed []Isolation     `json:"isolation_levels_used"`
	EvidenceDigest      string          `json:"evidence_digest"`
}

// ToolRef names a tool and its version.
type ToolRef struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// CommandRecord is one command the evaluator ran.
type CommandRecord struct {
	Command    string    `json:"command"`
	Cwd        string    `json:"cwd"`
	ExitStatus int       `json:"exit_status"`
	StartedAt  string    `json:"started_at"`
	EndedAt    string    `json:"ended_at"`
	LogRef     string    `json:"log_ref"`
	Isolation  Isolation `json:"isolation"`
}

// Environment describes the machine the evaluation ran on.
type Environment struct {
	OS              string            `json:"os"`
	Arch            string            `json:"arch"`
	RuntimeVersions map[string]string `json:"runtime_versions"`
}

// Learning records precedents consulted and reward records emitted.
type Learning struct {
	PrecedentsRetrieved  []PrecedentRef `json:"precedents_retrieved"`
	RewardRecordsCreated []string       `json:"reward_records_created"`
}

// PrecedentRef points at a prior evaluation and how similar it is.
type PrecedentRef struct {
	PrecedentID string  `json:"precedent_id"`
	Criterion   string  `json:"criterion"`
	Similarity  float64 `json:"similarity"`
}

// Evidence is a single piece of support cited for a check's verdict.
type Evidence struct {
	Kind          Kind               `json:"kind"`
	Locator       Locator            `json:"locator"`
	Observation   string             `json:"observation"`
	Provenance    EvidenceProvenance `json:"provenance"`
	Isolation     Isolation          `json:"isolation"`
	Origin        Origin             `json:"origin"`
	RubricVersion string             `json:"rubric_version,omitempty"` // required when Kind == judgment
	Model         string             `json:"model,omitempty"`          // "unknown" is a valid value
}

// EvidenceProvenance records what produced a piece of evidence and when.
type EvidenceProvenance struct {
	Tool      string `json:"tool"`
	Version   string `json:"version"`
	Revision  string `json:"revision"`
	Timestamp string `json:"timestamp"`
}
