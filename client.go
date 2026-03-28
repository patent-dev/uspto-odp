package odp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/patent-dev/uspto-odp/generated"
	oa "github.com/patent-dev/uspto-odp/generated/oa"
	tsdrgen "github.com/patent-dev/uspto-odp/generated/tsdr"
)

// Client is the main USPTO ODP API client
type Client struct {
	config     *Config
	httpClient *http.Client
	generated  *generated.ClientWithResponses
	oa         *oa.ClientWithResponses
	tsdr       *tsdrgen.ClientWithResponses
}

// Config holds client configuration.
// Note: Timeout applies to all APIs (ODP, OA, TSDR) uniformly. If TSDR document
// downloads need a longer timeout, create a separate Client with a higher Timeout.
type Config struct {
	BaseURL    string
	APIKey     string
	UserAgent  string
	MaxRetries int
	RetryDelay int // seconds
	Timeout    int // seconds

	// TSDR (Trademark Status & Document Retrieval) - separate server + API key
	TSDRBaseURL string // defaults to "https://tsdrapi.uspto.gov"
	TSDRAPIKey  string // from https://account.uspto.gov/profile/api-manager
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:    "https://api.uspto.gov",
		UserAgent:  "PatentDev/1.4",
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

	// Defensive copy to prevent mutation after construction
	cfg := *config
	config = &cfg

	httpClient := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	// Shared request editor for ODP + OA (same auth header)
	odpEditor := func(_ context.Context, req *http.Request) error {
		req.Header.Set("User-Agent", config.UserAgent)
		if config.APIKey != "" {
			req.Header.Set("X-API-Key", config.APIKey)
		}
		return nil
	}

	genClient, err := generated.NewClientWithResponses(
		config.BaseURL,
		generated.WithHTTPClient(httpClient),
		generated.WithRequestEditorFn(generated.RequestEditorFn(odpEditor)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	oaClient, err := oa.NewClientWithResponses(
		config.BaseURL,
		oa.WithHTTPClient(httpClient),
		oa.WithRequestEditorFn(oa.RequestEditorFn(odpEditor)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OA client: %w", err)
	}

	client := &Client{
		config:     config,
		httpClient: httpClient,
		generated:  genClient,
		oa:         oaClient,
	}

	// TSDR client (optional, only initialized if TSDRAPIKey is set)
	if config.TSDRAPIKey != "" {
		tsdrBaseURL := config.TSDRBaseURL
		if tsdrBaseURL == "" {
			tsdrBaseURL = "https://tsdrapi.uspto.gov"
		}

		tsdrEditor := tsdrgen.RequestEditorFn(func(_ context.Context, req *http.Request) error {
			req.Header.Set("User-Agent", config.UserAgent)
			req.Header.Set("USPTO-API-KEY", config.TSDRAPIKey)
			return nil
		})

		tsdrClient, err := tsdrgen.NewClientWithResponses(
			tsdrBaseURL,
			tsdrgen.WithHTTPClient(httpClient),
			tsdrgen.WithRequestEditorFn(tsdrEditor),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create TSDR client: %w", err)
		}
		client.tsdr = tsdrClient
	}

	return client, nil
}

// APIError represents an error returned by the USPTO API with status code
type APIError struct {
	StatusCode int
	Message    string
	Body       string // server response body for debugging
}

func (e *APIError) Error() string {
	return e.Message
}

// Detail returns the error message with the server response body, if available.
func (e *APIError) Detail() string {
	if e.Body != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Body)
	}
	return e.Message
}

// IsRetryable returns true for transient errors (429, 5xx) that should be retried
func (e *APIError) IsRetryable() bool {
	return e.StatusCode == http.StatusTooManyRequests || e.StatusCode >= 500
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.IsRetryable()
	}
	// Only retry on network-level transient errors (timeouts, connection resets)
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	// Connection refused, reset, etc.
	var opErr *net.OpError
	return errors.As(err, &opErr)
}

