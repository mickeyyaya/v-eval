package export_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mickeyyaya/v-eval/core/export"
	"github.com/mickeyyaya/v-eval/core/report"
)

// update rewrites the golden files instead of comparing against them. The
// goldens are read by eye, so regenerating them is a deliberate act.
var update = flag.Bool("update", false, "rewrite golden files")

// loadReport decodes one of the report package's worked fixtures by name.
func loadReport(t *testing.T, name string) report.Report {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "report", "testdata", name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	rep, err := report.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	return rep
}

// loadFixture decodes the worked example, the fixture most tests start from.
func loadFixture(t *testing.T) report.Report {
	t.Helper()
	return loadReport(t, "worked-example")
}

// mustSARIF maps a report, failing the test on any mapping error.
func mustSARIF(t *testing.T, rep report.Report) export.Log {
	t.Helper()
	log, err := export.ToSARIF(rep)
	if err != nil {
		t.Fatal(err)
	}
	return log
}

// resultsFor returns every result written against one rule id, in order.
func resultsFor(t *testing.T, log export.Log, ruleID string) []export.Result {
	t.Helper()
	if len(log.Runs) != 1 {
		t.Fatalf("runs = %d, want 1", len(log.Runs))
	}
	var out []export.Result
	for _, r := range log.Runs[0].Results {
		if r.RuleID == ruleID {
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		t.Fatalf("no result for rule %q", ruleID)
	}
	return out
}

// evidenceOf returns the evidence list a result carries in its property bag.
func evidenceOf(t *testing.T, r export.Result) []map[string]any {
	t.Helper()
	evidence, ok := r.Properties["evidence"].([]map[string]any)
	if !ok {
		t.Fatalf("result %s: properties.evidence = %#v", r.RuleID, r.Properties["evidence"])
	}
	return evidence
}

// ruleByID returns the driver rule with an id, or fails the test.
func ruleByID(t *testing.T, log export.Log, id string) export.Rule {
	t.Helper()
	for _, rule := range log.Runs[0].Tool.Driver.Rules {
		if rule.ID == id {
			return rule
		}
	}
	t.Fatalf("no rule %q", id)
	return export.Rule{}
}

// assertGolden compares got against testdata/name, or rewrites it under -update.
func assertGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("missing golden %s; run: go test ./core/export -update", path)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("%s differs from golden; run go test ./core/export -update and review the diff", name)
	}
}

