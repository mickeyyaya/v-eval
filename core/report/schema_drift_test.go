package report

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/mickeyyaya/v-eval/schema"
)

// The Go types in this package are the source of truth for the report; the
// published JSON Schema is a second statement of the same shape. These tests
// fail whenever the two drift apart: a field added to one and not the other,
// a requiredness that disagrees, or an enum value only one side knows.

// schemaDoc decodes the embedded JSON Schema document.
func schemaDoc(t *testing.T) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(schema.Report(), &doc); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}
	return doc
}

// node walks doc along keys, failing when a key is missing or not an object.
func node(t *testing.T, doc map[string]any, keys ...string) map[string]any {
	t.Helper()
	current := doc
	for i, key := range keys {
		next, ok := current[key].(map[string]any)
		if !ok {
			t.Fatalf("schema: %v is missing or not an object", keys[:i+1])
		}
		current = next
	}
	return current
}

// driftObject binds one schema object to the Go struct it must mirror.
type driftObject struct {
	name   string
	path   []string
	goType any
}

// driftObjects lists every schema object that has its own property set, with
// the Go type that defines it. TestSchemaCoversEveryObject keeps this list
// complete.
func driftObjects() []driftObject {
	return []driftObject{
		{"$", nil, Report{}},
		{"identity", []string{"properties", "identity"}, Identity{}},
		{"contract", []string{"properties", "contract"}, Contract{}},
		{"routing", []string{"properties", "routing"}, Routing{}},
		{"observations[]", []string{"properties", "observations", "items"}, Observation{}},
		{"claims[]", []string{"properties", "claims", "items"}, Claim{}},
		{"criteria[]", []string{"properties", "criteria", "items"}, CriterionResult{}},
		{"forensics[]", []string{"properties", "forensics", "items"}, Finding{}},
		{"dimensions[]", []string{"properties", "dimensions", "items"}, Dimension{}},
		{"counts", []string{"properties", "counts"}, Counts{}},
		{"status", []string{"properties", "status"}, Status{}},
		{"improvement[]", []string{"properties", "improvement", "items"}, Improvement{}},
		{"limitations", []string{"properties", "limitations"}, Limitations{}},
		{"provenance", []string{"properties", "provenance"}, Provenance{}},
		{"learning", []string{"properties", "learning"}, Learning{}},
		{"$defs.criterion", []string{"$defs", "criterion"}, Criterion{}},
		{"$defs.evidence", []string{"$defs", "evidence"}, Evidence{}},
		{"$defs.tally", []string{"$defs", "tally"}, Tally{}},
	}
}

// jsonTags returns every json tag name on a struct, and the subset declared
// without omitempty, both sorted.
func jsonTags(t *testing.T, goType any) (all, mandatory []string) {
	t.Helper()
	typ := reflect.TypeOf(goType)
	for i := 0; i < typ.NumField(); i++ {
		tag, ok := typ.Field(i).Tag.Lookup("json")
		if !ok || tag == "-" {
			t.Fatalf("%s.%s has no json tag", typ.Name(), typ.Field(i).Name)
		}
		name, options, _ := strings.Cut(tag, ",")
		all = append(all, name)
		if !strings.Contains(options, "omitempty") {
			mandatory = append(mandatory, name)
		}
	}
	sort.Strings(all)
	sort.Strings(mandatory)
	return all, mandatory
}

