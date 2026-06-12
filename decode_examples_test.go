package odp

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	generated "github.com/patent-dev/uspto-odp/generated"
)

// testdata/strictdecode holds one recorded raw USPTO ODP API response body per
// covered endpoint, captured live. These committed fixtures are the ground truth
// for the deterministic, network-free regression below: TestFixtures runs in the
// normal build (no integration tag, no credentials, no network) and reads ONLY
// these files. A human who wants to refresh them runs `make refresh-fixtures`,
// which re-captures from the live API (see Makefile).
//
// USPTO ODP endpoints fall into two response kinds for this regression:
//
//   - kindTyped: a structured JSON body the generated (or hand-written) type
//     models field-for-field. These get all three layers:
//     (a) STRICT decode (json.Decoder + DisallowUnknownFields) into the typed
//     value, so no field the API returns is silently dropped;
//     (b) GOLDEN round-trip: re-marshal the decoded typed value and deep-compare
//     it, after normalize(), against the raw fixture body, proving every modeled
//     field survives decode+re-marshal losslessly;
//     (c) TARGETED key-field assertions read from the re-marshaled TYPED value, so
//     a pass proves the value parsed into the correct generated field.
//   - kindLoose: a response whose body the library deliberately models with a
//     free-form container (Office Action /records Docs []map[string]any in
//     DSAPIResponse, and any interface{}/RawMessage/raw-bytes payloads). A
//     map/interface absorbs unknown keys, so DisallowUnknownFields and the golden
//     round-trip do not apply; these get a minimal sanity check only. None of the
//     committed fixtures are currently loose (the Office Action /records search
//     responses are exercised live in integration_test.go), but the kind exists so
//     a loose fixture can be added without weakening the typed guarantees.
//
// LIMITATION (shared with every sampled-response guard): a fixture only exercises
// the fields present in that one recorded response. A field the API can return but
// did not in the sample is not caught. This is a high-value smoke test against
// real data, not a proof of total completeness. Truncated search fixtures keep a
// single representative record (count reflects the real total).

// fixtureKind classifies a recorded response body.
type fixtureKind int

const (
	// kindTyped: structured JSON modeled field-for-field; gets strict decode +
	// golden round-trip + key-field checks.
	kindTyped fixtureKind = iota
	// kindLoose: free-form container (map/interface/raw bytes); gets a minimal
	// sanity check only (see the kind comment above).
	kindLoose
)

// endpoint is one row of the fixture table: exactly one row per recorded fixture
// file. Missing/extra files fail the count guard in TestFixtures.
type endpoint struct {
	file string // fixture filename under testdata/strictdecode (without .json)
	kind fixtureKind

	// target returns a fresh pointer to decode the raw body into (kindTyped only).
	// A constructor keeps each case's concrete type local. Inline 200 schemas
	// decode into the JSON200 field of their generated response wrapper, reached as
	// &(&Wrapper{}).JSON200 (a **payload that re-marshals to the payload JSON).
	target func() any

	// checks are the targeted key-field assertions, run against the map produced by
	// re-marshaling the decoded TYPED value (not the raw fixture). Paths use dot
	// notation with [i] for array indices.
	checks []check
}

// check is one key-field assertion against the decoded+re-marshaled typed value.
type check struct {
	path string
	// want, when non-nil, is the exact expected value at path (compared as the
	// JSON-decoded scalar, stringified). When nil, the assertion is
	// "present and non-empty".
	want *string
	// length, when non-nil, asserts the value at path is an array of this length.
	length *int
}

func eq(s string) *string { return &s }
func ln(n int) *int       { return &n }