func TestSARIFMapping(t *testing.T) {
	t.Parallel()
	log, err := export.ToSARIF(loadFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if log.Version != "2.1.0" || len(log.Runs) != 1 || len(log.Runs[0].Tool.Driver.Rules) != 5 {
		t.Fatalf("shape: %+v", log)
	}
	kinds := map[string]string{}
	for _, r := range log.Runs[0].Results {
		kinds[r.RuleID] = r.Kind
	}
	if kinds["C1"] != "fail" || kinds["C4"] != "pass" || kinds["C5"] != "review" {
		t.Fatalf("kinds = %v", kinds)
	}
	veval := log.Runs[0].Properties["veval"].(map[string]any)
	if veval["overall"] != "FAIL" {
		t.Fatal("run.properties.veval.overall must carry the overall status")
	}
}

func TestSARIFErrorCriterionMarksInvocationFailed(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Criteria[4].Result = report.ResultError
	rep.Provenance.Commands = []report.CommandRecord{{Command: "go test ./...", Cwd: ".", ExitStatus: 2, StartedAt: "2026-09-14T00:00:00Z", EndedAt: "2026-09-14T00:00:01Z", LogRef: "logs/1.txt", Isolation: report.IsolationNone}}
	log, err := export.ToSARIF(rep)
	if err != nil {
		t.Fatal(err)
	}
	inv := log.Runs[0].Invocations
	if len(inv) != 1 || inv[0].ExecutionSuccessful || len(inv[0].ToolExecutionNotifications) != 1 {
		t.Fatalf("invocations = %+v", inv)
	}
}

func TestSARIFSchemaAndToolIdentifiers(t *testing.T) {
	t.Parallel()
	log := mustSARIF(t, loadFixture(t))
	if log.Schema != "https://docs.oasis-open.org/sarif/sarif/v2.1.0/errata01/os/schemas/sarif-schema-2.1.0.json" {
		t.Fatalf("schema = %q", log.Schema)
	}
	driver := log.Runs[0].Tool.Driver
	if driver.Name != "veval" || driver.Version != "0.0.0-dev" || driver.InformationURI != "https://github.com/mickeyyaya/v-eval" {
		t.Fatalf("driver = %+v", driver)
	}
}

func TestSARIFRulesCarryContractText(t *testing.T) {
	t.Parallel()
	log := mustSARIF(t, loadReport(t, "extended-example"))
	rule := ruleByID(t, log, "E3")
	if rule.Name != "E3" {
		t.Fatalf("rule name = %q, want the id", rule.Name)
	}
	if rule.ShortDescription.Text != "Throughput stays at or above the stated target" {
		t.Fatalf("shortDescription = %q, want the requirement", rule.ShortDescription.Text)
	}
	if !strings.HasPrefix(rule.FullDescription.Text, "Measured throughput") {
		t.Fatalf("fullDescription = %q, want the acceptance rule", rule.FullDescription.Text)
	}
	if rule.Properties["veval_required"] != false || rule.Properties["veval_provisional"] != false {
		t.Fatalf("rule properties = %v", rule.Properties)
	}
}

func TestSARIFLevelSetOnlyOnFail(t *testing.T) {
	t.Parallel()
	log := mustSARIF(t, loadFixture(t))
	for _, r := range log.Runs[0].Results {
		wantLevel := ""
		if r.RuleID != "C4" && r.RuleID != "C5" {
			wantLevel = "error" // C1-C3 fail on required criteria
		}
		if r.Level != wantLevel {
			t.Fatalf("result %s: level = %q, want %q", r.RuleID, r.Level, wantLevel)
		}
	}
}

func TestSARIFOptionalFailIsWarning(t *testing.T) {
	t.Parallel()
	rep := loadReport(t, "extended-example")
	rep.Criteria[2].Result = report.ResultFail // E3 is the one optional criterion
	got := resultsFor(t, mustSARIF(t, rep), "E3")[0]
	if got.Kind != "fail" || got.Level != "warning" {
		t.Fatalf("optional fail = kind %q level %q, want fail/warning", got.Kind, got.Level)
	}
}

func TestSARIFEveryResultCarriesVevalResult(t *testing.T) {
	t.Parallel()
	log := mustSARIF(t, loadReport(t, "extended-example"))
	want := map[string]string{"E1": "PASS", "E2": "PASS", "E3": "UNKNOWN", "E4": "ERROR", "E5": "NOT_APPLICABLE", "E6": "UNKNOWN"}
	for id, result := range want {
		got := resultsFor(t, log, id)[0]
		if got.Properties["veval_result"] != result {
			t.Fatalf("result %s: veval_result = %v, want %q", id, got.Properties["veval_result"], result)
		}
	}
	if kind := resultsFor(t, log, "E4")[0].Kind; kind != "review" {
		t.Fatalf("ERROR criterion kind = %q, want review", kind)
	}
	if kind := resultsFor(t, log, "E5")[0].Kind; kind != "notApplicable" {
		t.Fatalf("NOT_APPLICABLE criterion kind = %q, want notApplicable", kind)
	}
}

func TestSARIFFileLocatorBecomesPhysicalLocation(t *testing.T) {
	t.Parallel()
	got := resultsFor(t, mustSARIF(t, loadFixture(t)), "C4")[0]
	if len(got.Locations) != 1 {
		t.Fatalf("locations = %+v", got.Locations)
	}
	physical := got.Locations[0].PhysicalLocation
	if physical.ArtifactLocation.URI != "examples/code-review/input.md" {
		t.Fatalf("uri = %q", physical.ArtifactLocation.URI)
	}
	if physical.Region.StartLine != 26 || physical.Region.EndLine != 27 {
		t.Fatalf("region = %+v", physical.Region)
	}
}

func TestSARIFNonFileLocatorsBecomeEvidenceProperties(t *testing.T) {
	t.Parallel()
	extended := mustSARIF(t, loadReport(t, "extended-example"))

	command := evidenceOf(t, resultsFor(t, extended, "E1")[0])[0]
	if command["command"] != "go test ./..." || command["cwd"] != "/tmp/veval-worktree" ||
		command["exit_status"] != 0 || command["log_ref"] != "logs/e1-go-test.log" {
		t.Fatalf("command evidence = %v", command)
	}
	passage := evidenceOf(t, resultsFor(t, extended, "E2")[0])[0]
	if passage["source_ref"] != "https://api.example.test/docs/rate-limits" ||
		passage["source_date_or_version"] != "2026-08-30" || passage["access_date"] != "2026-09-14" ||
		!strings.HasPrefix(passage["passage"].(string), "Clients may issue") {
		t.Fatalf("passage evidence = %v", passage)
	}
	note := evidenceOf(t, resultsFor(t, mustSARIF(t, loadFixture(t)), "C5")[0])[0]
	if note["note"] != "Candidate author's statement in input.md" {
		t.Fatalf("note evidence = %v", note)
	}
}

func TestSARIFEveryEvidenceRecordIsListed(t *testing.T) {
	t.Parallel()
	// E2 cites a passage and a file; the file one is a location, and both are
	// listed in properties.evidence so the record of what was seen is complete.
	got := resultsFor(t, mustSARIF(t, loadReport(t, "extended-example")), "E2")[0]
	evidence := evidenceOf(t, got)
	if len(evidence) != 2 || len(got.Locations) != 1 {
		t.Fatalf("evidence = %d, locations = %d, want 2 and 1", len(evidence), len(got.Locations))
	}
	for _, entry := range evidence {
		for _, key := range []string{"kind", "origin", "isolation", "observation"} {
			if _, ok := entry[key]; !ok {
				t.Fatalf("evidence entry %v is missing %q", entry, key)
			}
		}
	}
	if _, ok := evidence[1]["file"]; ok {
		t.Fatalf("file locator repeated in properties.evidence: %v", evidence[1])
	}
}

func TestSARIFCommandsBecomeInvocations(t *testing.T) {
	t.Parallel()
	log := mustSARIF(t, loadReport(t, "extended-example"))
	inv := log.Runs[0].Invocations
	if len(inv) != 3 {
		t.Fatalf("invocations = %d, want one per command", len(inv))
	}
	first := inv[0]
	if first.CommandLine != "go test ./..." || first.WorkingDirectory.URI != "/tmp/veval-worktree" ||
		first.ExitCode == nil || *first.ExitCode != 0 ||
		first.StartTimeUtc != "2026-09-14T00:03:10Z" || first.EndTimeUtc != "2026-09-14T00:03:12Z" {
		t.Fatalf("invocation = %+v", first)
	}
	if first.Properties["veval_isolation"] != "worktree" || first.Properties["veval_log_ref"] != "logs/e1-go-test.log" {
		t.Fatalf("invocation properties = %v", first.Properties)
	}
	// E4 is an ERROR criterion, so no invocation of this run is successful and
	// exactly one notification names it.
	notifications := 0
	for _, one := range inv {
		if one.ExecutionSuccessful {
			t.Fatalf("invocation %q reported success under an ERROR criterion", one.CommandLine)
		}
		notifications += len(one.ToolExecutionNotifications)
	}
	if notifications != 1 {
		t.Fatalf("notifications = %d, want one per ERROR criterion", notifications)
	}
	got := inv[0].ToolExecutionNotifications[0]
	if got.Level != "error" || !strings.Contains(got.Message.Text, "E4") {
		t.Fatalf("notification = %+v", got)
	}
}

func TestSARIFInvocationSuccessfulWhenNoErrorCriterion(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Provenance.Commands = []report.CommandRecord{
		{Command: "go test ./...", Cwd: ".", ExitStatus: 0, StartedAt: "a", EndedAt: "b", LogRef: "logs/1.txt", Isolation: report.IsolationWorktree},
		{Command: "go vet ./...", Cwd: ".", ExitStatus: 1, StartedAt: "a", EndedAt: "b", LogRef: "logs/2.txt", Isolation: report.IsolationWorktree},
	}
	inv := mustSARIF(t, rep).Runs[0].Invocations
	if len(inv) != 2 || !inv[0].ExecutionSuccessful || inv[1].ExecutionSuccessful {
		t.Fatalf("executionSuccessful must follow the exit status: %+v", inv)
	}
	if len(inv[0].ToolExecutionNotifications) != 0 {
		t.Fatalf("notifications without an ERROR criterion: %+v", inv[0])
	}
}

func TestSARIFErrorWithoutCommandsSynthesizesInvocation(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t) // the worked example runs no commands
	rep.Criteria[4].Result = report.ResultError
	inv := mustSARIF(t, rep).Runs[0].Invocations
	if len(inv) != 1 || inv[0].ExecutionSuccessful || len(inv[0].ToolExecutionNotifications) != 1 {
		t.Fatalf("an ERROR criterion must never be silent: %+v", inv)
	}
	if inv[0].CommandLine != "" {
		t.Fatalf("synthesized invocation invented a command line: %q", inv[0].CommandLine)
	}
}

