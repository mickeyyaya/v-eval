package report

import (
	"bytes"
	"os"
	"testing"
)

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
