package export

// The types below are the slice of SARIF 2.1.0 this export writes, and no
// more: a field absent here is a field the mapping never produces. Names and
// casing follow the OASIS SARIF 2.1.0 errata01 schema, so a reader can check
// this file against the specification without a translation table.

// Log is a SARIF log file: one v-eval report becomes one log with one run.
type Log struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []Run  `json:"runs"`
}

// Run is a single invocation of an analysis tool over an artifact.
type Run struct {
	Tool                     Tool         `json:"tool"`
	Invocations              []Invocation `json:"invocations,omitempty"`
	VersionControlProvenance []VCS        `json:"versionControlProvenance,omitempty"`
	Artifacts                []Artifact   `json:"artifacts,omitempty"`
	Results                  []Result     `json:"results"`
	Properties               Properties   `json:"properties,omitempty"`
}

// Properties is a SARIF property bag: the place for what SARIF has no field
// for. Every v-eval state that SARIF cannot express lands in one of these.
type Properties map[string]any

// Tool names the analysis tool that produced a run.
type Tool struct {
	Driver Driver `json:"driver"`
}

// Driver is the tool component that holds the rules a run reports against.
type Driver struct {
	Name           string `json:"name"`
	Version        string `json:"version,omitempty"`
	InformationURI string `json:"informationUri,omitempty"`
	Rules          []Rule `json:"rules"`
}

// Rule is a reportingDescriptor: one contract criterion, as SARIF states it.
type Rule struct {
	ID               string     `json:"id"`
	Name             string     `json:"name,omitempty"`
	ShortDescription Text       `json:"shortDescription"`
	FullDescription  Text       `json:"fullDescription"`
	Properties       Properties `json:"properties,omitempty"`
}

// Text is SARIF's message and multiformatMessageString: both carry a plain
// string under "text", which is all this export writes.
type Text struct {
	Text string `json:"text"`
}

// Result is one verdict reported against one rule.
type Result struct {
	RuleID     string     `json:"ruleId"`
	Kind       string     `json:"kind"`
	Level      string     `json:"level,omitempty"`
	Message    Text       `json:"message"`
	Locations  []Location `json:"locations,omitempty"`
	Properties Properties `json:"properties,omitempty"`
}

// Location points at the place in an artifact a result concerns.
type Location struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

// PhysicalLocation is a file and, optionally, a region within it.
type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation"`
	Region           *Region          `json:"region,omitempty"`
}

// ArtifactLocation is a URI for a file or directory.
type ArtifactLocation struct {
	URI string `json:"uri"`
}

// Region is a line range within an artifact. SARIF line numbers are 1-based,
// as a v-eval file locator's are.
type Region struct {
	StartLine int `json:"startLine"`
	EndLine   int `json:"endLine,omitempty"`
}

// Invocation is one command the evaluator ran, or, when nothing ran under a
// criterion that errored, the record that the run did not complete.
type Invocation struct {
	CommandLine                string            `json:"commandLine,omitempty"`
	WorkingDirectory           *ArtifactLocation `json:"workingDirectory,omitempty"`
	ExitCode                   *int              `json:"exitCode,omitempty"`
	StartTimeUtc               string            `json:"startTimeUtc,omitempty"`
	EndTimeUtc                 string            `json:"endTimeUtc,omitempty"`
	ExecutionSuccessful        bool              `json:"executionSuccessful"`
	ToolExecutionNotifications []Notification    `json:"toolExecutionNotifications,omitempty"`
	Properties                 Properties        `json:"properties,omitempty"`
}

// Notification is something the tool reports about its own execution.
type Notification struct {
	Level   string `json:"level,omitempty"`
	Message Text   `json:"message"`
}

// VCS is a versionControlDetails entry: which revision was analyzed.
type VCS struct {
	RevisionID string `json:"revisionId"`
}

// Artifact is a file the run analyzed, with the hashes that pin its content.
type Artifact struct {
	Location *ArtifactLocation `json:"location,omitempty"`
	Hashes   map[string]string `json:"hashes,omitempty"`
}