func TestSARIFNoInvocationsWhenNothingRan(t *testing.T) {
	t.Parallel()
	if inv := mustSARIF(t, loadFixture(t)).Runs[0].Invocations; len(inv) != 0 {
		t.Fatalf("invocations = %+v, want none", inv)
	}
}

func TestSARIFGitRevisionBecomesVersionControlProvenance(t *testing.T) {
	t.Parallel()
	run := mustSARIF(t, loadReport(t, "extended-example")).Runs[0]
	if len(run.VersionControlProvenance) != 1 || run.VersionControlProvenance[0].RevisionID != "9f2c1a4e5b6d7c8f9a0b1c2d3e4f5a6b7c8d9e0f" {
		t.Fatalf("versionControlProvenance = %+v", run.VersionControlProvenance)
	}
	if len(run.Artifacts) != 0 {
		t.Fatalf("a git revision carries no artifact hash: %+v", run.Artifacts)
	}
}

func TestSARIFContentRevisionBecomesArtifactHashes(t *testing.T) {
	t.Parallel()
	run := mustSARIF(t, loadFixture(t)).Runs[0]
	if len(run.Artifacts) != 1 {
		t.Fatalf("artifacts = %+v, want one per artifact path", run.Artifacts)
	}
	got := run.Artifacts[0]
	if got.Location == nil || got.Location.URI != "examples/code-review/input.md" {
		t.Fatalf("artifact location = %+v", got.Location)
	}
	if got.Hashes["sha-256"] != "befaf705c894d30d62d3d53dd5e6653c910063747fc7e508c1b9de2de306dfa8" {
		t.Fatalf("artifact hashes = %v", got.Hashes)
	}
	if len(run.VersionControlProvenance) != 0 {
		t.Fatalf("a content revision is not version control provenance: %+v", run.VersionControlProvenance)
	}
}

