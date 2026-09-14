// Package export turns a report into an interchange format another tool
// reads. The export presents: it never recomputes a verdict, a count, or a
// digest, and it adds nothing the report does not already state, so a SARIF
// log and the report it came from can only ever disagree about wording.
package export

import (
	"bytes"
	"encoding/json"
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
)

// ToSARIF maps a report onto a SARIF log. It fails rather than guess: a
// criterion result the schema does not define has no SARIF kind, and a result
// whose criterion is absent from the contract has no rule to report against.
func ToSARIF(rep report.Report) (Log, error) {
	results, err := criterionResults(rep)
	if err != nil {
		return Log{}, err
	}
	provenance, artifacts := revisionProvenance(rep.Identity.Artifact)
	run := Run{
		Tool:                     Tool{Driver: driver(rep)},
		Invocations:              invocations(rep),
		VersionControlProvenance: provenance,
		Artifacts:                artifacts,
		Results:                  append(results, forensicResults(rep.Forensics)...),
		Properties:               Properties{"veval": vevalProperties(rep)},
	}
	return Log{Schema: schemaURI, Version: sarifVersion, Runs: []Run{run}}, nil
}

// Marshal writes a log in the canonical form every v-eval JSON takes:
// two-space indent, no HTML escaping, one trailing newline.
func Marshal(log Log) ([]byte, error) {
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(log); err != nil {
		return nil, fmt.Errorf("export: marshal: %w", err)
	}
	return out.Bytes(), nil
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
			return nil, fmt.Errorf("export: criterion %q: no such criterion in the contract", verdict.ID)
		}
		kind, ok := kindFor(verdict.Result)
		if !ok {
			return nil, fmt.Errorf("export: criterion %q: result %q has no SARIF kind", verdict.ID, verdict.Result)
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
func forensicResults(findings []report.Finding) []Result {
	out := make([]Result, 0, len(findings))
	for _, finding := range findings {
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
	return out
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
// of what was seen survives the export whole: what kind of evidence it is,
// where it came from, what isolation it ran under, and what was observed.
func evidenceProperties(evidence []report.Evidence) []map[string]any {
	out := make([]map[string]any, 0, len(evidence))
	for _, item := range evidence {
		entry := locatorFields(item.Locator)
		entry["kind"] = string(item.Kind)
		entry["origin"] = string(item.Origin)
		entry["isolation"] = string(item.Isolation)
		entry["observation"] = item.Observation
		out = append(out, entry)
	}
	return out
}

// locatorFields spells out the locators SARIF cannot hold: a command, a
// passage, or a note. A file locator is already a physicalLocation and is not
// repeated here, so a reader follows locations and properties.evidence in the
// same order; a locator of no complete shape points nowhere and adds nothing.
// Shape reports a group only when it is complete, so a command shape always
// carries an exit status.
func locatorFields(locator report.Locator) map[string]any {
	switch locator.Shape() {
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

// invocations records what the evaluator ran. An ERROR criterion means the
// evaluation itself did not complete, so no invocation of the run counts as
// successful and each errored criterion is named in a notification. A report
// that errored without running anything still gets an invocation to carry
// them, because a failure SARIF does not show is a failure a reader misses.
func invocations(rep report.Report) []Invocation {
	notifications := errorNotifications(rep.Criteria)
	completed := len(notifications) == 0
	out := make([]Invocation, 0, len(rep.Provenance.Commands))
	for _, command := range rep.Provenance.Commands {
		exit := command.ExitStatus
		out = append(out, Invocation{
			CommandLine:         command.Command,
			WorkingDirectory:    &ArtifactLocation{URI: command.Cwd},
			ExitCode:            &exit,
			StartTimeUtc:        command.StartedAt,
			EndTimeUtc:          command.EndedAt,
			ExecutionSuccessful: completed && exit == 0,
			Properties: Properties{
				"veval_isolation": string(command.Isolation),
				"veval_log_ref":   command.LogRef,
			},
		})
	}
	if completed {
		return out
	}
	if len(out) == 0 {
		out = append(out, Invocation{})
	}
	out[0].ToolExecutionNotifications = notifications
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

// revisionProvenance places the artifact revision where SARIF keeps that kind
// of identity: a git revision is version control provenance, a content digest
// is an artifact hash. Neither branch nor repository URI is written, because
// a report does not carry them. Any other revision form is left alone, in
// run.properties.veval.revision, rather than reshaped into a claim the report
// never made.
func revisionProvenance(artifact report.Artifact) ([]VCS, []Artifact) {
	if revision, ok := strings.CutPrefix(artifact.Revision, gitRevisionPrefix); ok {
		return []VCS{{RevisionID: revision}}, nil
	}
	hash, ok := strings.CutPrefix(artifact.Revision, contentRevisionPrefix)
	if !ok {
		return nil, nil
	}
	if len(artifact.Paths) == 0 {
		return nil, []Artifact{{Hashes: map[string]string{sha256HashName: hash}}}
	}
	out := make([]Artifact, 0, len(artifact.Paths))
	for _, path := range artifact.Paths {
		out = append(out, Artifact{
			Location: &ArtifactLocation{URI: path},
			Hashes:   map[string]string{sha256HashName: hash},
		})
	}
	return nil, out
}

// vevalProperties carries the report-level facts SARIF has no field for: the
// verdict and the rule that produced it, the counts behind it, and the
// identity a reader needs to find the report this log came from.
func vevalProperties(rep report.Report) map[string]any {
	return map[string]any{
		"overall":         string(rep.Status.Overall),
		"rule_applied":    rep.Status.RuleApplied,
		"blocked_by":      rep.Status.BlockedBy,
		"advisory":        rep.Status.Advisory,
		"counts":          rep.Counts,
		"report_id":       rep.Identity.ReportID,
		"evidence_digest": rep.Provenance.EvidenceDigest,
		"schema_version":  rep.Identity.SchemaVersion,
		"revision":        rep.Identity.Artifact.Revision,
	}
}
