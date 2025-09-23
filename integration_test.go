//go:build integration
// +build integration

package odp

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/patent-dev/uspto-odp/generated"
)

// TestIntegrationWithRealAPI tests against the actual USPTO API
// Run with: go test -tags=integration -v
func TestIntegrationWithRealAPI(t *testing.T) {
	apiKey := os.Getenv("USPTO_API_KEY")
	if apiKey == "" {
		t.Fatal("USPTO_API_KEY environment variable is required. Set it before running tests")
	}

	config := &Config{
		BaseURL:    "https://api.uspto.gov",
		APIKey:     apiKey,
		MaxRetries: 2,
		RetryDelay: 1,
		Timeout:    30,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Run("GetStatusCodes - Working Endpoint", func(t *testing.T) {
		result, err := client.GetStatusCodes(ctx)
		if err != nil {
			t.Fatalf("GetStatusCodes failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		if result.Count == nil || *result.Count == 0 {
			t.Error("Expected status codes count > 0")
		}

		if result.StatusCodeBag != nil && len(*result.StatusCodeBag) > 0 {
			first := (*result.StatusCodeBag)[0]
			t.Logf("Success: Retrieved %d status codes", *result.Count)
			if first.ApplicationStatusCode != nil && first.ApplicationStatusDescriptionText != nil {
				t.Logf("   Example: Code %d = %s",
					*first.ApplicationStatusCode,
					*first.ApplicationStatusDescriptionText)
			}
		}
	})

	t.Run("SearchPatents", func(t *testing.T) {
		result, err := client.SearchPatents(ctx, "artificial intelligence", 0, 2)
		if err != nil {
			t.Fatalf("SearchPatents failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		if result.Count == nil || *result.Count == 0 {
			t.Error("Expected patents count > 0")
		}

		t.Logf("Success: Found %d patents matching 'artificial intelligence'", *result.Count)
	})

	t.Run("GetPatent", func(t *testing.T) {
		result, err := client.GetPatent(ctx, "17123456")
		if err != nil {
			t.Fatalf("GetPatent failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		t.Logf("Success: Retrieved patent 17123456")
	})

	t.Run("GetPatentAdjustment", func(t *testing.T) {
		result, err := client.GetPatentAdjustment(ctx, "17123456")
		if err != nil {
			t.Fatalf("GetPatentAdjustment failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		t.Logf("Success: Retrieved patent adjustment data")
	})

	t.Run("GetPatentContinuity", func(t *testing.T) {
		result, err := client.GetPatentContinuity(ctx, "17123456")
		if err != nil {
			t.Fatalf("GetPatentContinuity failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		t.Logf("Success: Retrieved patent continuity data")
	})

	t.Run("GetPatentDocuments", func(t *testing.T) {
		result, err := client.GetPatentDocuments(ctx, "17123456")
		if err != nil {
			t.Fatalf("GetPatentDocuments failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		t.Logf("Success: Retrieved patent documents")
	})

	t.Run("SearchBulkProducts", func(t *testing.T) {
		result, err := client.SearchBulkProducts(ctx, "patent grant", 0, 1)
		if err != nil {
			t.Fatalf("SearchBulkProducts failed: %v", err)
		}

		if result != nil && result.Count != nil {
			t.Logf("Success: Found %d bulk products", *result.Count)
		}
	})

	t.Run("GetBulkProduct", func(t *testing.T) {
		result, err := client.GetBulkProduct(ctx, "PTGRXML")
		if err != nil {
			t.Fatalf("GetBulkProduct failed: %v", err)
		}

		if result != nil && result.Count != nil {
			t.Logf("Success: Retrieved bulk product PTGRXML")

			// Check if we actually have products
			if result.BulkDataProductBag != nil && len(*result.BulkDataProductBag) > 0 {
				product := (*result.BulkDataProductBag)[0]
				if product.ProductIdentifier != nil {
					t.Logf("   Product ID: %s", *product.ProductIdentifier)
				}
				if product.ProductFileBag != nil && product.ProductFileBag.Count != nil {
					t.Logf("   File count: %d", *product.ProductFileBag.Count)
				}
			}
		}
	})

	t.Run("SearchPetitions", func(t *testing.T) {
		result, err := client.SearchPetitions(ctx, "revival", 0, 2)
		if err != nil {
			t.Fatalf("SearchPetitions failed: %v", err)
		}

		if result != nil && result.Count != nil {
			t.Logf("Success: Found %d petition decisions", *result.Count)
		}
	})

	// Test ALL remaining endpoints for complete coverage

	t.Run("GetPatentMetaData", func(t *testing.T) {
		result, err := client.GetPatentMetaData(ctx, "17123456")
		if err != nil {
			t.Fatalf("GetPatentMetaData failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		t.Logf("Success: Retrieved patent metadata")
	})

	t.Run("GetPatentAssignment", func(t *testing.T) {
		result, err := client.GetPatentAssignment(ctx, "17123456")
		if err != nil {
			t.Fatalf("GetPatentAssignment failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		t.Logf("Success: Retrieved patent assignment data")
	})

	t.Run("GetPatentAssociatedDocuments", func(t *testing.T) {
		result, err := client.GetPatentAssociatedDocuments(ctx, "17123456")
		if err != nil {
			t.Fatalf("GetPatentAssociatedDocuments failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		t.Logf("Success: Retrieved patent associated documents")
	})

	t.Run("GetPatentAttorney", func(t *testing.T) {
		result, err := client.GetPatentAttorney(ctx, "17123456")
		if err != nil {
			t.Fatalf("GetPatentAttorney failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		t.Logf("Success: Retrieved patent attorney information")
	})

	t.Run("GetPatentForeignPriority", func(t *testing.T) {
		result, err := client.GetPatentForeignPriority(ctx, "17123456")
		if err != nil {
			t.Fatalf("GetPatentForeignPriority failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		t.Logf("Success: Retrieved patent foreign priority data")
	})

	t.Run("GetPatentTransactions", func(t *testing.T) {
		result, err := client.GetPatentTransactions(ctx, "17123456")
		if err != nil {
			t.Fatalf("GetPatentTransactions failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		t.Logf("Success: Retrieved patent transactions")
	})

	t.Run("SearchPatentsDownload", func(t *testing.T) {
		format := generated.PatentDownloadRequestFormat("json")
		req := generated.PatentDownloadRequest{
			Q:      StringPtr("test"),
			Format: &format,
			Pagination: &generated.Pagination{
				Offset: Int32Ptr(0),
				Limit:  Int32Ptr(1),
			},
		}

		result, err := client.SearchPatentsDownload(ctx, req)
		if err != nil {
			t.Fatalf("SearchPatentsDownload failed: %v", err)
		}

		if result == nil || len(result) == 0 {
			t.Fatal("Expected download data")
		}

		t.Logf("Success: Downloaded patent search results (%d bytes)", len(result))
	})

	t.Run("GetPetitionDecision", func(t *testing.T) {
		// Use a sample record ID - this might not exist
		result, err := client.GetPetitionDecision(ctx, "9dc6b94a-afa0-5e66-beef-f26fa80992b8", false)
		if err != nil {
			t.Fatalf("GetPetitionDecision failed: %v", err)
		}

		if result != nil && result.Count != nil {
			t.Logf("Success: Retrieved petition decision with count %d", *result.Count)
		}
	})

	t.Run("SearchPetitionsDownload", func(t *testing.T) {
		format := generated.PetitionDecisionDownloadRequestFormat("json")
		req := generated.PetitionDecisionDownloadRequest{
			Q:      StringPtr("revival"),
			Format: &format,
			Pagination: &generated.Pagination{
				Offset: Int32Ptr(0),
				Limit:  Int32Ptr(1),
			},
		}

		result, err := client.SearchPetitionsDownload(ctx, req)
		if err != nil {
			t.Fatalf("SearchPetitionsDownload failed: %v", err)
		}

		if result == nil || len(result) == 0 {
			t.Fatal("Expected download data")
		}

		t.Logf("Success: Downloaded petition search results (%d bytes)", len(result))
	})

	t.Run("DownloadBulkFile", func(t *testing.T) {
		// Skip actual download test by default (files are very large)
		if os.Getenv("TEST_BULK_DOWNLOAD") != "true" {
			t.Skip("Skipping bulk file download test (set TEST_BULK_DOWNLOAD=true to run)")
		}

		// Get a real FileDownloadURI from the API
		result, err := client.GetBulkProduct(ctx, "PTGRXML")
		if err != nil {
			t.Fatalf("GetBulkProduct failed: %v", err)
		}

		if result.BulkDataProductBag == nil || len(*result.BulkDataProductBag) == 0 {
			t.Fatal("No product data found")
		}

		product := (*result.BulkDataProductBag)[0]
		if product.ProductFileBag == nil || product.ProductFileBag.FileDataBag == nil || len(*product.ProductFileBag.FileDataBag) == 0 {
			t.Fatal("No files found")
		}

		// Look for specific file ipg250916.zip
		var fileIndex int = -1
		for i, f := range *product.ProductFileBag.FileDataBag {
			if f.FileName != nil && strings.Contains(*f.FileName, "ipg250916.zip") {
				fileIndex = i
				break
			}
		}

		if fileIndex == -1 {
			// Fallback to first file if specific file not found
			fileIndex = 0
			t.Logf("Warning: ipg250916.zip not found, using first file")
		}

		file := (*product.ProductFileBag.FileDataBag)[fileIndex]
		if file.FileDownloadURI == nil {
			t.Fatal("No FileDownloadURI found")
		}

		// Use the new recommended API
		var buf bytes.Buffer
		err = client.DownloadBulkFile(ctx, *file.FileDownloadURI, &buf)
		if err != nil {
			t.Fatalf("DownloadBulkFile failed: %v", err)
		}

		if buf.Len() == 0 {
			t.Error("Downloaded file is empty")
		} else {
			t.Logf("Success: Downloaded bulk file using FileDownloadURI: %d bytes", buf.Len())
		}
	})

}

// TestEndpointCoverage documents which endpoints are implemented
func TestEndpointCoverage(t *testing.T) {
	endpoints := []struct {
		category string
		method   string
		path     string
		function string
	}{
		// Patent Application API (13 endpoints)
		{"Patent", "POST", "/api/v1/patent/applications/search", "SearchPatents"},
		{"Patent", "GET", "/api/v1/patent/applications/{applicationNumber}", "GetPatent"},
		{"Patent", "GET", "/api/v1/patent/applications/{applicationNumber}/meta-data", "GetPatentMetaData"},
		{"Patent", "GET", "/api/v1/patent/applications/{applicationNumber}/adjustment", "GetPatentAdjustment"},
		{"Patent", "GET", "/api/v1/patent/applications/{applicationNumber}/continuity", "GetPatentContinuity"},
		{"Patent", "GET", "/api/v1/patent/applications/{applicationNumber}/documents", "GetPatentDocuments"},
		{"Patent", "GET", "/api/v1/patent/applications/{applicationNumber}/assignment", "GetPatentAssignment"},
		{"Patent", "GET", "/api/v1/patent/applications/{applicationNumber}/associated-documents", "GetPatentAssociatedDocuments"},
		{"Patent", "GET", "/api/v1/patent/applications/{applicationNumber}/attorney", "GetPatentAttorney"},
		{"Patent", "GET", "/api/v1/patent/applications/{applicationNumber}/foreign-priority", "GetPatentForeignPriority"},
		{"Patent", "GET", "/api/v1/patent/applications/{applicationNumber}/transactions", "GetPatentTransactions"},
		{"Patent", "POST", "/api/v1/patent/applications/search/download", "SearchPatentsDownload"},
		{"Patent", "GET", "/api/v1/patent/status-codes", "GetStatusCodes"},

		// Bulk Data API (3 endpoints)
		{"Bulk", "GET", "/api/v1/datasets/products/search", "SearchBulkProducts"},
		{"Bulk", "GET", "/api/v1/datasets/products/{productId}", "GetBulkProduct"},
		{"Bulk", "GET", "/api/v1/datasets/products/files/{productId}/{fileName}", "DownloadBulkFile/DownloadBulkFileWithProgress"},

		// Petition API (3 endpoints)
		{"Petition", "POST", "/api/v1/petition/decisions/search", "SearchPetitions"},
		{"Petition", "GET", "/api/v1/petition/decisions/{recordId}", "GetPetitionDecision"},
		{"Petition", "POST", "/api/v1/petition/decisions/search/download", "SearchPetitionsDownload"},
	}

	t.Log("USPTO ODP API Client - Endpoint Coverage")
	t.Log("==========================================")

	for _, ep := range endpoints {
		t.Logf("[%s] %s %s -> %s()", ep.category, ep.method, ep.path, ep.function)
	}

	t.Logf("\nTotal endpoints implemented: %d", len(endpoints))
	t.Log("\nSuccess: ALL 19 USPTO ODP API endpoints are implemented and tested!")
}