// endpoints is the authoritative fixture table. Every row maps a committed
// testdata/strictdecode/<file>.json to the typed value it decodes into and the
// stable key-field assertions read from the recorded body. Search fixtures are
// truncated to one representative record; count reflects the real API total.
var endpoints = []endpoint{
	// --- Patent application + sub-resources (typed) ------------------------
	{
		file: "get_patent", kind: kindTyped,
		target: func() any { return new(generated.PatentDataResponse) },
		checks: []check{
			{path: "count", want: eq("1")},
			{path: "patentFileWrapperDataBag", length: ln(1)},
			{path: "patentFileWrapperDataBag[0].applicationNumberText", want: eq("17248024")},
			{path: "patentFileWrapperDataBag[0].applicationMetaData.patentNumber", want: eq("11646472")},
		},
	},
	{
		file: "search_patents", kind: kindTyped,
		target: func() any { return new(generated.PatentDataResponse) },
		checks: []check{
			{path: "count", want: eq("1")},
			{path: "patentFileWrapperDataBag[0].applicationNumberText", want: eq("17248024")},
		},
	},
	{
		file: "get_patent_adjustment", kind: kindTyped,
		target: func() any {
			return &(&generated.GetApiV1PatentApplicationsApplicationNumberTextAdjustmentResponse{}).JSON200
		},
		checks: []check{
			{path: "count", want: eq("1")},
			{path: "patentFileWrapperDataBag[0].patentTermAdjustmentData.adjustmentTotalQuantity", want: eq("304")},
		},
	},
	{
		file: "get_patent_assignment", kind: kindTyped,
		target: func() any {
			return &(&generated.GetApiV1PatentApplicationsApplicationNumberTextAssignmentResponse{}).JSON200
		},
		checks: []check{
			{path: "count", want: eq("1")},
			{path: "patentFileWrapperDataBag[0].assignmentBag[0].reelAndFrameNumber", want: eq("038323/0190")},
		},
	},
	{
		file: "get_patent_documents", kind: kindTyped,
		target: func() any { return new(generated.DocumentBag) },
		checks: []check{
			{path: "count", want: eq("45")},
			{path: "documentBag[0].applicationNumberText", want: eq("17248024")},
			{path: "documentBag[0].documentIdentifier", want: eq("LN4VBTHCXBLUEX2")},
		},
	},
	{
		file: "get_patent_continuity", kind: kindTyped,
		target: func() any {
			return &(&generated.GetApiV1PatentApplicationsApplicationNumberTextContinuityResponse{}).JSON200
		},
		checks: []check{
			{path: "count", want: eq("1")},
			{path: "patentFileWrapperDataBag[0].parentContinuityBag[0].parentApplicationNumberText", want: eq("16695054")},
		},
	},
	{
		file: "get_patent_foreign_priority", kind: kindTyped,
		target: func() any {
			return &(&generated.GetApiV1PatentApplicationsApplicationNumberTextForeignPriorityResponse{}).JSON200
		},
		checks: []check{
			{path: "count", want: eq("1")},
			{path: "patentFileWrapperDataBag[0].applicationNumberText", want: eq("17248024")},
		},
	},
	{
		file: "get_patent_transactions", kind: kindTyped,
		target: func() any {
			return &(&generated.GetApiV1PatentApplicationsApplicationNumberTextTransactionsResponse{}).JSON200
		},
		checks: []check{
			{path: "count", want: eq("1")},
			{path: "patentFileWrapperDataBag[0].eventDataBag[0].eventCode"},
		},
	},
	{
		file: "get_patent_attorney", kind: kindTyped,
		target: func() any {
			return &(&generated.GetApiV1PatentApplicationsApplicationNumberTextAttorneyResponse{}).JSON200
		},
		checks: []check{
			{path: "count", want: eq("1")},
			{path: "patentFileWrapperDataBag[0].recordAttorney.customerNumberCorrespondenceData.powerOfAttorneyAddressBag[0].cityName", want: eq("OAKLAND")},
		},
	},
	{
		file: "get_patent_associated_documents", kind: kindTyped,
		target: func() any {
			return &(&generated.GetApiV1PatentApplicationsApplicationNumberTextAssociatedDocumentsResponse{}).JSON200
		},
		checks: []check{
			{path: "count", want: eq("1")},
			{path: "patentFileWrapperDataBag[0].grantDocumentMetaData.productIdentifier", want: eq("PTGRXML")},
		},
	},

	// --- Status codes + bulk data (typed) ----------------------------------
	{
		file: "get_status_codes", kind: kindTyped,
		target: func() any { return new(generated.StatusCodeSearchResponse) },
		checks: []check{
			{path: "count", want: eq("241")},
			{path: "statusCodeBag[0].applicationStatusCode", want: eq("1")},
			{path: "statusCodeBag[0].applicationStatusDescriptionText", want: eq("Missassigned Application Number")},
		},
	},
	// The two bulk-data product endpoints are LOOSE, not because the body is
	// free-form, but because the committed swagger (swagger_fixed.yaml, mirroring
	// the published spec) mistypes two field families relative to the live API, so
	// the golden round-trip cannot be lossless until the spec is corrected and the
	// client regenerated:
	//   - productDataSetArrayText / productDataSetCategoryArrayText: the API field
	//     is spelled productDataset... (lowercase s). Go's case-insensitive decode
	//     accepts the live key, but the typed value re-marshals to the spec's
	//     productDataSet... casing, so the round-trip keys differ.
	//   - productTotalFileSize / fileSize: modeled as float32, which truncates the
	//     large byte counts the API returns (e.g. 25938442242 -> 2.5938442e+10).
	// A minimal sanity check (valid, non-nil JSON) is applied; the live response is
	// exercised in integration_test.go (TestIntegrationGetBulkProduct /
	// TestIntegrationSearchBulkProducts).
	{file: "search_bulk_products", kind: kindLoose},
	{file: "get_bulk_product", kind: kindLoose},

	// --- PTAB trials: decisions, documents, proceedings (typed) ------------
	{
		file: "search_trial_decisions", kind: kindTyped,
		target: func() any { return new(generated.DecisionDataResponse) },
		checks: []check{
			{path: "patentTrialDocumentDataBag[0].trialNumber"},
			{path: "patentTrialDocumentDataBag[0].trialDocumentCategory", want: eq("Decision")},
		},
	},
	{
		file: "get_trial_decisions", kind: kindTyped,
		target: func() any { return new(generated.DecisionDataResponse) },
		checks: []check{
			{path: "count", want: eq("2")},
			{path: "patentTrialDocumentDataBag[0].trialNumber", want: eq("PGR2025-00004")},
		},
	},
	{
		file: "search_trial_documents", kind: kindTyped,
		target: func() any { return new(generated.DocumentDataResponse) },
		checks: []check{
			{path: "patentTrialDocumentDataBag[0].trialDocumentCategory", want: eq("Document")},
		},
	},
	{
		file: "get_trial_documents", kind: kindTyped,
		target: func() any { return new(generated.DocumentDataResponse) },
		checks: []check{
			{path: "count", want: eq("99")},
			{path: "patentTrialDocumentDataBag[0].trialNumber", want: eq("PGR2026-00039")},
		},
	},
	{
		file: "search_trial_proceedings", kind: kindTyped,
		target: func() any { return new(generated.ProceedingDataResponse) },
		checks: []check{
			{path: "patentTrialProceedingDataBag[0].trialNumber"},
		},
	},
	{
		file: "get_trial_proceeding", kind: kindTyped,
		target: func() any { return new(generated.ProceedingDataResponse) },
		checks: []check{
			{path: "count", want: eq("1")},
			{path: "patentTrialProceedingDataBag[0].trialNumber", want: eq("PGR2026-00055")},
		},
	},

	// --- PTAB appeals + interferences (typed) ------------------------------
	{
		file: "search_appeal_decisions", kind: kindTyped,
		target: func() any { return new(generated.AppealDecisionDataResponse) },
		checks: []check{
			{path: "patentAppealDataBag[0].appealNumber"},
			{path: "patentAppealDataBag[0].appealDocumentCategory", want: eq("Decision")},
		},
	},
	{
		file: "get_appeal_decisions", kind: kindTyped,
		target: func() any { return new(generated.AppealDecisionDataResponse) },
		checks: []check{
			{path: "count", want: eq("1")},
			{path: "patentAppealDataBag[0].appealNumber", want: eq("2026001845")},
		},
	},
	{
		file: "search_interference_decisions", kind: kindTyped,
		target: func() any { return new(generated.InterferenceDecisionDataResponse) },
		checks: []check{
			{path: "patentInterferenceDataBag[0].interferenceNumber"},
		},
	},
	{
		file: "get_interference_decisions", kind: kindTyped,
		target: func() any { return new(generated.InterferenceDecisionDataResponse) },
		checks: []check{
			{path: "count", want: eq("2")},
			{path: "patentInterferenceDataBag[0].interferenceNumber", want: eq("106130")},
		},
	},

	// --- Petitions (typed) -------------------------------------------------
	{
		file: "search_petitions", kind: kindTyped,
		target: func() any { return new(generated.PetitionDecisionResponseBag) },
		checks: []check{
			{path: "petitionDecisionDataBag[0].applicationNumberText"},
		},
	},
	{
		file: "get_petition_decision", kind: kindTyped,
		target: func() any { return new(generated.PetitionDecisionIdentifierResponseBag) },
		checks: []check{
			{path: "count", want: eq("1")},
			{path: "petitionDecisionDataBag[0].applicationNumberText", want: eq("10347018")},
		},
	},

	// --- Office Action DSAPI /fields (typed; hand-written DSAPIFieldsResponse) ---
	{
		file: "get_office_action_fields", kind: kindTyped,
		target: func() any { return new(DSAPIFieldsResponse) },
		checks: []check{
			{path: "apiKey", want: eq("oa_actions")},
			{path: "fieldCount", want: eq("56")},
		},
	},
	{
		file: "get_office_action_citation_fields", kind: kindTyped,
		target: func() any { return new(DSAPIFieldsResponse) },
		checks: []check{
			{path: "apiKey", want: eq("oa_citations")},
			{path: "fieldCount", want: eq("16")},
		},
	},
	{
		file: "get_office_action_rejection_fields", kind: kindTyped,
		target: func() any { return new(DSAPIFieldsResponse) },
		checks: []check{
			{path: "apiKey", want: eq("oa_rejections")},
			{path: "fieldCount", want: eq("31")},
		},
	},
	{
		file: "get_enriched_citation_fields", kind: kindTyped,
		target: func() any { return new(DSAPIFieldsResponse) },
		checks: []check{
			{path: "apiKey", want: eq("enriched_cited_reference_metadata")},
			{path: "fieldCount", want: eq("22")},
		},
	},
}

