package report

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Decode parses a report, rejecting any field the schema does not define.
func Decode(raw []byte) (Report, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var rep Report
	if err := dec.Decode(&rep); err != nil {
		return Report{}, fmt.Errorf("report: decode: %w", err)
	}
	return rep, nil
}

// Encode writes a report in canonical form: two-space indent, trailing newline.
func Encode(rep Report) ([]byte, error) {
	out, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("report: encode: %w", err)
	}
	return append(out, '\n'), nil
}

// ReportID is the content digest a report identifies itself by.
//
// TODO(task-8): return "sha256:" + hex(sha256(Encode(rep with Identity.ReportID = ""))).
// Until then it returns the empty string, which is what a report that has not
// been aggregated carries, so identity.report_id can already be wired up.
func ReportID(rep Report) string { return "" }

// EvidenceDigest is the digest over every evidence array in a report.
//
// TODO(task-8): return "sha256:" + hex over the canonical JSON of the criteria,
// observations, claims, forensics, and dimensions evidence arrays in document
// order. Until then it returns the empty string, as ReportID does.
func EvidenceDigest(rep Report) string { return "" }