// keys returns the sorted key set of a schema object's named map.
func keys(t *testing.T, obj map[string]any, field string) []string {
	t.Helper()
	raw, ok := obj[field].(map[string]any)
	if !ok {
		t.Fatalf("schema object has no %s map", field)
	}
	names := make([]string, 0, len(raw))
	for name := range raw {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// stringList converts a schema string array, sorted when sortIt is set.
func stringList(t *testing.T, raw any, sortIt bool) []string {
	t.Helper()
	list, ok := raw.([]any)
	if !ok {
		t.Fatalf("schema: %v is not an array", raw)
	}
	out := make([]string, 0, len(list))
	for _, item := range list {
		text, ok := item.(string)
		if !ok {
			t.Fatalf("schema: %v is not a string", item)
		}
		out = append(out, text)
	}
	if sortIt {
		sort.Strings(out)
	}
	return out
}

func TestSchemaPropertiesMatchGoFields(t *testing.T) {
	t.Parallel()
	doc := schemaDoc(t)
	for _, object := range driftObjects() {
		t.Run(object.name, func(t *testing.T) {
			t.Parallel()
			obj := node(t, doc, object.path...)
			all, mandatory := jsonTags(t, object.goType)
			if got := keys(t, obj, "properties"); !reflect.DeepEqual(got, all) {
				t.Fatalf("schema properties %v, go json tags %v", got, all)
			}
			if got := stringList(t, obj["required"], true); !reflect.DeepEqual(got, mandatory) {
				t.Fatalf("schema required %v, go tags without omitempty %v", got, mandatory)
			}
		})
	}
}

// TestSchemaCoversEveryObject fails when the schema grows an object that no
// Go type is checked against, which would let that object drift unnoticed.
func TestSchemaCoversEveryObject(t *testing.T) {
	t.Parallel()
	doc := schemaDoc(t)
	covered := map[string]bool{}
	for _, object := range driftObjects() {
		covered[object.name] = true
	}

	for name, raw := range node(t, doc, "properties") {
		property, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("schema: properties.%s is not an object", name)
		}
		if key := objectKey(name, property); key != "" && !covered[key] {
			t.Fatalf("schema object %s is not bound to a Go type in driftObjects", key)
		}
	}
	for name, raw := range node(t, doc, "$defs") {
		definition, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("schema: $defs.%s is not an object", name)
		}
		if _, has := definition["properties"]; has && !covered["$defs."+name] {
			t.Fatalf("schema object $defs.%s is not bound to a Go type in driftObjects", name)
		}
	}
}

// objectKey names the driftObjects entry a schema property needs, or "" when
// the property carries no property set of its own (a scalar, enum, or $ref).
func objectKey(name string, property map[string]any) string {
	if _, ok := property["properties"]; ok {
		return name
	}
	if items, ok := property["items"].(map[string]any); ok {
		if _, ok := items["properties"]; ok {
			return name + "[]"
		}
	}
	return ""
}

func TestSchemaEnumsMatchGoConstants(t *testing.T) {
	t.Parallel()
	doc := schemaDoc(t)
	cases := []struct {
		name string
		path []string
		want []string
	}{
		{"result", []string{"$defs", "result"}, Result("").validValues()},
		{"method", []string{"$defs", "method"}, Method("").validValues()},
		{"isolation", []string{"$defs", "isolation"}, Isolation("").validValues()},
		{"status.overall", []string{"properties", "status", "properties", "overall"}, Overall("").validValues()},
		{"evidence.kind", []string{"$defs", "evidence", "properties", "kind"}, Kind("").validValues()},
		{"evidence.origin", []string{"$defs", "evidence", "properties", "origin"}, Origin("").validValues()},
		{"contract.status", []string{"properties", "contract", "properties", "status"}, ContractStatus("").validValues()},
		{"claims.status", []string{"properties", "claims", "items", "properties", "status"}, ClaimStatus("").validValues()},
		{"forensics.severity", []string{"properties", "forensics", "items", "properties", "severity"}, Severity("").validValues()},
		{"forensics.disposition", []string{"properties", "forensics", "items", "properties", "disposition"}, Disposition("").validValues()},
		{"dimensions.authority_type", []string{"properties", "dimensions", "items", "properties", "authority_type"}, Authority("").validValues()},
		{"observations.origin", []string{"properties", "observations", "items", "properties", "origin"}, ObservationOrigin("").validValues()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := stringList(t, node(t, doc, tc.path...)["enum"], false)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("schema enum %v, go constants %v", got, tc.want)
			}
		})
	}
}

// TestSchemaLocatorBranchesCoverEveryField checks the locator, whose shape the
// schema states as four exclusive branches rather than one property set.
func TestSchemaLocatorBranchesCoverEveryField(t *testing.T) {
	t.Parallel()
	doc := schemaDoc(t)
	branches, ok := node(t, doc, "$defs", "locator")["oneOf"].([]any)
	if !ok || len(branches) != 4 {
		t.Fatalf("locator oneOf = %v, want four branches", branches)
	}

	union := map[string]any{}
	for _, raw := range branches {
		branch, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("locator branch %v is not an object", raw)
		}
		for _, name := range keys(t, branch, "properties") {
			union[name] = struct{}{}
		}
	}

	got := keys(t, map[string]any{"properties": union}, "properties")
	all, _ := jsonTags(t, Locator{})
	if !reflect.DeepEqual(got, all) {
		t.Fatalf("locator branch properties %v, go json tags %v", got, all)
	}
}

func TestSchemaDeclaresPackageVersion(t *testing.T) {
	t.Parallel()
	doc := schemaDoc(t)
	got := node(t, doc, "properties", "identity", "properties", "schema_version")["const"]
	if got != schema.Version {
		t.Fatalf("identity.schema_version const = %v, want %q", got, schema.Version)
	}
}
