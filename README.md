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
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/patent-dev/uspto-odp"
)

func main() {
    // Create client
    client, err := odp.NewClient(&odp.Config{
        APIKey: "your-api-key",
    })
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Search patents
    results, err := client.SearchPatents(ctx, "artificial intelligence", 0, 10)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d patents\n", *results.Count)

    // Get status codes (most reliable endpoint)
    statuses, err := client.GetStatusCodes(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Retrieved %d status codes\n", *statuses.Count)
}
```

## API Methods - Complete Coverage (19 endpoints)

All 19 functional USPTO ODP API endpoints are fully implemented and tested.

### Patent Application API (13 endpoints)

```go
// Core Patent Data
SearchPatents(ctx, query string, offset, limit int32) (*PatentDataResponse, error)
GetPatent(ctx, applicationNumber string) (*PatentDataResponse, error)
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

## Bulk File Downloads

Download bulk data files using the FileDownloadURI provided by the API:

```go
// 1. Get the bulk product to access file metadata
product, err := client.GetBulkProduct(ctx, "PTGRXML")
if err != nil {
    log.Fatal(err)
}

// 2. Find your desired file and use its FileDownloadURI directly
files := *product.BulkDataProductBag[0].ProductFileBag.FileDataBag
for _, file := range files {
    if file.FileName != nil && strings.Contains(*file.FileName, "ipg250923.zip") {
        if file.FileDownloadURI != nil {
            // 3. Download using the FileDownloadURI directly
            err := client.DownloadBulkFileWithProgress(ctx, *file.FileDownloadURI, outputFile,
                func(bytesComplete, bytesTotal int64) {
                    percent := float64(bytesComplete) * 100 / float64(bytesTotal)
                    fmt.Printf("\rProgress: %.1f%%", percent)
                })
            if err != nil {
                log.Fatal(err)
            }
        }
        break
    }
}
```

**URL Validation**: The download methods validate that the provided URL is a valid FileDownloadURI from the USPTO API (must start with `https://api.uspto.gov/api/v1/datasets/products/files/`).

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
├── client.go           # Main client implementation (package odp)
├── client_test.go      # Unit tests with mock server
├── integration_test.go # Integration tests (real API)
├── generated/          # Auto-generated OpenAPI code
│   ├── client_gen.go   # Generated client (package generated)
│   └── types_gen.go    # Generated types (package generated)
└── swagger_fixed.yaml  # Fixed OpenAPI specification
```

## Implementation

This library provides a Go client for the USPTO ODP API through a multi-step process:

1. **API Specification**: Started with the official [USPTO ODP Swagger specification](https://data.uspto.gov/swagger/index.html#/)
2. **Fix Mismatches**: Fixed type mismatches between swagger and actual API responses (see [Swagger Fixes](#swagger-fixes-applied))
3. **Code Generation**: Generate types and client code using [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) into `generated/` package
4. **Idiomatic Wrapper**: Wrap generated code in a clean, idiomatic Go client with retry logic (main `odp` package)

## Testing

This library includes two types of tests serving different purposes:

### Unit Tests (Mock Server)

Offline tests using a mock HTTP server with hardcoded responses based on swagger & real API. These tests verify the client's parsing logic without making actual API calls.

**Run unit tests:**
```bash
# Run all unit tests
go test -v

# Run specific test
go test -v -run TestClientWithActualResponses/SearchPatents

# Run with coverage
go test -v -cover
```

### Integration Tests (Real API)

Tests that make actual HTTP requests to `https://api.uspto.gov` to validate our swagger fixes and ensure compatibility with the real USPTO API.

**Run integration tests:**
```bash
# Set your API key (add to ~/.zshrc for persistence)
export USPTO_API_KEY=your-api-key

# Run all integration tests
go test -tags=integration -v

# Run specific endpoint test
go test -tags=integration -v -run TestIntegrationWithRealAPI/GetStatusCodes

# Test endpoint coverage documentation
go test -tags=integration -v -run TestEndpointCoverage

# Test bulk file download (skipped by default due to large file size)
TEST_BULK_DOWNLOAD=true go test -tags=integration -v -run DownloadBulkFile
```

**Note**:
- Integration tests require `USPTO_API_KEY` environment variable and will fail with a clear error message if not set.
- Bulk file download test is skipped by default to avoid downloading large files (can be several GB). Set `TEST_BULK_DOWNLOAD=true` to run it.

## Project Structure

```
.
├── README.md           # This file
├── swagger_fixed.yaml  # Fixed OpenAPI spec (corrected from original)
├── swagger_original.yaml # Original USPTO swagger (for reference)
├── types_gen.go        # Generated types (DO NOT EDIT)
├── client_gen.go       # Generated HTTP client (DO NOT EDIT)
├── client.go           # Clean wrapper with retry logic
├── client_test.go      # Unit tests with mock
├── integration_test.go # Real API tests
├── go.mod              # Module definition
└── go.sum              # Module dependencies
```

## Endpoint Coverage Status

**100% Coverage**: All 19 functional USPTO ODP API endpoints are implemented and tested:
- 13 Patent Application API endpoints
- 3 Bulk Data API endpoints
- 3 Petition API endpoints

## Swagger Fixes Applied

The original USPTO swagger specification had several type mismatches with the actual API responses. We maintain a fixed version (`swagger_fixed.yaml`) with the following corrections:

### 1. Type Corrections
- **`applicationConfirmationNumber`**: Changed from `string` to `number`
  - API returns numeric value (e.g., 1061) not string
  - Fixed in PatentData schema

- **`prosecutionStatusCode`**: Changed from `string` to `number`
  - API returns numeric status codes
  - Fixed in PatentData schema

- **`customerNumberCorrespondenceData`**: Changed from array to object
  - API returns a single object, not an array
  - Fixed in PatentData schema

### 2. Structure Fixes
- **`BulkDataProductBag`**: Changed from array type alias to proper object
  - Was causing unmarshaling errors with array type alias
  - Fixed to be an object containing the array

- **`petitionIssueConsideredTextBag`**: Changed from array of objects to array of strings
  - API returns simple string array
  - Fixed in PetitionDecision schema

### 3. Format Removals
- **Date fields**: Removed `format: date` from non-ISO date fields
  - Fields like `createDateTime`, `mailDateTime` return custom format "YYYY-MM-DD HH:MM:SS"
  - Removed format constraint to allow proper parsing

### 4. Endpoint Removals
- **`/api/v1/patent/applications/text-to-search`**: Removed entirely
  - Endpoint defined but has no operations (no GET/POST/etc methods)
  - Cannot be implemented or used

### 5. Endpoint Additions
- **`/api/v1/datasets/products/files/{productIdentifier}/{fileName}`**: Added missing endpoint
  - Not present in original swagger but exists in the actual API
  - Returns 302 redirect to download bulk data files
  - Essential for downloading files referenced in bulk product responses

These fixes ensure the generated Go client correctly unmarshals all API responses without type errors and provides access to all functional endpoints.


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
