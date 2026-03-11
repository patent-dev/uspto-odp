# USPTO Open Data Portal (ODP) Go Client

[![Go Reference](https://pkg.go.dev/badge/github.com/patent-dev/uspto-odp.svg)](https://pkg.go.dev/github.com/patent-dev/uspto-odp)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A complete Go client library for the USPTO Open Data Portal API.

## Getting Started

### API Key Required

You need an API key to use the USPTO ODP API:
- **Details**: https://data.uspto.gov/apis/getting-started
- **Note**: Video verification is required during registration
- **Rate limits**: Check the documentation for current rate limits and usage guidelines

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

## API Methods - Complete Coverage (38 endpoints)

All 38 USPTO ODP API endpoints are fully implemented and tested.

### Patent Application API (13 endpoints)

```go
// Core Patent Data
SearchPatents(ctx, query string, offset, limit int32) (*PatentDataResponse, error)
GetPatent(ctx, patentNumber string) (*PatentDataResponse, error)  // Accepts any patent number format
GetPatentMetaData(ctx, applicationNumber string) (interface{}, error)

// Patent Details
GetPatentAdjustment(ctx, applicationNumber string) (*AdjustmentResponse, error)
GetPatentContinuity(ctx, applicationNumber string) (*ContinuityResponse, error)
GetPatentDocuments(ctx, applicationNumber string) (*DocumentBag, error)
GetPatentAssignment(ctx, applicationNumber string) (*AssignmentResponse, error)
GetPatentAssociatedDocuments(ctx, applicationNumber string) (any, error)
GetPatentAttorney(ctx, applicationNumber string) (any, error)
GetPatentForeignPriority(ctx, applicationNumber string) (any, error)
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
SearchPetitions(ctx, query string, offset, limit int32) (*PetitionDecisionResponseBag, error)
GetPetitionDecision(ctx, recordID string, includeDocuments bool) (*PetitionDecisionIdentifierResponseBag, error)
SearchPetitionsDownload(ctx, req PetitionDecisionDownloadRequest) ([]byte, error)
```

### PTAB (Patent Trial and Appeal Board) API (19 endpoints)

```go
// Trial Proceedings (IPR, PGR, CBM)
SearchTrialProceedings(ctx, query string, offset, limit int32) (*ProceedingDataResponse, error)
GetTrialProceeding(ctx, trialNumber string) (*ProceedingDataResponse, error)
SearchTrialProceedingsDownload(ctx, req DownloadRequest) ([]byte, error)

// Trial Decisions
SearchTrialDecisions(ctx, query string, offset, limit int32) (*DecisionDataResponse, error)
GetTrialDecision(ctx, documentIdentifier string) (*DecisionDataResponse, error)
GetTrialDecisionsByTrialNumber(ctx, trialNumber string) (*DecisionDataResponse, error)
SearchTrialDecisionsDownload(ctx, req DownloadRequest) ([]byte, error)

// Trial Documents
SearchTrialDocuments(ctx, query string, offset, limit int32) (*DocumentDataResponse, error)
GetTrialDocument(ctx, documentIdentifier string) (*DocumentDataResponse, error)
GetTrialDocumentsByTrialNumber(ctx, trialNumber string) (*DocumentDataResponse, error)
SearchTrialDocumentsDownload(ctx, req DownloadRequest) ([]byte, error)

// Appeal Decisions
SearchAppealDecisions(ctx, query string, offset, limit int32) (*AppealDecisionDataResponse, error)
GetAppealDecision(ctx, documentIdentifier string) (*AppealDecisionDataResponse, error)
GetAppealDecisionsByAppealNumber(ctx, appealNumber string) (*AppealDecisionDataResponse, error)
SearchAppealDecisionsDownload(ctx, req DownloadRequest) ([]byte, error)

// Interference Decisions
SearchInterferenceDecisions(ctx, query string, offset, limit int32) (*InterferenceDecisionDataResponse, error)
GetInterferenceDecision(ctx, documentIdentifier string) (*InterferenceDecisionDataResponse, error)
GetInterferenceDecisionsByNumber(ctx, interferenceNumber string) (*InterferenceDecisionDataResponse, error)
SearchInterferenceDecisionsDownload(ctx, req PatentDownloadRequest) ([]byte, error)
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
- Publications: `20250087686`, `US20250087686A1`

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
    RetryDelay: 1,                       // Seconds between retries
    Timeout:    30,                      // Request timeout in seconds
}

client, err := odp.NewClient(config)
```

## Package Structure

```
├── client.go            # Main client implementation (package odp)
├── types.go             # Typed response structs (continuity, assignment, adjustment, transactions)
├── patent_number.go     # Patent number normalization
├── xml.go               # XML full text parsing (ICE DTD 4.6/4.7)
├── client_test.go       # Unit tests with mock server
├── types_test.go        # Typed response tests with real fixtures
├── patent_number_test.go# Patent number normalization tests
├── xml_test.go          # XML parsing tests
├── integration_test.go  # Integration tests (real API)
├── generated/           # Auto-generated OpenAPI code
│   ├── client_gen.go    # Generated client (package generated)
│   └── types_gen.go     # Generated types (package generated)
├── cmd/gen/             # Code generation tool (pure Go)
│   └── main.go          # Bundles swagger files and applies fixes
├── demo/                # Usage examples with saved responses
│   └── main.go          # Demo runner for all API services
├── swagger/             # Official USPTO OpenAPI specs (DO NOT EDIT)
│   ├── swagger.yaml     # Main API specification
│   ├── odp-common-base.yaml  # Shared type definitions
│   └── trial-*.yaml     # PTAB API specifications
├── swagger_fixed.yaml   # Processed spec with fixes (auto-generated)
└── dtd/                 # ICE DTD documentation
    └── README.md        # DTD structure and information
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

Requires `USPTO_API_KEY` environment variable:
```bash
# Set your API key (add to ~/.zshrc for persistence)
export USPTO_API_KEY=your-api-key

# Run all integration tests
go test -tags=integration -v

# Run specific endpoint test
go test -tags=integration -v -run TestIntegrationWithRealAPI/GetStatusCodes

# Test endpoint coverage documentation
go test -tags=integration -v -run TestEndpointCoverage

# Test XML parsing with real API data
go test -tags=integration -v -run TestXMLParsing

# Test bulk file download (skipped by default due to large file size)
TEST_BULK_DOWNLOAD=true go test -tags=integration -v -run DownloadBulkFile
```

Integration tests require `USPTO_API_KEY` environment variable. Bulk file download test skipped by default (set `TEST_BULK_DOWNLOAD=true` to run).

## Endpoint Coverage

All 38 USPTO ODP API endpoints are implemented and tested:
- 13 Patent Application API endpoints
- 3 Bulk Data API endpoints
- 3 Petition API endpoints
- 19 PTAB (Patent Trial and Appeal Board) API endpoints

## Swagger Processing

### Source Files

The USPTO ODP API specification is distributed as multiple YAML files with `$ref` references between them. The original files are downloaded from [USPTO ODP Swagger](https://data.uspto.gov/swagger/index.html#/) and stored in `swagger/`:

```
swagger/
├── swagger.yaml           # Main API spec (Patent, Bulk, Petition endpoints)
├── odp-common-base.yaml   # Shared type definitions
├── trial-proceedings.yaml # PTAB trial proceedings
├── trial-decisions.yaml   # PTAB trial decisions
├── trial-documents.yaml   # PTAB trial documents
├── trial-appeal-decisions.yaml  # PTAB appeal decisions
├── trial-interferences.yaml     # PTAB interference decisions
└── trial-common.yaml      # Shared PTAB types
```

**Important:** Do not edit files in `swagger/` - these are the original USPTO specifications.

### Code Generation

The `cmd/gen` tool (pure Go, no external dependencies) processes these files:

```bash
go run ./cmd/gen
```

This tool:
1. **Bundles** all YAML files, resolving `$ref` references between files
2. **Applies fixes** for mismatches between swagger spec and actual API responses
3. **Generates** `swagger_fixed.yaml` (processed OpenAPI spec)
4. **Generates** Go code in `generated/` using oapi-codegen

### Fixes Applied

The USPTO swagger specification has several mismatches with actual API responses:

**Type Corrections:**
- `frameNumber`, `reelNumber`: string → integer (API returns numeric values)
- `documentNumber`: string → integer (PTAB API returns numbers)
- Error response `code`: integer → string (API returns `"404"` not `404`)

**Structure Fixes:**
- `petitionIssueConsideredTextBag`: array of objects → array of strings
- `correspondenceAddress`: array → object (Assignment API returns object)
- `DecisionData.statuteAndRuleBag`, `issueTypeBag`: string → array (PTAB API returns arrays)
- `GetPatentAssignment.assignmentBag`: single object → array (API returns array of assignments)

**Field Name Fixes:**
- `InterferenceDecisionRecord.decisionDocumentData` → `documentData` (API uses different field name)

**Format Fixes:**
- Removed `format: date-time` from datetime fields that return non-RFC3339 formats (e.g., `lastModifiedDateTime` returns `"2025-11-26T23:58:00"` without timezone)
- Removed `format: date` from datetime fields (e.g., `appealLastModifiedDateTime` returns datetime, not date)
- Removed `format: date` from fields returning non-ISO dates (e.g., `fileReleaseDate` returns `"2025-09-23 00:57:53"`)

**Endpoint Fixes:**
- Removed `/api/v1/patent/applications/text-to-search` (defined in spec but has no operations)

## Version History

### v1.3.0 - Strongly-Typed Response Parsing
- `GetPatentContinuity` returns `*ContinuityResponse` with Parents/Children and relationship types
- `GetPatentAssignment` returns `*AssignmentResponse` with assignors, assignees, reel/frame
- `GetPatentAdjustment` returns `*AdjustmentResponse` with PTA delay breakdown
- `GetPatentTransactions` returns `*TransactionsResponse` with event date/code/description
- Comprehensive unit tests with real JSON fixtures and edge cases
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
