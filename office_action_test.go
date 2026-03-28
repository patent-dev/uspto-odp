package odp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupOAMockServer(t *testing.T) (*httptest.Server, *Client) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {

		// Office Action Text Retrieval
		case "/api/v1/patent/oa/oa_actions/v1/records":
			if r.Method != "POST" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"response": map[string]any{
					"numFound": 5432,
					"start":    0,
					"docs": []any{
						map[string]any{
							"patentApplicationNumber": "16123456",
							"actionType":              "Non-Final Rejection",
							"mailedDate":              "2021-05-15",
							"id":                      "abc123",
						},
					},
				},
			})

		case "/api/v1/patent/oa/oa_actions/v1/fields":
			if r.Method != "GET" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"apiKey":              "oa_actions",
				"apiVersionNumber":    "v1",
				"apiUrl":              "https://api.uspto.gov/api/v1/patent/oa/oa_actions/v1/fields",
				"apiDocumentationUrl": "https://data.uspto.gov/swagger",
				"apiStatus":           "PUBLISHED",
				"fieldCount":          5,
				"fields":              []string{"patentApplicationNumber", "actionType", "mailedDate", "id", "createDateTime"},
			})

		// Office Action Citations
		case "/api/v1/patent/oa/oa_citations/v2/records":
			if r.Method != "POST" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"response": map[string]any{
					"numFound": 12000,
					"start":    0,
					"docs": []any{
						map[string]any{
							"patentApplicationNumber": "16123456",
							"legalSectionCode":        "103",
							"form892":                 true,
						},
					},
				},
			})

		case "/api/v1/patent/oa/oa_citations/v2/fields":
			if r.Method != "GET" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"apiKey":     "oa_citations",
				"fieldCount": 3,
				"fields":     []string{"patentApplicationNumber", "legalSectionCode", "form892"},
			})

		// Office Action Rejections
		case "/api/v1/patent/oa/oa_rejections/v2/records":
			if r.Method != "POST" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"response": map[string]any{
					"numFound": 86973947,
					"start":    0,
					"docs": []any{
						map[string]any{
							"patentApplicationNumber": "12190351",
							"hasRej101":               float64(0),
							"hasRej102":               float64(0),
							"hasRej103":               float64(1),
							"hasRej112":               float64(1),
							"hasRejDP":                float64(0),
							"aliceIndicator":          false,
							"bilskiIndicator":         false,
							"groupArtUnitNumber":      "1713",
							"legalSectionCode":        "112",
						},
					},
				},
			})

		case "/api/v1/patent/oa/oa_rejections/v2/fields":
			if r.Method != "GET" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"apiKey":     "oa_rejections",
				"fieldCount": 31,
				"fields": []string{
					"bilskiIndicator", "hasRej101", "hasRej103", "hasRejDP",
					"hasRej102", "hasRej112", "patentApplicationNumber",
					"legalSectionCode", "groupArtUnitNumber", "aliceIndicator",
				},
			})

		// Enriched Citations
		case "/api/v1/patent/oa/enriched_cited_reference_metadata/v3/records":
			if r.Method != "POST" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"response": map[string]any{
					"numFound": 50000,
					"start":    0,
					"docs": []any{
						map[string]any{
							"patentApplicationNumber": "15123456",
							"citedPatentNumber":       "9876543",
							"rejectedClaimNumbers":    "1,2,3",
						},
					},
				},
			})

		case "/api/v1/patent/oa/enriched_cited_reference_metadata/v3/fields":
			if r.Method != "GET" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"apiKey":     "enriched_cited_reference_metadata",
				"fieldCount": 3,
				"fields":     []string{"patentApplicationNumber", "citedPatentNumber", "rejectedClaimNumbers"},
			})

		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": "Not Found",
			})
		}
	}))

	config := &Config{
		BaseURL:    server.URL,
		APIKey:     "test-key",
		MaxRetries: 0,
		Timeout:    10,
		UserAgent:  "test",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	return server, client
}

// --- Happy Path Tests ---

