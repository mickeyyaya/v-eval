package report

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
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

// Encode writes a report in canonical form: two-space indent, trailing
// newline, and empty containers where a report left them nil. Canonical form
// is what the digests are taken over, so a report built in Go and the same
// report decoded from JSON must encode to the same bytes.
func Encode(rep Report) ([]byte, error) {
	out, err := json.MarshalIndent(normalize(rep), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("report: encode: %w", err)
	}
	return append(out, '\n'), nil
}

// normalize returns a copy of rep in which every nil slice and nil map has
// become an empty one, so canonical JSON writes [] and {} where the schema
// says array and object, never null. Nil pointers stay nil: an absent
// dimension value and an absent learning section are states of their own,
// not empty containers.
//
// The walk is generic so that a field added to any report type is covered
// without a second edit here. It copies every slice, map, and pointer, so the
// result shares no memory with rep.
func normalize(rep Report) Report {
	out := reflect.New(reflect.TypeOf(rep)).Elem()
	normalizeInto(out, reflect.ValueOf(rep))
	return out.Interface().(Report)
}

// normalizeInto writes src into the settable value dst, substituting empty
// containers for nil ones. Report types are plain structs of exported fields,
// so every destination field is settable.
func normalizeInto(dst, src reflect.Value) {
	switch src.Kind() {
	case reflect.Slice:
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Len()))
		for i := 0; i < src.Len(); i++ {
			normalizeInto(dst.Index(i), src.Index(i))
		}
	case reflect.Map:
		dst.Set(reflect.MakeMapWithSize(src.Type(), src.Len()))
		for _, key := range src.MapKeys() {
			value := reflect.New(src.Type().Elem()).Elem()
			normalizeInto(value, src.MapIndex(key))
			dst.SetMapIndex(key, value)
		}
	case reflect.Pointer:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.New(src.Type().Elem()))
		normalizeInto(dst.Elem(), src.Elem())
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			if !src.Type().Field(i).IsExported() {
				continue
			}
			normalizeInto(dst.Field(i), src.Field(i))
		}
	default:
		dst.Set(src)
	}
}

// ReportID is the content digest a report identifies itself by: the digest of
// its own canonical JSON with the id field blanked, so that naming a report
// does not change what is named. Every other field is covered, which is why
// Aggregate computes it last.
//
// A report JSON cannot represent has no id and returns the empty string; only
// a NaN or infinite dimension value makes one, and identity.report_id then
// reports the mismatch rather than accepting a report with no name.
func ReportID(rep Report) string {
	rep.Identity.ReportID = ""
	raw, err := Encode(rep)
	if err != nil {
		return ""
	}
	return digest(raw)
}

// EvidenceDigest is the digest over every evidence array in a report, grouped
// by the entry that cites it and taken in the document order walkEvidence
// documents: criteria, observations, claims, forensics, dimensions. It moves
// when the evidence moves and stands still when only a verdict changes, so
// two reports can be compared on what was seen rather than what was concluded.
// Evidence JSON cannot represent yields the empty string, as ReportID does.
func EvidenceDigest(rep Report) string {
	rep = normalize(rep)
	lists := evidenceLists(rep)
	raw, err := json.Marshal(lists)
	if err != nil {
		return ""
	}
	return digest(raw)
}

// evidenceLists collects a report's evidence arrays, one per citing entry, in
// document order. Grouping is kept rather than flattened so that moving a
// piece of evidence from one criterion to another changes the digest.
func evidenceLists(rep Report) [][]Evidence {
	lists := make([][]Evidence, 0,
		len(rep.Criteria)+len(rep.Observations)+len(rep.Claims)+len(rep.Forensics)+len(rep.Dimensions))
	for _, result := range rep.Criteria {
		lists = append(lists, result.Evidence)
	}
	for _, observation := range rep.Observations {
		lists = append(lists, observation.Evidence)
	}
	for _, claim := range rep.Claims {
		lists = append(lists, claim.Verification)
	}
	for _, finding := range rep.Forensics {
		lists = append(lists, finding.Evidence)
	}
	for _, dimension := range rep.Dimensions {
		lists = append(lists, dimension.Evidence)
	}
	return lists
}

// digest is the "sha256:<hex>" form every digest a report carries takes.
func digest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
