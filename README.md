# USPTO Open Data Portal (ODP) Go Client

[![CI](https://github.com/patent-dev/uspto-odp/actions/workflows/ci.yml/badge.svg)](https://github.com/patent-dev/uspto-odp/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/patent-dev/uspto-odp.svg)](https://pkg.go.dev/github.com/patent-dev/uspto-odp)
[![Go Report Card](https://goreportcard.com/badge/github.com/patent-dev/uspto-odp)](https://goreportcard.com/report/github.com/patent-dev/uspto-odp)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A Go client for the USPTO Open Data Portal (ODP) REST API, the Office Action
APIs (DSAPI), and TSDR (Trademark Status & Document Retrieval).

## Overview

- **Patent Application API** (13 endpoints) - search, metadata, continuity,
  assignment, adjustment, attorney, foreign priority, transactions, documents.
- **Bulk Data API** (3 endpoints) - product search and resumable file downloads
  with progress callbacks.
- **Petition API** (3 endpoints) - petition decision search, retrieval, download.
- **Trademark Decisions and Proceedings API** (4 endpoints) - expungement and
  reexamination proceedings, petitions to the Director, and administrative
  decisions (sanctions, show-cause orders), plus the decision PDFs.
- **PTAB API** (19 endpoints) - trial proceedings, decisions, documents, appeal
  decisions, interference decisions.
- **Office Action APIs** (8 endpoints) - text retrieval, citations, rejections,
  enriched citations, via the DSAPI Lucene/Solr pattern.
- **Office Action Search API** (5 endpoints) - ODP office action search,
  retrieval, download and full raw text (applications filed 2001 or later).
- **TSDR** (7 wrapper methods) - trademark status and document retrieval on a
  separate server with its own key.
- **Patent number normalization** - accepts application, grant, publication, and
  PCT formats; resolves grant/publication numbers to application numbers.
- **XML full text** - parses ICE DTD 4.6/4.7 grant and application documents.
- Retry logic with exponential backoff, `Retry-After` honoring, and configurable
  timeouts.

### ODP releases & spec versions

USPTO bumps the **portal version** (currently 5.1) for any change to the
ODP -- UI updates, new bulk datasets, ecosystem additions -- even when the
underlying **API specs** don't change. Each swagger file carries its own
`info.version`, and that's the contract this client binds to.

| Spec file | Version | Surface |
|---|---|---|
| `swagger.yaml` | 1.0.0 | Patent Application, Bulk Data, PTAB |
| `odp-common-base.yaml` | 1.0.0 | Shared schemas |
| `petition-decisions.yaml` | 1.0.0 | Petition decisions |
| `trademark-decisions.yaml` | 1.0.0 | Trademark decisions and proceedings |
| `office-actions.yaml` | 1.0.0 | ODP Office Action Search |
| `trial-proceedings.yaml` | 1.0.0 | PTAB trial proceedings |
| `trial-decisions.yaml` | 1.0.0 | PTAB trial decisions |
| `trial-documents.yaml` | 1.0.0 | PTAB trial documents |
| `trial-appeal-decisions.yaml` | 1.0.0 | PTAB appeal decisions |
| `trial-interferences.yaml` | 1.0.0 | PTAB interference decisions |
| `trial-common.yaml` | 1.1.0 | Shared PTAB types |
| `oa-text-retrieval.yaml` | 1.0.0 | Office Action text |
| `oa-citations.yaml` | 1.0.0 | Office Action citations |
| `oa-rejections.yaml` | 1.0.0 | Office Action rejections |
| `oa-enriched-citations.yaml` | 1.0.0 | Office Action enriched citations |
| `tsdr-swagger.json` | version 1 | TSDR (separate server + key) |

What this client implements from each ODP portal release:

| Portal release | Date | Changes affecting this client |
|---|---|---|
| 3.0 | 2025-11-21 | Initial PTAB datasets via ODP -- wrapped (19 endpoints). |
| 3.2 | 2026-01-16 | Office Action Weekly Archives surfaced as bulk product `OACT` -- accessible via the existing `SearchBulkProducts` / `GetBulkProduct` wrappers. PTAB RSS feed not wrapped (separate host, no API key). |
| 3.3 | 2026-03-04 | Trademark (TM) Decisions & Proceedings reachable only through the portal SPA at `data.uspto.gov/ui/trademark/tm-decisions-api/`, using cookie + WAF-token auth. No public API spec at the time; wrapped since 5.1. |
| 3.4 | 2026-03-13 | PatentsView bulk products surfaced via the existing bulk-data wrappers. |
| 3.5 | 2026-03-24 | Office Action APIs migrated to ODP -- wrapped (8 endpoints). The OA wrappers use the ODP host (`Config.OABaseURL` default `https://api.uspto.gov`). |
| 3.6 | 2026-04-10 | US PCT numbers standardized to 15-char (`PCTUS...`); display form is 17-char with slashes -- fully supported. Patent Assignments API: `countryOrStateCode` deprecated, location consolidated under `geographicRegionCode` (the field left the spec in 5.0 and the fallback was dropped). Daily/yearly bulk file resumption is dataset coverage, no client change. |
| 4.0 | | Portal-level changes; no API spec change. |
| 5.0 | 2026-10-02 | Petition decisions moved to their own `petition-decisions.yaml` with a reworked record: `decisionTypeCode`/`decisionTypeCodeDescriptionText` replaced by `decidedPetitionTypeCode`/`decidedPetitionTypeCodeDescriptionText`, new `applicationStatusDescriptionText`, `applicationStatusDate`, `decisionMailDate`, `finalPetitionDecisionIndicator`, `ocrSearchText`; `decisionPetitionTypeCode` is now a string; `includeDocuments` is gone (documents always come back). Breaking for `GetPetitionDecision` (see Petition API). The new ODP Office Action Search (`/api/v1/patent/office-action/...`) is wrapped (5 endpoints) in `generated/oasearch`; the DSAPI Office Action wrappers keep working. |
| 5.1 | 2026-11-13 | Trademark Decisions and Proceedings API published (`trademark-decisions.yaml`) -- wrapped (4 endpoints) in `generated/tmd`. |

When a release changes a path or schema, drop the new swagger files into
`swagger/` and run `go run ./cmd/gen`; the coverage test catches anything that
goes silently unwrapped.

## Installation

```bash
go get github.com/patent-dev/uspto-odp
```

## Getting access

The ODP API requires a free MyUSPTO (USPTO.gov) account and an ODP API key
issued from data.uspto.gov. Key issuance involves identity verification: your
USPTO.gov account must be linked to a validated ID.me identity. The same key
covers the Patent, PTAB, Petition, Bulk Data, and Office Action APIs and is sent
in the `X-API-Key` header. TSDR uses a separate key (see [TSDR](#tsdr-trademark-status--document-retrieval-api)).

1. Sign in at [MyODP](https://data.uspto.gov/myodp), or create a USPTO.gov
   account from the sign-in page if you don't have one.

2. Verify your identity with ID.me and link it to your USPTO.gov account
   (required before a key can be issued).

3. Request an API key from the [Getting Started](https://data.uspto.gov/apis/getting-started)
   page and copy it.

4. Export it for the client and demo:
   ```bash
   export USPTO_API_KEY=...
   # optional, for TSDR (https://account.uspto.gov/profile/api-manager):
   export USPTO_TSDR_API_KEY=...
   ```

> Unused ODP keys are deleted after 90 days of inactivity.


## Quick start

```go
config := odp.DefaultConfig()
config.APIKey = os.Getenv("USPTO_API_KEY")
client, err := odp.NewClient(config)
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()
results, err := client.SearchPatents(ctx, "artificial intelligence", 0, 10)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Found %d patents\n", *results.Count)
```

## Usage

All ODP API endpoints are wrapped, plus the Office Action and TSDR APIs
(53 wrapper methods total).

### Patent Application API (13 endpoints)

```go
// Core Patent Data
SearchPatents(ctx, query string, offset, limit int) (*PatentDataResponse, error)
GetPatent(ctx, patentNumber string) (*PatentDataResponse, error)  // Accepts any patent number format
GetPatentMetaData(ctx, applicationNumber string) (*MetaDataResponse, error)

// Patent Details
GetPatentAdjustment(ctx, applicationNumber string) (*AdjustmentResponse, error)
GetPatentContinuity(ctx, applicationNumber string) (*ContinuityResponse, error)
GetPatentDocuments(ctx, applicationNumber string) (*DocumentBag, error)
GetPatentAssignment(ctx, applicationNumber string) (*AssignmentResponse, error)
GetPatentAssociatedDocuments(ctx, applicationNumber string) (*AssociatedDocumentsResponse, error)
GetPatentAttorney(ctx, applicationNumber string) (*RecordAttorney, error)
GetPatentForeignPriority(ctx, applicationNumber string) (*ForeignPriorityResponse, error)
GetPatentTransactions(ctx, applicationNumber string) (*TransactionsResponse, error)

// Downloads & Utilities
SearchPatentsDownload(ctx, req PatentDownloadRequest) ([]byte, error)
GetStatusCodes(ctx) (*StatusCodeSearchResponse, error)
```

### Bulk Data API (3 endpoints)

```go
SearchBulkProducts(ctx, query string, offset, limit int) (*BdssResponseBag, error)
GetBulkProduct(ctx, productID string) (*BdssResponseProductBag, error)

// File download methods (use FileDownloadURI directly):
DownloadBulkFile(ctx, fileDownloadURI string, w io.Writer) error
DownloadBulkFileWithProgress(ctx, fileDownloadURI string, w io.Writer,
    progress func(bytesComplete, bytesTotal int64)) error
```

```go
product, err := client.GetBulkProduct(ctx, "PTGRXML")
files := *product.BulkDataProductBag[0].ProductFileBag.FileDataBag

for _, file := range files {
    if file.FileName != nil && strings.Contains(*file.FileName, "ipg250923.zip") {
        if file.FileDownloadURI != nil {
            err := client.DownloadBulkFileWithProgress(ctx, *file.FileDownloadURI, outputFile,
                func(bytesComplete, bytesTotal int64) {
                    percent := float64(bytesComplete) * 100 / float64(bytesTotal)
                    fmt.Printf("\rProgress: %.1f%%", percent)
                })
        }
        break
    }
}
```

### Petition API (3 endpoints)

```go
SearchPetitions(ctx, query string, offset, limit int) (*PetitionDecisionResponseBag, error)
GetPetitionDecision(ctx, recordID string) (*PetitionDecisionResponseBag, error)
SearchPetitionsDownload(ctx, req PetitionDecisionDownloadRequest) ([]byte, error)
```

`GetPetitionDecision` returns the record as the single entry of
`PetitionDecisionDataBag`, with the decision documents in its `DocumentBag`.

### Trademark Decisions and Proceedings API (4 endpoints)

One dataset with three `decisionProceedingCategory` values: `ExpungReexam`,
`PetitionsToDirector` and `Administrative`. Types live in `generated/tmd`.

```go
SearchTrademarkDecisions(ctx, query string, offset, limit int) (*tmd.SearchResponse, error)
SearchTrademarkDecisionsWithRequest(ctx, req tmd.SearchRequest) (*tmd.SearchResponse, error)
GetTrademarkDecision(ctx, documentIdentifier string) (*tmd.TrademarkDecisionResponse, error)
SearchTrademarkDecisionsDownload(ctx, req tmd.DownloadRequest) ([]byte, error)
DownloadTrademarkDecisionFile(ctx, fileDownloadURI string, w io.Writer) error
```

```go
limit := 10
res, err := client.SearchTrademarkDecisionsWithRequest(ctx, tmd.SearchRequest{
    Q:          odp.StringPtr("expungement"),
    Filters:    &[]tmd.FilterClause{{Name: "decisionProceedingCategory", Value: []string{odp.TrademarkDecisionCategoryExpungReexam}}},
    Pagination: &tmd.Pagination{Limit: &limit},
})
```

The API accepts `limit` 1-100. Records link to the TSDR case through
`serialNumber` / `registrationNumber`; Administrative records carry a
`fileDownloadURI` for the decision PDF.

### PTAB (Patent Trial and Appeal Board) API (19 endpoints)

```go
// Trial Proceedings (IPR, PGR, CBM)
SearchTrialProceedings(ctx, query string, offset, limit int) (*ProceedingDataResponse, error)
GetTrialProceeding(ctx, trialNumber string) (*ProceedingDataResponse, error)
SearchTrialProceedingsDownload(ctx, req DownloadRequest) ([]byte, error)

// Trial Decisions
SearchTrialDecisions(ctx, query string, offset, limit int) (*DecisionDataResponse, error)
GetTrialDecision(ctx, documentIdentifier string) (*DecisionDataResponse, error)
GetTrialDecisionsByTrialNumber(ctx, trialNumber string) (*DecisionDataResponse, error)
SearchTrialDecisionsDownload(ctx, req DownloadRequest) ([]byte, error)

// Trial Documents
SearchTrialDocuments(ctx, query string, offset, limit int) (*DocumentDataResponse, error)
GetTrialDocument(ctx, documentIdentifier string) (*DocumentDataResponse, error)
GetTrialDocumentsByTrialNumber(ctx, trialNumber string) (*DocumentDataResponse, error)
SearchTrialDocumentsDownload(ctx, req DownloadRequest) ([]byte, error)

// Appeal Decisions
SearchAppealDecisions(ctx, query string, offset, limit int) (*AppealDecisionDataResponse, error)
GetAppealDecision(ctx, documentIdentifier string) (*AppealDecisionDataResponse, error)
GetAppealDecisionsByAppealNumber(ctx, appealNumber string) (*AppealDecisionDataResponse, error)
SearchAppealDecisionsDownload(ctx, req DownloadRequest) ([]byte, error)

// Interference Decisions
SearchInterferenceDecisions(ctx, query string, offset, limit int) (*InterferenceDecisionDataResponse, error)
GetInterferenceDecision(ctx, documentIdentifier string) (*InterferenceDecisionDataResponse, error)
GetInterferenceDecisionsByNumber(ctx, interferenceNumber string) (*InterferenceDecisionDataResponse, error)
SearchInterferenceDecisionsDownload(ctx, req PatentDownloadRequest) ([]byte, error)
```

### Office Action APIs (8 endpoints)

These use the DSAPI pattern (form-encoded POST with Lucene/Solr queries). Same
API key as the main ODP API.

```go
// Office Action Text Retrieval
SearchOfficeActions(ctx, criteria string, start, rows int) (*DSAPIResponse, error)
GetOfficeActionFields(ctx) (*DSAPIFieldsResponse, error)

// Office Action Citations (Forms PTO-892 & PTO-1449)
SearchOfficeActionCitations(ctx, criteria string, start, rows int) (*DSAPIResponse, error)
GetOfficeActionCitationFields(ctx) (*DSAPIFieldsResponse, error)

// Office Action Rejections (101, 102, 103, 112, DP)
SearchOfficeActionRejections(ctx, criteria string, start, rows int) (*DSAPIResponse, error)
GetOfficeActionRejectionFields(ctx) (*DSAPIFieldsResponse, error)

// Enriched Citations (AI/ML extracted from office actions)
SearchEnrichedCitations(ctx, criteria string, start, rows int) (*DSAPIResponse, error)
GetEnrichedCitationFields(ctx) (*DSAPIFieldsResponse, error)
```

Office Action APIs use Lucene query syntax:

```go
// Search by patent application number
result, err := client.SearchOfficeActionRejections(ctx, "patentApplicationNumber:12190351", 0, 10)

// Search with boolean operators
result, err := client.SearchOfficeActions(ctx, "hasRej103:1 AND groupArtUnitNumber:1713", 0, 10)

// Get available fields
fields, err := client.GetOfficeActionRejectionFields(ctx)
fmt.Printf("Searchable fields: %v\n", fields.Fields)
```

### Office Action Search API (5 endpoints)

The ODP Office Action Search (release 5.0): office actions of applications filed
in 2001 or later, searchable by metadata and full text. Separate from the DSAPI
wrappers above. Types live in `generated/oasearch`.

```go
SearchOfficeActionRecords(ctx, query string, offset, limit int) (*oasearch.OfficeActionSearchResponse, error)
SearchOfficeActionRecordsWithRequest(ctx, req oasearch.OfficeActionSearchRequest) (*oasearch.OfficeActionSearchResponse, error)
GetOfficeActionRecord(ctx, recordID string) (*oasearch.OfficeActionResponse, error)
SearchOfficeActionRecordsDownload(ctx, req oasearch.OfficeActionDownloadRequest) ([]byte, error)
GetOfficeActionText(ctx, recordID string) ([]byte, error)
```

Search and by-id records carry an `officeActionText` excerpt; `GetOfficeActionText`
returns the full raw text.

### TSDR (Trademark Status & Document Retrieval) API

Separate server (`tsdrapi.uspto.gov`) and API key. Requires `TSDRAPIKey` in
config, sent in the `USPTO-API-KEY` header.

```go
config := odp.DefaultConfig()
config.APIKey = "your-odp-key"
config.TSDRAPIKey = "your-tsdr-key"  // from https://account.uspto.gov/profile/api-manager
client, err := odp.NewClient(config)

// Get trademark case status (raw XML response)
xmlResp, err := client.GetTrademarkStatus(ctx, "97123456")

// Get trademark case status (JSON via content negotiation)
status, err := client.GetTrademarkStatusJSON(ctx, "97123456")
fmt.Printf("Trademarks: %d\n", len(status.Trademarks))

// Get document list (XML)
docs, err := client.GetTrademarkDocumentsXML(ctx, "97123456")

// Get info about a specific document (XML) - docID format: {TypeCode}{YYYYMMDD}
info, err := client.GetTrademarkDocumentInfo(ctx, "97123456", "NOA20230322")

// Download a document as PDF
var buf bytes.Buffer
client.DownloadTrademarkDocument(ctx, "97123456", "NOA20230322", &buf)

// Get last update time
update, err := client.GetTrademarkLastUpdate(ctx, "97123456")

// Multi-status lookup
result, err := client.GetTrademarkMultiStatus(ctx, "sn", []string{"97123456", "97654321"})
```

### Patent number normalization

The library handles various patent number formats and automatically resolves
them to application numbers:

```go
// GetPatent accepts any patent number format
doc, err := client.GetPatent(ctx, "US 11,646,472 B2")  // Grant number
doc, err := client.GetPatent(ctx, "17/248,024")        // Application number
doc, err := client.GetPatent(ctx, "US20250087686A1")   // Publication number

// For other methods, resolve to application number first
appNumber, err := client.ResolvePatentNumber(ctx, "US 11,646,472 B2")
// appNumber = "17248024" (the actual application number)

// Low-level normalization (formatting only, doesn't resolve)
pn, err := odp.NormalizePatentNumber("US 11,646,472 B2")
fmt.Println(pn.Type)                  // PatentNumberTypeGrant
fmt.Println(pn.Normalized)            // "11646472" (normalized, not application number!)
fmt.Println(pn.FormatAsGrant())       // "11,646,472"
```

**Note:** Grant and publication numbers are **not** the same as application
numbers. The library uses the search API to resolve grant/publication numbers to
their corresponding application numbers.

Supported formats:
- Applications: `17248024`, `17/248,024`, `US 17/248,024`
- Grants: `11646472`, `11,646,472`, `US 11,646,472 B2`
- Publications: `20250087686`, `US20250087686A1` (kind code preserved when supplied: `A2`, `A9`, ...)
- PCT: `PCTUS2025058371` (15-char API form), `PCT/US2025/058371` (17-char display), `PCTUS0719317` (12-char legacy). Use `pn.FormatAsPCT()` for the display form.

**Note:** 8-digit numbers (like `11646472`) are ambiguous - they could be either
grant or application numbers. Use formatting (commas, kind codes) to
disambiguate.

### XML full text retrieval

Parse full patent text (ICE DTD 4.6/4.7):

```go
doc, err := client.GetPatentXML(ctx, "US 11,646,472 B2")

title := doc.GetTitle()
abstract := doc.GetAbstract().ExtractAbstractText()
claims := doc.GetClaims().ExtractAllClaimsTextFormatted()
description := doc.GetDescription().ExtractDescriptionText()
```

Advanced usage:

```go
// Get XML URL and type
xmlURL, docType, err := client.GetXMLURLForApplication(ctx, "17248024")

// Download with type hint
doc, err := client.DownloadXMLWithType(ctx, xmlURL, docType)

// Parse raw XML
data := []byte(/* XML content */)
doc, err = odp.ParseGrantXML(data)  // or ParseApplicationXML
```

### Configuration

```go
config := &odp.Config{
    BaseURL:    "https://api.uspto.gov", // Default
    APIKey:     "your-api-key",
    UserAgent:  "YourApp/1.0",
    MaxRetries: 3,                       // Retry failed requests
    RetryDelay: 1 * time.Second,         // Base backoff between retries
    Timeout:    30 * time.Second,        // Request timeout
    MaxRetryAfter: 60 * time.Second,     // Longest Retry-After the client will honor

    // Office Action DSAPI host (defaults to the ODP host)
    OABaseURL:  "https://api.uspto.gov", // Default (Office Action endpoints on the ODP host)

    // TSDR (optional - separate server and API key)
    TSDRAPIKey: "your-tsdr-key",         // From https://account.uspto.gov/profile/api-manager
    TSDRBaseURL: "https://tsdrapi.uspto.gov", // Default
}

client, err := odp.NewClient(config)
```

`DefaultConfig()` fills in sane defaults; override only the fields you need.

## Error handling

Non-2xx responses surface as `*APIError`, carrying the status code, a message,
and a truncated response body for debugging. Use `errors.As` to inspect them:

```go
results, err := client.SearchPatents(ctx, query, 0, 10)
if err != nil {
    var apiErr *odp.APIError
    if errors.As(err, &apiErr) {
        fmt.Println("status:", apiErr.StatusCode)
        fmt.Println("detail:", apiErr.Detail()) // message + server body
        if apiErr.IsRetryable() {               // 429 or 5xx
            // back off and retry, or let the client's own retry handle it
        }
    }
    return err
}
```

On HTTP 429 the client honors the `Retry-After` header (capped by
`Config.MaxRetryAfter`, default 60s) instead of its exponential backoff. When
the server asks for a longer wait than the cap, the request fails with a
non-retryable `*APIError` so the caller decides rather than the client blocking.

## Testing

```bash
make test    # unit tests with the race detector
make lint    # gofmt + go vet
```

Integration tests require `USPTO_API_KEY` (and optionally `USPTO_TSDR_API_KEY`):

```bash
export USPTO_API_KEY=your-odp-key
export USPTO_TSDR_API_KEY=your-tsdr-key  # from https://account.uspto.gov/profile/api-manager

make test-integration
```

## Swagger processing

### Source files

The USPTO ODP API specification is distributed as multiple YAML files with `$ref`
references between them. The originals are downloaded from
[USPTO ODP Swagger](https://data.uspto.gov/swagger/index.html#/) and stored in `swagger/`:

```
swagger/
  swagger.yaml                  # Main ODP API spec (Patent, Bulk, PTAB)
  odp-common-base.yaml          # Shared type definitions
  petition-decisions.yaml       # Petition decisions (bundled into generated/)
  trademark-decisions.yaml      # Trademark decisions and proceedings
  office-actions.yaml           # ODP Office Action Search
  trial-proceedings.yaml        # PTAB trial proceedings
  trial-decisions.yaml          # PTAB trial decisions
  trial-documents.yaml          # PTAB trial documents
  trial-appeal-decisions.yaml   # PTAB appeal decisions
  trial-interferences.yaml      # PTAB interference decisions
  trial-common.yaml             # Shared PTAB types
  oa-text-retrieval.yaml        # Office Action Text Retrieval
  oa-citations.yaml             # Office Action Citations
  oa-rejections.yaml            # Office Action Rejections
  oa-enriched-citations.yaml    # Enriched Citations
  tsdr-swagger.json             # TSDR (Trademark) API
```

Do not edit files in `swagger/` - these are the original USPTO specifications.

### Code generation

The `cmd/gen` tool (pure Go, no external dependencies) bundles and fixes the specs:

```bash
go run ./cmd/gen
```

1. Bundles ODP + PTAB YAML, resolving `$ref` -> `swagger_fixed.yaml` -> `generated/`
2. Bundles the 4 Office Action specs -> `swagger_oa_fixed.yaml` -> `generated/oa/`
3. Copies the TSDR spec -> `swagger_tsdr_fixed.json` -> `generated/tsdr/`
4. Fixes the trademark decisions spec -> `swagger_tmd_fixed.yaml` -> `generated/tmd/`
5. Fixes the Office Action Search spec -> `swagger_oas_fixed.yaml` -> `generated/oasearch/`
6. Applies fixes for mismatches between the specs and actual API responses

### Fixes applied

The USPTO swagger has several mismatches with actual API responses:

**Type corrections:**
- `frameNumber`, `reelNumber`: string -> integer (API returns numbers)
- `documentNumber`: string -> integer (PTAB API returns numbers)
- Error response `code`: integer -> string (API returns `"404"`, not `404`)

**Structure fixes:**
- `correspondenceAddress`: array -> object (Assignment API returns an object)
- `DecisionData.statuteAndRuleBag`, `issueTypeBag`: string -> array
- `GetPatentAssignment.assignmentBag`: single object -> array

**Field-name fixes:**
- `InterferenceDecisionRecord.decisionDocumentData` -> `documentData`
- `DecisionDataResponse.patentTrialDecisionDataBag` -> `patentTrialDocumentDataBag`
  (the trial-decisions search returns the array under `patentTrialDocumentDataBag`;
  without the fix the slice silently unmarshals to nil while `count` parses fine)

**Format fixes:**
- Removed `format: date-time` from fields returning non-RFC3339 values
  (e.g. `lastModifiedDateTime` returns `"2025-11-26T23:58:00"` without a timezone)
- Removed `format: date` from fields that return datetimes or non-ISO dates
  (e.g. `fileReleaseDate` returns `"2025-09-23 00:57:53"`)

**Petition decision fixes:**
- `petition-decisions.yaml` is no longer `$ref`'d from `swagger.yaml`; its paths
  are merged into the main bundle, and the stale petition `$ref`s left in
  `swagger.yaml` are dropped
- `PetitionDecisionSearchResponse` renamed `PetitionDecisionResponseBag`
- `GET /api/v1/petition/decisions/{id}` returns the search envelope, not a bare record
- `decisionPetitionTypeCode`: integer -> string (API returns `"505"`)
- `petitionDecisionRecordIdentifier`: dropped `format: uuid` (kept a string)
- Added `allSearchableFields` and the download options' `downloadUrl` /
  `pageTotalQuantity` (returned but not in the spec)

**Trademark decisions fixes:**
- `serialNumber`, `registrationNumber`: string -> integer (API returns numbers)
- Dropped the incomplete enums on `documentTypeDescriptionText` and
  `proceedingTypeCategory` (the API returns e.g. `REEXAMINATION`)
- Added `fileDownloadURI`, `textractStatus` (returned but not in the spec)
- The by-identifier response is named `TrademarkDecisionResponse` and gains `count`
- `POST .../search/download` takes `DownloadRequest` (with `format`), not `SearchRequest`

**Office Action Search fixes:**
- Error responses: `code` integer -> string, plus `message` / `detailedMessage`
  (the API returns `{"code":"404","message":...}`, also for a search without matches)
- `lastModifiedDateTime`: dropped `format: date` (the API returns a datetime)
- Added `inventionTitle`, `publicStatusIndicator` (returned but not in the spec)
- The by-identifier response is named `OfficeActionResponse`

**Endpoint fixes:**
- Removed `/api/v1/patent/applications/text-to-search` (defined but has no operations)

**Office Action (DSAPI) fixes:**
- Removed phantom path parameters (`dataset`, `version` declared `in: path` but the
  paths are static)
- Made `operationId` values unique across the four bundled specs (all four used
  identical IDs)
- The Office Action APIs are served from the ODP host (`https://api.uspto.gov`,
  the `route Office Action APIs to ODP` migration) under `/api/v1/patent/oa/...`;
  override the host with `Config.OABaseURL` if needed.

**TSDR fixes:**
- Fixed the protocol-relative server URL (`//tsdrapi.uspto.gov/` -> `https://tsdrapi.uspto.gov`)
- `GetDocumentInfoXml` / `GetCaseDocsInfoXml` return 406 with `Accept: application/json`
  (content negotiation is unsupported despite XML/JSON paths in the spec)
- Removed `format: date-time` / `format: date` from 64 fields - the API returns
  inconsistent date formats that break `time.Time` parsing in generated code

**Upstream bugs worked around:**
- `trial-appeal-decisions.yaml`: `appelantData` -> `appellantData`, `realPartyName` ->
  `realPartyInInterestName`, `techCenterNumber` -> `technologyCenterNumber`,
  `requestorData` -> `thirdPartyRequesterData`, `documentTypeCategory` ->
  `documentTypeDescriptionText`, `downloadURI` -> `fileDownloadURI`
- `trial-common.yaml`: `downloadURI` -> `fileDownloadURI`, `statuteAndRuleBag` /
  `issueTypeBag` string -> array, `documentNumber` string -> integer

## Related projects

Part of the [patent.dev](https://patent.dev) open-source patent data ecosystem:

- [epo-ops](https://github.com/patent-dev/epo-ops) - EPO Open Patent Services client (bibliographic, full text, families, legal status, images)
- [epo-bdds](https://github.com/patent-dev/epo-bdds) - EPO Bulk Data Distribution Service client (DOCDB, INPADOC, EP full text)
- [dpma-connect-plus](https://github.com/patent-dev/dpma-connect-plus) - DPMA Connect Plus client (patents, designs, trademarks)

The [bulk-file-loader](https://github.com/patent-dev/bulk-file-loader) uses these libraries for automated patent data downloads.

## License

MIT - Funktionslust GmbH / patent.dev.
