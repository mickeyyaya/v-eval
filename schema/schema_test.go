package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/mickeyyaya/v-eval/schema"
)

func TestEmbeddedSchemaParsesAndDeclaresVersion(t *testing.T) {
	t.Parallel()
	var doc map[string]any
	if err := json.Unmarshal(schema.Report(), &doc); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}
	if got := doc["$id"]; got != "https://github.com/mickeyyaya/v-eval/schema/report/"+schema.Version {
		t.Fatalf("$id = %v, want it to end with %s", got, schema.Version)
	}
	if schema.Version != "0.1.0" {
		t.Fatalf("Version = %q", schema.Version)
	}
}

func TestReportReturnsIndependentCopy(t *testing.T) {
	t.Parallel()
	first := schema.Report()
	if len(first) == 0 {
		t.Fatalf("Report() returned no bytes")
	}
	original := first[0]
	first[0] = original + 1 // mutate the caller's copy

	second := schema.Report()
	if second[0] != original {
		t.Fatalf("Report() byte 0 = %v after caller mutation, want unchanged %v (Report must return a copy)", second[0], original)
	}
}