// TestFixtures is the single deterministic, network-free regression over the
// committed testdata/strictdecode fixtures, one named subtest per covered
// endpoint. It runs in the normal build (no integration tag, no credentials).
//
// Per typed endpoint it performs three layered checks: (a) strict decode,
// (b) golden round-trip after normalize(), (c) targeted key-field assertions.
// See the file-level comment for the full rationale.
func TestFixtures(t *testing.T) {
	dir := filepath.Join("testdata", "strictdecode")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read fixtures dir: %v", err)
	}
	// Count guard: exactly one fixture file per table row. Fails if a fixture is
	// added/removed without updating the table.
	if len(entries) != len(endpoints) {
		t.Fatalf("fixture count %d != endpoint table rows %d (table/testdata drifted)",
			len(entries), len(endpoints))
	}

	byFile := make(map[string]bool, len(entries))
	for _, e := range entries {
		byFile[e.Name()] = true
	}

	for _, ep := range endpoints {
		name := ep.file + ".json"
		if !byFile[name] {
			t.Fatalf("table references missing fixture %q", name)
		}
		t.Run(ep.file, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatalf("read %s: %v", name, err)
			}
			if len(raw) == 0 {
				t.Fatalf("fixture %s is empty", name)
			}
			switch ep.kind {
			case kindTyped:
				runTyped(t, ep, raw)
			case kindLoose:
				runLoose(t, ep, raw)
			}
		})
	}
}

