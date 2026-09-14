package report

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mickeyyaya/v-eval/schema"
)

func TestAggregateFillsDerivedFieldsAndIsIdempotent(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	rep.Counts, rep.Status, rep.Identity.ReportID, rep.Provenance.EvidenceDigest = Counts{}, Status{}, "", ""
	first := Aggregate(rep)
	if first.Status.Overall != OverallFail || first.Counts.Required.Fail != 3 {
		t.Fatalf("aggregate: %+v %+v", first.Status, first.Counts)
	}
	if !strings.HasPrefix(first.Identity.ReportID, "sha256:") || !strings.HasPrefix(first.Provenance.EvidenceDigest, "sha256:") {
		t.Fatalf("digests: %q %q", first.Identity.ReportID, first.Provenance.EvidenceDigest)
	}
	if second := Aggregate(first); second.Identity.ReportID != first.Identity.ReportID {
		t.Fatal("Aggregate is not idempotent")
	}
	changed := first
	changed.Criteria = append([]CriterionResult(nil), first.Criteria...)
	changed.Criteria[3].Result = ResultFail
	if Aggregate(changed).Identity.ReportID == first.Identity.ReportID {
		t.Fatal("report_id must change when a result changes")
	}
	if Aggregate(changed).Provenance.EvidenceDigest != first.Provenance.EvidenceDigest {
		t.Fatal("evidence digest must not change when only a result changes")
	}
}

// minimalReport is the smallest report that breaks no structural rule: an
// empty contract, judged by no criteria. Every slice and map in it is nil,
// which is the point of the test below.
func minimalReport() Report {
	return Report{
		Identity: Identity{
			SchemaVersion: schema.Version,
			VevalVersion:  "0.1.0",
			SkillRevision: "skill-rev-1",
			Artifact:      Artifact{Kind: "repository", Revision: "0000000"},
			Task:          Task{RequestedOutcome: "evaluate an empty contract"},
			CreatedAt:     "2026-09-14T00:00:00Z",
			Host:          Host{CLI: "v-eval", OS: "darwin", Arch: "arm64"},
		},
		Contract: Contract{ContractID: "K-empty", ContractVersion: "1", Status: ContractStatusUserSpecified},
		Routing:  Routing{Rationale: "nothing was supplied, so no adapter ran"},
	}
}

// TestAggregateNormalisesNilContainers pins the shape of canonical JSON for a
// report built in Go rather than decoded: a nil slice is an empty array and a
// nil map an empty object, never null, and the result still validates.
func TestAggregateNormalisesNilContainers(t *testing.T) {
	t.Parallel()
	raw, err := Encode(Aggregate(minimalReport()))
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if bytes.Contains(raw, []byte("null")) {
		t.Fatalf("canonical JSON must not contain null:\n%s", raw)
	}
	if !bytes.Contains(raw, []byte("[]")) || !bytes.Contains(raw, []byte("{}")) {
		t.Fatalf("canonical JSON must write empty containers as [] and {}:\n%s", raw)
	}
	_, violations, err := Validate(raw)
	if err != nil || len(violations) != 0 {
		t.Fatalf("err=%v violations=%v", err, violations)
	}
}
