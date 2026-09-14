// Package schema embeds the published report contract. The Go types in core/report are the
// source of truth; core/report/schema_drift_test.go keeps this file equal to them.
package schema

import (
	"bytes"
	_ "embed"
)

// Version is the report schema version every valid report declares.
const Version = "0.1.0"

//go:embed report.schema.json
var raw []byte

// Report returns the JSON Schema document. Each call returns a fresh copy owned by the
// caller, so mutating the returned bytes cannot corrupt the embedded schema for later calls.
func Report() []byte { return bytes.Clone(raw) }
