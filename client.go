package usptoapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ODPClient is the main USPTO ODP API client
type ODPClient struct {
	config    *Config
	generated *ClientWithResponses
}

// Config holds client configuration
type Config struct {
	BaseURL    string
	APIKey     string
	UserAgent  string
	MaxRetries int
	RetryDelay int // seconds
	Timeout    int // seconds
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:    "https://api.uspto.gov",
		UserAgent:  "PatentDev/1.0",
		MaxRetries: 3,
		RetryDelay: 1,
		Timeout:    30,
	}
}

// NewODPClient creates a new USPTO ODP API client
func NewODPClient(config *Config) (*ODPClient, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	// Add authentication headers
	requestEditor := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("User-Agent", config.UserAgent)
		if config.APIKey != "" {
			req.Header.Set("X-API-Key", config.APIKey)
		}
		return nil
	}

	// Create generated client
	genClient, err := NewClientWithResponses(
		config.BaseURL,
		WithHTTPClient(httpClient),
		WithRequestEditorFn(requestEditor),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &ODPClient{
		config:    config,
		generated: genClient,
	}, nil
}

// retryableRequest wraps requests with retry logic
func (c *ODPClient) retryableRequest(fn func() error) error {
	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err

		if attempt < c.config.MaxRetries {
			time.Sleep(time.Duration(c.config.RetryDelay*(attempt+1)) * time.Second)
		}
	}
	return fmt.Errorf("failed after %d retries: %w", c.config.MaxRetries, lastErr)
}

// Patent Application API Methods