func TestSearchOfficeActions(t *testing.T) {
	server, client := setupOAMockServer(t)
	defer server.Close()

	result, err := client.SearchOfficeActions(context.Background(), "patentApplicationNumber:16123456", 0, 10)
	if err != nil {
		t.Fatalf("SearchOfficeActions failed: %v", err)
	}
	if result.Response.NumFound != 5432 {
		t.Errorf("Expected numFound 5432, got %d", result.Response.NumFound)
	}
	if len(result.Response.Docs) != 1 {
		t.Fatalf("Expected 1 doc, got %d", len(result.Response.Docs))
	}
	if result.Response.Docs[0]["patentApplicationNumber"] != "16123456" {
		t.Errorf("Expected patentApplicationNumber 16123456, got %v", result.Response.Docs[0]["patentApplicationNumber"])
	}
}

func TestSearchOfficeActionsEmptyCriteria(t *testing.T) {
	server, client := setupOAMockServer(t)
	defer server.Close()

	result, err := client.SearchOfficeActions(context.Background(), "", 0, 10)
	if err != nil {
		t.Fatalf("SearchOfficeActions with empty criteria failed: %v", err)
	}
	if result.Response.NumFound != 5432 {
		t.Errorf("Expected numFound 5432, got %d", result.Response.NumFound)
	}
}

func TestGetOfficeActionFields(t *testing.T) {
	server, client := setupOAMockServer(t)
	defer server.Close()

	result, err := client.GetOfficeActionFields(context.Background())
	if err != nil {
		t.Fatalf("GetOfficeActionFields failed: %v", err)
	}
	if result.APIKey != "oa_actions" {
		t.Errorf("Expected apiKey oa_actions, got %s", result.APIKey)
	}
	if result.FieldCount != 5 {
		t.Errorf("Expected fieldCount 5, got %d", result.FieldCount)
	}
	if len(result.Fields) != 5 {
		t.Errorf("Expected 5 fields, got %d", len(result.Fields))
	}
}

func TestSearchOfficeActionCitations(t *testing.T) {
	server, client := setupOAMockServer(t)
	defer server.Close()

	result, err := client.SearchOfficeActionCitations(context.Background(), "*:*", 0, 10)
	if err != nil {
		t.Fatalf("SearchOfficeActionCitations failed: %v", err)
	}
	if result.Response.NumFound != 12000 {
		t.Errorf("Expected numFound 12000, got %d", result.Response.NumFound)
	}
}

func TestGetOfficeActionCitationFields(t *testing.T) {
	server, client := setupOAMockServer(t)
	defer server.Close()

	result, err := client.GetOfficeActionCitationFields(context.Background())
	if err != nil {
		t.Fatalf("GetOfficeActionCitationFields failed: %v", err)
	}
	if result.APIKey != "oa_citations" {
		t.Errorf("Expected apiKey oa_citations, got %s", result.APIKey)
	}
}

func TestSearchOfficeActionRejections(t *testing.T) {
	server, client := setupOAMockServer(t)
	defer server.Close()

	result, err := client.SearchOfficeActionRejections(context.Background(), "patentApplicationNumber:12190351", 0, 1)
	if err != nil {
		t.Fatalf("SearchOfficeActionRejections failed: %v", err)
	}
	if result.Response.NumFound != 86973947 {
		t.Errorf("Expected numFound 86973947, got %d", result.Response.NumFound)
	}
	doc := result.Response.Docs[0]
	if doc["groupArtUnitNumber"] != "1713" {
		t.Errorf("Expected groupArtUnitNumber 1713, got %v", doc["groupArtUnitNumber"])
	}
}

func TestGetOfficeActionRejectionFields(t *testing.T) {
	server, client := setupOAMockServer(t)
	defer server.Close()

	result, err := client.GetOfficeActionRejectionFields(context.Background())
	if err != nil {
		t.Fatalf("GetOfficeActionRejectionFields failed: %v", err)
	}
	if result.FieldCount != 31 {
		t.Errorf("Expected fieldCount 31, got %d", result.FieldCount)
	}
}

func TestSearchEnrichedCitations(t *testing.T) {
	server, client := setupOAMockServer(t)
	defer server.Close()

	result, err := client.SearchEnrichedCitations(context.Background(), "*:*", 0, 5)
	if err != nil {
		t.Fatalf("SearchEnrichedCitations failed: %v", err)
	}
	if result.Response.NumFound != 50000 {
		t.Errorf("Expected numFound 50000, got %d", result.Response.NumFound)
	}
}

