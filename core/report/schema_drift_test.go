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

// driftTypes binds every schema object that declares its own property set to
// the Go struct it must mirror. A key is the object's location in the schema:
// "$" is the report itself, a nested object is dotted onto its parent, "[]"
// marks an array's items, and "$defs.x" is a named definition. Objects reached
// through a $ref are bound once, under the definition they come from.
func driftTypes() map[string]any {
	return map[string]any{
		"$":                               Report{},
		"identity":                        Identity{},
		"identity.artifact":               Artifact{},
		"identity.task":                   Task{},
		"identity.host":                   Host{},
		"contract":                        Contract{},
		"contract.conflicts[]":            Conflict{},
		"routing":                         Routing{},
		"routing.supplied[]":              SuppliedSource{},
		"routing.profiles[]":              ProfileRef{},
		"routing.adapters_run[]":          AdapterRun{},
		"routing.adapters_skipped[]":      AdapterSkipped{},
		"routing.ambiguity[]":             Ambiguity{},
		"observations[]":                  Observation{},
		"claims[]":                        Claim{},
		"criteria[]":                      CriterionResult{},
		"forensics[]":                     Finding{},
		"dimensions[]":                    Dimension{},
		"counts":                          Counts{},
		"counts.coverage":                 Coverage{},
		"status":                          Status{},
		"improvement[]":                   Improvement{},
		"limitations":                     Limitations{},
		"provenance":                      Provenance{},
		"provenance.tools[]":              ToolRef{},
		"provenance.commands[]":           CommandRecord{},
		"provenance.environment":          Environment{},
		"learning":                        Learning{},
		"learning.precedents_retrieved[]": PrecedentRef{},
		"$defs.criterion":                 Criterion{},
		"$defs.evidence":                  Evidence{},
		"$defs.evidence.provenance":       EvidenceProvenance{},
		"$defs.tally":                     Tally{},
	}
}

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

// schemaObject is one object found in the schema, with where it was found.
type schemaObject struct {
	key string
	obj map[string]any
}

// collectObjects returns every schema object that declares its own property
// set: the report, each named definition, and everything nested inside them,
// array items included. An object with no "properties" is a leaf, which is
// what an enum, a $ref to a definition collected in its own right, and a
// free-form map such as runtime_versions all are.
func collectObjects(t *testing.T, doc map[string]any) []schemaObject {
	t.Helper()
	var found []schemaObject

	var visit func(key string, obj map[string]any)
	visit = func(key string, obj map[string]any) {
		found = append(found, schemaObject{key: key, obj: obj})
		properties := node(t, obj, "properties")
		for _, name := range sortedKeys(properties) {
			property, ok := properties[name].(map[string]any)
			if !ok {
				t.Fatalf("schema: %s.%s is not an object", key, name)
			}
			if child, suffix := objectOf(property); child != nil {
				visit(childKey(key, name+suffix), child)
			}
		}
	}

	visit("$", doc)
	for _, name := range sortedKeys(node(t, doc, "$defs")) {
		definition := node(t, doc, "$defs", name)
		if _, has := definition["properties"]; has {
			visit("$defs."+name, definition)
		}
	}
	return found
}

// objectOf returns the object a schema property declares inline — itself, or
// its array items — with the suffix its key carries, or nil for a leaf.
func objectOf(property map[string]any) (map[string]any, string) {
	if _, ok := property["properties"]; ok {
		return property, ""
	}
	if items, ok := property["items"].(map[string]any); ok {
		if _, ok := items["properties"]; ok {
			return items, "[]"
		}
	}
	return nil, ""
}

// childKey names an object nested inside another: a property of the report
// keeps its own name, anything deeper is dotted onto its parent's key.
func childKey(parent, name string) string {
	if parent == "$" {
		return name
	}
	return parent + "." + name
}

// jsonTags returns every json tag name on a struct, and the subset declared
// without omitempty, both sorted and never nil.
func jsonTags(t *testing.T, goType any) (all, mandatory []string) {
	t.Helper()
	all, mandatory = []string{}, []string{}
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

// sortedKeys returns a schema map's key set, sorted.
func sortedKeys(raw map[string]any) []string {
	names := make([]string, 0, len(raw))
	for name := range raw {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// requiredNames returns an object's required list, empty when it states none.
func requiredNames(t *testing.T, obj map[string]any) []string {
	t.Helper()
	raw, ok := obj["required"]
	if !ok {
		return []string{}
	}
	return stringList(t, raw, true)
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
	types := driftTypes()
	for _, object := range collectObjects(t, schemaDoc(t)) {
		t.Run(object.key, func(t *testing.T) {
			t.Parallel()
			goType, ok := types[object.key]
			if !ok {
				t.Fatalf("schema object %s is not bound to a Go type in driftTypes", object.key)
			}
			all, mandatory := jsonTags(t, goType)
			if got := sortedKeys(node(t, object.obj, "properties")); !reflect.DeepEqual(got, all) {
				t.Fatalf("schema properties %v, go json tags %v", got, all)
			}
			if got := requiredNames(t, object.obj); !reflect.DeepEqual(got, mandatory) {
				t.Fatalf("schema required %v, go tags without omitempty %v", got, mandatory)
			}
		})
	}
}

// TestSchemaCoversEveryObject fails when the schema grows an object that no Go
// type is checked against, which would let that object drift unnoticed, and
// when driftTypes keeps a binding the schema no longer has.
func TestSchemaCoversEveryObject(t *testing.T) {
	t.Parallel()
	types := driftTypes()
	found := map[string]bool{}
	var unmapped []string
	for _, object := range collectObjects(t, schemaDoc(t)) {
		found[object.key] = true
		if _, ok := types[object.key]; !ok {
			unmapped = append(unmapped, object.key)
		}
	}
	if len(unmapped) > 0 {
		sort.Strings(unmapped)
		t.Errorf("schema objects not bound to a Go type in driftTypes: %v", unmapped)
	}

	var stale []string
	for key := range types {
		if !found[key] {
			stale = append(stale, key)
		}
	}
	if len(stale) > 0 {
		sort.Strings(stale)
		t.Errorf("driftTypes binds objects the schema does not contain: %v", stale)
	}
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
// schema states as four exclusive branches rather than one property set. Each
// branch is partial by design, so only their union is compared.
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
		for _, name := range sortedKeys(node(t, branch, "properties")) {
			union[name] = struct{}{}
		}
	}

	all, _ := jsonTags(t, Locator{})
	if got := sortedKeys(union); !reflect.DeepEqual(got, all) {
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