// SearchPatents searches for patent applications
func (c *ODPClient) SearchPatents(ctx context.Context, query string, offset, limit int32) (*PatentDataResponse, error) {
	req := PatentSearchRequest{
		Q: StringPtr(query),
		Pagination: &Pagination{
			Offset: Int32Ptr(offset),
			Limit:  Int32Ptr(limit),
		},
	}

	var resp *PostApiV1PatentApplicationsSearchResponse
	err := c.retryableRequest(func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentApplicationsSearchWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("API returned status %d", resp.StatusCode())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetPatent retrieves a specific patent application
func (c *ODPClient) GetPatent(ctx context.Context, applicationNumber string) (*PatentDataResponse, error) {
	var resp *GetApiV1PatentApplicationsApplicationNumberTextResponse
	err := c.retryableRequest(func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("API returned status %d", resp.StatusCode())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetPatentAdjustment retrieves patent term adjustment data
func (c *ODPClient) GetPatentAdjustment(ctx context.Context, applicationNumber string) (interface{}, error) {
	var resp *GetApiV1PatentApplicationsApplicationNumberTextAdjustmentResponse
	err := c.retryableRequest(func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextAdjustmentWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("API returned status %d", resp.StatusCode())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetPatentContinuity retrieves patent continuity data
func (c *ODPClient) GetPatentContinuity(ctx context.Context, applicationNumber string) (interface{}, error) {
	var resp *GetApiV1PatentApplicationsApplicationNumberTextContinuityResponse
	err := c.retryableRequest(func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextContinuityWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("API returned status %d", resp.StatusCode())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetPatentDocuments retrieves patent documents list
func (c *ODPClient) GetPatentDocuments(ctx context.Context, applicationNumber string) (*DocumentBag, error) {
	params := &GetApiV1PatentApplicationsApplicationNumberTextDocumentsParams{}
	var resp *GetApiV1PatentApplicationsApplicationNumberTextDocumentsResponse
	err := c.retryableRequest(func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextDocumentsWithResponse(ctx, applicationNumber, params)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("API returned status %d", resp.StatusCode())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetStatusCodes retrieves all patent status codes
func (c *ODPClient) GetStatusCodes(ctx context.Context) (*StatusCodeSearchResponse, error) {
	params := &GetApiV1PatentStatusCodesParams{}
	var resp *GetApiV1PatentStatusCodesResponse
	err := c.retryableRequest(func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentStatusCodesWithResponse(ctx, params)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("API returned status %d", resp.StatusCode())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Bulk Data API Methods

// SearchBulkProducts searches for bulk data products
func (c *ODPClient) SearchBulkProducts(ctx context.Context, query string, offset, limit int) (*BdssResponseBag, error) {
	params := &GetApiV1DatasetsProductsSearchParams{
		Q:      StringPtr(query),
		Offset: IntPtr(offset),
		Limit:  IntPtr(limit),
	}

	var resp *GetApiV1DatasetsProductsSearchResponse
	err := c.retryableRequest(func() error {
		var err error
		resp, err = c.generated.GetApiV1DatasetsProductsSearchWithResponse(ctx, params)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("API returned status %d", resp.StatusCode())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetBulkProduct retrieves a specific bulk data product
func (c *ODPClient) GetBulkProduct(ctx context.Context, productID string) (*BdssResponseProductBag, error) {
	params := &GetApiV1DatasetsProductsProductIdentifierParams{}
	var resp *GetApiV1DatasetsProductsProductIdentifierResponse
	err := c.retryableRequest(func() error {
		var err error
		resp, err = c.generated.GetApiV1DatasetsProductsProductIdentifierWithResponse(ctx, productID, params)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("API returned status %d", resp.StatusCode())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// DownloadBulkFile downloads a specific bulk data file and writes it to the provided writer
// This is memory-efficient for large files as it streams directly to the writer
func (c *ODPClient) DownloadBulkFile(ctx context.Context, productID string, fileName string, w io.Writer) error {
	// First, get the redirect URL
	httpClient := &http.Client{
		Timeout: time.Duration(c.config.Timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Stop following redirects on first request
			return http.ErrUseLastResponse
		},
	}

	// Create request to get redirect URL
	url := fmt.Sprintf("%s/api/v1/datasets/products/files/%s/%s", c.config.BaseURL, productID, fileName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	// Add headers
	req.Header.Set("User-Agent", c.config.UserAgent)
	if c.config.APIKey != "" {
		req.Header.Set("X-API-Key", c.config.APIKey)
	}

	// Execute request to get redirect
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("getting redirect URL: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle 302 redirect
	if resp.StatusCode != http.StatusFound {
		return fmt.Errorf("expected 302 redirect, got status %d", resp.StatusCode)
	}

	downloadURL := resp.Header.Get("Location")
	if downloadURL == "" {
		return fmt.Errorf("302 redirect without Location header")
	}

	// Now download the actual file (allow redirects for the actual download)
	downloadClient := &http.Client{
		Timeout: 0, // No timeout for large file downloads
	}

	downloadReq, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("creating download request: %w", err)
	}

	downloadResp, err := downloadClient.Do(downloadReq)
	if err != nil {
		return fmt.Errorf("downloading file: %w", err)
	}
	defer func() { _ = downloadResp.Body.Close() }()

	if downloadResp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", downloadResp.StatusCode)
	}

	// Get expected size if available
	expectedSize := downloadResp.ContentLength

	// Stream the file content to the writer
	bytesWritten, err := io.Copy(w, downloadResp.Body)
	if err != nil {
		return fmt.Errorf("writing file data: %w", err)
	}

	// Validate we got the expected amount of data
	if expectedSize > 0 && bytesWritten != expectedSize {
		return fmt.Errorf("incomplete download: got %d bytes, expected %d", bytesWritten, expectedSize)
	}

	return nil
}

// DownloadBulkFileWithProgress downloads a file with optional progress callback
// The progress callback receives bytes written and total bytes (if known, -1 otherwise)
func (c *ODPClient) DownloadBulkFileWithProgress(ctx context.Context, productID string, fileName string, w io.Writer, progress func(bytesComplete int64, bytesTotal int64)) error {
	// First, get the redirect URL
	httpClient := &http.Client{
		Timeout: time.Duration(c.config.Timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	url := fmt.Sprintf("%s/api/v1/datasets/products/files/%s/%s", c.config.BaseURL, productID, fileName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.config.UserAgent)
	if c.config.APIKey != "" {
		req.Header.Set("X-API-Key", c.config.APIKey)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("getting redirect URL: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusFound {
		return fmt.Errorf("expected 302 redirect, got status %d", resp.StatusCode)
	}

	downloadURL := resp.Header.Get("Location")
	if downloadURL == "" {
		return fmt.Errorf("302 redirect without Location header")
	}

	// Download with progress
	downloadClient := &http.Client{Timeout: 0}
	downloadReq, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("creating download request: %w", err)
	}

	downloadResp, err := downloadClient.Do(downloadReq)
	if err != nil {
		return fmt.Errorf("downloading file: %w", err)
	}
	defer func() { _ = downloadResp.Body.Close() }()

	if downloadResp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", downloadResp.StatusCode)
	}

	expectedSize := downloadResp.ContentLength

	// Create progress reader if callback provided
	var reader io.Reader = downloadResp.Body
	if progress != nil {
		reader = &progressReader{
			reader:   downloadResp.Body,
			total:    expectedSize,
			callback: progress,
		}
	}

	bytesWritten, err := io.Copy(w, reader)
	if err != nil {
		return fmt.Errorf("writing file data: %w", err)
	}

	if expectedSize > 0 && bytesWritten != expectedSize {
		return fmt.Errorf("incomplete download: got %d bytes, expected %d", bytesWritten, expectedSize)
	}

	return nil
}

// progressReader wraps an io.Reader to report progress
type progressReader struct {
	reader   io.Reader
	total    int64
	written  int64
	callback func(bytesComplete int64, bytesTotal int64)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.written += int64(n)
		pr.callback(pr.written, pr.total)
	}
	return n, err
}

// GetBulkFileURL gets the download URL for a bulk data file
// Note: The URL is temporary and will expire
func (c *ODPClient) GetBulkFileURL(ctx context.Context, productID string, fileName string) (string, error) {
	// Create a custom HTTP client that doesn't follow redirects
	httpClient := &http.Client{
		Timeout: time.Duration(c.config.Timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Stop following redirects
			return http.ErrUseLastResponse
		},
	}

	// Create request
	url := fmt.Sprintf("%s/api/v1/datasets/products/files/%s/%s", c.config.BaseURL, productID, fileName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	// Add headers
	req.Header.Set("User-Agent", c.config.UserAgent)
	if c.config.APIKey != "" {
		req.Header.Set("X-API-Key", c.config.APIKey)
	}

	// Execute request
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle 302 redirect
	if resp.StatusCode == http.StatusFound {
		location := resp.Header.Get("Location")
		if location == "" {
			return "", fmt.Errorf("302 redirect without Location header")
		}
		return location, nil
	}

	return "", fmt.Errorf("API returned status %d", resp.StatusCode)
}

// Petition API Methods

// SearchPetitions searches for petition decisions
func (c *ODPClient) SearchPetitions(ctx context.Context, query string, offset, limit int32) (*PetitionDecisionResponseBag, error) {
	req := PetitionDecisionSearchRequest{
		Q: StringPtr(query),
		Pagination: &Pagination{
			Offset: Int32Ptr(offset),
			Limit:  Int32Ptr(limit),
		},
	}

	var resp *PostApiV1PetitionDecisionsSearchResponse
	err := c.retryableRequest(func() error {
		var err error
		resp, err = c.generated.PostApiV1PetitionDecisionsSearchWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("API returned status %d", resp.StatusCode())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Additional Patent Application API Methods

// GetPatentAssignment retrieves patent assignment data
func (c *ODPClient) GetPatentAssignment(ctx context.Context, applicationNumber string) (interface{}, error) {
	var resp *GetApiV1PatentApplicationsApplicationNumberTextAssignmentResponse
	err := c.retryableRequest(func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextAssignmentWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("API returned status %d", resp.StatusCode())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetPatentAssociatedDocuments retrieves associated documents
func (c *ODPClient) GetPatentAssociatedDocuments(ctx context.Context, applicationNumber string) (interface{}, error) {
	var resp *GetApiV1PatentApplicationsApplicationNumberTextAssociatedDocumentsResponse
	err := c.retryableRequest(func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextAssociatedDocumentsWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("API returned status %d", resp.StatusCode())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetPatentAttorney retrieves patent attorney information
func (c *ODPClient) GetPatentAttorney(ctx context.Context, applicationNumber string) (interface{}, error) {
	var resp *GetApiV1PatentApplicationsApplicationNumberTextAttorneyResponse
	err := c.retryableRequest(func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextAttorneyWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("API returned status %d", resp.StatusCode())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetPatentForeignPriority retrieves foreign priority data
func (c *ODPClient) GetPatentForeignPriority(ctx context.Context, applicationNumber string) (interface{}, error) {
	var resp *GetApiV1PatentApplicationsApplicationNumberTextForeignPriorityResponse
	err := c.retryableRequest(func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextForeignPriorityWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("API returned status %d", resp.StatusCode())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetPatentMetaData retrieves patent metadata
func (c *ODPClient) GetPatentMetaData(ctx context.Context, applicationNumber string) (interface{}, error) {
	var resp *GetApiV1PatentApplicationsApplicationNumberTextMetaDataResponse
	err := c.retryableRequest(func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextMetaDataWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("API returned status %d", resp.StatusCode())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetPatentTransactions retrieves patent transaction history
func (c *ODPClient) GetPatentTransactions(ctx context.Context, applicationNumber string) (interface{}, error) {
	var resp *GetApiV1PatentApplicationsApplicationNumberTextTransactionsResponse
	err := c.retryableRequest(func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextTransactionsWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("API returned status %d", resp.StatusCode())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// SearchPatentsDownload downloads patent search results
func (c *ODPClient) SearchPatentsDownload(ctx context.Context, req PatentDownloadRequest) ([]byte, error) {
	var resp *PostApiV1PatentApplicationsSearchDownloadResponse
	err := c.retryableRequest(func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentApplicationsSearchDownloadWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("API returned status %d", resp.StatusCode())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// GetPetitionDecision retrieves a specific petition decision
func (c *ODPClient) GetPetitionDecision(ctx context.Context, recordID string, includeDocuments bool) (*PetitionDecisionIdentifierResponseBag, error) {
	params := &GetApiV1PetitionDecisionsPetitionDecisionRecordIdentifierParams{
		IncludeDocuments: &includeDocuments,
	}
	var resp *GetApiV1PetitionDecisionsPetitionDecisionRecordIdentifierResponse
	err := c.retryableRequest(func() error {
		var err error
		resp, err = c.generated.GetApiV1PetitionDecisionsPetitionDecisionRecordIdentifierWithResponse(ctx, recordID, params)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("API returned status %d", resp.StatusCode())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// SearchPetitionsDownload downloads petition search results
func (c *ODPClient) SearchPetitionsDownload(ctx context.Context, req PetitionDecisionDownloadRequest) ([]byte, error) {
	var resp *PostApiV1PetitionDecisionsSearchDownloadResponse
	err := c.retryableRequest(func() error {
		var err error
		resp, err = c.generated.PostApiV1PetitionDecisionsSearchDownloadWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("API returned status %d", resp.StatusCode())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// Helper functions

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}

// Int32Ptr returns a pointer to an int32
func Int32Ptr(i int32) *int32 {
	return &i
}
