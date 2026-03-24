package jasynthex_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	finder "github.com/MensahPrince/Jasynthex"
)

// ---- Fix #3: Case-insensitive and partial matching ------------------------

func TestFix3_CaseInsensitive(t *testing.T) {
	// Same concept, different casing across endpoints — all must be found
	raw := []byte(`{
		"endpointA": { "email":  "a@clinic.com" },
		"endpointB": { "Email":  "b@clinic.com" },
		"endpointC": { "EMAIL":  "c@clinic.com" }
	}`)

	results, err := finder.Find(raw, "email", finder.CaseInsensitive)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("\n=== Fix #3a: CaseInsensitive ===")
	for _, r := range results {
		fmt.Printf("  Path: %-30s  Value: %v\n", r.PathString, r.Value)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestFix3_Contains(t *testing.T) {
	// Inconsistent field naming across endpoints — all email-related keys found
	raw := []byte(`{
		"user": {
			"email":         "direct@clinic.com",
			"email_address": "address@clinic.com",
			"contactEmail":  "contact@clinic.com",
			"EmailId":       "id@clinic.com",
			"phone":         "0244000001"
		}
	}`)

	results, err := finder.Find(raw, "email", finder.Contains)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("\n=== Fix #3b: Contains (substring match) ===")
	for _, r := range results {
		fmt.Printf("  Path: %-35s  Value: %v\n", r.PathString, r.Value)
	}
	// "phone" must NOT appear — only email-related keys
	if len(results) != 4 {
		t.Errorf("expected 4 results, got %d", len(results))
	}
}

func TestFix3_ExactStillDefault(t *testing.T) {
	// Calling Find with no mode must still behave exactly as before (backward compat)
	raw := []byte(`{
		"email":   "match@clinic.com",
		"Email":   "no-match@clinic.com",
		"emailId": "no-match@clinic.com"
	}`)

	results, err := finder.Find(raw, "email") // no mode arg
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("\n=== Fix #3c: Exact is still the default ===")
	for _, r := range results {
		fmt.Printf("  Path: %-20s  Value: %v\n", r.PathString, r.Value)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 exact result, got %d", len(results))
	}
}

// ---- Fix #1: Duplicate keys -----------------------------------------------

func TestFix1_DuplicateKeys(t *testing.T) {
	// Both "status" keys must be found — standard json.Unmarshal would lose the first
	raw := []byte(`{
		"request":  { "status": "pending" },
		"response": { "status": "approved" }
	}`)

	results, err := finder.Find(raw, "status")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("\n=== Fix #1: Duplicate keys ===")
	for _, r := range results {
		fmt.Printf("  Path: %-45s  Value: %v\n", r.PathString, r.Value)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

// ---- Fix #2: Document order ------------------------------------------------

func TestFix2_DocumentOrder(t *testing.T) {
	// Fields appear in a specific order; results must come back in that order
	raw := []byte(`{
		"alpha": { "score": 1 },
		"beta":  { "score": 2 },
		"gamma": { "score": 3 }
	}`)

	results, err := finder.Find(raw, "score")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("\n=== Fix #2: Document order ===")
	for _, r := range results {
		fmt.Printf("  Path: %-30s  Value: %v\n", r.PathString, r.Value)
	}

	// Verify order: alpha → beta → gamma
	expected := []string{
		"$.alpha.score",
		"$.beta.score",
		"$.gamma.score",
	}
	for i, r := range results {
		if r.PathString != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], r.PathString)
		}
	}
}

// ---- Fix #4: Streaming large JSON ------------------------------------------

func TestFix4_StreamLargeJSON(t *testing.T) {
	// Simulate receiving a large JSON payload as an io.Reader (e.g. HTTP body)
	// FindStream never loads the full payload into []byte
	raw := []byte(`{
		"batch": [
			{ "id": 1, "email": "a@example.com" },
			{ "id": 2, "email": "b@example.com" },
			{ "id": 3, "email": "c@example.com" }
		]
	}`)

	dec := json.NewDecoder(bytes.NewReader(raw))
	results, err := finder.FindStream(dec, "email")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("\n=== Fix #4: Streaming decode ===")
	for _, r := range results {
		fmt.Printf("  Path: %-35s  Value: %v\n", r.PathString, r.Value)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 emails, got %d", len(results))
	}
}

// ---- Fix #5: Nested array path notation ------------------------------------

func TestFix5_NestedArrayPaths(t *testing.T) {
	// Matrix: array of arrays — path must be $.matrix[0][1] not $.matrix[0].[1]
	raw := []byte(`{
		"matrix": [
			[{ "val": 10 }, { "val": 20 }],
			[{ "val": 30 }, { "val": 40 }]
		]
	}`)

	results, err := finder.Find(raw, "val")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("\n=== Fix #5: Nested array path notation ===")
	for _, r := range results {
		fmt.Printf("  Path: %-30s  Value: %v\n", r.PathString, r.Value)
	}

	// Spot-check: second element of first row
	if results[1].PathString != "$.matrix[0][1].val" {
		t.Errorf("expected $.matrix[0][1].val, got %s", results[1].PathString)
	}
}

// ---- Fix #6: Structured PathSegments ---------------------------------------

func TestFix6_StructuredPathSegments(t *testing.T) {
	raw := []byte(`{
		"hospital": {
			"departments": [
				{
					"name": "Physiotherapy",
					"head": { "email": "pt@hospital.com" }
				}
			]
		}
	}`)

	results, err := finder.Find(raw, "email")
	if err != nil {
		t.Fatal(err)
	}

	r := results[0]
	fmt.Println("\n=== Fix #6: Structured PathSegments ===")
	fmt.Printf("  PathString:  %s\n", r.PathString)
	fmt.Printf("  Depth:       %d\n", r.Depth())
	fmt.Printf("  ParentPath:  %s\n", r.ParentPath())
	fmt.Printf("  Segments:\n")
	for i, seg := range r.PathSegments {
		fmt.Printf("    [%d] kind=%d  key=%q  index=%d\n", i, seg.Kind, seg.Key, seg.Index)
	}

	// Programmatic manipulation: go up one level without string parsing
	if r.ParentPath() != "$.hospital.departments[0].head" {
		t.Errorf("unexpected parent path: %s", r.ParentPath())
	}

	// Check individual segments
	segs := r.PathSegments
	if segs[0].Key != "hospital" {
		t.Errorf("seg[0] expected 'hospital', got %q", segs[0].Key)
	}
	if segs[2].Kind != finder.IndexSegment || segs[2].Index != 0 {
		t.Errorf("seg[2] expected IndexSegment 0, got %+v", segs[2])
	}
}

// ---- Real-world scenario: API interception ---------------------------------

func TestRealWorld_APIInterception(t *testing.T) {
	// Simulated intercepted API response from an engineering endpoint
	intercepted := []byte(`{
		"meta": { "status": "ok", "version": "2.1" },
		"data": {
			"users": [
				{
					"id": "u001",
					"profile": {
						"name": "Ama Boateng",
						"contact": {
							"email": "ama@clinic.com",
							"phone": "0244000001"
						}
					},
					"billing": { "email": "ama-billing@clinic.com" }
				},
				{
					"id": "u002",
					"profile": {
						"name": "Kofi Asante",
						"contact": {
							"email": "kofi@clinic.com",
							"phone": "0244000002"
						}
					},
					"billing": { "email": "kofi-billing@clinic.com" }
				}
			]
		}
	}`)

	results, err := finder.Find(intercepted, "email")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("\n=== Real-world: all email paths in intercepted JSON ===")
	for _, r := range results {
		fmt.Printf("  %-55s → %v\n", r.PathString, r.Value)
	}
	if len(results) != 4 {
		t.Errorf("expected 4 emails, got %d", len(results))
	}
}
