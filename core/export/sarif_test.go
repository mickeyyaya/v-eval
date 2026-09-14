package export_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
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

// vevalOf returns the run.properties.veval map a mapped run carries.
func vevalOf(t *testing.T, run export.Run) map[string]any {
	t.Helper()
	veval, ok := run.Properties["veval"].(map[string]any)
	if !ok {
		t.Fatalf("run properties = %v", run.Properties)
	}
	return veval
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
	veval := vevalOf(t, log.Runs[0])
	if veval["overall"] != "FAIL" {
		t.Fatal("run.properties.veval.overall must carry the overall status")
	}
}

func TestSARIFErrorCriterionGetsAnInvocationOfItsOwn(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Criteria[4].Result = report.ResultError
	rep.Provenance.Commands = []report.CommandRecord{{Command: "go test ./...", Cwd: ".", ExitStatus: 2, StartedAt: "2026-09-14T00:00:00Z", EndedAt: "2026-09-14T00:00:01Z", LogRef: "logs/1.txt", Isolation: report.IsolationNone}}
	log, err := export.ToSARIF(rep)
	if err != nil {
		t.Fatal(err)
	}
	inv := log.Runs[0].Invocations
	if len(inv) != 2 {
		t.Fatalf("invocations = %+v, want the command and a synthesized carrier", inv)
	}
	if inv[0].CommandLine != "go test ./..." || inv[0].ExecutionSuccessful ||
		len(inv[0].ToolExecutionNotifications) != 0 {
		t.Fatalf("the command reports its own exit status and nothing else: %+v", inv[0])
	}
	if inv[1].CommandLine != "" || inv[1].ExecutionSuccessful ||
		len(inv[1].ToolExecutionNotifications) != 1 {
		t.Fatalf("synthesized invocation = %+v", inv[1])
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
	// E2 cites a passage and a file; the file one is also a location, and both
	// are listed in properties.evidence so the record of what was seen is
	// complete.
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
}

func TestSARIFFileEvidenceObjectIsSelfContained(t *testing.T) {
	t.Parallel()
	// A reader holding one evidence object must learn where it points from that
	// object alone, without counting positions in locations[].
	got := resultsFor(t, mustSARIF(t, loadReport(t, "extended-example")), "E2")[0]
	file := evidenceOf(t, got)[1]
	if file["file"] != "examples/extended/service.go" || file["line_start"] != 48 || file["line_end"] != 52 {
		t.Fatalf("file evidence = %v, want its own file and line range", file)
	}
	provenance, ok := file["provenance"].(map[string]any)
	if !ok {
		t.Fatalf("file evidence carries no provenance: %v", file)
	}
	if provenance["tool"] != "assistant" || provenance["version"] != "fixture" ||
		provenance["revision"] != "git:9f2c1a4e5b6d7c8f9a0b1c2d3e4f5a6b7c8d9e0f" ||
		provenance["timestamp"] != "2026-09-14T00:04:00Z" {
		t.Fatalf("provenance = %v, want what produced the evidence and when", provenance)
	}
	if len(got.Locations) != 1 {
		t.Fatalf("locations = %+v, want the file locator kept as a physicalLocation too", got.Locations)
	}
}

func TestSARIFJudgmentEvidenceCarriesRubricAndModel(t *testing.T) {
	t.Parallel()
	// E6 cites a rubric judgment; the rubric version and the model are what
	// produced it, and a reader cannot weigh the judgment without them.
	got := resultsFor(t, mustSARIF(t, loadReport(t, "extended-example")), "E6")[0]
	judgment := evidenceOf(t, got)[1]
	if judgment["rubric_version"] != "integrity-rubric-0.2" || judgment["model"] != "unknown" {
		t.Fatalf("judgment evidence = %v, want the rubric version and the model the report states", judgment)
	}
	provenance, ok := judgment["provenance"].(map[string]any)
	if !ok {
		t.Fatalf("judgment evidence carries no provenance: %v", judgment)
	}
	if _, ok := provenance["rubric_version"]; ok {
		t.Fatalf("provenance = %v, want only what produced the evidence and when", provenance)
	}
}

func TestSARIFEvidenceOmitsAbsentRubricAndModel(t *testing.T) {
	t.Parallel()
	// Inspection evidence names no rubric and no model, and an empty string
	// would read as one that is unknown rather than one that does not apply.
	got := resultsFor(t, mustSARIF(t, loadFixture(t)), "C1")[0]
	entry := evidenceOf(t, got)[0]
	if _, ok := entry["rubric_version"]; ok {
		t.Fatalf("evidence entry = %v, want no empty rubric_version", entry)
	}
	if _, ok := entry["model"]; ok {
		t.Fatalf("evidence entry = %v, want no empty model", entry)
	}
}

func TestSARIFIncompleteLocatorPointsNowhere(t *testing.T) {
	t.Parallel()
	// The export assumes a report that passed report.Validate, where a locator
	// is one complete group. A locator that is not says where nothing is, and
	// the export states the evidence without inventing a place for it.
	rep := oneCriterionReport(report.Evidence{
		Kind:        report.KindInspection,
		Locator:     report.Locator{File: "service.go"}, // no line range: no complete shape
		Observation: "half a locator",
		Origin:      report.OriginObserved,
		Isolation:   report.IsolationNone,
	})
	got := resultsFor(t, mustSARIF(t, rep), "C1")[0]
	if len(got.Locations) != 0 {
		t.Fatalf("locations = %+v, want none for a locator of no complete shape", got.Locations)
	}
	entry := evidenceOf(t, got)[0]
	want := []string{"kind", "origin", "isolation", "observation", "provenance"}
	if len(entry) != len(want) {
		t.Fatalf("evidence entry = %v, want only %v", entry, want)
	}
	for _, key := range want {
		if _, ok := entry[key]; !ok {
			t.Fatalf("evidence entry %v is missing %q", entry, key)
		}
	}
}

func TestSARIFCommandsBecomeInvocations(t *testing.T) {
	t.Parallel()
	log := mustSARIF(t, loadReport(t, "extended-example"))
	inv := log.Runs[0].Invocations
	if len(inv) != 4 {
		t.Fatalf("invocations = %d, want one per command plus the errored criterion's own", len(inv))
	}
	first := inv[0]
	if first.CommandLine != "go test ./..." || first.WorkingDirectory.URI != "/tmp/veval-worktree" ||
		first.ExitCode == nil || *first.ExitCode != 0 ||
		first.StartTimeUTC != "2026-09-14T00:03:10Z" || first.EndTimeUTC != "2026-09-14T00:03:12Z" {
		t.Fatalf("invocation = %+v", first)
	}
	if first.Properties["veval_isolation"] != "worktree" || first.Properties["veval_log_ref"] != "logs/e1-go-test.log" {
		t.Fatalf("invocation properties = %v", first.Properties)
	}
	// Each command reports its own truth: go test exited 0, and both docker
	// commands exited 127.
	if !inv[0].ExecutionSuccessful {
		t.Fatalf("the command that exited 0 must report success: %+v", inv[0])
	}
	for _, one := range inv[1:3] {
		if one.ExecutionSuccessful {
			t.Fatalf("a command that exited 127 must not report success: %+v", one)
		}
		if len(one.ToolExecutionNotifications) != 0 {
			t.Fatalf("a command carries no notification about a criterion: %+v", one)
		}
	}
	// E4 is an ERROR criterion: its notification sits on an invocation of its
	// own, appended last, which names no command because none of them errored.
	last := inv[len(inv)-1]
	if last.CommandLine != "" || last.ExecutionSuccessful || len(last.ToolExecutionNotifications) != 1 {
		t.Fatalf("synthesized invocation = %+v", last)
	}
	got := last.ToolExecutionNotifications[0]
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

func TestSARIFGitRevisionStaysInProperties(t *testing.T) {
	t.Parallel()
	// SARIF requires a repositoryUri on every versionControlProvenance entry
	// and a report carries none, so a git revision is stated in the property
	// bag rather than in an entry no validator would accept.
	log := mustSARIF(t, loadReport(t, "extended-example"))
	veval := vevalOf(t, log.Runs[0])
	if veval["vcs_revision_id"] != "9f2c1a4e5b6d7c8f9a0b1c2d3e4f5a6b7c8d9e0f" {
		t.Fatalf("vcs_revision_id = %v", veval["vcs_revision_id"])
	}
	if veval["revision"] != "git:9f2c1a4e5b6d7c8f9a0b1c2d3e4f5a6b7c8d9e0f" {
		t.Fatalf("revision = %v, want the revision as the report states it", veval["revision"])
	}
	if len(log.Runs[0].Artifacts) != 0 {
		t.Fatalf("a git revision carries no artifact hash: %+v", log.Runs[0].Artifacts)
	}
	out, err := export.Marshal(log)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(out, []byte("versionControlProvenance")) {
		t.Fatal("versionControlProvenance cannot be completed from a report and must not be written")
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
	if _, ok := vevalOf(t, mustSARIF(t, loadFixture(t)).Runs[0])["vcs_revision_id"]; ok {
		t.Fatal("a content revision is not a version control revision id")
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

// TestSARIFIncompleteContentDigestYieldsNoArtifact covers the revision that
// names the right kind of identity without carrying one: an artifacts entry
// built from it would state a hash nothing hashes to.
func TestSARIFIncompleteContentDigestYieldsNoArtifact(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"empty digest":    "content:sha256:",
		"short digest":    "content:sha256:befaf705",
		"upper case":      "content:sha256:BEFAF705C894D30D62D3D53DD5E6653C910063747FC7E508C1B9DE2DE306DFA8",
		"not hexadecimal": "content:sha256:zzzzf705c894d30d62d3d53dd5e6653c910063747fc7e508c1b9de2de306dfa8",
	}
	for name, revision := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rep := loadFixture(t)
			rep.Identity.Artifact.Revision = revision
			run := mustSARIF(t, rep).Runs[0]
			if len(run.Artifacts) != 0 {
				t.Fatalf("artifacts = %+v, want none for a digest that pins nothing", run.Artifacts)
			}
			if vevalOf(t, run)["revision"] != revision {
				t.Fatalf("the revision must still be stated as the report states it: %v", vevalOf(t, run))
			}
		})
	}
}

func TestSARIFUnrecognizedRevisionStaysInProperties(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Identity.Artifact.Revision = "svn:r1234"
	run := mustSARIF(t, rep).Runs[0]
	if len(run.Artifacts) != 0 {
		t.Fatalf("an unrecognized revision must not be reshaped: %+v", run.Artifacts)
	}
	veval := vevalOf(t, run)
	if veval["revision"] != "svn:r1234" {
		t.Fatalf("veval properties = %v", veval)
	}
	if _, ok := veval["vcs_revision_id"]; ok {
		t.Fatalf("an unrecognized revision is not a git revision id: %v", veval["vcs_revision_id"])
	}
}

func TestSARIFRunPropertiesCarryTheVerdict(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	veval := vevalOf(t, mustSARIF(t, rep).Runs[0])
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
	if _, ok := finding.Properties["veval_result"]; ok {
		t.Fatalf("a finding is not a criterion verdict and states no result: %v", finding.Properties)
	}
}

// assertMappingError requires a mapping to fail, naming the package and the
// operation before the detail, so that an error read far from here says which
// export refused and why.
func assertMappingError(t *testing.T, rep report.Report, wantPrefix string) {
	t.Helper()
	_, err := export.ToSARIF(rep)
	if err == nil {
		t.Fatalf("mapping must fail with an error starting %q", wantPrefix)
	}
	if !strings.HasPrefix(err.Error(), wantPrefix) {
		t.Errorf("error = %q, want it to start %q", err, wantPrefix)
	}
}

func TestSARIFRejectsResultOutsideTheSchema(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Criteria[0].Result = report.Result("PROBABLY")
	// A result with no SARIF kind is an error, not a blank kind.
	assertMappingError(t, rep, `export: sarif: criterion "C1": result "PROBABLY"`)
}

func TestSARIFRejectsCriterionMissingFromContract(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Criteria[0].ID = "C9"
	// A result with no contract criterion has no rule to report against.
	assertMappingError(t, rep, `export: sarif: criterion "C9":`)
}

// blockedByOf returns the blocked_by list a run states.
func blockedByOf(t *testing.T, log export.Log) []string {
	t.Helper()
	veval := vevalOf(t, log.Runs[0])
	blocked, ok := veval["blocked_by"].([]string)
	if !ok {
		t.Fatalf("blocked_by = %#v, want a list", veval["blocked_by"])
	}
	return blocked
}

func TestSARIFBlockedByIsNeverNull(t *testing.T) {
	t.Parallel()
	// A report built in Go, not decoded from JSON, can leave the list nil.
	rep := report.Report{Status: report.Status{Overall: report.OverallPass}}
	if rep.Status.BlockedBy != nil {
		t.Fatal("this test is only meaningful while a zero Status leaves BlockedBy nil")
	}
	log := mustSARIF(t, rep)
	if blocked := blockedByOf(t, log); blocked == nil || len(blocked) != 0 {
		t.Fatalf("blocked_by = %#v, want an empty list", blocked)
	}
	out, err := export.Marshal(log)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out, []byte(`"blocked_by": []`)) {
		t.Fatalf("blocked_by must serialize as an empty array:\n%s", out)
	}
}