func TestSARIFContentRevisionWithoutPaths(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Identity.Artifact.Paths = nil
	run := mustSARIF(t, rep).Runs[0]
	if len(run.Artifacts) != 1 || run.Artifacts[0].Location != nil {
		t.Fatalf("artifacts = %+v, want one hash entry with no location", run.Artifacts)
	}
}

func TestSARIFUnrecognizedRevisionStaysInProperties(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Identity.Artifact.Revision = "svn:r1234"
	run := mustSARIF(t, rep).Runs[0]
	if len(run.Artifacts) != 0 || len(run.VersionControlProvenance) != 0 {
		t.Fatalf("an unrecognized revision must not be reshaped: %+v %+v", run.Artifacts, run.VersionControlProvenance)
	}
	if run.Properties["veval"].(map[string]any)["revision"] != "svn:r1234" {
		t.Fatalf("veval properties = %v", run.Properties["veval"])
	}
}

func TestSARIFRunPropertiesCarryTheVerdict(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	veval := mustSARIF(t, rep).Runs[0].Properties["veval"].(map[string]any)
	if veval["rule_applied"] != "required applicable criterion failed" || veval["advisory"] != false {
		t.Fatalf("veval = %v", veval)
	}
	if blocked, ok := veval["blocked_by"].([]string); !ok || len(blocked) != 3 {
		t.Fatalf("blocked_by = %v", veval["blocked_by"])
	}
	if veval["counts"] != rep.Counts {
		t.Fatalf("counts = %v, want the report's own counts", veval["counts"])
	}
	if veval["report_id"] != rep.Identity.ReportID || veval["evidence_digest"] != rep.Provenance.EvidenceDigest ||
		veval["schema_version"] != "0.1.0" || veval["revision"] != rep.Identity.Artifact.Revision {
		t.Fatalf("identity fields missing from veval = %v", veval)
	}
}

