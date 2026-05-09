# USPTO Open Data Portal (ODP) Go Client

[![Go Reference](https://pkg.go.dev/badge/github.com/patent-dev/uspto-odp.svg)](https://pkg.go.dev/github.com/patent-dev/uspto-odp)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A complete Go client library for the USPTO Open Data Portal API, Office Action APIs, and TSDR (Trademark Status & Document Retrieval).

## ODP releases & spec versions

USPTO bumps the **portal version** (currently 4.0) for any change to the
ODP -- UI updates, new bulk datasets, ecosystem additions -- even when the
underlying **API specs** don't change. Each swagger file carries its own
`info.version`, and that's the contract this client binds to.

| Spec file | Version | Surface |
|---|---|---|
| `swagger.yaml` | 1.0.0 | Patent Application, Bulk Data, Petition, PTAB |
| `odp-common-base.yaml` | 1.0.0 | Shared schemas |
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
| 3.3 | 2026-03-04 | Trademark (TM) Decisions & Proceedings reachable only through the portal SPA at `data.uspto.gov/ui/trademark/tm-decisions-api/`, using cookie + WAF-token auth. No public API spec. Not wrapped; will revisit if USPTO publishes one. |
| 3.4 | 2026-03-13 | PatentsView bulk products surfaced via the existing bulk-data wrappers. |
| 3.5 | 2026-03-24 | Office Action APIs migrated to ODP -- wrapped (8 endpoints). The migration is incomplete on USPTO's side: the OA routes don't yet exist at `api.uspto.gov`; the legacy Developer Hub at `developer.uspto.gov/ds-api/...` still serves them with no API key. The client points the OA wrappers at the legacy host via `Config.OABaseURL` (default `https://developer.uspto.gov`); override when USPTO completes the migration. |
| 3.6 | 2026-04-10 | US PCT numbers standardized to 15-char (`PCTUS...`); display form is 17-char with slashes -- fully supported. Patent Assignments API: `countryOrStateCode` deprecated, location consolidated under `geographicRegionCode` -- handled with fallback. Daily/yearly bulk file resumption is dataset coverage, no client change. |
| 4.0 | (latest) | Portal-level changes; the API specs in this repo are still 1.0.x. If 4.0 ever introduces a path or schema change, drop the new swagger files into `swagger/` and run `go run ./cmd/gen`; the coverage test catches anything that goes silently unwrapped. |

## Getting Started

### API Keys

**ODP API Key** (Patent, PTAB, Petition, Bulk Data, Office Action):
- Register at https://data.uspto.gov/apis/getting-started
- Video verification required during registration
- Used via `X-API-KEY` header

**TSDR API Key** (Trademark Status & Document Retrieval):
- Register at https://account.uspto.gov/profile/api-manager
- Separate key from ODP
- Used via `USPTO-API-KEY` header

## Installation

```bash
go get github.com/patent-dev/uspto-odp
```

## Quick Start

```go
config := odp.DefaultConfig()
config.APIKey = "your-api-key"
client, err := odp.NewClient(config)
ctx := context.Background()

results, err := client.SearchPatents(ctx, "artificial intelligence", 0, 10)
fmt.Printf("Found %d patents\n", *results.Count)
```

## API Methods - Complete Coverage (53 wrapper methods)

All USPTO ODP API endpoints are fully implemented and tested, plus Office Action and TSDR APIs.

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

### Petition API (3 endpoints)

```go
SearchPetitions(ctx, query string, offset, limit int) (*PetitionDecisionResponseBag, error)
GetPetitionDecision(ctx, recordID string, includeDocuments bool) (*PetitionDecisionIdentifierResponseBag, error)
SearchPetitionsDownload(ctx, req PetitionDecisionDownloadRequest) ([]byte, error)
```

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

These use the DSAPI pattern (form-encoded POST with Lucene/Solr queries). Same API key as the main ODP API.

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

### TSDR (Trademark Status & Document Retrieval) API

Separate server (`tsdrapi.uspto.gov`) and API key. Requires `TSDRAPIKey` in config.

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

## Patent Full Text & Advanced Features

### Patent Number Normalization

The library handles various patent number formats and automatically resolves them to application numbers:

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

**Note:** Grant and publication numbers are **not** the same as application numbers. The library uses the search API to resolve grant/publication numbers to their corresponding application numbers.

Supported formats:
- Applications: `17248024`, `17/248,024`, `US 17/248,024`
- Grants: `11646472`, `11,646,472`, `US 11,646,472 B2`
- Publications: `20250087686`, `US20250087686A1` (kind code preserved when supplied: `A2`, `A9`, ...)
- PCT: `PCTUS2025058371` (15-char API form), `PCT/US2025/058371` (17-char display), `PCTUS0719317` (12-char legacy). Use `pn.FormatAsPCT()` for the display form.

**Note:** 8-digit numbers (like `11646472`) are ambiguous - they could be either grant or application numbers. Use formatting (commas, kind codes) to disambiguate.

### XML Full Text Retrieval

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

## Bulk File Downloads

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

## Configuration

```go
config := &odp.Config{
    BaseURL:    "https://api.uspto.gov", // Default
    APIKey:     "your-api-key",
    UserAgent:  "YourApp/1.0",
    MaxRetries: 3,                       // Retry failed requests
    RetryDelay: 1 * time.Second,         // Base backoff between retries
    Timeout:    30 * time.Second,        // Request timeout

    // Office Action DSAPI - separate host while USPTO completes its migration
    OABaseURL:  "https://developer.uspto.gov", // Default; no API key required at this host

    // TSDR (optional - separate server and API key)
    TSDRAPIKey: "your-tsdr-key",         // From https://account.uspto.gov/profile/api-manager
    TSDRBaseURL: "https://tsdrapi.uspto.gov", // Default
}

client, err := odp.NewClient(config)
```

When the server returns 429 with a `Retry-After` header, the client honors the
header value (capped at 60s) instead of the exponential backoff.

## Package Structure

```
├── client.go              # Main client implementation (package odp)
├── office_action.go       # Office Action API wrappers (DSAPI pattern)
├── tsdr.go                # TSDR (Trademark) API wrappers
├── types.go               # Typed response structs (continuity, assignment, adjustment, transactions)
├── patent_number.go       # Patent number normalization
├── xml.go                 # XML full text parsing (ICE DTD 4.6/4.7)
├── *_test.go              # Unit tests, integration tests
├── generated/             # Auto-generated OpenAPI code
│   ├── client_gen.go      # ODP client (package generated)
│   ├── types_gen.go       # ODP types (package generated)
│   ├── oa/                # Office Action DSAPI (package oa)
│   │   ├── client_gen.go
│   │   └── types_gen.go
│   └── tsdr/              # TSDR Trademark (package tsdr)
│       ├── client_gen.go
│       └── types_gen.go
├── cmd/gen/               # Code generation tool (pure Go)
│   └── main.go            # Bundles swagger files and applies fixes
├── demo/                  # Usage examples with saved responses
├── swagger/               # Official USPTO OpenAPI specs (DO NOT EDIT)
│   ├── swagger.yaml       # Main ODP API specification
│   ├── odp-common-base.yaml # Shared type definitions
│   ├── trial-*.yaml       # PTAB API specifications
│   ├── oa-*.yaml          # Office Action DSAPI specifications
│   └── tsdr-swagger.json  # TSDR API specification
├── swagger_fixed.yaml     # Processed ODP spec (auto-generated)
├── swagger_oa_fixed.yaml  # Processed OA spec (auto-generated)
├── swagger_tsdr_fixed.json# Processed TSDR spec (auto-generated)
└── dtd/                   # ICE DTD documentation
```

## Implementation

This library provides a Go client for the USPTO ODP API through a multi-step process:

1. **API Specification**: Started with the official [USPTO ODP Swagger specification](https://data.uspto.gov/swagger/index.html#/)
2. **Fix Mismatches**: Fixed type mismatches between swagger and actual API responses (see [Swagger Fixes](#swagger-fixes-applied))
3. **Code Generation**: Generate types and client code using [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) into `generated/` package
4. **Idiomatic Wrapper**: Wrap generated code in a clean, idiomatic Go client with retry logic (main `odp` package)

## Testing

### Unit Tests

```bash
go test -v
go test -v -cover
```

### Integration Tests

Requires `USPTO_API_KEY` and optionally `USPTO_TSDR_API_KEY`:
```bash
# Set your API keys (add to ~/.zshrc for persistence)
export USPTO_API_KEY=your-odp-key
export USPTO_TSDR_API_KEY=your-tsdr-key  # from https://account.uspto.gov/profile/api-manager

# Run all integration tests
go test -tags=integration -v

# Office Action APIs
go test -tags=integration -v -run TestOfficeActionAPIs

# TSDR API
go test -tags=integration -v -run TestTSDRAPIs

# Run specific endpoint test
go test -tags=integration -v -run TestIntegrationWithRealAPI/GetStatusCodes

# Test bulk file download (skipped by default due to large file size)
TEST_BULK_DOWNLOAD=true go test -tags=integration -v -run DownloadBulkFile
```

## Endpoint Coverage

All USPTO ODP API endpoints are implemented and tested:
- 13 Patent Application API endpoints
- 3 Bulk Data API endpoints
- 3 Petition API endpoints
- 19 PTAB (Patent Trial and Appeal Board) API endpoints
- 8 Office Action DSAPI endpoints (Text Retrieval, Citations, Rejections, Enriched Citations)
- 7 TSDR (Trademark Status & Document Retrieval) wrapper methods

## Swagger Processing

### Source Files

The USPTO ODP API specification is distributed as multiple YAML files with `$ref` references between them. The original files are downloaded from [USPTO ODP Swagger](https://data.uspto.gov/swagger/index.html#/) and stored in `swagger/`:

```
swagger/
├── swagger.yaml                 # Main ODP API spec (Patent, Bulk, Petition)
├── odp-common-base.yaml         # Shared type definitions
├── trial-proceedings.yaml       # PTAB trial proceedings
├── trial-decisions.yaml         # PTAB trial decisions
├── trial-documents.yaml         # PTAB trial documents
├── trial-appeal-decisions.yaml  # PTAB appeal decisions
├── trial-interferences.yaml     # PTAB interference decisions
├── trial-common.yaml            # Shared PTAB types
├── oa-text-retrieval.yaml       # Office Action Text Retrieval
├── oa-citations.yaml            # Office Action Citations
├── oa-rejections.yaml           # Office Action Rejections
├── oa-enriched-citations.yaml   # Enriched Citations
└── tsdr-swagger.json            # TSDR (Trademark) API
```

**Important:** Do not edit files in `swagger/` - these are the original USPTO specifications.

### Code Generation

The `cmd/gen` tool (pure Go, no external dependencies) processes these files:

```bash
go run ./cmd/gen
```

This tool:
1. **Bundles** ODP + PTAB YAML files, resolving `$ref` references -> `swagger_fixed.yaml` -> `generated/`
2. **Bundles** 4 Office Action specs -> `swagger_oa_fixed.yaml` -> `generated/oa/`
3. **Copies** TSDR spec -> `swagger_tsdr_fixed.json` -> `generated/tsdr/`
4. **Applies fixes** for mismatches between swagger specs and actual API responses

### Fixes Applied

The USPTO swagger specification has several mismatches with actual API responses:

**Type Corrections:**
- `frameNumber`, `reelNumber`: string -> integer (API returns numeric values)
- `documentNumber`: string -> integer (PTAB API returns numbers)
- Error response `code`: integer -> string (API returns `"404"` not `404`)

**Structure Fixes:**
- `petitionIssueConsideredTextBag`: array of objects -> array of strings
- `correspondenceAddress`: array -> object (Assignment API returns object)
- `DecisionData.statuteAndRuleBag`, `issueTypeBag`: string -> array (PTAB API returns arrays)
- `GetPatentAssignment.assignmentBag`: single object -> array (API returns array of assignments)

**Field Name Fixes:**
- `InterferenceDecisionRecord.decisionDocumentData` -> `documentData` (API uses different field name)
- `DecisionDataResponse.patentTrialDecisionDataBag` -> `patentTrialDocumentDataBag` (PTAB trial-decisions search returns the array under `patentTrialDocumentDataBag`, same as trial-documents; without this fix the slice silently unmarshals to nil while `count` parses fine)

**Format Fixes:**
- Removed `format: date-time` from datetime fields that return non-RFC3339 formats (e.g., `lastModifiedDateTime` returns `"2025-11-26T23:58:00"` without timezone)
- Removed `format: date` from datetime fields (e.g., `appealLastModifiedDateTime` returns datetime, not date)
- Removed `format: date` from fields returning non-ISO dates (e.g., `fileReleaseDate` returns `"2025-09-23 00:57:53"`)

**Endpoint Fixes:**
- Removed `/api/v1/patent/applications/text-to-search` (defined in spec but has no operations)

**Office Action DSAPI Fixes:**
- Removed phantom path parameters (`dataset`, `version` declared as `in: path` but paths are static)
- Made `operationId` values unique across bundled specs (all 4 specs used identical IDs)
- Rewrote OA paths from `/api/v1/patent/oa/<api>/<v>/<x>` to `/ds-api/<api>/<v>/<x>` and replaced the spec server URL from `https://api.uspto.gov` to `https://developer.uspto.gov`. The ODP 3.5 (2026-03-24) migration is incomplete on USPTO's side; the documented endpoints 404 today, while the legacy Developer Hub at `developer.uspto.gov/ds-api/...` still serves the same payloads. The fix lives in `applyOAFixes` and runs at bundle time. Override at runtime via `Config.OABaseURL` when USPTO completes the migration.

**TSDR Fixes:**
- Fixed protocol-relative server URL (`//tsdrapi.uspto.gov/` -> `https://tsdrapi.uspto.gov`)
- `GetDocumentInfoXml` and `GetCaseDocsInfoXml` endpoints return 406 with `Accept: application/json` - content negotiation is not supported despite XML/JSON server paths in spec
- Removed `format: date-time` and `format: date` from 64 fields - API returns inconsistent date formats (date-only `"2021-11-19"` in fields declared as `date-time`, non-ISO formats elsewhere), causing `time.Time` parsing failures in generated code

**Bugs Fixed by USPTO:**
- `trial-appeal-decisions.yaml`: `appelantData` -> `appellantData` (spelling), `realPartyName` -> `realPartyInInterestName`, `techCenterNumber` -> `technologyCenterNumber`, `requestorData` -> `thirdPartyRequesterData`, `documentTypeCategory` -> `documentTypeDescriptionText`, `downloadURI` -> `fileDownloadURI`, added `decisionData` block, added `requestIdentifier`
- `trial-common.yaml`: `downloadURI` -> `fileDownloadURI`, `statuteAndRuleBag`/`issueTypeBag` string -> array, `documentNumber` string -> integer, added `RegularPetitionerData` fields, added `appealOutcomeCategory`
- `trial-decisions.yaml`: now uses full inline schemas instead of `allOf` refs

## Version History

### v1.5.1 - PTAB decisions field name fix

- `DecisionDataResponse.PatentTrialDecisionDataBag` renamed to `PatentTrialDocumentDataBag` to match the actual API response. Previously the slice always unmarshaled to nil while `count` parsed fine. Mechanically breaking for consumers, but the old field never had data.
- `cmd/gen` now sorts source files deterministically; previously the bundled spec depended on Go map iteration order.

### v1.5.0 - PCT support, typed responses, ergonomics

Breaking changes (consumers must update in lockstep):
- `Config.Timeout` and `Config.RetryDelay` are `time.Duration` instead of `int` seconds.
- Search wrappers (`SearchPatents`, `SearchPetitions`, all PTAB `Search*`,
  all Office Action `Search*`) take `int` instead of `int32` for
  offsets/limits.
- `GetPatentMetaData` returns `*MetaDataResponse`, covering every field on
  `ApplicationMetaData` including applicants, inventors, entity status,
  classification bags, and PCT data. Numeric and boolean fields are
  pointer-typed where the API distinguishes absent from zero.
- `GetPatentForeignPriority` returns `*ForeignPriorityResponse`.
- `AssignmentEntry.Assignors []Assignor` and `Assignees []Assignee` are
  separate types: assignors carry name + execution date; assignees carry
  name + address. New fields: `MailedDate`, `ReceivedDate`,
  `DocumentLocationURI`.
- `AdjustmentResponse` drops `FilingDate` / `GrantDate` (the adjustment
  endpoint does not return them). Use `GetPatentMetaData` instead.

New:
- PCT number support in `NormalizePatentNumber`, `ResolvePatentNumber`, and
  `GetPatent`: 15-char API form (`PCTUS2025058371`), 17-char display form
  (`PCT/US2025/058371`), and 12-char legacy form. `FormatAsPCT()` returns
  the display form.
- Publication kind codes (`A1`, `A2`, `A9`, ...) flow through
  `ResolvePatentNumber` when supplied (previously forced to `A1`).
- `APIError.RetryAfter` (`time.Duration`) populated from the `Retry-After`
  response header. `retryableRequest` honors it. When the requested wait
  exceeds `retryAfterCap` (60s), `IsRetryable()` returns false so the
  caller can decide rather than the client truncating. `Retry-After: 0`
  triggers an immediate retry.
- Assignment address parsing reads `geographicRegionCode` (current 3.6
  field) and falls back to `countryOrStateCode` for older records.
- `Config.OABaseURL` (default `https://developer.uspto.gov`) routes the
  Office Action DSAPI wrappers to the legacy host that still serves them
  while ODP 3.5's migration to `api.uspto.gov` is in progress.
- `cmd/gen` rewrites the OA spec's paths from `/api/v1/patent/oa/...` to
  `/ds-api/...` and sets the spec server URL to `developer.uspto.gov` so
  the generated client targets the working endpoints.
- `Config.MaxRetryAfter` (default 60s) caps how long the client will wait
  for a server-requested retry; longer waits surface as non-retryable
  `*APIError` so the caller can decide.
- `Config.UserAgent` defaults to `uspto-odp/<version> (patent.dev; +<repo>)`
  exposed via `DefaultUserAgent` and `Version` constants.
- Search wrappers validate `offset`/`limit` against `int32` bounds at the
  call boundary instead of silently truncating.
- `DownloadBulkFile`/`DownloadBulkFileWithProgress` retry the
  connection-setup phase (request, status check) through `retryableRequest`;
  mid-stream errors propagate without retry to avoid overwriting partial
  writes on non-seekable writers.
- `xml.go`'s `DownloadXMLWithType` uses the configured `c.httpClient` so
  `Config.Timeout` actually applies.
- `MetaDataResponse.ApplicationConfirmationNumber` is `*int` (the swagger
  types it as `number`/`float32`; we coerce to a sane Go type in the
  wrapper).
- `TestGeneratedClientCoverage` matches generated client methods used as
  function calls in the wrapper sources, asserting every endpoint is
  either wrapped or pattern-allow-listed.

### v1.4.0 - Full ODP Coverage (Office Action + TSDR)
- Office Action APIs: Text Retrieval, Citations, Rejections, Enriched Citations (8 endpoints)
- TSDR (Trademark Status & Document Retrieval) API (24 endpoints, separate server + key)
- Updated PTAB swagger specs with USPTO's field name fixes
- Separate generated packages: `generated/oa/`, `generated/tsdr/`
- DSAPI pattern support (form-encoded POST, Lucene/Solr queries)
- TSDR content negotiation (JSON via Accept header, XML default)
- Integration tests for all new APIs

### v1.3.0 - Strongly-Typed Response Parsing
- `GetPatentContinuity` returns `*ContinuityResponse` with Parents/Children and relationship types
- `GetPatentAssignment` returns `*AssignmentResponse` with assignors, assignees, reel/frame
- `GetPatentAdjustment` returns `*AdjustmentResponse` with PTA delay breakdown
- `GetPatentTransactions` returns `*TransactionsResponse` with event date/code/description
- Unit tests with real JSON fixtures and edge cases
- Enhanced integration tests with typed assertions and fixture update mechanism

### v1.2.0 - PTAB API Complete (2025-11-27)
- Support for USPTO ODP 3.0 (released 2025-11-21) which added PTAB datasets
- Added 19 PTAB (Patent Trial and Appeal Board) API endpoints
- Trial Proceedings, Decisions, Documents, Appeal Decisions, Interference Decisions
- Pure Go code generation tool (`cmd/gen`) with no external dependencies
- Multi-file swagger processing (USPTO distributes spec as multiple YAML files)
- Demo with example saving (request/response pairs for documentation)
- Fixed API/swagger mismatches for PTAB endpoints

### v1.1.0 - Patent Number Normalization & XML Parsing
- Patent number normalization (accepts any format: grant, application, publication)
- `ResolvePatentNumber()` to convert grant/publication numbers to application numbers
- XML full text parsing (ICE DTD 4.6/4.7)
- Refactored demo suite

### v1.0.0 - Initial Release
- Complete USPTO ODP API client with 19 endpoints
- Patent Application API (13 endpoints)
- Bulk Data API (3 endpoints)
- Petition API (3 endpoints)
- Retry logic and configurable timeouts

## Related Projects

Part of the [patent.dev](https://patent.dev) open-source patent data ecosystem:

- [epo-ops](https://github.com/patent-dev/epo-ops) - EPO Open Patent Services client (search, biblio, legal status, family, images)
- [epo-bdds](https://github.com/patent-dev/epo-bdds) - EPO Bulk Data Distribution Service client
- [dpma-connect-plus](https://github.com/patent-dev/dpma-connect-plus) - DPMA Connect Plus client (patents, designs, trademarks)

The [bulk-file-loader](https://github.com/patent-dev/bulk-file-loader) uses these libraries for automated patent data downloads.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Credits

**Developed by:**
- Wolfgang Stark - [patent.dev](https://patent.dev) - [Funktionslust GmbH](https://funktionslust.digital)

## Acknowledgments

- USPTO for providing the Open Data Portal API
- [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) for code generation