func TestSARIFRejectsFindingMissingFromContract(t *testing.T) {
	t.Parallel()
	rep := loadReport(t, "extended-example")
	rep.Forensics[0].CriterionID = "E9"
	// A finding on no contract criterion has no rule to report against.
	assertMappingError(t, rep, `export: sarif: finding "F1":`)
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

// oneCriterionReport is the smallest report that carries one piece of
// evidence: one required criterion, failed, citing it.
func oneCriterionReport(evidence report.Evidence) report.Report {
	return report.Report{
		Contract: report.Contract{Criteria: []report.Criterion{{ID: "C1", Requirement: "r", Required: true}}},
		Criteria: []report.CriterionResult{{ID: "C1", Result: report.ResultFail, Evidence: []report.Evidence{evidence}}},
	}
}

// assertNoJSONNull decodes out and walks every map and slice looking for a
// JSON null. Matching on the token itself, rather than the bytes "null",
// means a passage of prose that happens to contain the word is never
// mistaken for the hole a consumer would otherwise have to guess at: an
// empty list and a list nobody wrote read the same once they are both null,
// so a log must state the empty container instead.
func assertNoJSONNull(t *testing.T, name string, out []byte) {
	t.Helper()
	var decoded any
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	walkForNull(t, name, "$", decoded)
}

// walkForNull recurses through a decoded JSON value, failing the test at the
// first null it finds, named by its path.
func walkForNull(t *testing.T, name, path string, v any) {
	t.Helper()
	switch value := v.(type) {
	case nil:
		t.Errorf("%s: %s is null, where an empty container says what is meant", name, path)
	case map[string]any:
		for key, child := range value {
			walkForNull(t, name, path+"."+key, child)
		}
	case []any:
		for i, child := range value {
			walkForNull(t, name, fmt.Sprintf("%s[%d]", path, i), child)
		}
	}
}

func TestSARIFGoldens(t *testing.T) {
	// Not parallel: the goldens are written under -update.
	for _, name := range []string{"worked-example", "extended-example"} {
		out, err := export.Marshal(mustSARIF(t, loadReport(t, name)))
		if err != nil {
			t.Fatal(err)
		}
		assertNoJSONNull(t, name, out)
		assertGolden(t, name+".sarif.json", out)
	}
}