// checkStatusWithBody returns an APIError for non-2xx responses, including the response body for debugging.
func checkStatusWithBody(statusCode int, body []byte) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}
	apiErr := &APIError{
		StatusCode: statusCode,
		Message:    fmt.Sprintf("API returned status %d", statusCode),
	}
	if len(body) > 0 {
		apiErr.Body = truncatePreview(string(body), 512)
	}
	return apiErr
}

// truncatePreview returns s truncated to maxLen with "..." appended if truncated.
func truncatePreview(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func checkStatus(statusCode int) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}
	return &APIError{StatusCode: statusCode, Message: fmt.Sprintf("API returned status %d", statusCode)}
}

// retryableRequest wraps requests with retry logic, respecting context cancellation.
func (c *Client) retryableRequest(ctx context.Context, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err

		if !isRetryableError(err) {
			return err
		}

		if attempt < c.config.MaxRetries {
			// Exponential backoff: base * 2^attempt, with jitter
			base := float64(c.config.RetryDelay)
			delay := base * math.Pow(2, float64(attempt))
			jitter := delay * 0.25 * rand.Float64()
			wait := time.Duration(delay+jitter) * time.Second

			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return fmt.Errorf("request cancelled during retry: %w", ctx.Err())
			}
		}
	}
	return fmt.Errorf("failed after %d retries: %w", c.config.MaxRetries, lastErr)
}

// drainClose reads remaining body bytes (for HTTP connection reuse) and closes.
func drainClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
	_ = body.Close()
}

