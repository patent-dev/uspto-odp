//go:build integration
// +build integration

package odp

import (
	"bytes"
	"errors"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/patent-dev/uspto-odp/generated"
)

// Per-endpoint live integration tests: exactly one TestIntegration<Method> for
// every exported Client method that maps to a USPTO endpoint (the coverage tool
// scripts/check-integration-coverage.sh enforces the 1:1 mapping). Each test runs
// against the real API and is honest: it PASSES on a good response or SKIPS
// cleanly, and never FAILS on a documented operational condition.
//
// Skips happen when:
//   - the required credential (USPTO_API_KEY, USPTO_TSDR_API_KEY) is absent;
//   - the API returns 401/403 (key not provisioned for this dataset),
//     404 (the sampled id aged out of the index), or 429 (rate limited);
//   - the response carries the known account limit "No permission of user=sip"
//     (an Office Action / TSDR entitlement this account lacks).
//
// Inputs are the stable, real ids used across the fixture set (application
// 17248024 = US 11,646,472 B2; trial PGR2025-00004; appeal 2026001845;
// interference 106130; petition 10347018; product PTGRXML), or are chained from a
// search/list call in the same test. Each test makes at most one extra "find an
// id" call, respecting the one-call-per-endpoint rate-limit guidance.
//
// Shared helpers (testCtx, testTimeout, the *flag and saveFixture) live in
// integration_test.go, which holds the original broader-scenario tests; this file
// adds the canonical per-endpoint layer.

const (
	// itApp is the primary patent test application (US 11,646,472 B2).
	itApp = "17248024"
	// itForeignApp has foreign-priority data.
	itForeignApp = "15000001"
	// itTrialNumber, itAppealNumber, itInterferenceNumber are stable PTAB ids.
	itTrialNumber        = "PGR2025-00004"
	itAppealNumber       = "2026001845"
	itInterferenceNumber = "106130"
	// itBulkProduct is a stable bulk-data product identifier.
	itBulkProduct = "PTGRXML"
	// itSerial is a trademark serial number for TSDR tests.
	itSerial = "97123456"
)