// runTyped performs the strict-decode, golden round-trip and key-field checks for
// a typed endpoint.
func runTyped(t *testing.T, ep endpoint, body []byte) {
	t.Helper()

	// (a) strict decode into the typed value: DisallowUnknownFields fails if the
	// recorded body carries any key the type does not model (a dropped field).
	decoded := ep.target()
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(decoded); err != nil {
		t.Fatalf("strict decode dropped a field the API returns: %v", err)
	}

	// (b) golden round-trip: re-marshal the typed value and deep-compare the
	// normalized forms against the raw fixture body, proving every modeled field
	// survives decode+re-marshal losslessly.
	remarshaled, err := json.Marshal(decoded)
	if err != nil {
		t.Fatalf("re-marshal decoded typed value: %v", err)
	}
	var fixtureMap, typedMap any
	if err := json.Unmarshal(body, &fixtureMap); err != nil {
		t.Fatalf("unmarshal fixture body to map: %v", err)
	}
	if err := json.Unmarshal(remarshaled, &typedMap); err != nil {
		t.Fatalf("unmarshal re-marshaled typed value to map: %v", err)
	}
	if !reflect.DeepEqual(normalize(fixtureMap), normalize(typedMap)) {
		t.Fatalf("golden round-trip mismatch: a modeled field did not survive decode+re-marshal losslessly\n fixture (normalized): %#v\n typed   (normalized): %#v",
			normalize(fixtureMap), normalize(typedMap))
	}

	// (c) targeted key-field assertions, read from the typed value's re-marshal so
	// a pass proves the value parsed into the correct typed field.
	for _, c := range ep.checks {
		got, ok := lookup(typedMap, c.path)
		if !ok {
			t.Errorf("key-field %q: not present in parsed value", c.path)
			continue
		}
		switch {
		case c.length != nil:
			a, ok := got.([]any)
			if !ok {
				t.Errorf("key-field %q: expected array, got %T", c.path, got)
				continue
			}
			if len(a) != *c.length {
				t.Errorf("key-field %q: array length = %d, want %d", c.path, len(a), *c.length)
			}
		case c.want != nil:
			if s := scalarString(got); s != *c.want {
				t.Errorf("key-field %q = %q, want %q", c.path, s, *c.want)
			}
		default:
			if s := scalarString(got); s == "" {
				t.Errorf("key-field %q: expected non-empty value, got %#v", c.path, got)
			}
		}
	}
}

