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
