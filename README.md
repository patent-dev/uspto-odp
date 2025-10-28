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
client, err := odp.NewClient(&odp.Config{APIKey: "your-api-key"})
ctx := context.Background()

results, err := client.SearchPatents(ctx, "artificial intelligence", 0, 10)
fmt.Printf("Found %d patents\n", *results.Count)
```

## API Methods - Complete Coverage (19 endpoints)

All 19 functional USPTO ODP API endpoints are fully implemented and tested.

### Patent Application API (13 endpoints)

```go
// Core Patent Data
SearchPatents(ctx, query string, offset, limit int32) (*PatentDataResponse, error)
GetPatent(ctx, patentNumber string) (*PatentDataResponse, error)  // Accepts any patent number format
GetPatentMetaData(ctx, applicationNumber string) (interface{}, error)

// Patent Details
GetPatentAdjustment(ctx, applicationNumber string) (interface{}, error)
GetPatentContinuity(ctx, applicationNumber string) (interface{}, error)
GetPatentDocuments(ctx, applicationNumber string) (*DocumentBag, error)
GetPatentAssignment(ctx, applicationNumber string) (interface{}, error)
GetPatentAssociatedDocuments(ctx, applicationNumber string) (interface{}, error)
GetPatentAttorney(ctx, applicationNumber string) (interface{}, error)
GetPatentForeignPriority(ctx, applicationNumber string) (interface{}, error)
GetPatentTransactions(ctx, applicationNumber string) (interface{}, error)

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
├── patent_number.go     # Patent number normalization
├── xml.go               # XML full text parsing (ICE DTD 4.6/4.7)
├── client_test.go       # Unit tests with mock server
├── patent_number_test.go# Patent number normalization tests
├── xml_test.go          # XML parsing tests
├── integration_test.go  # Integration tests (real API)
├── generated/           # Auto-generated OpenAPI code
│   ├── client_gen.go    # Generated client (package generated)
│   └── types_gen.go     # Generated types (package generated)
├── dtd/                 # ICE DTD documentation
│   └── README.md        # DTD structure and information
└── swagger_fixed.yaml   # Fixed OpenAPI specification
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

All 19 functional USPTO ODP API endpoints are implemented and tested:
- 13 Patent Application API endpoints
- 3 Bulk Data API endpoints
- 3 Petition API endpoints

## Swagger Fixes Applied

Fixed type mismatches in USPTO swagger specification (`swagger_fixed.yaml`):

### Type Corrections
- `applicationConfirmationNumber`, `prosecutionStatusCode`: string → number
- `frameNumber`, `reelNumber`: string → integer (API returns numeric values)

### Structure Fixes
- `BulkDataProductBag`: array alias → object with array field
- `assignmentBag`: single object → array of Assignment objects
- `petitionIssueConsideredTextBag`: array of objects → array of strings

### Format Fixes
- Removed `format: date` from non-ISO date fields (e.g., `createDateTime`, `mailDateTime`)

### Endpoint Changes
- Removed: `/api/v1/patent/applications/text-to-search` (defined but has no operations)
- Added: `/api/v1/datasets/products/files/{productIdentifier}/{fileName}` (missing from original swagger)

## Development

### Regenerating from Swagger

If the swagger spec is updated:

```bash
# Install generator
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# Generate types (DO NOT EDIT generated/types_gen.go directly)
oapi-codegen -package generated -generate types swagger_fixed.yaml > generated/types_gen.go

# Generate client (DO NOT EDIT generated/client_gen.go directly)
oapi-codegen -package generated -generate client swagger_fixed.yaml > generated/client_gen.go
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Credits

**Developed by:**
- Wolfgang Stark - [patent.dev](https://patent.dev) - [Funktionslust GmbH](https://funktionslust.digital)

## Acknowledgments

- USPTO for providing the Open Data Portal API
- [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) for code generation