// runLoose performs the minimal sanity check for a free-form (map/interface/raw)
// endpoint: the body is valid JSON and non-empty. The library models these with a
// container that absorbs unknown keys, so strict decode and the golden round-trip
// do not apply (see the kind comment).
func runLoose(t *testing.T, ep endpoint, body []byte) {
	t.Helper()
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatalf("loose fixture %s is not valid JSON: %v", ep.file, err)
	}
	if v == nil {
		t.Fatalf("loose fixture %s decoded to nil", ep.file)
	}
}

// scalarString renders a JSON-decoded scalar (string/number/bool) as a string for
// the key-field compare. Numbers decode into float64; integral values render
// without a trailing ".0" so checks read naturally (e.g. "304", "56").
func scalarString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case float64:
		if x == float64(int64(x)) {
			b, _ := json.Marshal(int64(x))
			return string(b)
		}
		b, _ := json.Marshal(x)
		return string(b)
	default:
		return ""
	}
}

// normalize canonicalizes a JSON value (decoded into map/slice/scalar) for the
// golden round-trip deep-compare. The generated types use *T fields with
// `omitempty`, so a re-marshal can legitimately differ from the raw fixture only
// in ways that carry no information. normalize erases exactly those differences:
//
//   - null is dropped (a nil pointer marshals to nothing, omitempty or not).
//   - empty string "" is dropped (a fixture's "" maps to *string(""); whether it
//     survives marshaling is immaterial, the value carries no data either way).
//   - empty object {} and empty array [] are dropped (after recursion), so an
//     all-empty subtree on one side matches its absence on the other.
//
// It does NOT touch any non-empty scalar, so every modeled value that carries data
// must appear identically on both sides or the compare fails. Both sides pass
// through encoding/json, so any numeric canonicalization is symmetric.
func normalize(v any) any {
	switch t := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(t))
		for k, val := range t {
			nv := normalize(val)
			if nv == nil {
				continue
			}
			if s, ok := nv.(string); ok && s == "" {
				continue
			}
			m[k] = nv
		}
		if len(m) == 0 {
			return nil
		}
		return m
	case []any:
		s := make([]any, 0, len(t))
		for _, e := range t {
			s = append(s, normalize(e))
		}
		if len(s) == 0 {
			return nil
		}
		return s
	default:
		return v
	}
}

// lookup walks a decoded JSON value (map[string]any / []any) by a dotted path with
// optional [i] array indices, e.g. "patentFileWrapperDataBag[0].applicationNumberText".
// It returns the value at the path and whether it was found.
func lookup(root any, path string) (any, bool) {
	cur := root
	for _, seg := range splitPath(path) {
		switch s := seg.(type) {
		case string:
			m, ok := cur.(map[string]any)
			if !ok {
				return nil, false
			}
			cur, ok = m[s]
			if !ok {
				return nil, false
			}
		case int:
			a, ok := cur.([]any)
			if !ok || s < 0 || s >= len(a) {
				return nil, false
			}
			cur = a[s]
		}
	}
	return cur, true
}

// splitPath turns "a.b[0].c" into ["a","b",0,"c"] (strings for keys, ints for
// array indices).
func splitPath(path string) []any {
	var out []any
	field := make([]rune, 0, len(path))
	flush := func() {
		if len(field) > 0 {
			out = append(out, string(field))
			field = field[:0]
		}
	}
	for i := 0; i < len(path); i++ {
		c := path[i]
		switch c {
		case '.':
			flush()
		case '[':
			flush()
			j := i + 1
			n := 0
			for j < len(path) && path[j] >= '0' && path[j] <= '9' {
				n = n*10 + int(path[j]-'0')
				j++
			}
			out = append(out, n)
			i = j // skip past the closing ']'
		default:
			field = append(field, rune(c))
		}
	}
	flush()
	return out
}
