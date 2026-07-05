package odp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const testGrantXML = `<?xml version="1.0" encoding="UTF-8"?>
<us-patent-grant lang="EN" dtd-version="v4.7 2022-02-17" country="US">
<us-bibliographic-data-grant>
<invention-title id="d2e53">Test Invention</invention-title>
</us-bibliographic-data-grant>
<abstract id="abstract"><p id="p-0001">An abstract.</p></abstract>
<claims id="claims"><claim id="CLM-00001" num="00001"><claim-text>A device.</claim-text></claim></claims>
</us-patent-grant>`

func newDownloadTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	cfg := DefaultConfig()
	cfg.BaseURL = baseURL
	cfg.APIKey = "secret-key"
	cfg.MaxRetries = 0
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

// DownloadXML/DownloadXMLWithType must reject URLs outside the configured host's
// dataset file path: the request carries the API key, so an attacker-supplied
// FileLocationURI would otherwise exfiltrate it.
func TestDownloadXML_RejectsForeignURL(t *testing.T) {
	var attackerHits atomic.Int32
	attacker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attackerHits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer attacker.Close()

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer api.Close()

	client := newDownloadTestClient(t, api.URL)

	tests := []struct {
		name string
		url  string
	}{
		{"foreign host", attacker.URL + "/api/v1/datasets/products/files/PTGRXML-SPLT/2023/x.xml"},
		{"own host wrong path", api.URL + "/steal/creds.xml"},
		{"empty URL", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := client.DownloadXML(context.Background(), tt.url); err == nil {
				t.Errorf("DownloadXML(%q) succeeded, want validation error", tt.url)
			}
			if _, err := client.DownloadXMLWithType(context.Background(), tt.url, DocumentTypeGrant); err == nil {
				t.Errorf("DownloadXMLWithType(%q) succeeded, want validation error", tt.url)
			}
		})
	}
	if n := attackerHits.Load(); n != 0 {
		t.Errorf("attacker host received %d request(s); the API key was sent off-host", n)
	}
}

// A valid FileLocationURI under the configured host must still download, carry
// the API key, and parse.
func TestDownloadXML_FetchesValidatedURI(t *testing.T) {
	const filePath = "/api/v1/datasets/products/files/PTGRXML-SPLT/2023/ipg230509/17248024_11646472.xml"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != filePath {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if got := r.Header.Get("X-API-Key"); got != "secret-key" {
			t.Errorf("X-API-Key = %q, want secret-key", got)
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(testGrantXML))
	}))
	defer server.Close()

	client := newDownloadTestClient(t, server.URL)
	doc, err := client.DownloadXML(context.Background(), server.URL+filePath)
	if err != nil {
		t.Fatalf("DownloadXML: %v", err)
	}
	if doc.GetDocumentType() != DocumentTypeGrant {
		t.Errorf("GetDocumentType() = %v, want grant", doc.GetDocumentType())
	}
	if doc.GetTitle() != "Test Invention" {
		t.Errorf("GetTitle() = %q, want Test Invention", doc.GetTitle())
	}
}