func TestSARIFForensicFindingBecomesResult(t *testing.T) {
	t.Parallel()
	log := mustSARIF(t, loadReport(t, "extended-example"))
	onE6 := resultsFor(t, log, "E6")
	if len(onE6) != 2 {
		t.Fatalf("results on E6 = %d, want the criterion and its finding", len(onE6))
	}
	finding := onE6[1]
	if finding.Kind != "review" || finding.Level != "" {
		t.Fatalf("finding = kind %q level %q", finding.Kind, finding.Level)
	}
	want := map[string]any{
		"veval_finding_id":  "F1",
		"veval_detector":    "mtime-drift",
		"veval_severity":    "suspicious",
		"veval_disposition": "open",
	}
	for key, value := range want {
		if finding.Properties[key] != value {
			t.Fatalf("finding property %s = %v, want %v", key, finding.Properties[key], value)
		}
	}
	if !strings.HasPrefix(finding.Properties["veval_benign_alternative"].(string), "A checkout or a formatter") {
		t.Fatalf("benign alternative = %v", finding.Properties["veval_benign_alternative"])
	}
	if len(finding.Locations) != 1 || len(evidenceOf(t, finding)) != 1 {
		t.Fatalf("finding evidence was not mapped: %+v", finding)
	}
}

func TestSARIFRejectsResultOutsideTheSchema(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Criteria[0].Result = report.Result("PROBABLY")
	if _, err := export.ToSARIF(rep); err == nil {
		t.Fatal("a result with no SARIF kind must be an error, not a blank kind")
	}
}

func TestSARIFRejectsCriterionMissingFromContract(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Criteria[0].ID = "C9"
	if _, err := export.ToSARIF(rep); err == nil {
		t.Fatal("a result with no contract criterion has no rule and must be an error")
	}
}

func TestMarshalIsCanonicalJSON(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Contract.Criteria[0].Requirement = "a < b && c > d"
	out, err := export.Marshal(mustSARIF(t, rep))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasSuffix(out, []byte("\n")) {
		t.Fatal("canonical JSON ends with a newline")
	}
	if !bytes.Contains(out, []byte("\n  \"version\": \"2.1.0\"")) {
		t.Fatalf("canonical JSON is indented by two spaces:\n%s", out[:120])
	}
	if !bytes.Contains(out, []byte("a < b && c > d")) {
		t.Fatal("canonical JSON does not escape HTML")
	}
}

func TestSARIFGoldens(t *testing.T) {
	// Not parallel: the goldens are written under -update.
	for _, name := range []string{"worked-example", "extended-example"} {
		out, err := export.Marshal(mustSARIF(t, loadReport(t, name)))
		if err != nil {
			t.Fatal(err)
		}
		assertGolden(t, name+".sarif.json", out)
	}
}
