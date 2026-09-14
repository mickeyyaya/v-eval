package report

import (
	"os"
	"testing"
)

// loadFixture decodes the worked-example report used across the report tests.
func loadFixture(t *testing.T) Report {
	t.Helper()
	raw, err := os.ReadFile("testdata/worked-example.json")
	if err != nil {
		t.Fatal(err)
	}
	rep, err := Decode(raw)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return rep
}
