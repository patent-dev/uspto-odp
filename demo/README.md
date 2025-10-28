# USPTO ODP API Client Demo

Demonstration program for the USPTO Open Data Portal (ODP) API client library.

## Features

The demo covers all major API endpoints:

- **Patent API** (13 endpoints) - Search patents, retrieve metadata, documents, assignments, etc.
- **Petition API** (3 endpoints) - Search and retrieve petition decisions
- **Bulk Data** - Browse and download bulk data products with progress tracking
- **XML Full Text** - Retrieve and parse full patent text (grants and applications)

## Prerequisites

1. USPTO API Key (required)
   - Get your key from [USPTO Developer Portal](https://developer.uspto.gov/)
   - Set via environment variable: `export USPTO_API_KEY=your_key_here`
   - Or use `-key` flag

2. Go 1.25 or later

## Quick Start

```bash
# Set your API key
export USPTO_API_KEY=your_key_here

# Run all non-interactive demonstrations
cd demo
go run .

# Run specific service
go run . -service=patent
go run . -service=petition
go run . -service=xml
go run . -service=bulk

# Interactive mode
go run . -interactive

# Use specific patent number
go run . -service=patent -patent="US 11,646,472 B2"
go run . -service=patent -patent="17/248,024"
```

## Usage

### Command Line Flags

```
-key string
    USPTO API key (default: $USPTO_API_KEY)

-patent string
    Patent number to use for demonstrations (default: "17248024")
    Supports formats: "17248024", "17/248,024", "US 11,646,472 B2"

-service string
    Run specific service: patent, petition, xml, bulk
    If not specified, runs all non-interactive services

-interactive
    Run in interactive menu mode
```

### Examples

#### Test All Patent Endpoints

```bash
go run . -service=patent -patent="17248024"
```

This demonstrates all 13 Patent API endpoints:
- SearchPatents
- GetPatent
- GetPatentMetaData
- GetPatentAdjustment
- GetPatentContinuity
- GetPatentDocuments
- GetPatentAssignment
- GetPatentAssociatedDocuments
- GetPatentAttorney
- GetPatentForeignPriority
- GetPatentTransactions
- SearchPatentsDownload
- GetStatusCodes

#### Test Petition Endpoints

```bash
go run . -service=petition
```

This demonstrates all 3 Petition API endpoints:
- SearchPetitions
- GetPetitionDecision
- SearchPetitionsDownload

#### Retrieve Patent Full Text XML

```bash
go run . -service=xml
```

Interactive mode that:
- Prompts for patent number
- Fetches full XML document
- Parses and displays title, abstract, claims, description
- Shows document statistics

#### Browse and Download Bulk Data

```bash
go run . -service=bulk
```

Interactive mode that:
- Lists all available bulk data products
- Displays paginated file listings
- Supports file search
- Downloads files with progress tracking

#### Interactive Menu

```bash
go run . -interactive
```

Provides a menu to select and run any service interactively.

## Code Organization

```
demo/
├── main.go          # Entry point, CLI flags, routing
├── patent.go        # 13 Patent API endpoint demonstrations
├── petition.go      # 3 Petition API endpoint demonstrations
├── bulk.go          # Bulk data browser and downloader
├── xmldemo.go       # XML full text retrieval and parsing
├── utils.go         # Helper functions for display
├── testdata.go      # Test patent numbers
├── go.mod           # Module definition
└── README.md        # This file
```

## API Key Security

The demo accepts API keys via:
1. Environment variable `USPTO_API_KEY` (recommended)
2. Command line flag `-key`

For production use, always use environment variables or secure key management systems. Never hardcode API keys in source code.

## Error Handling

The demo displays errors from API calls directly. Common errors:
- `404`: Patent/petition not found
- `401`: Invalid or missing API key
- `429`: Rate limit exceeded
- Network errors: Check connectivity and API status

## Testing with Real Data

Test patent numbers included:
- Application: `17248024` or `17/248,024`
- Grant: `US 11,646,472 B2`
- Publication: `US20250087686A1`

These are for the same patent family and demonstrate various number formats the library supports.
