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
