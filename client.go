package odp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/patent-dev/uspto-odp/generated"
)

// Client is the main USPTO ODP API client
type Client struct {
	config    *Config
	generated *generated.ClientWithResponses
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

// NewClient creates a new USPTO ODP API client
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	httpClient := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	requestEditor := generated.RequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("User-Agent", config.UserAgent)
		if config.APIKey != "" {
			req.Header.Set("X-API-Key", config.APIKey)
		}
		return nil
	})

	genClient, err := generated.NewClientWithResponses(
		config.BaseURL,
		generated.WithHTTPClient(httpClient),
		generated.WithRequestEditorFn(requestEditor),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Client{
		config:    config,
		generated: genClient,
	}, nil
}

// retryableRequest wraps requests with retry logic
func (c *Client) retryableRequest(fn func() error) error {
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

// SearchPatents searches for patent applications
func (c *Client) SearchPatents(ctx context.Context, query string, offset, limit int32) (*generated.PatentDataResponse, error) {
	req := generated.PatentSearchRequest{
		Q: StringPtr(query),
		Pagination: &generated.Pagination{
			Offset: Int32Ptr(offset),
			Limit:  Int32Ptr(limit),
		},
	}

	var resp *generated.PostApiV1PatentApplicationsSearchResponse
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

// resolveGrantToApplicationNumber searches for a grant number and returns its application number
func (c *Client) resolveGrantToApplicationNumber(ctx context.Context, grantNumber string) (string, error) {
	query := fmt.Sprintf("applicationMetaData.patentNumber:%s", grantNumber)

	result, err := c.SearchPatents(ctx, query, 0, 1)
	if err != nil {
		return "", fmt.Errorf("failed to search for grant number %s: %w", grantNumber, err)
	}

	if result.PatentFileWrapperDataBag == nil || len(*result.PatentFileWrapperDataBag) == 0 {
		return "", fmt.Errorf("no application found for grant number %s", grantNumber)
	}

	patent := (*result.PatentFileWrapperDataBag)[0]
	if patent.ApplicationNumberText == nil {
		return "", fmt.Errorf("application number not found in response for grant number %s", grantNumber)
	}

	return *patent.ApplicationNumberText, nil
}

// resolvePublicationToApplicationNumber searches for a publication number and returns its application number
func (c *Client) resolvePublicationToApplicationNumber(ctx context.Context, publicationNumber string) (string, error) {
	// Format publication number for search (e.g., 20250087686 -> US20250087686A1)
	formattedPub := publicationNumber
	if len(publicationNumber) == 11 && !strings.HasPrefix(publicationNumber, "US") {
		formattedPub = "US" + publicationNumber + "A1"
	}

	query := fmt.Sprintf("applicationMetaData.earliestPublicationNumber:%s", formattedPub)

	result, err := c.SearchPatents(ctx, query, 0, 1)
	if err != nil {
		return "", fmt.Errorf("failed to search for publication number %s: %w", publicationNumber, err)
	}

	if result.PatentFileWrapperDataBag == nil || len(*result.PatentFileWrapperDataBag) == 0 {
		return "", fmt.Errorf("no application found for publication number %s", publicationNumber)
	}

	patent := (*result.PatentFileWrapperDataBag)[0]
	if patent.ApplicationNumberText == nil {
		return "", fmt.Errorf("application number not found in response for publication number %s", publicationNumber)
	}

	return *patent.ApplicationNumberText, nil
}

// ResolvePatentNumber resolves any patent number format (application, grant, or publication)
// to its application number by searching the USPTO API when necessary.
// For application numbers, returns the normalized number directly.
// For grant and publication numbers, performs an API search to find the corresponding application number.
func (c *Client) ResolvePatentNumber(ctx context.Context, patentNumber string) (string, error) {
	pn, err := NormalizePatentNumber(patentNumber)
	if err != nil {
		return "", fmt.Errorf("invalid patent number: %w", err)
	}

	switch pn.Type {
	case PatentNumberTypeGrant:
		return c.resolveGrantToApplicationNumber(ctx, pn.Normalized)
	case PatentNumberTypePublication:
		return c.resolvePublicationToApplicationNumber(ctx, pn.Normalized)
	case PatentNumberTypeApplication:
		return pn.ToApplicationNumber(), nil
	default:
		return "", fmt.Errorf("unknown patent number type")
	}
}

// GetPatent retrieves patent data by application, grant, or publication number
func (c *Client) GetPatent(ctx context.Context, patentNumber string) (*generated.PatentDataResponse, error) {
	applicationNumber, err := c.ResolvePatentNumber(ctx, patentNumber)
	if err != nil {
		return nil, err
	}

	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextResponse
	err = c.retryableRequest(func() error {
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
func (c *Client) GetPatentAdjustment(ctx context.Context, applicationNumber string) (any, error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextAdjustmentResponse
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
func (c *Client) GetPatentContinuity(ctx context.Context, applicationNumber string) (any, error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextContinuityResponse
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
func (c *Client) GetPatentDocuments(ctx context.Context, applicationNumber string) (*generated.DocumentBag, error) {
	params := &generated.GetApiV1PatentApplicationsApplicationNumberTextDocumentsParams{}
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextDocumentsResponse
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
func (c *Client) GetStatusCodes(ctx context.Context) (*generated.StatusCodeSearchResponse, error) {
	params := &generated.GetApiV1PatentStatusCodesParams{}
	var resp *generated.GetApiV1PatentStatusCodesResponse
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

// SearchBulkProducts searches for bulk data products
func (c *Client) SearchBulkProducts(ctx context.Context, query string, offset, limit int) (*generated.BdssResponseBag, error) {
	params := &generated.GetApiV1DatasetsProductsSearchParams{
		Q:      StringPtr(query),
		Offset: IntPtr(offset),
		Limit:  IntPtr(limit),
	}

	var resp *generated.GetApiV1DatasetsProductsSearchResponse
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
func (c *Client) GetBulkProduct(ctx context.Context, productID string) (*generated.BdssResponseProductBag, error) {
	params := &generated.GetApiV1DatasetsProductsProductIdentifierParams{}
	var resp *generated.GetApiV1DatasetsProductsProductIdentifierResponse
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

// validateFileDownloadURI validates that the URL is a proper FileDownloadURI from the USPTO API
func (c *Client) validateFileDownloadURI(fileDownloadURI string) error {
	if fileDownloadURI == "" {
		return fmt.Errorf("fileDownloadURI cannot be empty")
	}

	expectedPrefix := "https://api.uspto.gov/api/v1/datasets/products/files/"
	if !strings.HasPrefix(fileDownloadURI, expectedPrefix) {
		return fmt.Errorf("invalid FileDownloadURI: must start with %s (got: %s)", expectedPrefix, fileDownloadURI)
	}

	return nil
}

// DownloadBulkFile downloads a file directly using the FileDownloadURI from the API response
func (c *Client) DownloadBulkFile(ctx context.Context, fileDownloadURI string, w io.Writer) error {
	return c.DownloadBulkFileWithProgress(ctx, fileDownloadURI, w, nil)
}

// DownloadBulkFileWithProgress downloads a file directly using FileDownloadURI with progress tracking
func (c *Client) DownloadBulkFileWithProgress(ctx context.Context, fileDownloadURI string, w io.Writer, progress func(bytesComplete int64, bytesTotal int64)) error {
	if err := c.validateFileDownloadURI(fileDownloadURI); err != nil {
		return err
	}

	// Create HTTP client that follows redirects (since we're using the direct URL)
	httpClient := &http.Client{
		Timeout: time.Duration(c.config.Timeout) * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fileDownloadURI, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	// Add authentication headers
	req.Header.Set("User-Agent", c.config.UserAgent)
	if c.config.APIKey != "" {
		req.Header.Set("X-API-Key", c.config.APIKey)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("downloading file: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	expectedSize := resp.ContentLength

	// Stream with progress tracking if callback provided
	var bytesWritten int64
	if progress != nil {
		// Create a pipe to track progress
		pr, pw := io.Pipe()

		// Track bytes in a goroutine
		go func() {
			defer pw.Close()
			var written int64
			buf := make([]byte, 32*1024) // 32KB buffer

			for {
				n, err := resp.Body.Read(buf)
				if n > 0 {
					written += int64(n)
					progress(written, expectedSize)
					if _, writeErr := pw.Write(buf[:n]); writeErr != nil {
						return
					}
				}
				if err != nil {
					return
				}
			}
		}()

		bytesWritten, err = io.Copy(w, pr)
	} else {
		bytesWritten, err = io.Copy(w, resp.Body)
	}
	if err != nil {
		return fmt.Errorf("writing file data: %w", err)
	}

	if expectedSize > 0 && bytesWritten != expectedSize {
		return fmt.Errorf("incomplete download: got %d bytes, expected %d", bytesWritten, expectedSize)
	}

	return nil
}

// SearchPetitions searches for petition decisions
func (c *Client) SearchPetitions(ctx context.Context, query string, offset, limit int32) (*generated.PetitionDecisionResponseBag, error) {
	req := generated.PetitionDecisionSearchRequest{
		Q: StringPtr(query),
		Pagination: &generated.Pagination{
			Offset: Int32Ptr(offset),
			Limit:  Int32Ptr(limit),
		},
	}

	var resp *generated.PostApiV1PetitionDecisionsSearchResponse
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

// GetPatentAssignment retrieves patent assignment data
func (c *Client) GetPatentAssignment(ctx context.Context, applicationNumber string) (any, error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextAssignmentResponse
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
func (c *Client) GetPatentAssociatedDocuments(ctx context.Context, applicationNumber string) (any, error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextAssociatedDocumentsResponse
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
func (c *Client) GetPatentAttorney(ctx context.Context, applicationNumber string) (any, error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextAttorneyResponse
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
func (c *Client) GetPatentForeignPriority(ctx context.Context, applicationNumber string) (any, error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextForeignPriorityResponse
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
func (c *Client) GetPatentMetaData(ctx context.Context, applicationNumber string) (any, error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextMetaDataResponse
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
func (c *Client) GetPatentTransactions(ctx context.Context, applicationNumber string) (any, error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextTransactionsResponse
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
func (c *Client) SearchPatentsDownload(ctx context.Context, req generated.PatentDownloadRequest) ([]byte, error) {
	var resp *generated.PostApiV1PatentApplicationsSearchDownloadResponse
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
func (c *Client) GetPetitionDecision(ctx context.Context, recordID string, includeDocuments bool) (*generated.PetitionDecisionIdentifierResponseBag, error) {
	params := &generated.GetApiV1PetitionDecisionsPetitionDecisionRecordIdentifierParams{
		IncludeDocuments: &includeDocuments,
	}
	var resp *generated.GetApiV1PetitionDecisionsPetitionDecisionRecordIdentifierResponse
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
func (c *Client) SearchPetitionsDownload(ctx context.Context, req generated.PetitionDecisionDownloadRequest) ([]byte, error) {
	var resp *generated.PostApiV1PetitionDecisionsSearchDownloadResponse
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