// progressReader wraps an io.Reader with a progress callback.
type progressReader struct {
	r       io.Reader
	written int64
	total   int64
	fn      func(bytesComplete, bytesTotal int64)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	if n > 0 {
		pr.written += int64(n)
		pr.fn(pr.written, pr.total)
	}
	return n, err
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
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentApplicationsSearchWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
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
	err = c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetPatentAdjustment retrieves patent term adjustment data.
func (c *Client) GetPatentAdjustment(ctx context.Context, applicationNumber string) (*AdjustmentResponse, error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextAdjustmentResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextAdjustmentWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	result := &AdjustmentResponse{ApplicationNumber: applicationNumber}
	if resp.JSON200 != nil && resp.JSON200.PatentFileWrapperDataBag != nil && len(*resp.JSON200.PatentFileWrapperDataBag) > 0 {
		bag := (*resp.JSON200.PatentFileWrapperDataBag)[0]
		if bag.PatentTermAdjustmentData != nil {
			pta := bag.PatentTermAdjustmentData
			result.TotalAdjustmentDays = derefInt(pta.AdjustmentTotalQuantity)
			result.ADelays = derefInt(pta.ADelayQuantity)
			result.BDelays = derefInt(pta.BDelayQuantity)
			result.CDelays = derefInt(pta.CDelayQuantity)
		}
	}
	return result, nil
}

// GetPatentContinuity retrieves patent continuity data.
func (c *Client) GetPatentContinuity(ctx context.Context, applicationNumber string) (*ContinuityResponse, error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextContinuityResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextContinuityWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	result := &ContinuityResponse{
		ApplicationNumber: applicationNumber,
		Parents:           []ContinuityParent{},
		Children:          []ContinuityChild{},
	}
	if resp.JSON200 != nil && resp.JSON200.PatentFileWrapperDataBag != nil && len(*resp.JSON200.PatentFileWrapperDataBag) > 0 {
		bag := (*resp.JSON200.PatentFileWrapperDataBag)[0]
		if bag.ParentContinuityBag != nil {
			for _, p := range *bag.ParentContinuityBag {
				result.Parents = append(result.Parents, ContinuityParent{
					ApplicationNumber: derefStr(p.ParentApplicationNumberText),
					PatentNumber:      derefStr(p.ParentPatentNumber),
					FilingDate:        derefStr(p.ParentApplicationFilingDate),
					Status:            derefStr(p.ParentApplicationStatusDescriptionText),
					RelationshipType:  mapRelationshipType(derefStr(p.ClaimParentageTypeCode), derefStr(p.ClaimParentageTypeCodeDescriptionText)),
				})
			}
		}
		if bag.ChildContinuityBag != nil {
			for _, ch := range *bag.ChildContinuityBag {
				result.Children = append(result.Children, ContinuityChild{
					ApplicationNumber: derefStr(ch.ChildApplicationNumberText),
					PatentNumber:      derefStr(ch.ChildPatentNumber),
					FilingDate:        derefStr(ch.ChildApplicationFilingDate),
					Status:            derefStr(ch.ChildApplicationStatusDescriptionText),
					RelationshipType:  mapRelationshipType(derefStr(ch.ClaimParentageTypeCode), derefStr(ch.ClaimParentageTypeCodeDescriptionText)),
				})
			}
		}
	}
	return result, nil
}

// GetPatentDocuments retrieves patent documents list
func (c *Client) GetPatentDocuments(ctx context.Context, applicationNumber string) (*generated.DocumentBag, error) {
	params := &generated.GetApiV1PatentApplicationsApplicationNumberTextDocumentsParams{}
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextDocumentsResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextDocumentsWithResponse(ctx, applicationNumber, params)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
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
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentStatusCodesWithResponse(ctx, params)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
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
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1DatasetsProductsSearchWithResponse(ctx, params)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
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
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1DatasetsProductsProductIdentifierWithResponse(ctx, productID, params)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
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

	expectedPrefix := c.config.BaseURL + "/api/v1/datasets/products/files/"
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

	req, err := http.NewRequestWithContext(ctx, "GET", fileDownloadURI, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	// Add authentication headers
	req.Header.Set("User-Agent", c.config.UserAgent)
	if c.config.APIKey != "" {
		req.Header.Set("X-API-Key", c.config.APIKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("downloading file: %w", err)
	}
	defer drainClose(resp.Body)

	if err := checkStatus(resp.StatusCode); err != nil {
		return err
	}

	expectedSize := resp.ContentLength

	// Wrap with progress tracking if callback provided
	var src io.Reader = resp.Body
	if progress != nil {
		src = &progressReader{r: resp.Body, total: expectedSize, fn: progress}
	}

	bytesWritten, err := io.Copy(w, src)
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
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PetitionDecisionsSearchWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetPatentAssignment retrieves patent assignment data.
func (c *Client) GetPatentAssignment(ctx context.Context, applicationNumber string) (*AssignmentResponse, error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextAssignmentResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextAssignmentWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	result := &AssignmentResponse{
		ApplicationNumber: applicationNumber,
		Assignments:       []AssignmentEntry{},
	}
	if resp.JSON200 != nil && resp.JSON200.PatentFileWrapperDataBag != nil && len(*resp.JSON200.PatentFileWrapperDataBag) > 0 {
		bag := (*resp.JSON200.PatentFileWrapperDataBag)[0]
		if bag.AssignmentBag != nil {
			for _, a := range *bag.AssignmentBag {
				entry := AssignmentEntry{
					RecordedDate: derefStr(a.AssignmentRecordedDate),
					Conveyance:   derefStr(a.ConveyanceText),
					ReelFrame:    derefStr(a.ReelAndFrameNumber),
				}
				// Join multiple assignors
				if a.AssignorBag != nil {
					var names []string
					for _, assignor := range *a.AssignorBag {
						if name := derefStr(assignor.AssignorName); name != "" {
							names = append(names, name)
						}
						if entry.ExecutionDate == "" {
							entry.ExecutionDate = derefStr(assignor.ExecutionDate)
						}
					}
					entry.Assignor = strings.Join(names, ", ")
				}
				// Join multiple assignees
				if a.AssigneeBag != nil {
					var names []string
					for _, assignee := range *a.AssigneeBag {
						if name := derefStr(assignee.AssigneeNameText); name != "" {
							names = append(names, name)
						}
					}
					entry.Assignee = strings.Join(names, ", ")
				}
				result.Assignments = append(result.Assignments, entry)
			}
		}
	}
	return result, nil
}

// GetPatentAssociatedDocuments retrieves patent grant and publication XML file metadata.
func (c *Client) GetPatentAssociatedDocuments(ctx context.Context, applicationNumber string) (*AssociatedDocumentsResponse, error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextAssociatedDocumentsResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextAssociatedDocumentsWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	result := &AssociatedDocumentsResponse{ApplicationNumber: applicationNumber}
	if resp.JSON200 != nil && resp.JSON200.PatentFileWrapperDataBag != nil && len(*resp.JSON200.PatentFileWrapperDataBag) > 0 {
		bag := (*resp.JSON200.PatentFileWrapperDataBag)[0]
		result.GrantDocumentMetaData = bag.GrantDocumentMetaData
		result.PgpubDocumentMetaData = bag.PgpubDocumentMetaData
	}
	return result, nil
}

// GetPatentAttorney retrieves patent attorney information.
func (c *Client) GetPatentAttorney(ctx context.Context, applicationNumber string) (*generated.RecordAttorney, error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextAttorneyResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextAttorneyWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	if resp.JSON200 != nil && resp.JSON200.PatentFileWrapperDataBag != nil && len(*resp.JSON200.PatentFileWrapperDataBag) > 0 {
		return (*resp.JSON200.PatentFileWrapperDataBag)[0].RecordAttorney, nil
	}
	return nil, nil
}

// GetPatentForeignPriority retrieves foreign priority data.
func (c *Client) GetPatentForeignPriority(ctx context.Context, applicationNumber string) ([]generated.ForeignPriority, error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextForeignPriorityResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextForeignPriorityWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	if resp.JSON200 != nil && resp.JSON200.PatentFileWrapperDataBag != nil && len(*resp.JSON200.PatentFileWrapperDataBag) > 0 {
		bag := (*resp.JSON200.PatentFileWrapperDataBag)[0]
		if bag.ForeignPriorityBag != nil {
			return *bag.ForeignPriorityBag, nil
		}
	}
	return nil, nil
}

// GetPatentMetaData retrieves patent metadata (status, filing date, examiner, classification, etc.).
func (c *Client) GetPatentMetaData(ctx context.Context, applicationNumber string) (*generated.ApplicationMetaData, error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextMetaDataResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextMetaDataWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	if resp.JSON200 != nil && resp.JSON200.PatentFileWrapperDataBag != nil && len(*resp.JSON200.PatentFileWrapperDataBag) > 0 {
		return (*resp.JSON200.PatentFileWrapperDataBag)[0].ApplicationMetaData, nil
	}
	return nil, nil
}

// GetPatentTransactions retrieves patent transaction history.
func (c *Client) GetPatentTransactions(ctx context.Context, applicationNumber string) (*TransactionsResponse, error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextTransactionsResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextTransactionsWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	result := &TransactionsResponse{
		ApplicationNumber: applicationNumber,
		Events:            []TransactionEvent{},
	}
	if resp.JSON200 != nil && resp.JSON200.PatentFileWrapperDataBag != nil && len(*resp.JSON200.PatentFileWrapperDataBag) > 0 {
		bag := (*resp.JSON200.PatentFileWrapperDataBag)[0]
		if bag.EventDataBag != nil {
			for _, e := range *bag.EventDataBag {
				result.Events = append(result.Events, TransactionEvent{
					Date:        derefStr(e.EventDate),
					Code:        derefStr(e.EventCode),
					Description: derefStr(e.EventDescriptionText),
				})
			}
		}
	}
	return result, nil
}

// SearchPatentsDownload downloads patent search results
func (c *Client) SearchPatentsDownload(ctx context.Context, req generated.PatentDownloadRequest) ([]byte, error) {
	var resp *generated.PostApiV1PatentApplicationsSearchDownloadResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentApplicationsSearchDownloadWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
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
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PetitionDecisionsPetitionDecisionRecordIdentifierWithResponse(ctx, recordID, params)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
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
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PetitionDecisionsSearchDownloadWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
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

// ==============================================================================
// PTAB (Patent Trial and Appeal Board) Methods
// ==============================================================================

// SearchTrialProceedings searches PTAB trial proceedings
func (c *Client) SearchTrialProceedings(ctx context.Context, query string, offset, limit int32) (*generated.ProceedingDataResponse, error) {
	req := generated.PostApiV1PatentTrialsProceedingsSearchJSONRequestBody{
		Q: StringPtr(query),
		Pagination: &generated.Pagination{
			Offset: Int32Ptr(offset),
			Limit:  Int32Ptr(limit),
		},
	}

	var resp *generated.PostApiV1PatentTrialsProceedingsSearchResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentTrialsProceedingsSearchWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetTrialProceeding retrieves a specific PTAB trial proceeding by trial number
func (c *Client) GetTrialProceeding(ctx context.Context, trialNumber string) (*generated.ProceedingDataResponse, error) {
	var resp *generated.GetApiV1PatentTrialsProceedingsTrialNumberResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentTrialsProceedingsTrialNumberWithResponse(ctx, trialNumber)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// SearchTrialDecisions searches PTAB trial decisions
func (c *Client) SearchTrialDecisions(ctx context.Context, query string, offset, limit int32) (*generated.DecisionDataResponse, error) {
	req := generated.PostApiV1PatentTrialsDecisionsSearchJSONRequestBody{
		Q: StringPtr(query),
		Pagination: &generated.Pagination{
			Offset: Int32Ptr(offset),
			Limit:  Int32Ptr(limit),
		},
	}

	var resp *generated.PostApiV1PatentTrialsDecisionsSearchResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentTrialsDecisionsSearchWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetTrialDecision retrieves a specific PTAB trial decision by document identifier
func (c *Client) GetTrialDecision(ctx context.Context, documentIdentifier string) (*generated.DecisionDataResponse, error) {
	var resp *generated.GetApiV1PatentTrialsDecisionsDocumentIdentifierResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentTrialsDecisionsDocumentIdentifierWithResponse(ctx, documentIdentifier)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// SearchTrialDocuments searches PTAB trial documents
func (c *Client) SearchTrialDocuments(ctx context.Context, query string, offset, limit int32) (*generated.DocumentDataResponse, error) {
	req := generated.PostApiV1PatentTrialsDocumentsSearchJSONRequestBody{
		Q: StringPtr(query),
		Pagination: &generated.Pagination{
			Offset: Int32Ptr(offset),
			Limit:  Int32Ptr(limit),
		},
	}

	var resp *generated.PostApiV1PatentTrialsDocumentsSearchResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentTrialsDocumentsSearchWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetTrialDocument retrieves a specific PTAB trial document by document identifier
func (c *Client) GetTrialDocument(ctx context.Context, documentIdentifier string) (*generated.DocumentDataResponse, error) {
	var resp *generated.GetApiV1PatentTrialsDocumentsDocumentIdentifierResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentTrialsDocumentsDocumentIdentifierWithResponse(ctx, documentIdentifier)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// SearchAppealDecisions searches PTAB appeal decisions
func (c *Client) SearchAppealDecisions(ctx context.Context, query string, offset, limit int32) (*generated.AppealDecisionDataResponse, error) {
	req := generated.PostApiV1PatentAppealsDecisionsSearchJSONRequestBody{
		Q: StringPtr(query),
		Pagination: &generated.Pagination{
			Offset: Int32Ptr(offset),
			Limit:  Int32Ptr(limit),
		},
	}

	var resp *generated.PostApiV1PatentAppealsDecisionsSearchResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentAppealsDecisionsSearchWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetAppealDecision retrieves a specific PTAB appeal decision by document identifier
func (c *Client) GetAppealDecision(ctx context.Context, documentIdentifier string) (*generated.AppealDecisionDataResponse, error) {
	var resp *generated.GetApiV1PatentAppealsDecisionsDocumentIdentifierResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentAppealsDecisionsDocumentIdentifierWithResponse(ctx, documentIdentifier)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetAppealDecisionsByAppealNumber retrieves all decisions for a specific appeal number
func (c *Client) GetAppealDecisionsByAppealNumber(ctx context.Context, appealNumber string) (*generated.AppealDecisionDataResponse, error) {
	var resp *generated.GetApiV1PatentAppealsAppealNumberDecisionsResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentAppealsAppealNumberDecisionsWithResponse(ctx, appealNumber)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// SearchInterferenceDecisions searches PTAB interference decisions
func (c *Client) SearchInterferenceDecisions(ctx context.Context, query string, offset, limit int32) (*generated.InterferenceDecisionDataResponse, error) {
	req := generated.PostApiV1PatentInterferencesDecisionsSearchJSONRequestBody{
		Q: StringPtr(query),
		Pagination: &generated.Pagination{
			Offset: Int32Ptr(offset),
			Limit:  Int32Ptr(limit),
		},
	}

	var resp *generated.PostApiV1PatentInterferencesDecisionsSearchResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentInterferencesDecisionsSearchWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetInterferenceDecision retrieves a specific PTAB interference decision by document identifier
func (c *Client) GetInterferenceDecision(ctx context.Context, documentIdentifier string) (*generated.InterferenceDecisionDataResponse, error) {
	var resp *generated.GetApiV1PatentInterferencesDecisionsDocumentIdentifierResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentInterferencesDecisionsDocumentIdentifierWithResponse(ctx, documentIdentifier)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetInterferenceDecisionsByNumber retrieves all decisions for a specific interference number
func (c *Client) GetInterferenceDecisionsByNumber(ctx context.Context, interferenceNumber string) (*generated.InterferenceDecisionDataResponse, error) {
	var resp *generated.GetApiV1PatentInterferencesInterferenceNumberDecisionsResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentInterferencesInterferenceNumberDecisionsWithResponse(ctx, interferenceNumber)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetTrialDecisionsByTrialNumber retrieves all decisions for a specific trial number
func (c *Client) GetTrialDecisionsByTrialNumber(ctx context.Context, trialNumber string) (*generated.DecisionDataResponse, error) {
	var resp *generated.GetApiV1PatentTrialsTrialNumberDecisionsResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentTrialsTrialNumberDecisionsWithResponse(ctx, trialNumber)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetTrialDocumentsByTrialNumber retrieves all documents for a specific trial number
func (c *Client) GetTrialDocumentsByTrialNumber(ctx context.Context, trialNumber string) (*generated.DocumentDataResponse, error) {
	var resp *generated.GetApiV1PatentTrialsTrialNumberDocumentsResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentTrialsTrialNumberDocumentsWithResponse(ctx, trialNumber)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// SearchTrialProceedingsDownload downloads trial proceedings search results
func (c *Client) SearchTrialProceedingsDownload(ctx context.Context, req generated.DownloadRequest) ([]byte, error) {
	var resp *generated.PostApiV1PatentTrialsProceedingsSearchDownloadResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentTrialsProceedingsSearchDownloadWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// SearchTrialDecisionsDownload downloads trial decisions search results
func (c *Client) SearchTrialDecisionsDownload(ctx context.Context, req generated.DownloadRequest) ([]byte, error) {
	var resp *generated.PostApiV1PatentTrialsDecisionsSearchDownloadResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentTrialsDecisionsSearchDownloadWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// SearchTrialDocumentsDownload downloads trial documents search results
func (c *Client) SearchTrialDocumentsDownload(ctx context.Context, req generated.DownloadRequest) ([]byte, error) {
	var resp *generated.PostApiV1PatentTrialsDocumentsSearchDownloadResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentTrialsDocumentsSearchDownloadWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// SearchAppealDecisionsDownload downloads appeal decisions search results
func (c *Client) SearchAppealDecisionsDownload(ctx context.Context, req generated.DownloadRequest) ([]byte, error) {
	var resp *generated.PostApiV1PatentAppealsDecisionsSearchDownloadResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentAppealsDecisionsSearchDownloadWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// SearchInterferenceDecisionsDownload downloads interference decisions search results
func (c *Client) SearchInterferenceDecisionsDownload(ctx context.Context, req generated.PatentDownloadRequest) ([]byte, error) {
	var resp *generated.PostApiV1PatentInterferencesDecisionsSearchDownloadResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentInterferencesDecisionsSearchDownloadWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkStatusWithBody(resp.StatusCode(), resp.Body); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