// newITClient builds a client from USPTO_API_KEY (skipping if absent). For tests
// that also need TSDR, set needTSDR so the TSDR key is required too.
func newITClient(t *testing.T, needTSDR bool) *Client {
	t.Helper()
	apiKey := os.Getenv("USPTO_API_KEY")
	if apiKey == "" {
		t.Skip("USPTO_API_KEY environment variable is required")
	}
	cfg := &Config{
		BaseURL:    "https://api.uspto.gov",
		APIKey:     apiKey,
		MaxRetries: 2,
		RetryDelay: 1 * time.Second,
		Timeout:    90 * time.Second,
	}
	if needTSDR {
		tsdrKey := os.Getenv("USPTO_TSDR_API_KEY")
		if tsdrKey == "" {
			t.Skip("USPTO_TSDR_API_KEY environment variable is required")
		}
		cfg.TSDRAPIKey = tsdrKey
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

// skipExpected turns a documented operational error into a clean t.Skip and
// leaves any other error for the caller to fail on. It returns true when it
// skipped (the caller should return).
func skipExpected(t *testing.T, err error) bool {
	t.Helper()
	if err == nil {
		return false
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case 401, 403:
			t.Skipf("skip: %d not provisioned for this account: %s", apiErr.StatusCode, apiErr.Message)
			return true
		case 404:
			t.Skipf("skip: 404 sampled id aged out of the index: %s", apiErr.Message)
			return true
		case 429:
			t.Skipf("skip: 429 rate limited: %s", apiErr.Message)
			return true
		}
	}
	// Known account entitlement limit surfaced in the body, not the status.
	if s := err.Error(); strings.Contains(s, "No permission of user=sip") {
		t.Skipf("skip: known account limit: %s", s)
		return true
	}
	return false
}

// --- Patent application + sub-resources --------------------------------------

func TestIntegrationSearchPatents(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.SearchPatents(testCtx(t), "artificial intelligence", 0, 2)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchPatents: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationResolvePatentNumber(t *testing.T) {
	c := newITClient(t, false)
	app, err := c.ResolvePatentNumber(testCtx(t), "US 11,646,472 B2")
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("ResolvePatentNumber: %v", err)
	}
	if app != itApp {
		t.Errorf("resolved app = %q, want %q", app, itApp)
	}
}

func TestIntegrationGetPatent(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetPatent(testCtx(t), itApp)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetPatent: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetPatentMetaData(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetPatentMetaData(testCtx(t), itApp)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetPatentMetaData: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetPatentAdjustment(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetPatentAdjustment(testCtx(t), itApp)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetPatentAdjustment: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetPatentContinuity(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetPatentContinuity(testCtx(t), itApp)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetPatentContinuity: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetPatentDocuments(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetPatentDocuments(testCtx(t), itApp)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetPatentDocuments: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetPatentAssignment(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetPatentAssignment(testCtx(t), itApp)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetPatentAssignment: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetPatentAssociatedDocuments(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetPatentAssociatedDocuments(testCtx(t), itApp)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetPatentAssociatedDocuments: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetPatentAttorney(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetPatentAttorney(testCtx(t), itApp)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetPatentAttorney: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetPatentForeignPriority(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetPatentForeignPriority(testCtx(t), itForeignApp)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetPatentForeignPriority: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetPatentTransactions(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetPatentTransactions(testCtx(t), itApp)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetPatentTransactions: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationSearchPatentsDownload(t *testing.T) {
	c := newITClient(t, false)
	format := generated.PatentDownloadRequestFormat("json")
	req := generated.PatentDownloadRequest{
		Q:          StringPtr("artificial intelligence"),
		Format:     &format,
		Pagination: &generated.Pagination{Offset: Int32Ptr(0), Limit: Int32Ptr(1)},
	}
	data, err := c.SearchPatentsDownload(testCtx(t), req)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchPatentsDownload: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty download bytes")
	}
}

func TestIntegrationGetStatusCodes(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetStatusCodes(testCtx(t))
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetStatusCodes: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

// --- Patent XML --------------------------------------------------------------

func TestIntegrationGetXMLURLForApplication(t *testing.T) {
	c := newITClient(t, false)
	url, _, err := c.GetXMLURLForApplication(testCtx(t), itApp)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetXMLURLForApplication: %v", err)
	}
	if !strings.HasPrefix(url, "https://") {
		t.Errorf("expected https URL, got %q", url)
	}
}

func TestIntegrationGetPatentXML(t *testing.T) {
	c := newITClient(t, false)
	doc, err := c.GetPatentXML(testCtx(t), itApp)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetPatentXML: %v", err)
	}
	if doc == nil || doc.GetTitle() == "" {
		t.Fatal("expected a parsed XML document with a title")
	}
}

func TestIntegrationDownloadXML(t *testing.T) {
	c := newITClient(t, false)
	url, _, err := c.GetXMLURLForApplication(testCtx(t), itApp)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetXMLURLForApplication (chain): %v", err)
	}
	doc, err := c.DownloadXML(testCtx(t), url)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("DownloadXML: %v", err)
	}
	if doc == nil {
		t.Fatal("expected non-nil document")
	}
}

func TestIntegrationDownloadXMLWithType(t *testing.T) {
	c := newITClient(t, false)
	url, dt, err := c.GetXMLURLForApplication(testCtx(t), itApp)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetXMLURLForApplication (chain): %v", err)
	}
	doc, err := c.DownloadXMLWithType(testCtx(t), url, dt)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("DownloadXMLWithType: %v", err)
	}
	if doc == nil {
		t.Fatal("expected non-nil document")
	}
}

func TestIntegrationDownloadPatentDocument(t *testing.T) {
	c := newITClient(t, false)
	// Chain: find a document with a PDF download URL from the documents listing.
	docs, err := c.GetPatentDocuments(testCtx(t), itApp)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetPatentDocuments (chain): %v", err)
	}
	url := firstPatentDocURL(docs)
	if url == "" {
		t.Skip("skip: no PDF download URL available in documents listing")
	}
	var buf bytes.Buffer
	err = c.DownloadPatentDocument(testCtx(t), url, &buf)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("DownloadPatentDocument: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty document bytes")
	}
}

// firstPatentDocURL returns the first PDF downloadUrl in a documents bag, or "".
func firstPatentDocURL(docs *generated.DocumentBag) string {
	if docs == nil || docs.DocumentBag == nil {
		return ""
	}
	for _, d := range *docs.DocumentBag {
		if d.DownloadOptionBag == nil {
			continue
		}
		for _, o := range *d.DownloadOptionBag {
			if o.DownloadUrl != nil && *o.DownloadUrl != "" {
				return *o.DownloadUrl
			}
		}
	}
	return ""
}

// --- Bulk data ---------------------------------------------------------------

func TestIntegrationSearchBulkProducts(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.SearchBulkProducts(testCtx(t), "patent grant", 0, 1)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchBulkProducts: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetBulkProduct(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetBulkProduct(testCtx(t), itBulkProduct)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetBulkProduct: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationDownloadBulkFile(t *testing.T) {
	c := newITClient(t, false)
	// Bulk files are multi-hundred-MB ZIPs; only run the full download when
	// explicitly enabled. The id-resolution call still exercises the chain.
	res, err := c.GetBulkProduct(testCtx(t), itBulkProduct)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetBulkProduct (chain): %v", err)
	}
	uri := firstBulkFileURI(res)
	if uri == "" {
		t.Skip("skip: no FileDownloadURI available")
	}
	if os.Getenv("TEST_BULK_DOWNLOAD") != "true" {
		t.Skip("skip: bulk file download is large; set TEST_BULK_DOWNLOAD=true to run")
	}
	var buf bytes.Buffer
	err = c.DownloadBulkFile(testCtx(t), uri, &buf)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("DownloadBulkFile: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty file bytes")
	}
}

func TestIntegrationDownloadBulkFileWithProgress(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetBulkProduct(testCtx(t), itBulkProduct)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetBulkProduct (chain): %v", err)
	}
	uri := firstBulkFileURI(res)
	if uri == "" {
		t.Skip("skip: no FileDownloadURI available")
	}
	if os.Getenv("TEST_BULK_DOWNLOAD") != "true" {
		t.Skip("skip: bulk file download is large; set TEST_BULK_DOWNLOAD=true to run")
	}
	var buf bytes.Buffer
	err = c.DownloadBulkFileWithProgress(testCtx(t), uri, &buf, func(_, _ int64) {})
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("DownloadBulkFileWithProgress: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty file bytes")
	}
}

// firstBulkFileURI returns the first FileDownloadURI in a product bag, or "".
func firstBulkFileURI(res *generated.BdssResponseProductBag) string {
	if res == nil || res.BulkDataProductBag == nil {
		return ""
	}
	for _, p := range *res.BulkDataProductBag {
		if p.ProductFileBag == nil || p.ProductFileBag.FileDataBag == nil {
			continue
		}
		for _, f := range *p.ProductFileBag.FileDataBag {
			if f.FileDownloadURI != nil && *f.FileDownloadURI != "" {
				return *f.FileDownloadURI
			}
		}
	}
	return ""
}

// --- Petitions ---------------------------------------------------------------

func TestIntegrationSearchPetitions(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.SearchPetitions(testCtx(t), "revival", 0, 2)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchPetitions: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetPetitionDecision(t *testing.T) {
	c := newITClient(t, false)
	// Chain: find a petition record identifier from a search.
	search, err := c.SearchPetitions(testCtx(t), "revival", 0, 1)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchPetitions (chain): %v", err)
	}
	recordID := firstPetitionRecordID(search)
	if recordID == "" {
		t.Skip("skip: no petition record identifier available")
	}
	res, err := c.GetPetitionDecision(testCtx(t), recordID, false)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetPetitionDecision: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func firstPetitionRecordID(res *generated.PetitionDecisionResponseBag) string {
	if res == nil || res.PetitionDecisionDataBag == nil {
		return ""
	}
	for _, p := range *res.PetitionDecisionDataBag {
		if p.PetitionDecisionRecordIdentifier != nil && *p.PetitionDecisionRecordIdentifier != "" {
			return *p.PetitionDecisionRecordIdentifier
		}
	}
	return ""
}

func TestIntegrationSearchPetitionsDownload(t *testing.T) {
	c := newITClient(t, false)
	format := generated.PetitionDecisionDownloadRequestFormat("json")
	req := generated.PetitionDecisionDownloadRequest{
		Q:          StringPtr("revival"),
		Format:     &format,
		Pagination: &generated.Pagination{Offset: Int32Ptr(0), Limit: Int32Ptr(1)},
	}
	data, err := c.SearchPetitionsDownload(testCtx(t), req)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchPetitionsDownload: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty download bytes")
	}
}

// --- PTAB trial proceedings --------------------------------------------------

func TestIntegrationSearchTrialProceedings(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.SearchTrialProceedings(testCtx(t), "*:*", 0, 1)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchTrialProceedings: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetTrialProceeding(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetTrialProceeding(testCtx(t), itTrialNumber)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetTrialProceeding: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationSearchTrialProceedingsDownload(t *testing.T) {
	c := newITClient(t, false)
	data, err := c.SearchTrialProceedingsDownload(testCtx(t), trialDownloadReq())
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchTrialProceedingsDownload: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty download bytes")
	}
}

// trialDownloadReq builds a minimal PTAB download request (json, one row).
func trialDownloadReq() generated.DownloadRequest {
	format := generated.PatentDownloadRequestFormat("json")
	return generated.DownloadRequest{
		Q:          StringPtr("*:*"),
		Format:     &format,
		Pagination: &generated.Pagination{Offset: Int32Ptr(0), Limit: Int32Ptr(1)},
	}
}

// --- PTAB trial decisions ----------------------------------------------------

func TestIntegrationSearchTrialDecisions(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.SearchTrialDecisions(testCtx(t), "*:*", 0, 1)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchTrialDecisions: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetTrialDecision(t *testing.T) {
	c := newITClient(t, false)
	// Chain: pull a decision documentIdentifier from a by-number lookup.
	byNum, err := c.GetTrialDecisionsByTrialNumber(testCtx(t), itTrialNumber)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetTrialDecisionsByTrialNumber (chain): %v", err)
	}
	id := firstTrialDocID(byNum)
	if id == "" {
		t.Skip("skip: no trial decision documentIdentifier available")
	}
	res, err := c.GetTrialDecision(testCtx(t), id)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetTrialDecision: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetTrialDecisionsByTrialNumber(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetTrialDecisionsByTrialNumber(testCtx(t), itTrialNumber)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetTrialDecisionsByTrialNumber: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationSearchTrialDecisionsDownload(t *testing.T) {
	c := newITClient(t, false)
	data, err := c.SearchTrialDecisionsDownload(testCtx(t), trialDownloadReq())
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchTrialDecisionsDownload: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty download bytes")
	}
}

// firstTrialDocID returns the first documentData.documentIdentifier in a trial
// decision/document response, or "".
func firstTrialDocID(res *generated.DecisionDataResponse) string {
	if res == nil || res.PatentTrialDocumentDataBag == nil {
		return ""
	}
	for _, d := range *res.PatentTrialDocumentDataBag {
		if d.DocumentData != nil && d.DocumentData.DocumentIdentifier != nil && *d.DocumentData.DocumentIdentifier != "" {
			return *d.DocumentData.DocumentIdentifier
		}
	}
	return ""
}

// --- PTAB trial documents ----------------------------------------------------

func TestIntegrationSearchTrialDocuments(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.SearchTrialDocuments(testCtx(t), "*:*", 0, 1)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchTrialDocuments: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetTrialDocument(t *testing.T) {
	c := newITClient(t, false)
	byNum, err := c.GetTrialDocumentsByTrialNumber(testCtx(t), itTrialNumber)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetTrialDocumentsByTrialNumber (chain): %v", err)
	}
	id := firstTrialDocumentID(byNum)
	if id == "" {
		t.Skip("skip: no trial document documentIdentifier available")
	}
	res, err := c.GetTrialDocument(testCtx(t), id)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetTrialDocument: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetTrialDocumentsByTrialNumber(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetTrialDocumentsByTrialNumber(testCtx(t), itTrialNumber)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetTrialDocumentsByTrialNumber: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationSearchTrialDocumentsDownload(t *testing.T) {
	c := newITClient(t, false)
	data, err := c.SearchTrialDocumentsDownload(testCtx(t), trialDownloadReq())
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchTrialDocumentsDownload: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty download bytes")
	}
}

func firstTrialDocumentID(res *generated.DocumentDataResponse) string {
	if res == nil || res.PatentTrialDocumentDataBag == nil {
		return ""
	}
	for _, d := range *res.PatentTrialDocumentDataBag {
		if d.DocumentData != nil && d.DocumentData.DocumentIdentifier != nil && *d.DocumentData.DocumentIdentifier != "" {
			return *d.DocumentData.DocumentIdentifier
		}
	}
	return ""
}

// --- PTAB appeal decisions ---------------------------------------------------

func TestIntegrationSearchAppealDecisions(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.SearchAppealDecisions(testCtx(t), "*:*", 0, 1)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchAppealDecisions: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetAppealDecision(t *testing.T) {
	c := newITClient(t, false)
	byNum, err := c.GetAppealDecisionsByAppealNumber(testCtx(t), itAppealNumber)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetAppealDecisionsByAppealNumber (chain): %v", err)
	}
	id := firstAppealDocID(byNum)
	if id == "" {
		t.Skip("skip: no appeal documentIdentifier available")
	}
	res, err := c.GetAppealDecision(testCtx(t), id)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetAppealDecision: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetAppealDecisionsByAppealNumber(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetAppealDecisionsByAppealNumber(testCtx(t), itAppealNumber)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetAppealDecisionsByAppealNumber: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationSearchAppealDecisionsDownload(t *testing.T) {
	c := newITClient(t, false)
	data, err := c.SearchAppealDecisionsDownload(testCtx(t), trialDownloadReq())
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchAppealDecisionsDownload: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty download bytes")
	}
}

func firstAppealDocID(res *generated.AppealDecisionDataResponse) string {
	if res == nil || res.PatentAppealDataBag == nil {
		return ""
	}
	for _, d := range *res.PatentAppealDataBag {
		if d.DocumentData != nil && d.DocumentData.DocumentIdentifier != nil && *d.DocumentData.DocumentIdentifier != "" {
			return *d.DocumentData.DocumentIdentifier
		}
	}
	return ""
}

// --- PTAB interference decisions ---------------------------------------------

func TestIntegrationSearchInterferenceDecisions(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.SearchInterferenceDecisions(testCtx(t), "*:*", 0, 1)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchInterferenceDecisions: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetInterferenceDecision(t *testing.T) {
	c := newITClient(t, false)
	byNum, err := c.GetInterferenceDecisionsByNumber(testCtx(t), itInterferenceNumber)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetInterferenceDecisionsByNumber (chain): %v", err)
	}
	id := firstInterferenceDocID(byNum)
	if id == "" {
		t.Skip("skip: no interference documentIdentifier available")
	}
	res, err := c.GetInterferenceDecision(testCtx(t), id)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetInterferenceDecision: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetInterferenceDecisionsByNumber(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetInterferenceDecisionsByNumber(testCtx(t), itInterferenceNumber)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetInterferenceDecisionsByNumber: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationSearchInterferenceDecisionsDownload(t *testing.T) {
	c := newITClient(t, false)
	data, err := c.SearchInterferenceDecisionsDownload(testCtx(t), trialDownloadReq())
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchInterferenceDecisionsDownload: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty download bytes")
	}
}

func firstInterferenceDocID(res *generated.InterferenceDecisionDataResponse) string {
	if res == nil {
		return ""
	}
	for _, d := range res.PatentInterferenceDataBag {
		if d.DocumentData != nil && d.DocumentData.DocumentIdentifier != nil && *d.DocumentData.DocumentIdentifier != "" {
			return *d.DocumentData.DocumentIdentifier
		}
	}
	return ""
}

// --- Office Action DSAPI -----------------------------------------------------

func TestIntegrationSearchOfficeActions(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.SearchOfficeActions(testCtx(t), "*:*", 0, 1)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchOfficeActions: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetOfficeActionFields(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetOfficeActionFields(testCtx(t))
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetOfficeActionFields: %v", err)
	}
	if res == nil || res.FieldCount == 0 {
		t.Fatal("expected non-nil response with fields")
	}
}

func TestIntegrationSearchOfficeActionCitations(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.SearchOfficeActionCitations(testCtx(t), "*:*", 0, 1)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchOfficeActionCitations: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetOfficeActionCitationFields(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetOfficeActionCitationFields(testCtx(t))
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetOfficeActionCitationFields: %v", err)
	}
	if res == nil || res.FieldCount == 0 {
		t.Fatal("expected non-nil response with fields")
	}
}

func TestIntegrationSearchOfficeActionRejections(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.SearchOfficeActionRejections(testCtx(t), "patentApplicationNumber:12190351", 0, 3)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchOfficeActionRejections: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetOfficeActionRejectionFields(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetOfficeActionRejectionFields(testCtx(t))
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetOfficeActionRejectionFields: %v", err)
	}
	if res == nil || res.FieldCount == 0 {
		t.Fatal("expected non-nil response with fields")
	}
}

func TestIntegrationSearchEnrichedCitations(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.SearchEnrichedCitations(testCtx(t), "*:*", 0, 1)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("SearchEnrichedCitations: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetEnrichedCitationFields(t *testing.T) {
	c := newITClient(t, false)
	res, err := c.GetEnrichedCitationFields(testCtx(t))
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetEnrichedCitationFields: %v", err)
	}
	if res == nil || res.FieldCount == 0 {
		t.Fatal("expected non-nil response with fields")
	}
}

// --- TSDR (trademark) --------------------------------------------------------

func TestIntegrationGetTrademarkStatus(t *testing.T) {
	c := newITClient(t, true)
	res, err := c.GetTrademarkStatus(testCtx(t), itSerial)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetTrademarkStatus: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetTrademarkStatusJSON(t *testing.T) {
	c := newITClient(t, true)
	res, err := c.GetTrademarkStatusJSON(testCtx(t), itSerial)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetTrademarkStatusJSON: %v", err)
	}
	if res == nil || len(res.Trademarks) == 0 {
		t.Fatal("expected non-empty trademarks")
	}
}

func TestIntegrationGetTrademarkDocumentsXML(t *testing.T) {
	c := newITClient(t, true)
	data, err := c.GetTrademarkDocumentsXML(testCtx(t), itSerial)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetTrademarkDocumentsXML: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty XML")
	}
}

// firstTSDRDocID derives a {DocumentTypeCode}{YYYYMMDD} docID from the documents
// XML listing (e.g. NOA20230322), used to chain document-level TSDR tests.
func firstTSDRDocID(t *testing.T, c *Client) string {
	t.Helper()
	data, err := c.GetTrademarkDocumentsXML(testCtx(t), itSerial)
	if skipExpected(t, err) {
		return ""
	}
	if err != nil {
		t.Fatalf("GetTrademarkDocumentsXML (chain): %v", err)
	}
	// The mailRoomDate sits next to the type code; build the docID as code+date.
	// Elements are e.g. <DocumentTypeCode>NOA</DocumentTypeCode> and
	// <MailRoomDate>2023-03-22-04:00</MailRoomDate> (a timezone suffix follows the
	// day, so the date pattern stops at the day).
	codeRe := regexp.MustCompile(`<(?:[A-Za-z0-9]+:)?DocumentTypeCode>([A-Za-z0-9]+)</`)
	dateRe := regexp.MustCompile(`<(?:[A-Za-z0-9]+:)?MailRoomDate>(\d{4})-(\d{2})-(\d{2})`)
	code := codeRe.FindSubmatch(data)
	date := dateRe.FindSubmatch(data)
	if code == nil || date == nil {
		return ""
	}
	return string(code[1]) + string(date[1]) + string(date[2]) + string(date[3])
}

func TestIntegrationGetTrademarkDocumentInfo(t *testing.T) {
	c := newITClient(t, true)
	docID := firstTSDRDocID(t, c)
	if docID == "" {
		t.Skip("skip: no TSDR docID derivable from documents listing")
	}
	data, err := c.GetTrademarkDocumentInfo(testCtx(t), itSerial, docID)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetTrademarkDocumentInfo: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty XML")
	}
}

func TestIntegrationDownloadTrademarkDocument(t *testing.T) {
	c := newITClient(t, true)
	docID := firstTSDRDocID(t, c)
	if docID == "" {
		t.Skip("skip: no TSDR docID derivable from documents listing")
	}
	var buf bytes.Buffer
	err := c.DownloadTrademarkDocument(testCtx(t), itSerial, docID, &buf)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("DownloadTrademarkDocument: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
}

func TestIntegrationGetTrademarkLastUpdate(t *testing.T) {
	c := newITClient(t, true)
	res, err := c.GetTrademarkLastUpdate(testCtx(t), itSerial)
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetTrademarkLastUpdate: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIntegrationGetTrademarkMultiStatus(t *testing.T) {
	c := newITClient(t, true)
	res, err := c.GetTrademarkMultiStatus(testCtx(t), "sn", []string{itSerial})
	if skipExpected(t, err) {
		return
	}
	if err != nil {
		t.Fatalf("GetTrademarkMultiStatus: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}
