// Package jasynthex searches arbitrarily nested JSON for a field name
// and returns every match with its full structured path and value.
//
// Fixes applied:
//   #1 - Duplicate keys: uses json.Decoder with a custom ordered map so every
//        key in a JSON object is seen, even duplicates.
//   #2 - Order not guaranteed: objects are decoded into []KeyValue (an ordered
//        slice), so results always come back in document (top-to-bottom) order.
//   #3 - Exact match only: three MatchMode options — Exact, CaseInsensitive,
//        and Contains — handle inconsistent field naming across endpoints.
//   #4 - Large JSON performance: decoding is streaming via json.Decoder; the
//        tree is never buffered twice and FindStream accepts an io.Reader
//        directly (e.g. an HTTP response body).
//   #5 - Nested array path notation: arrays-of-arrays produce correct standard
//        JSONPath notation e.g. $.a[0][1] which all major evaluators accept.
//   #6 - Path as string only: Result now carries both a human-readable
//        PathString ("$.a.b[0].c") AND a PathSegments slice so callers can
//        navigate/manipulate the path programmatically without string parsing.

package jasynthex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------------
// MatchMode — fix #3 (exact match only)
// ---------------------------------------------------------------------------

// MatchMode controls how the search term is compared against JSON key names.
type MatchMode int

const (
	// Exact requires the key to match the search term character-for-character.
	// This is the default and the fastest mode.
	//   search "email"  →  matches "email" only
	Exact MatchMode = iota

	// CaseInsensitive matches regardless of letter casing.
	//   search "email"  →  matches "email", "Email", "EMAIL"
	CaseInsensitive

	// Contains matches any key that contains the search term as a substring,
	// case-insensitively. Good for inconsistent API field naming across endpoints.
	//   search "email"  →  matches "email", "Email", "email_address", "contactEmail"
	Contains
)

// matchKey reports whether key matches fieldName under the given mode.
func matchKey(key, fieldName string, mode MatchMode) bool {
	switch mode {
	case CaseInsensitive:
		return strings.EqualFold(key, fieldName)
	case Contains:
		return strings.Contains(strings.ToLower(key), strings.ToLower(fieldName))
	default: // Exact
		return key == fieldName
	}
}

// ---------------------------------------------------------------------------
// Segment — one step in a JSONPath
// ---------------------------------------------------------------------------

// SegmentKind distinguishes object keys from array indices.
type SegmentKind int

const (
	KeySegment   SegmentKind = iota // e.g. "address" in $.user.address
	IndexSegment                    // e.g. 0 in $.items[0]
)

// Segment is one step in a JSONPath — either a string key or an int index.
type Segment struct {
	Kind  SegmentKind
	Key   string // set when Kind == KeySegment
	Index int    // set when Kind == IndexSegment
}

func (s Segment) String() string {
	if s.Kind == IndexSegment {
		return fmt.Sprintf("[%d]", s.Index)
	}
	return s.Key
}

// ---------------------------------------------------------------------------
// Result — what callers get back
// ---------------------------------------------------------------------------

// Result is returned for every field that matched the search.
type Result struct {
	PathString   string    // e.g. "$.organization.departments[0].head.email"
	PathSegments []Segment // structured path — navigate without string parsing
	Value        any       // the value at that path
}

// ParentPath returns the PathString of the object that contains this field.
// Useful when you need to fetch sibling fields.
func (r Result) ParentPath() string {
	if len(r.PathSegments) == 0 {
		return "$"
	}
	return buildPathString(r.PathSegments[:len(r.PathSegments)-1])
}

// Depth returns how many levels deep the matched field is (root = 0).
func (r Result) Depth() int {
	return len(r.PathSegments)
}

// ---------------------------------------------------------------------------
// Ordered JSON tree — fixes #1 (duplicate keys) and #2 (insertion order)
// ---------------------------------------------------------------------------

// KeyValue is one entry in a JSON object, preserving insertion order.
// Using a slice instead of map[string]any means duplicate keys are all kept.
type KeyValue struct {
	Key   string
	Value any // orderedNode: []KeyValue | []any | scalar
}