// The XML download path now goes through streamDownload, so failures surface as
// typed *APIError values with body and Retry-After preserved.
func TestDownloadXMLWithType_TypedError(t *testing.T) {
	const filePath = "/api/v1/datasets/products/files/PTGRXML-SPLT/2023/ipg230509/x.xml"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "7")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"slow down"}`))
	}))
	defer server.Close()

	client := newDownloadTestClient(t, server.URL)
	_, err := client.DownloadXMLWithType(context.Background(), server.URL+filePath, DocumentTypeGrant)
	if err == nil {
		t.Fatal("expected error for 429 response")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusTooManyRequests {
		t.Errorf("StatusCode = %d, want 429", apiErr.StatusCode)
	}
	if !strings.Contains(apiErr.Body, "slow down") {
		t.Errorf("Body = %q, want server body preserved", apiErr.Body)
	}
	if apiErr.RetryAfter != 7*time.Second {
		t.Errorf("RetryAfter = %v, want 7s", apiErr.RetryAfter)
	}
}

// A trailing slash on BaseURL must not break download-URL prefix validation
// (BaseURL+path would produce "//api/..." which never matches).
func TestValidateDownloadURLs_TrailingSlashBaseURL(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BaseURL = "https://api.uspto.gov/"
	cfg.APIKey = "test"
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	tests := []struct {
		name     string
		validate func(string) error
		url      string
		wantErr  bool
	}{
		{
			"file URI accepted",
			client.validateFileDownloadURI,
			"https://api.uspto.gov/api/v1/datasets/products/files/PTGRXML-SPLT/2023/x.xml",
			false,
		},
		{
			"file URI foreign host rejected",
			client.validateFileDownloadURI,
			"https://evil.example/api/v1/datasets/products/files/x.xml",
			true,
		},
		{
			"document URL accepted",
			client.validateDocumentDownloadURL,
			"https://api.uspto.gov/api/v1/download/applications/17248024/ABC123.pdf",
			false,
		},
		{
			"document URL foreign host rejected",
			client.validateDocumentDownloadURL,
			"https://evil.example/api/v1/download/applications/17248024/ABC123.pdf",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validate(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

// Streaming downloads must not be cut off by Config.Timeout once the body is
// flowing: the timeout bounds only the wait for response headers.
func TestStreamDownload_NotBoundByOverallTimeout(t *testing.T) {
	const filePath = "/api/v1/datasets/products/files/PTGRXML-SPLT/2023/big.zip"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		flusher := w.(http.Flusher)
		w.WriteHeader(http.StatusOK)
		for range 3 {
			_, _ = w.Write(bytes.Repeat([]byte("d"), 1024))
			flusher.Flush()
			time.Sleep(150 * time.Millisecond)
		}
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.APIKey = "test"
	cfg.MaxRetries = 0
	cfg.Timeout = 200 * time.Millisecond // shorter than the ~450ms transfer
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	var buf bytes.Buffer
	if err := client.DownloadBulkFile(context.Background(), server.URL+filePath, &buf); err != nil {
		t.Fatalf("DownloadBulkFile: %v", err)
	}
	if buf.Len() != 3*1024 {
		t.Errorf("downloaded %d bytes, want %d", buf.Len(), 3*1024)
	}
}

func TestCheckJSONPayload(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		body          string
		decoded       bool
		wantErr       bool
		wantRetryable bool
	}{
		{"decoded payload", 200, `{"count":1}`, true, false, false},
		{"empty body", 200, "", false, true, true},
		{"whitespace body", 204, " \n", false, true, true},
		{"non-JSON body", 200, "<html>gateway</html>", false, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkJSONPayload(tt.statusCode, []byte(tt.body), tt.decoded)
			if (err != nil) != tt.wantErr {
				t.Fatalf("checkJSONPayload() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				return
			}
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected *APIError, got %T", err)
			}
			if apiErr.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.statusCode)
			}
			if apiErr.IsRetryable() != tt.wantRetryable {
				t.Errorf("IsRetryable() = %v, want %v", apiErr.IsRetryable(), tt.wantRetryable)
			}
		})
	}
}

// A 2xx whose body did not decode as JSON must produce a typed error, not a
// silent (nil, nil) success.
func TestSearchAndGet_NonJSON2xxIsTypedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html>upstream proxy error</html>"))
	}))
	defer server.Close()

	client := newDownloadTestClient(t, server.URL)
	ctx := context.Background()

	t.Run("SearchPatents", func(t *testing.T) {
		result, err := client.SearchPatents(ctx, "test", 0, 1)
		if err == nil {
			t.Fatalf("expected error, got result %v", result)
		}
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected *APIError, got %T: %v", err, err)
		}
		if !strings.Contains(apiErr.Body, "upstream proxy error") {
			t.Errorf("Body = %q, want body preview", apiErr.Body)
		}
	})

	t.Run("GetPatent", func(t *testing.T) {
		result, err := client.GetPatent(ctx, "17123456")
		if err == nil {
			t.Fatalf("expected error, got result %v", result)
		}
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected *APIError, got %T: %v", err, err)
		}
	})
}

// Publication resolution must match any kind code via a wildcard:
// earliestPublicationNumber carries the earliest (usually A1) publication, so a
// caller-supplied A2/A9 kind - or an assumed A1 - would miss records.
func TestResolvePatentNumber_PublicationKindWildcard(t *testing.T) {
	inputs := []string{"20250087686", "US20250087686A9", "US20250087686A1"}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			var gotQuery string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/patent/applications/search" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				body, _ := io.ReadAll(r.Body)
				var req struct {
					Q string `json:"q"`
				}
				_ = json.Unmarshal(body, &req)
				gotQuery = req.Q
				w.Header().Set("Content-Type", "application/json")
				writeWrapperBag(w, "18123456", "SOME TITLE")
			}))
			defer server.Close()

			client := newDownloadTestClient(t, server.URL)
			app, err := client.ResolvePatentNumber(context.Background(), input)
			if err != nil {
				t.Fatalf("ResolvePatentNumber(%q): %v", input, err)
			}
			if app != "18123456" {
				t.Errorf("resolved application = %q, want 18123456", app)
			}
			want := fmt.Sprintf("applicationMetaData.earliestPublicationNumber:%s", "US20250087686A*")
			if gotQuery != want {
				t.Errorf("query = %q, want %q", gotQuery, want)
			}
		})
	}
}
