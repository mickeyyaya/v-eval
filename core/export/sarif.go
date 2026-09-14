// Package export turns a report into an interchange format another tool
// reads. The export presents: it never recomputes a verdict, a count, or a
// digest, and it adds nothing the report does not already state, so a SARIF
// log and the report it came from can only ever disagree about wording.
package export

import (
	"fmt"
	"strings"

	"github.com/mickeyyaya/v-eval/core/report"
)

const (
	// schemaURI and sarifVersion name the exact specification this export
	// targets: OASIS SARIF 2.1.0 errata01.
	schemaURI    = "https://docs.oasis-open.org/sarif/sarif/v2.1.0/errata01/os/schemas/sarif-schema-2.1.0.json"
	sarifVersion = "2.1.0"

	toolName           = "veval"
	toolInformationURI = "https://github.com/mickeyyaya/v-eval"

	// gitRevisionPrefix and contentRevisionPrefix are the two revision forms
	// SARIF has a home for; any other form stays in the property bag.
	gitRevisionPrefix     = "git:"
	contentRevisionPrefix = "content:sha256:"
	sha256HashName        = "sha-256"

	// sha256HexLength is how many characters a SHA-256 digest takes in
	// hexadecimal: 32 bytes, two characters each.
	sha256HexLength = 64
)

// ToSARIF maps a report onto a SARIF log. It expects a report that passed
// report.Validate -- every locator one complete group, every result naming a
// criterion the contract defines -- and it fails rather than guess where that
// does not hold: a criterion result the schema does not define has no SARIF
// kind, and a result or a finding whose criterion is absent from the contract
// has no rule to report against.
func ToSARIF(rep report.Report) (Log, error) {
	results, err := criterionResults(rep)
	if err != nil {
		return Log{}, err
	}
	findings, err := forensicResults(rep)
	if err != nil {
		return Log{}, err
	}
	run := Run{
		Tool:        Tool{Driver: driver(rep)},
		Invocations: invocations(rep),
		Artifacts:   artifactHashes(rep.Identity.Artifact),
		Results:     append(results, findings...),
		Properties:  Properties{"veval": vevalProperties(rep)},
	}
	return Log{Schema: schemaURI, Version: sarifVersion, Runs: []Run{run}}, nil
}

// Marshal writes a log in the canonical form every v-eval JSON takes. It is
// the report's own encoder, so a log and the report it came from are written
// by one set of rules rather than two that could drift apart.
func Marshal(log Log) ([]byte, error) {
	out, err := report.CanonicalJSON(log)
	if err != nil {
		return nil, fmt.Errorf("export: marshal: %w", err)
	}
	return out, nil
}

// driver states the tool and turns every contract criterion into a rule, so
// that a result can be read against the requirement it was judged by.
func driver(rep report.Report) Driver {
	rules := make([]Rule, 0, len(rep.Contract.Criteria))
	for _, criterion := range rep.Contract.Criteria {
		rules = append(rules, Rule{
			ID:               criterion.ID,
			Name:             criterion.ID,
			ShortDescription: Text{Text: criterion.Requirement},
			FullDescription:  Text{Text: criterion.AcceptanceRule},
			Properties: Properties{
				"veval_required":    criterion.Required,
				"veval_provisional": criterion.Provisional,
			},
		})
	}
	return Driver{Name: toolName, Version: rep.Identity.VevalVersion, InformationURI: toolInformationURI, Rules: rules}
}

// criterionResults maps one report result to one SARIF result, in report order.
func criterionResults(rep report.Report) ([]Result, error) {
	out := make([]Result, 0, len(rep.Criteria))
	for _, verdict := range rep.Criteria {
		criterion, ok := rep.Contract.CriterionByID(verdict.ID)
		if !ok {
			return nil, fmt.Errorf("export: sarif: criterion %q: no such criterion in the contract", verdict.ID)
		}
		kind, ok := kindFor(verdict.Result)
		if !ok {
			return nil, fmt.Errorf("export: sarif: criterion %q: result %q has no SARIF kind", verdict.ID, verdict.Result)
		}
		out = append(out, Result{
			RuleID:    verdict.ID,
			Kind:      kind,
			Level:     levelFor(verdict.Result, criterion.Required),
			Message:   Text{Text: verdict.Reasoning},
			Locations: locations(verdict.Evidence),
			Properties: Properties{
				"veval_result": string(verdict.Result),
				"evidence":     evidenceProperties(verdict.Evidence),
			},
		})
	}
	return out, nil
}