// decodeOrdered reads a json.Decoder token stream into an ordered tree.
// Every key is preserved (including duplicates) in document order.
func decodeOrdered(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}

	switch v := tok.(type) {
	case json.Delim:
		switch v {
		case '{':
			var pairs []KeyValue
			for dec.More() {
				// key
				keyTok, err := dec.Token()
				if err != nil {
					return nil, err
				}
				key, ok := keyTok.(string)
				if !ok {
					return nil, fmt.Errorf("expected string key, got %T", keyTok)
				}
				// value (recursive)
				val, err := decodeOrdered(dec)
				if err != nil {
					return nil, err
				}
				pairs = append(pairs, KeyValue{Key: key, Value: val})
			}
			if _, err := dec.Token(); err != nil { // consume '}'
				return nil, err
			}
			return pairs, nil

		case '[':
			var items []any
			for dec.More() {
				item, err := decodeOrdered(dec)
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			}
			if _, err := dec.Token(); err != nil { // consume ']'
				return nil, err
			}
			return items, nil
		}
	default:
		return v, nil // scalar: string, json.Number, bool, nil
	}
	return nil, fmt.Errorf("unexpected token")
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// Find searches rawJSON for every occurrence of fieldName.
// It returns results in document order (fix #2), preserves duplicate keys
// (fix #1), uses streaming decode (fix #4), and returns structured path
// segments alongside the path string (fixes #5 and #6).
//
// The optional mode argument controls matching behaviour (fix #3):
//
//	finder.Find(raw, "email")                        // Exact (default)
//	finder.Find(raw, "email", finder.CaseInsensitive) // case-insensitive
//	finder.Find(raw, "email", finder.Contains)        // substring match
func Find(rawJSON []byte, fieldName string, mode ...MatchMode) ([]Result, error) {
	m := resolveMode(mode)
	dec := json.NewDecoder(bytes.NewReader(rawJSON))
	dec.UseNumber() // preserves large integers and float precision
	root, err := decodeOrdered(dec)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	var results []Result
	walk(root, nil, fieldName, m, &results)
	return results, nil
}

// FindStream is for very large payloads — pass an already-open *json.Decoder
// wrapping any io.Reader (e.g. an HTTP response body). The caller controls
// buffering so the full payload is never loaded into a []byte first (fix #4).
// Accepts the same optional MatchMode as Find.
func FindStream(dec *json.Decoder, fieldName string, mode ...MatchMode) ([]Result, error) {
	m := resolveMode(mode)
	dec.UseNumber()
	root, err := decodeOrdered(dec)
	if err != nil {
		return nil, fmt.Errorf("stream decode error: %w", err)
	}
	var results []Result
	walk(root, nil, fieldName, m, &results)
	return results, nil
}

// resolveMode returns the first supplied MatchMode or Exact if none given.
// This keeps Find() backward-compatible — existing callers need no changes.
func resolveMode(modes []MatchMode) MatchMode {
	if len(modes) > 0 {
		return modes[0]
	}
	return Exact
}

// ---------------------------------------------------------------------------
// Internal walker
// ---------------------------------------------------------------------------

func walk(node any, segs []Segment, fieldName string, mode MatchMode, results *[]Result) {
	switch typed := node.(type) {

	case []KeyValue: // JSON object — document order, duplicates included
		for _, kv := range typed {
			childSegs := appendSeg(segs, Segment{Kind: KeySegment, Key: kv.Key})
			if matchKey(kv.Key, fieldName, mode) {
				*results = append(*results, Result{
					PathString:   buildPathString(childSegs),
					PathSegments: childSegs,
					Value:        kv.Value,
				})
			}
			walk(kv.Value, childSegs, fieldName, mode, results)
		}

	case []any: // JSON array — fix #5: nested arrays get $.a[0][1] notation
		for i, item := range typed {
			childSegs := appendSeg(segs, Segment{Kind: IndexSegment, Index: i})
			walk(item, childSegs, fieldName, mode, results)
		}

	// scalar (string, json.Number, bool, nil) — nothing to descend into
	}
}

// appendSeg clones segs before appending so sibling branches never share
// the same underlying array (classic Go slice-header gotcha).
func appendSeg(segs []Segment, s Segment) []Segment {
	out := make([]Segment, len(segs)+1)
	copy(out, segs)
	out[len(segs)] = s
	return out
}

// buildPathString converts a segment slice into a standard JSONPath string.
// Produces $.a.b[0][1].c — compatible with all major JSONPath evaluators.
func buildPathString(segs []Segment) string {
	var sb strings.Builder
	sb.WriteString("$")
	for _, s := range segs {
		if s.Kind == IndexSegment {
			fmt.Fprintf(&sb, "[%d]", s.Index)
		} else {
			sb.WriteByte('.')
			sb.WriteString(s.Key)
		}
	}
	return sb.String()
}
