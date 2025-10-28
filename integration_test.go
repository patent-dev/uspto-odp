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

// TestXMLParsing tests XML download and parsing with real API data
func TestXMLParsing(t *testing.T) {
	apiKey := os.Getenv("USPTO_API_KEY")
	if apiKey == "" {
		t.Fatal("USPTO_API_KEY environment variable is required. Set it before running tests")
	}

	config := &Config{
		BaseURL:    "https://api.uspto.gov",
		APIKey:     apiKey,
		MaxRetries: 2,
		RetryDelay: 1,
		Timeout:    60, // Longer timeout for XML downloads
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Run("GetPatentXML - Grant Document", func(t *testing.T) {
		// US 11,646,472 B2 (Application 17/248,024) - Known granted patent
		appNumber := "17248024"

		t.Logf("Fetching XML for application %s (US 11,646,472 B2)...", appNumber)

		doc, err := client.GetPatentXML(ctx, appNumber)
		if err != nil {
			t.Fatalf("GetPatentXML failed: %v", err)
		}

		if doc == nil {
			t.Fatal("Expected document, got nil")
		}

		// Check document type
		docType := doc.GetDocumentType()
		t.Logf("Document type: %v", docType)

		if docType == DocumentTypeUnknown {
			t.Fatal("Document type is unknown")
		}

		// Extract title
		title := doc.GetTitle()
		if title == "" {
			t.Error("Title is empty")
		} else {
			t.Logf("Title: %s", title)
		}

		// Extract abstract
		abstract := doc.GetAbstract()
		if abstract == nil {
			t.Error("Abstract is nil")
		} else {
			abstractText := abstract.ExtractAbstractText()
			if abstractText == "" {
				t.Error("Abstract text is empty")
			} else {
				t.Logf("Abstract length: %d characters", len(abstractText))
				// Log first 200 chars
				preview := abstractText
				if len(preview) > 200 {
					preview = preview[:200] + "..."
				}
				t.Logf("Abstract preview: %s", preview)
			}
		}

		// Extract description
		description := doc.GetDescription()
		if description == nil {
			t.Error("Description is nil")
		} else {
			descText := description.ExtractDescriptionText()
			if descText == "" {
				t.Error("Description text is empty")
			} else {
				t.Logf("Description length: %d characters", len(descText))
			}
		}

		// Extract claims
		claims := doc.GetClaims()
		if claims == nil {
			t.Error("Claims is nil")
		} else {
			if len(claims.ClaimList) == 0 {
				t.Error("No claims found")
			} else {
				t.Logf("Total claims: %d", len(claims.ClaimList))

				// Extract first claim
				firstClaim := claims.ClaimList[0]
				claimText := firstClaim.ExtractClaimText()
				if claimText == "" {
					t.Error("First claim text is empty")
				} else {
					t.Logf("First claim length: %d characters", len(claimText))
					preview := claimText
					if len(preview) > 200 {
						preview = preview[:200] + "..."
					}
					t.Logf("First claim preview: %s", preview)
				}

				// Extract all claims formatted
				allClaimsText := claims.ExtractAllClaimsTextFormatted()
				if allClaimsText == "" {
					t.Error("All claims text is empty")
				} else {
					t.Logf("All claims formatted length: %d characters", len(allClaimsText))
				}
			}
		}

		t.Log("Success: XML parsed and all text extracted successfully")
	})

	t.Run("GetXMLURLForApplication", func(t *testing.T) {
		appNumber := "17248024"

		xmlURL, docType, err := client.GetXMLURLForApplication(ctx, appNumber)
		if err != nil {
			t.Fatalf("GetXMLURLForApplication failed: %v", err)
		}

		if xmlURL == "" {
			t.Fatal("XML URL is empty")
		}

		t.Logf("XML URL: %s", xmlURL)
		t.Logf("Document type: %v", docType)

		// Verify URL format
		if !strings.HasPrefix(xmlURL, "https://") {
			t.Errorf("Expected HTTPS URL, got: %s", xmlURL)
		}
	})

	t.Run("DownloadXML - Direct URL", func(t *testing.T) {
		appNumber := "17248024"

		// Get XML URL first
		xmlURL, _, err := client.GetXMLURLForApplication(ctx, appNumber)
		if err != nil {
			t.Fatalf("GetXMLURLForApplication failed: %v", err)
		}

		// Download and parse
		doc, err := client.DownloadXML(ctx, xmlURL)
		if err != nil {
			t.Fatalf("DownloadXML failed: %v", err)
		}

		if doc == nil {
			t.Fatal("Expected document, got nil")
		}

		// Verify we can extract data
		title := doc.GetTitle()
		if title == "" {
			t.Error("Title is empty")
		}

		claims := doc.GetClaims()
		if claims == nil || len(claims.ClaimList) == 0 {
			t.Error("No claims found")
		}

		t.Log("Success: XML downloaded and parsed from direct URL")
	})

	t.Run("ParseXML - Real Document Structure", func(t *testing.T) {
		appNumber := "17248024"

		// Get the document
		doc, err := client.GetPatentXML(ctx, appNumber)
		if err != nil {
			t.Fatalf("GetPatentXML failed: %v", err)
		}

		// Test comprehensive extraction
		title := doc.GetTitle()
		abstract := doc.GetAbstract()
		description := doc.GetDescription()
		claims := doc.GetClaims()

		// Calculate total text extracted
		totalLength := len(title)
		if abstract != nil {
			totalLength += len(abstract.ExtractAbstractText())
		}
		if description != nil {
			totalLength += len(description.ExtractDescriptionText())
		}
		if claims != nil {
			totalLength += len(claims.ExtractAllClaimsTextFormatted())
		}

		t.Logf("Total text extracted: %d characters", totalLength)

		// Verify we extracted significant content
		if totalLength < 1000 {
			t.Errorf("Total text seems too short (%d chars), may indicate parsing issues", totalLength)
		}

		t.Log("Success: Comprehensive text extraction completed")
	})

	t.Run("ResolvePatentNumber - Grant Number", func(t *testing.T) {
		// Grant 11646472 corresponds to application 17248024
		appNumber, err := client.ResolvePatentNumber(ctx, "US 11,646,472 B2")
		if err != nil {
			t.Fatalf("ResolvePatentNumber failed: %v", err)
		}

		expectedAppNumber := "17248024"
		if appNumber != expectedAppNumber {
			t.Errorf("Expected application number %s, got %s", expectedAppNumber, appNumber)
		}

		t.Logf("Success: Grant 11646472 resolved to application %s", appNumber)
	})

	t.Run("ResolvePatentNumber - Application Number", func(t *testing.T) {
		// Application number should return as-is (normalized)
		appNumber, err := client.ResolvePatentNumber(ctx, "17/248,024")
		if err != nil {
			t.Fatalf("ResolvePatentNumber failed: %v", err)
		}

		expectedAppNumber := "17248024"
		if appNumber != expectedAppNumber {
			t.Errorf("Expected application number %s, got %s", expectedAppNumber, appNumber)
		}

		t.Logf("Success: Application 17/248,024 normalized to %s", appNumber)
	})

	t.Run("GetPatent - Grant Number Format", func(t *testing.T) {
		// Test GetPatent with grant number - should resolve to application number
		result, err := client.GetPatent(ctx, "US 11,646,472 B2")
		if err != nil {
			t.Fatalf("GetPatent with grant number failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		if result.PatentFileWrapperDataBag == nil || len(*result.PatentFileWrapperDataBag) == 0 {
			t.Fatal("Expected patent data, got none")
		}

		patent := (*result.PatentFileWrapperDataBag)[0]
		if patent.ApplicationNumberText == nil {
			t.Fatal("Application number missing from response")
		}

		expectedAppNumber := "17248024"
		if *patent.ApplicationNumberText != expectedAppNumber {
			t.Errorf("Expected application number %s, got %s", expectedAppNumber, *patent.ApplicationNumberText)
		}

		// Verify it's the correct patent
		if patent.ApplicationMetaData != nil && patent.ApplicationMetaData.PatentNumber != nil {
			if *patent.ApplicationMetaData.PatentNumber != "11646472" {
				t.Errorf("Expected patent number 11646472, got %s", *patent.ApplicationMetaData.PatentNumber)
			}
		}

		t.Logf("Success: GetPatent with grant number 11646472 returned application %s", *patent.ApplicationNumberText)
	})

	t.Run("GetPatent - Various Formats", func(t *testing.T) {
		// Test that all these formats resolve to the same patent
		formats := []string{
			"17248024",
			"17/248,024",
			"US 11,646,472 B2",
			"11,646,472",
		}

		var results []*generated.PatentDataResponse
		for _, format := range formats {
			result, err := client.GetPatent(ctx, format)
			if err != nil {
				t.Errorf("GetPatent failed for format %s: %v", format, err)
				continue
			}
			results = append(results, result)
			t.Logf("Successfully retrieved patent using format: %s", format)
		}

		// Verify all results point to the same application
		if len(results) > 1 {
			firstAppNum := (*results[0].PatentFileWrapperDataBag)[0].ApplicationNumberText
			for i, result := range results[1:] {
				appNum := (*result.PatentFileWrapperDataBag)[0].ApplicationNumberText
				if *appNum != *firstAppNum {
					t.Errorf("Format %s returned different application number: %s vs %s",
						formats[i+1], *appNum, *firstAppNum)
				}
			}
		}

		t.Log("Success: All patent number formats resolved to the same application")
	})
}