// forensicResults reports each finding against the criterion it bears on.
// A finding is a detector hit awaiting judgment, so it is always for review.
func forensicResults(rep report.Report) ([]Result, error) {
	out := make([]Result, 0, len(rep.Forensics))
	for _, finding := range rep.Forensics {
		if _, ok := rep.Contract.CriterionByID(finding.CriterionID); !ok {
			return nil, fmt.Errorf("export: sarif: finding %q: no such criterion %q in the contract", finding.FindingID, finding.CriterionID)
		}
		out = append(out, Result{
			RuleID:    finding.CriterionID,
			Kind:      kindReview,
			Message:   Text{Text: "forensic finding " + finding.FindingID + " from detector " + finding.Detector},
			Locations: locations(finding.Evidence),
			Properties: Properties{
				"veval_finding_id":         finding.FindingID,
				"veval_detector":           finding.Detector,
				"veval_severity":           string(finding.Severity),
				"veval_disposition":        string(finding.Disposition),
				"veval_benign_alternative": finding.BenignAlternative,
				"evidence":                 evidenceProperties(finding.Evidence),
			},
		})
	}
	return out, nil
}

// The SARIF result kinds this export writes. UNKNOWN and ERROR have no kind
// of their own; both are for review, and properties.veval_result keeps them
// apart.
const (
	kindPass          = "pass"
	kindFail          = "fail"
	kindNotApplicable = "notApplicable"
	kindReview        = "review"
)

// kindFor returns the SARIF kind for a result, and whether there is one.
func kindFor(result report.Result) (string, bool) {
	switch result {
	case report.ResultPass:
		return kindPass, true
	case report.ResultFail:
		return kindFail, true
	case report.ResultNotApplicable:
		return kindNotApplicable, true
	case report.ResultUnknown, report.ResultError:
		return kindReview, true
	}
	return "", false
}

// levelFor grades a failure by whether the criterion was required. SARIF
// allows a level only on a failing result and defaults the rest to "none",
// which is written by leaving the field out.
func levelFor(result report.Result, required bool) string {
	switch {
	case result != report.ResultFail:
		return ""
	case required:
		return "error"
	default:
		return "warning"
	}
}

// locations carries the file evidence, the only locator SARIF has a place
// for. Everything else is in properties.evidence.
func locations(evidence []report.Evidence) []Location {
	var out []Location
	for _, item := range evidence {
		if item.Locator.Shape() != report.ShapeFile {
			continue
		}
		out = append(out, Location{PhysicalLocation: PhysicalLocation{
			ArtifactLocation: ArtifactLocation{URI: item.Locator.File},
			Region:           &Region{StartLine: item.Locator.LineStart, EndLine: item.Locator.LineEnd},
		}})
	}
	return out
}

// evidenceProperties lists every piece of evidence cited, so that the record
// of what was seen survives the export whole and in the shape the report
// states it: what kind of evidence it is, where it came from, what isolation
// it ran under, what was observed, what produced it, and -- for a judgment --
// the rubric and model behind it. An empty rubric version or model is left
// out: it would read as one nobody can name rather than as a kind of evidence
// none applies to.
func evidenceProperties(evidence []report.Evidence) []map[string]any {
	out := make([]map[string]any, 0, len(evidence))
	for _, item := range evidence {
		entry := locatorFields(item.Locator)
		entry["kind"] = string(item.Kind)
		entry["origin"] = string(item.Origin)
		entry["isolation"] = string(item.Isolation)
		entry["observation"] = item.Observation
		entry["provenance"] = provenanceFields(item.Provenance)
		if item.RubricVersion != "" {
			entry["rubric_version"] = item.RubricVersion
		}
		if item.Model != "" {
			entry["model"] = item.Model
		}
		out = append(out, entry)
	}
	return out
}

// provenanceFields records what produced a piece of evidence and when, and
// holds what the report's own provenance object holds and nothing else. A
// rubric version and a model sit on the evidence beside it, where the report
// puts them.
func provenanceFields(provenance report.EvidenceProvenance) map[string]any {
	return map[string]any{
		"tool":      provenance.Tool,
		"version":   provenance.Version,
		"revision":  provenance.Revision,
		"timestamp": provenance.Timestamp,
	}
}

// locatorFields spells out where a piece of evidence came from, in the report's
// own field names. Every evidence object stands on its own, a file locator
// included: a file locator is also a physicalLocation, but a reader who has one
// object in hand should not have to count positions in locations[] to learn
// which file it names. Shape reports a group only when it is complete, so a
// command shape always carries an exit status and a file shape always carries
// a line range; a locator of no complete shape points nowhere and adds nothing.
func locatorFields(locator report.Locator) map[string]any {
	switch locator.Shape() {
	case report.ShapeFile:
		return map[string]any{
			"file":       locator.File,
			"line_start": locator.LineStart,
			"line_end":   locator.LineEnd,
		}
	case report.ShapeCommand:
		return map[string]any{
			"command":     locator.Command,
			"cwd":         locator.Cwd,
			"exit_status": *locator.ExitStatus,
			"log_ref":     locator.LogRef,
		}
	case report.ShapePassage:
		return map[string]any{
			"passage":                locator.Passage,
			"source_ref":             locator.SourceRef,
			"source_date_or_version": locator.SourceDateOrVersion,
			"access_date":            locator.AccessDate,
		}
	case report.ShapeNote:
		return map[string]any{"note": locator.Note}
	default:
		return map[string]any{}
	}
}

