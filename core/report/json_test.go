package report

import (
	"bytes"
	"os"
	"testing"
)

func TestEncodeWritesWhatCanonicalJSONWrites(t *testing.T) {
	t.Parallel()
	// One encoder, so a report and anything exported beside it cannot drift
	// apart on indentation, escaping, or the trailing newline.
	rep := loadFixture(t)
	encoded, err := Encode(rep)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := CanonicalJSON(rep)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, canonical) {
		t.Fatalf("Encode and CanonicalJSON disagree:\n%s\n%s", encoded, canonical)
	}
}

func TestFixtureDecodesAndEncodesCanonically(t *testing.T) {
	t.Parallel()
	rep := loadFixture(t)
	if len(rep.Criteria) != 5 || rep.Criteria[3].ID != "C4" || rep.Criteria[3].Result != ResultPass || rep.Criteria[4].Result != ResultUnknown {
		t.Fatalf("unexpected criteria: %+v", rep.Criteria)
	}
	once, err := Encode(rep)
	if err != nil {
		t.Fatal(err)
	}
	again, _ := Encode(rep)
	if !bytes.Equal(once, again) {
		t.Fatal("Encode is not deterministic")
	}
	raw, _ := os.ReadFile("testdata/worked-example.json")
	if !bytes.Equal(once, raw) {
		t.Fatalf("fixture is not in canonical form; write Encode output back to testdata")
	}
}

func TestDecodeRejectsUnknownFields(t *testing.T) {
	t.Parallel()
	if _, err := Decode([]byte(`{"identity":{"unexpected":1}}`)); err == nil {
		t.Fatal("unknown field must be rejected")
	}
}

// TestEncodeLeavesHTMLCharactersAlone pins canonical form against Go's
// default HTML escaping: a report quotes source, and the angle brackets and
// ampersands it quotes must survive encoding as themselves rather than as
// numeric escapes. The digests are taken over these bytes, so escaping would
// otherwise be baked into every report id a reader has to compare by eye.
func TestEncodeLeavesHTMLCharactersAlone(t *testing.T) {
	t.Parallel()
	const quoted = "the template writes <b> & </b> raw"
	rep := loadFixture(t)
	rep.Criteria[0].Reasoning = quoted
	raw, err := Encode(rep)
	if err != nil {
		t.Fatal(err)
	}
	// Escaping any of the three characters would leave the literal absent.
	if !bytes.Contains(raw, []byte(quoted)) {
		t.Fatalf("canonical JSON must keep angle brackets and ampersands literal:\n%s", raw)
	}
}

// TestEvidenceDigestDistinguishesSections requires the digest to record which
// section cited a piece of evidence: the same evidence under criteria and
// under observations are two different claims about a report.
func TestEvidenceDigestDistinguishesSections(t *testing.T) {
	t.Parallel()
	evidence := []Evidence{{Kind: KindInspection, Origin: OriginObserved, Observation: "read the file",
		Locator: Locator{File: "main.go", LineStart: 1, LineEnd: 2}}}

	underCriteria := Report{Criteria: []CriterionResult{{ID: "C1", Evidence: evidence}}}
	underObservations := Report{Observations: []Observation{{ID: "O1", Evidence: evidence}}}

	if EvidenceDigest(underCriteria) == EvidenceDigest(underObservations) {
		t.Fatal("the same evidence in two sections must not digest the same")
	}
}

// TestEncodeNormalisesNestedNilEvidence covers normalisation below the top
// level: a criterion result built in Go with no evidence must still write an
// empty array, as the schema requires.
func TestEncodeNormalisesNestedNilEvidence(t *testing.T) {
	t.Parallel()
	raw, err := Encode(Report{Criteria: []CriterionResult{{ID: "C1"}}})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte(`"evidence": []`)) {
		t.Fatalf("nested nil evidence must encode as []:\n%s", raw)
	}
	if bytes.Contains(raw, []byte(`"evidence": null`)) {
		t.Fatalf("nested nil evidence must not encode as null:\n%s", raw)
	}
}
