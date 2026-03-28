package odp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupTSDRMockServer(t *testing.T) (*httptest.Server, *Client) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {

		// LoadXML - returns JSON when Accept header requests it, XML otherwise
		case "/ts/cd/casestatus/97123456/info":
			if r.Header.Get("Accept") == "application/json" {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"trademarks": []any{
						map[string]any{
							"status": map[string]any{
								"markInfo": map[string]any{
									"markText": "TEST MARK",
								},
								"prosecution": map[string]any{
									"statusText": "REGISTERED",
									"statusDate": "2024-01-15",
								},
							},
						},
					},
				})
			} else {
				w.Header().Set("Content-Type", "application/xml")
				_, _ = w.Write([]byte(`<TrademarkCase><SerialNumber>97123456</SerialNumber></TrademarkCase>`))
			}

		// GetCaseDocsInfoXml - returns JSON or XML based on Accept header
		case "/ts/cd/casedocs/97123456/info":
			if r.Header.Get("Accept") == "application/json" {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode([]map[string]any{
					{"documentId": "DOC001", "documentType": "Office Action"},
					{"documentId": "DOC002", "documentType": "Response"},
				})
			} else {
				w.Header().Set("Content-Type", "application/xml")
				_, _ = w.Write([]byte(`<DocumentList><Document><DocumentID>DOC001</DocumentID></Document></DocumentList>`))
			}

		// GetDocumentInfoXml - returns XML (endpoint does not support JSON)
		case "/ts/cd/casedoc/97123456/NOA20230322/info":
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Document><SerialNumber>97123456</SerialNumber><DocumentTypeCode>NOA</DocumentTypeCode><MailRoomDate>2023-03-22</MailRoomDate></Document>`))

		// GetDocumentContentPdf
		case "/ts/cd/casedoc/97123456/NOA20230322/content.pdf":
			w.Header().Set("Content-Type", "application/pdf")
			_, _ = w.Write([]byte("%PDF-1.4 fake pdf content"))

		// GetcaseUpdateInfo - matches CaseUpdateInfoList schema
		case "/last-update/info.json":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"caseUpdateInfo": []any{
					map[string]any{
						"name":  "lastModifiedDate",
						"value": "2024-03-15T10:30:00Z",
					},
				},
			})

		// GetList (multi-status) - matches TransactionBag schema
		case "/ts/cd/caseMultiStatus/sn":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"size": 2,
				"transactionList": []any{
					map[string]any{
						"searchId": "97123456",
					},
					map[string]any{
						"searchId": "97654321",
					},
				},
			})

		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found", "path": r.URL.Path})
		}
	}))

	config := &Config{
		BaseURL:     server.URL,
		APIKey:      "test-key",
		TSDRBaseURL: server.URL,
		TSDRAPIKey:  "test-tsdr-key",
		MaxRetries:  0,
		Timeout:     10,
		UserAgent:   "test",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	return server, client
}

func TestGetTrademarkStatusJSON(t *testing.T) {
	server, client := setupTSDRMockServer(t)
	defer server.Close()

	result, err := client.GetTrademarkStatusJSON(context.Background(), "97123456")
	if err != nil {
		t.Fatalf("GetTrademarkStatusJSON failed: %v", err)
	}
	if len(result.Trademarks) != 1 {
		t.Fatalf("Expected 1 trademark, got %d", len(result.Trademarks))
	}
	tm := result.Trademarks[0]
	status, ok := tm["status"].(map[string]any)
	if !ok {
		t.Fatal("Expected status map")
	}
	markInfo, ok := status["markInfo"].(map[string]any)
	if !ok {
		t.Fatal("Expected markInfo map")
	}
	if markInfo["markText"] != "TEST MARK" {
		t.Errorf("Expected markText 'TEST MARK', got %v", markInfo["markText"])
	}
}

func TestGetTrademarkStatus(t *testing.T) {
	server, client := setupTSDRMockServer(t)
	defer server.Close()

	resp, err := client.GetTrademarkStatus(context.Background(), "97123456")
	if err != nil {
		t.Fatalf("GetTrademarkStatus failed: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected non-nil response")
	}
	if resp.StatusCode() != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode())
	}
	if resp.Body == nil {
		t.Fatal("Expected non-nil body")
	}
}

func TestGetTrademarkDocumentsXML(t *testing.T) {
	server, client := setupTSDRMockServer(t)
	defer server.Close()

	result, err := client.GetTrademarkDocumentsXML(context.Background(), "97123456")
	if err != nil {
		t.Fatalf("GetTrademarkDocuments failed: %v", err)
	}
	if !bytes.Contains(result, []byte("DocumentList")) {
		t.Error("Expected XML with DocumentList element")
	}
}

func TestGetTrademarkDocumentInfo(t *testing.T) {
	server, client := setupTSDRMockServer(t)
	defer server.Close()

	result, err := client.GetTrademarkDocumentInfo(context.Background(), "97123456", "NOA20230322")
	if err != nil {
		t.Fatalf("GetTrademarkDocumentInfo failed: %v", err)
	}
	if !bytes.Contains(result, []byte("NOA")) {
		t.Error("Expected XML with NOA document type")
	}
	if !bytes.Contains(result, []byte("97123456")) {
		t.Error("Expected XML with serial number")
	}
}

func TestDownloadTrademarkDocument(t *testing.T) {
	server, client := setupTSDRMockServer(t)
	defer server.Close()

	var buf bytes.Buffer
	err := client.DownloadTrademarkDocument(context.Background(), "97123456", "NOA20230322", &buf)
	if err != nil {
		t.Fatalf("DownloadTrademarkDocument failed: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF")) {
		t.Error("Expected PDF content")
	}
}

func TestDownloadTrademarkDocument_WithRetries(t *testing.T) {
	server, _ := setupTSDRMockServer(t)
	defer server.Close()

	// Create client with retries enabled to exercise the buffered path
	config := &Config{
		BaseURL:     server.URL,
		APIKey:      "test-key",
		TSDRBaseURL: server.URL,
		TSDRAPIKey:  "test-tsdr-key",
		MaxRetries:  1,
		RetryDelay:  0,
		Timeout:     10,
		UserAgent:   "test",
	}
	client, err := NewClient(config)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = client.DownloadTrademarkDocument(context.Background(), "97123456", "NOA20230322", &buf)
	if err != nil {
		t.Fatalf("DownloadTrademarkDocument with retries failed: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF")) {
		t.Error("Expected PDF content")
	}
}

func TestGetTrademarkLastUpdate(t *testing.T) {
	server, client := setupTSDRMockServer(t)
	defer server.Close()

	result, err := client.GetTrademarkLastUpdate(context.Background(), "97123456")
	if err != nil {
		t.Fatalf("GetTrademarkLastUpdate failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.CaseUpdateInfo == nil || len(*result.CaseUpdateInfo) == 0 {
		t.Fatal("Expected non-empty CaseUpdateInfo")
	}
	info := (*result.CaseUpdateInfo)[0]
	if info.Name == nil || *info.Name != "lastModifiedDate" {
		t.Errorf("Expected name 'lastModifiedDate', got %v", info.Name)
	}
}

func TestGetTrademarkMultiStatus(t *testing.T) {
	server, client := setupTSDRMockServer(t)
	defer server.Close()

	result, err := client.GetTrademarkMultiStatus(context.Background(), "sn", []string{"97123456", "97654321"})
	if err != nil {
		t.Fatalf("GetTrademarkMultiStatus failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.TransactionList == nil || len(*result.TransactionList) != 2 {
		t.Errorf("Expected 2 transactions, got %v", result.TransactionList)
	}
	if result.Size == nil || *result.Size != 2 {
		t.Errorf("Expected size 2, got %v", result.Size)
	}
}

func TestGetTrademarkMultiStatus_InvalidType(t *testing.T) {
	server, client := setupTSDRMockServer(t)
	defer server.Close()

	_, err := client.GetTrademarkMultiStatus(context.Background(), "invalid", []string{"97123456"})
	if err == nil {
		t.Fatal("Expected error for invalid pType")
	}
	if !strings.Contains(err.Error(), "invalid type") {
		t.Errorf("Expected 'invalid type' error, got: %v", err)
	}
}

func TestTSDR_NotConfigured(t *testing.T) {
	// Client without TSDR key - tsdr field should be nil
	config := &Config{
		BaseURL:    "http://localhost",
		APIKey:     "test-key",
		MaxRetries: 0,
		Timeout:    10,
	}
	client, err := NewClient(config)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetTrademarkStatusJSON(context.Background(), "97123456")
	if err == nil {
		t.Fatal("Expected error for unconfigured TSDR")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("TSDR client not configured")) {
		t.Errorf("Expected 'TSDR client not configured' error, got: %v", err)
	}
}

func TestTSDR_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
	}))
	defer server.Close()

	config := &Config{
		BaseURL:     server.URL,
		APIKey:      "test",
		TSDRBaseURL: server.URL,
		TSDRAPIKey:  "test",
		MaxRetries:  0,
		Timeout:     10,
	}
	client, _ := NewClient(config)

	_, err := client.GetTrademarkStatusJSON(context.Background(), "97123456")
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
}

func TestTSDR_ContextCancellation(t *testing.T) {
	server, client := setupTSDRMockServer(t)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetTrademarkStatusJSON(ctx, "97123456")
	if err == nil {
		t.Fatal("Expected error for cancelled context")
	}
}

func TestTSDR_InvalidSerialNumber(t *testing.T) {
	server, client := setupTSDRMockServer(t)
	defer server.Close()

	tests := []struct {
		name   string
		serial string
	}{
		{"too short", "1234567"},
		{"too long", "123456789"},
		{"letters", "abcdefgh"},
		{"empty", ""},
		{"with slash", "9712/456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GetTrademarkStatusJSON(context.Background(), tt.serial)
			if err == nil {
				t.Fatal("Expected error for invalid serial number")
			}
			if !strings.Contains(err.Error(), "invalid serial number") {
				t.Errorf("Expected 'invalid serial number' error, got: %v", err)
			}
		})
	}
}
