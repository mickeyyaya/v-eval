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
	out, err := CanonicalJSON(normalize(rep))
	if err != nil {
		return nil, fmt.Errorf("report: encode: %w", err)
	}
	return out, nil
}

// CanonicalJSON writes any value in the canonical form a report takes:
// two-space indent, no HTML escaping, one trailing newline. It is exported so
// that anything written beside a report -- an export, a fixture -- goes
// through this encoder rather than a second one that could drift from it.
//
// It encodes the value it is given and nothing more: a nil slice or a nil map
// still writes as null. Turning those into the empty containers the schema
// asks for is normalize's job, which Encode does on the way in, so a caller
// writing something other than a report -- an export log, say -- is the one
// answerable for the nils it hands over.
func CanonicalJSON(v any) ([]byte, error) {
	return canonicalJSON(v, true)
}

// canonicalJSON is the one encoder every written report and every digest goes
// through. HTML escaping is off: a report quotes source, and "<" belongs in a
// report as "<", not as an escape a reader has to decode. json.Encoder ends
// its output with a newline, which is the trailing newline canonical form
// wants; the compact form carries it too, harmlessly, since a digest only
// needs the bytes to be the same bytes every time.
func canonicalJSON(v any, indent bool) ([]byte, error) {
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	if indent {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
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
// containers for nil ones.
//
// Report types must declare only exported fields: they mirror the JSON schema,
// and a field JSON cannot carry has no business in one. An unexported field
// added to any of them panics here rather than being skipped, because a field
// silently left out of canonical form would move every digest in the project
// without saying so.
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

// EvidenceDigest is the digest over every evidence array in a report, each
// tagged with the section and the path that cites it and taken in the
// document order evidenceSections defines: criteria, observations, claims,
// forensics, dimensions. Tagging is what makes it a digest of where the
// evidence sits and not merely of what it says, so moving a piece of evidence
// from a criterion to an observation changes it. It stands still when only a
// verdict changes, so two reports can be compared on what was seen rather
// than on what was concluded. Evidence JSON cannot represent yields the empty
// string, as ReportID does.
func EvidenceDigest(rep Report) string {
	raw, err := canonicalJSON(evidenceSections(normalize(rep)), false)
	if err != nil {
		return ""
	}
	return digest(raw)
}

// digest is the "sha256:<hex>" form every digest a report carries takes.
func digest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