// invocations records what the evaluator ran, and what the evaluation could
// not decide. Each command invocation reports its own truth: its
// executionSuccessful follows that command's exit status and nothing else,
// because a command that exited zero did exit zero however the rest of the
// evaluation went.
//
// An ERROR criterion is a fact about the run rather than about any one
// command, so it is never attached to one. Every errored criterion is named in
// a notification on a single synthesized invocation, appended last, which
// names no command line and reports no success. A report that errored without
// running anything still gets that invocation, because a failure SARIF does
// not show is a failure a reader misses; a report with no errored criterion
// gets none.
func invocations(rep report.Report) []Invocation {
	out := make([]Invocation, 0, len(rep.Provenance.Commands)+1)
	for _, command := range rep.Provenance.Commands {
		exit := command.ExitStatus
		out = append(out, Invocation{
			CommandLine:         command.Command,
			WorkingDirectory:    &ArtifactLocation{URI: command.Cwd},
			ExitCode:            &exit,
			StartTimeUTC:        command.StartedAt,
			EndTimeUTC:          command.EndedAt,
			ExecutionSuccessful: exit == 0,
			Properties: Properties{
				"veval_isolation": string(command.Isolation),
				"veval_log_ref":   command.LogRef,
			},
		})
	}
	if notifications := errorNotifications(rep.Criteria); len(notifications) > 0 {
		out = append(out, Invocation{ToolExecutionNotifications: notifications})
	}
	return out
}

// errorNotifications names every criterion the evaluation could not decide
// because something went wrong while deciding it.
func errorNotifications(criteria []report.CriterionResult) []Notification {
	var out []Notification
	for _, verdict := range criteria {
		if verdict.Result != report.ResultError {
			continue
		}
		out = append(out, Notification{
			Level:   "error",
			Message: Text{Text: "criterion " + verdict.ID + " could not be evaluated: the evaluation errored"},
		})
	}
	return out
}

// artifactHashes places a content digest where SARIF keeps that kind of
// identity: an artifacts entry per path the report names, or a single entry
// when it names none. Any other revision form, a git revision included, is
// left alone in run.properties.veval rather than reshaped into a claim the
// report never made. A git revision in particular is not written as
// versionControlProvenance: SARIF requires a repositoryUri on every entry and
// a report carries none, so such an entry could only ever be invalid.
//
// A revision that announces a content digest without carrying one -- an empty
// digest, or anything that is not a SHA-256 written the one way a report
// writes it -- yields no entry either: an artifacts entry states what the file
// hashes to, and half a digest states nothing a reader could check.
func artifactHashes(artifact report.Artifact) []Artifact {
	hash, ok := strings.CutPrefix(artifact.Revision, contentRevisionPrefix)
	if !ok || !isSHA256Hex(hash) {
		return nil
	}
	if len(artifact.Paths) == 0 {
		return []Artifact{{Hashes: map[string]string{sha256HashName: hash}}}
	}
	out := make([]Artifact, 0, len(artifact.Paths))
	for _, path := range artifact.Paths {
		out = append(out, Artifact{
			Location: &ArtifactLocation{URI: path},
			Hashes:   map[string]string{sha256HashName: hash},
		})
	}
	return out
}

// isSHA256Hex reports whether s is a SHA-256 digest in the one spelling a
// report writes: exactly sha256HexLength lowercase hexadecimal characters.
func isSHA256Hex(s string) bool {
	if len(s) != sha256HexLength {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// vevalProperties carries the report-level facts SARIF has no field for: the
// verdict and the rule that produced it, the counts behind it, the identity a
// reader needs to find the report this log came from, and the revision it
// judged, including a git revision id that has no valid SARIF home of its own.
func vevalProperties(rep report.Report) map[string]any {
	properties := map[string]any{
		"overall":         string(rep.Status.Overall),
		"rule_applied":    rep.Status.RuleApplied,
		"blocked_by":      orEmpty(rep.Status.BlockedBy),
		"advisory":        rep.Status.Advisory,
		"counts":          rep.Counts,
		"report_id":       rep.Identity.ReportID,
		"evidence_digest": rep.Provenance.EvidenceDigest,
		"schema_version":  rep.Identity.SchemaVersion,
		"revision":        rep.Identity.Artifact.Revision,
	}
	if revision, ok := strings.CutPrefix(rep.Identity.Artifact.Revision, gitRevisionPrefix); ok {
		properties["vcs_revision_id"] = revision
	}
	return properties
}

// orEmpty returns a list JSON writes as an array, never as null. A report that
// blocks on nothing states an empty list, and a reader of the export should
// find the same empty list rather than an absence to interpret.
func orEmpty(list []string) []string {
	if list == nil {
		return []string{}
	}
	return list
}