func TestGetEnrichedCitationFields(t *testing.T) {
	server, client := setupOAMockServer(t)
	defer server.Close()

	result, err := client.GetEnrichedCitationFields(context.Background())
	if err != nil {
		t.Fatalf("GetEnrichedCitationFields failed: %v", err)
	}
	if result.APIKey != "enriched_cited_reference_metadata" {
		t.Errorf("Expected apiKey enriched_cited_reference_metadata, got %s", result.APIKey)
	}
}

// --- Error Path Tests ---

func TestSearchOfficeActions_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal server error","message":"database unavailable"}`))
	}))
	defer server.Close()

	client, err := NewClient(&Config{BaseURL: server.URL, APIKey: "test", MaxRetries: 0, Timeout: 10})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.SearchOfficeActions(context.Background(), "*:*", 0, 10)
	if err == nil {
		t.Fatal("Expected error for 500 response")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("Expected APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("Expected status 500, got %d", apiErr.StatusCode)
	}
	if apiErr.Body == "" {
		t.Error("Expected error body to be populated")
	}
}

func TestSearchOfficeActions_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL, APIKey: "test", MaxRetries: 0, Timeout: 10})

	_, err := client.SearchOfficeActions(context.Background(), "*:*", 0, 10)
	if err == nil {
		t.Fatal("Expected error for 404 response")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("Expected APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", apiErr.StatusCode)
	}
}

func TestSearchOfficeActions_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL, APIKey: "test", MaxRetries: 0, Timeout: 10})

	_, err := client.SearchOfficeActions(context.Background(), "*:*", 0, 10)
	if err == nil {
		t.Fatal("Expected error for malformed JSON")
	}
}

func TestSearchOfficeActions_EmptyDocs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"response": map[string]any{
				"numFound": 0,
				"start":    0,
				"docs":     []any{},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL, APIKey: "test", MaxRetries: 0, Timeout: 10})

	result, err := client.SearchOfficeActions(context.Background(), "nonexistent:query", 0, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Response.NumFound != 0 {
		t.Errorf("Expected numFound 0, got %d", result.Response.NumFound)
	}
	if len(result.Response.Docs) != 0 {
		t.Errorf("Expected 0 docs, got %d", len(result.Response.Docs))
	}
}

func TestSearchOfficeActions_ContextCancellation(t *testing.T) {
	server, client := setupOAMockServer(t)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := client.SearchOfficeActions(ctx, "*:*", 0, 10)
	if err == nil {
		t.Fatal("Expected error for cancelled context")
	}
}

func TestGetOfficeActionFields_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid API key"}`))
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL, APIKey: "bad-key", MaxRetries: 0, Timeout: 10})

	_, err := client.GetOfficeActionFields(context.Background())
	if err == nil {
		t.Fatal("Expected error for 401 response")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("Expected APIError, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("Expected status 401, got %d", apiErr.StatusCode)
	}
}

func TestGetOfficeActionFields_WrongMethod(t *testing.T) {
	server, client := setupOAMockServer(t)
	defer server.Close()

	// The mock validates GET method; this test verifies the client sends GET
	result, err := client.GetOfficeActionFields(context.Background())
	if err != nil {
		t.Fatalf("GetOfficeActionFields should use GET: %v", err)
	}
	if result.APIKey != "oa_actions" {
		t.Errorf("Unexpected apiKey: %s", result.APIKey)
	}
}

func TestNormalizeDSAPIApplicationNumber(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"digits only", "17248024", "17248024"},
		{"with slash and comma", "17/248,024", "17248024"},
		{"US prefix", "US17/248,024", "17248024"},
		{"US prefix with space", "US 17/248,024", "17248024"},
		{"lowercase us prefix", "us17248024", "17248024"},
		{"leading/trailing spaces", " 17248024 ", "17248024"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeDSAPIApplicationNumber(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeDSAPIApplicationNumber(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDSAPIApplicationCriteria(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"application number", "17/248,024", "patentApplicationNumber:17248024"},
		{"with US prefix", "US17248024", "patentApplicationNumber:17248024"},
		{"empty defaults to match all", "", "*:*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DSAPIApplicationCriteria(tt.input)
			if got != tt.want {
				t.Errorf("DSAPIApplicationCriteria(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
