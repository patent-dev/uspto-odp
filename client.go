// Package odp provides a Go client for the USPTO Open Data Portal (ODP),
// covering patent and trademark application data, the Office Action APIs, and
// TSDR (Trademark Status & Document Retrieval).
package odp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"github.com/patent-dev/uspto-odp/generated"
	oa "github.com/patent-dev/uspto-odp/generated/oa"
	tsdrgen "github.com/patent-dev/uspto-odp/generated/tsdr"
)

// Version is the library version. Bumped per release; surfaces through the
// default User-Agent.
const Version = "1.5.0"

// DefaultUserAgent identifies this library in outbound requests. The
// product token is the library name so the request is grepable in USPTO
// logs by either the library or the project that maintains it. Consumers
// are encouraged to prepend their own identity, e.g.
// "MyApp/2.3 uspto-odp/1.5".
const DefaultUserAgent = "uspto-odp/" + Version + " (patent.dev; +https://github.com/patent-dev/uspto-odp)"

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
	RetryDelay time.Duration // base backoff between retries
	Timeout    time.Duration // request timeout for the underlying http.Client

	// MaxRetryAfter is the longest Retry-After the client will honor. If the
	// server requests a longer wait, the resulting *APIError reports
	// IsRetryable=false so the caller can decide. Zero means "use the
	// DefaultMaxRetryAfter constant".
	MaxRetryAfter time.Duration

	// OABaseURL is the host serving the Office Action APIs. Defaults to
	// the ODP host (https://api.uspto.gov); override to point elsewhere.
	OABaseURL string

	// TSDR (Trademark Status & Document Retrieval) - separate server + API key
	TSDRBaseURL string // defaults to "https://tsdrapi.uspto.gov"
	TSDRAPIKey  string // from https://account.uspto.gov/profile/api-manager
}

// DefaultOABaseURL is the ODP host serving the Office Action APIs.
const DefaultOABaseURL = "https://api.uspto.gov"

// DefaultMaxRetryAfter is the cap applied when Config.MaxRetryAfter is zero.
const DefaultMaxRetryAfter = 60 * time.Second

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:       "https://api.uspto.gov",
		UserAgent:     DefaultUserAgent,
		MaxRetries:    3,
		RetryDelay:    1 * time.Second,
		Timeout:       30 * time.Second,
		MaxRetryAfter: DefaultMaxRetryAfter,
		OABaseURL:     DefaultOABaseURL,
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
		Timeout: config.Timeout,
	}

	// ODP and the OA APIs both authenticate with X-API-Key on api.uspto.gov.
	odpEditor := func(_ context.Context, req *http.Request) error {
		req.Header.Set("User-Agent", config.UserAgent)
		if config.APIKey != "" {
			req.Header.Set("X-API-Key", config.APIKey)
		}
		return nil
	}
	oaEditor := func(_ context.Context, req *http.Request) error {
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

	oaBaseURL := config.OABaseURL
	if oaBaseURL == "" {
		oaBaseURL = DefaultOABaseURL
	}
	oaClient, err := oa.NewClientWithResponses(
		oaBaseURL,
		oa.WithHTTPClient(httpClient),
		oa.WithRequestEditorFn(oa.RequestEditorFn(oaEditor)),
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

// maxRetryAfter returns the configured Retry-After cap, falling back to the
// default for zero (the un-set value).
func (c *Client) maxRetryAfter() time.Duration {
	if c.config.MaxRetryAfter > 0 {
		return c.config.MaxRetryAfter
	}
	return DefaultMaxRetryAfter
}

// headerOf returns the Header of a possibly-nil *http.Response.
func headerOf(r *http.Response) http.Header {
	if r == nil {
		return nil
	}
	return r.Header
}

// validatePagination returns an error if offset or limit fall outside the
// non-negative int32 range expected by the upstream API. This sits at the
// boundary so callers using the wider int signature don't silently truncate
// when crossing into negative or 32-bit-overflow territory.
func validatePagination(offset, limit int) error {
	if offset < 0 {
		return fmt.Errorf("offset must be >= 0, got %d", offset)
	}
	if limit < 0 {
		return fmt.Errorf("limit must be >= 0, got %d", limit)
	}
	if offset > math.MaxInt32 {
		return fmt.Errorf("offset must fit in int32, got %d", offset)
	}
	if limit > math.MaxInt32 {
		return fmt.Errorf("limit must fit in int32, got %d", limit)
	}
	return nil
}

// truncatePreview returns s truncated to maxLen with "..." appended if truncated.
func truncatePreview(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
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

		// If the server requested a wait longer than the configured cap,
		// surface the error so the caller can decide.
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.RetryAfter > c.maxRetryAfter() {
			return err
		}

		if attempt < c.config.MaxRetries {
			// If the server told us to wait via Retry-After, honor that;
			// otherwise fall back to exponential backoff with jitter.
			wait := time.Duration(0)
			if apiErr != nil && apiErr.RetryAfter > 0 {
				wait = apiErr.RetryAfter
			} else {
				// RetryDelay is a time.Duration (nanoseconds under the hood);
				// the float64 round-trip stays in nanos, so the final
				// time.Duration cast carries the right unit.
				base := float64(c.config.RetryDelay)
				delay := base * math.Pow(2, float64(attempt))
				jitter := delay * 0.25 * rand.Float64()
				wait = time.Duration(delay + jitter)
			}

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

// SearchPatents searches for patent applications. It is the simple form of
// SearchPatentsWithOptions (query + pagination, no sort or field projection).
func (c *Client) SearchPatents(ctx context.Context, query string, offset, limit int) (*generated.PatentDataResponse, error) {
	return c.SearchPatentsWithOptions(ctx, query, offset, limit, nil)
}

// PatentSearchSort is one sort key for a patent search: a document field and an
// optional order ("Asc"/"Desc", case-insensitive; empty applies the API default).
type PatentSearchSort struct {
	Field string
	Order string
}

// PatentSearchFilter is a server-side term filter: a document field constrained
// to one or more exact values (OR-ed within the field, AND-ed across filters).
type PatentSearchFilter struct {
	Field  string
	Values []string
}

// PatentSearchRange is a server-side range filter on a date/number field. An
// empty From or To leaves that bound open.
type PatentSearchRange struct {
	Field string
	From  string
	To    string
}

// PatentSearchOptions carries the optional refinements for
// SearchPatentsWithOptions on top of the query and pagination: sort keys, a
// response field projection, and the server-side term/range filters the ODP
// search endpoint supports (faceted filtering and date/number windows).
type PatentSearchOptions struct {
	Sort         []PatentSearchSort
	Fields       []string
	Filters      []PatentSearchFilter
	RangeFilters []PatentSearchRange
}

// SearchPatentsWithOptions searches for patent applications with optional sort
// and field-projection refinements. A nil opts behaves exactly like SearchPatents.
func (c *Client) SearchPatentsWithOptions(ctx context.Context, query string, offset, limit int, opts *PatentSearchOptions) (*generated.PatentDataResponse, error) {
	if err := validatePagination(offset, limit); err != nil {
		return nil, err
	}
	req := generated.PatentSearchRequest{
		Q: StringPtr(query),
		Pagination: &generated.Pagination{
			Offset: Int32Ptr(int32(offset)),
			Limit:  Int32Ptr(int32(limit)),
		},
	}
	if opts != nil {
		if sort := buildSearchSort(opts.Sort); len(sort) > 0 {
			req.Sort = &sort
		}
		if len(opts.Fields) > 0 {
			fields := append([]string(nil), opts.Fields...)
			req.Fields = &fields
		}
		if filters := buildSearchFilters(opts.Filters); len(filters) > 0 {
			req.Filters = &filters
		}
		if ranges := buildSearchRanges(opts.RangeFilters); len(ranges) > 0 {
			req.RangeFilters = &ranges
		}
	}

	var resp *generated.PostApiV1PatentApplicationsSearchResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentApplicationsSearchWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// buildSearchSort maps the public sort keys onto the generated Sort entries,
// skipping keys with no field. An empty order is left unset so the API applies
// its default; otherwise it is normalized to the API's "Asc"/"Desc" spelling.
func buildSearchSort(keys []PatentSearchSort) []generated.Sort {
	out := make([]generated.Sort, 0, len(keys))
	for _, k := range keys {
		if k.Field == "" {
			continue
		}
		s := generated.Sort{Field: StringPtr(k.Field)}
		switch strings.ToLower(strings.TrimSpace(k.Order)) {
		case "asc":
			order := generated.SortOrderAsc
			s.Order = &order
		case "desc":
			order := generated.SortOrderDesc
			s.Order = &order
		}
		out = append(out, s)
	}
	return out
}

// buildSearchFilters maps the public term filters onto the generated Filter
// entries, skipping filters with no field or no values.
func buildSearchFilters(filters []PatentSearchFilter) []generated.Filter {
	out := make([]generated.Filter, 0, len(filters))
	for _, f := range filters {
		if f.Field == "" || len(f.Values) == 0 {
			continue
		}
		values := append([]string(nil), f.Values...)
		out = append(out, generated.Filter{Name: StringPtr(f.Field), Value: &values})
	}
	return out
}

// buildSearchRanges maps the public range filters onto the generated Range
// entries, skipping ranges with no field or no bound at all.
func buildSearchRanges(ranges []PatentSearchRange) []generated.Range {
	out := make([]generated.Range, 0, len(ranges))
	for _, r := range ranges {
		if r.Field == "" || (r.From == "" && r.To == "") {
			continue
		}
		entry := generated.Range{Field: StringPtr(r.Field)}
		if r.From != "" {
			entry.ValueFrom = StringPtr(r.From)
		}
		if r.To != "" {
			entry.ValueTo = StringPtr(r.To)
		}
		out = append(out, entry)
	}
	return out
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

// resolvePublicationToApplicationNumber searches for a publication number and returns its application number.
// kindCode is the publication kind suffix when supplied by the caller (e.g., "A1", "A2", "A9");
// empty string defaults to "A1".
func (c *Client) resolvePublicationToApplicationNumber(ctx context.Context, publicationNumber, kindCode string) (string, error) {
	if kindCode == "" {
		kindCode = "A1"
	}
	// Format publication number for search (e.g., 20250087686 -> US20250087686A1)
	formattedPub := publicationNumber
	if len(publicationNumber) == 11 && !strings.HasPrefix(publicationNumber, "US") {
		formattedPub = "US" + publicationNumber + kindCode
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

// ResolvePatentNumber resolves any patent number format (application, grant, or
// publication) to its application number, searching the USPTO API when necessary.
// A bare 8-digit number with no kind code is ambiguous (grant vs application); this
// entry point probes both interpretations and returns an *AmbiguousPatentNumberError
// when they differ, so a caller acting on user input can surface a "did you mean"
// choice. Call this with raw user input. To resolve a number you already know to be an
// application number, see resolveApplicationNumberLenient.
func (c *Client) ResolvePatentNumber(ctx context.Context, patentNumber string) (string, error) {
	pn, err := NormalizePatentNumber(patentNumber)
	if err != nil {
		return "", fmt.Errorf("invalid patent number: %w", err)
	}
	if pn.Type == PatentNumberTypeApplication && pn.Ambiguous {
		return c.resolveAmbiguousNumber(ctx, pn.Normalized)
	}
	return c.resolveNormalized(ctx, pn)
}

// resolveApplicationNumberLenient resolves a patent number without ambiguity probing:
// a bare 8-digit value is taken at face value as an application number. This is the
// correct resolver for numbers that are already known to be application numbers - for
// example one produced by a prior grant lookup - where probing for a same-digit grant
// would wrongly flag the application as ambiguous. Grant and publication inputs are
// still resolved normally.
func (c *Client) resolveApplicationNumberLenient(ctx context.Context, patentNumber string) (string, error) {
	pn, err := NormalizePatentNumber(patentNumber)
	if err != nil {
		return "", fmt.Errorf("invalid patent number: %w", err)
	}
	return c.resolveNormalized(ctx, pn)
}

// resolveNormalized maps a parsed patent number to its application number without any
// ambiguity probing.
func (c *Client) resolveNormalized(ctx context.Context, pn *PatentNumber) (string, error) {
	switch pn.Type {
	case PatentNumberTypeGrant:
		return c.resolveGrantToApplicationNumber(ctx, pn.Normalized)
	case PatentNumberTypePublication:
		return c.resolvePublicationToApplicationNumber(ctx, pn.Normalized, pn.KindCode)
	case PatentNumberTypeApplication, PatentNumberTypePCT:
		// PCT numbers (15-char or 12-char legacy) are accepted directly as the
		// application path parameter; no round-trip needed.
		return pn.ToApplicationNumber(), nil
	default:
		return "", fmt.Errorf("unknown patent number type")
	}
}

// PatentCandidate describes one interpretation of an ambiguous bare patent number,
// resolved far enough to show the caller what it points at.
type PatentCandidate struct {
	Type              PatentNumberType // PatentNumberTypeGrant or PatentNumberTypeApplication
	Number            string           // the bare digits as entered
	ApplicationNumber string           // application number this interpretation resolves to
	Title             string           // invention title, for a "did you mean" prompt
}

// AmbiguousPatentNumberError is returned when a bare number resolves to both a grant
// and a different application. It carries the candidates so the caller can ask the user
// which one they meant instead of guessing.
type AmbiguousPatentNumberError struct {
	Input      string
	Candidates []PatentCandidate
}

func (e *AmbiguousPatentNumberError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "patent number %q is ambiguous (matches both a grant and an application); did you mean", e.Input)
	for i, c := range e.Candidates {
		sep := ";"
		if i == 0 {
			sep = ""
		}
		kind := "application"
		if c.Type == PatentNumberTypeGrant {
			kind = "grant"
		}
		fmt.Fprintf(&b, "%s %s %s (application %s) %q", sep, kind, c.Number, c.ApplicationNumber, c.Title)
	}
	b.WriteString("? Specify a kind code (e.g. B2) for the grant or a slashed application number.")
	return b.String()
}

// resolveAmbiguousNumber probes both interpretations of a bare 8-digit number:
// the grant search and the direct application lookup. If both resolve to different
// applications it returns an *AmbiguousPatentNumberError; otherwise it returns the one
// that exists (fixing both the spurious 404 and the silent wrong-patent resolution).
func (c *Client) resolveAmbiguousNumber(ctx context.Context, digits string) (string, error) {
	grantApp, grantTitle, grantFound, err := c.findGrantCandidate(ctx, digits)
	if err != nil {
		return "", err
	}
	appTitle, appFound, err := c.findApplicationCandidate(ctx, digits)
	if err != nil {
		return "", err
	}

	switch {
	case grantFound && appFound && grantApp != digits:
		// Both interpretations exist and point at different applications.
		return "", &AmbiguousPatentNumberError{
			Input: digits,
			Candidates: []PatentCandidate{
				{Type: PatentNumberTypeGrant, Number: digits, ApplicationNumber: grantApp, Title: grantTitle},
				{Type: PatentNumberTypeApplication, Number: digits, ApplicationNumber: digits, Title: appTitle},
			},
		}
	case grantFound:
		return grantApp, nil
	case appFound:
		return digits, nil
	default:
		return "", fmt.Errorf("no grant or application found for number %s", digits)
	}
}

// findGrantCandidate searches for a grant with the given number and returns the
// application it maps to plus its title. found is false (with nil error) when no grant
// matches.
func (c *Client) findGrantCandidate(ctx context.Context, grantNumber string) (appNumber, title string, found bool, err error) {
	query := fmt.Sprintf("applicationMetaData.patentNumber:%s", grantNumber)
	result, err := c.SearchPatents(ctx, query, 0, 1)
	if err != nil {
		if isNotFoundErr(err) {
			return "", "", false, nil
		}
		return "", "", false, fmt.Errorf("failed to search for grant number %s: %w", grantNumber, err)
	}
	if result == nil || result.PatentFileWrapperDataBag == nil || len(*result.PatentFileWrapperDataBag) == 0 {
		return "", "", false, nil
	}
	w := (*result.PatentFileWrapperDataBag)[0]
	if w.ApplicationNumberText == nil {
		return "", "", false, nil
	}
	return *w.ApplicationNumberText, titleOf(w.ApplicationMetaData), true, nil
}

// findApplicationCandidate looks up the number directly as an application number and
// returns its title. found is false (with nil error) when no such application exists.
// It calls the generated endpoint directly to avoid recursing back into resolution.
func (c *Client) findApplicationCandidate(ctx context.Context, appNumber string) (title string, found bool, err error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextResponse
	err = c.retryableRequest(ctx, func() error {
		var reqErr error
		resp, reqErr = c.generated.GetApiV1PatentApplicationsApplicationNumberTextWithResponse(ctx, appNumber)
		if reqErr != nil {
			return reqErr
		}
		if resp.StatusCode() == http.StatusNotFound {
			return nil // not found is a definitive answer, not a transient failure
		}
		return checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse))
	})
	if err != nil {
		return "", false, err
	}
	if resp.StatusCode() == http.StatusNotFound || resp.JSON200 == nil {
		return "", false, nil
	}
	if resp.JSON200.PatentFileWrapperDataBag == nil || len(*resp.JSON200.PatentFileWrapperDataBag) == 0 {
		return "", false, nil
	}
	return titleOf((*resp.JSON200.PatentFileWrapperDataBag)[0].ApplicationMetaData), true, nil
}

// titleOf pulls the invention title from application metadata, if present.
func titleOf(meta *generated.ApplicationMetaData) string {
	if meta != nil && meta.InventionTitle != nil {
		return *meta.InventionTitle
	}
	return ""
}

// isNotFoundErr reports whether err is an APIError with HTTP 404 (used to treat an
// empty ODP search, which returns 404, as "no match" rather than a hard failure).
func isNotFoundErr(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// GetPatent retrieves patent data by application, grant, or publication number.
// It resolves leniently (no grant-vs-application ambiguity probing) because it is
// commonly called with an already-resolved application number; callers that need to
// disambiguate raw user input should call ResolvePatentNumber first.
func (c *Client) GetPatent(ctx context.Context, patentNumber string) (*generated.PatentDataResponse, error) {
	applicationNumber, err := c.resolveApplicationNumberLenient(ctx, patentNumber)
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
	if err := validatePagination(offset, limit); err != nil {
		return nil, err
	}
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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

// DownloadBulkFileWithProgress downloads a bulk dataset file using its
// FileDownloadURI, reporting progress through the optional callback. See
// streamDownload for the retry and mid-stream-failure semantics.
func (c *Client) DownloadBulkFileWithProgress(ctx context.Context, fileDownloadURI string, w io.Writer, progress func(bytesComplete int64, bytesTotal int64)) error {
	if err := c.validateFileDownloadURI(fileDownloadURI); err != nil {
		return err
	}
	return c.streamDownload(ctx, fileDownloadURI, w, progress)
}

// streamDownload performs an authenticated streaming GET of uri into w.
//
// Retry behavior: the connection-setup phase (request creation, transport
// errors, non-2xx status) goes through retryableRequest with full backoff
// and Retry-After honoring. Mid-stream errors (connection reset after the
// 200 response started flowing) propagate without retry -- restarting from
// zero would silently overwrite however many bytes the caller already
// committed to its writer. URI validation is the caller's responsibility.
func (c *Client) streamDownload(ctx context.Context, uri string, w io.Writer, progress func(bytesComplete int64, bytesTotal int64)) error {
	var resp *http.Response
	err := c.retryableRequest(ctx, func() error {
		// Discard any prior attempt's response before retrying.
		if resp != nil {
			drainClose(resp.Body)
			resp = nil
		}
		req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("User-Agent", c.config.UserAgent)
		if c.config.APIKey != "" {
			req.Header.Set("X-API-Key", c.config.APIKey)
		}
		r, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		if r.StatusCode < 200 || r.StatusCode >= 300 {
			// Read a bounded prefix of the error body for the APIError.
			body, _ := io.ReadAll(io.LimitReader(r.Body, 4096))
			drainClose(r.Body)
			return checkResponseStatus(r.StatusCode, body, r.Header)
		}
		resp = r
		return nil
	})
	if err != nil {
		return err
	}
	defer drainClose(resp.Body)

	expectedSize := resp.ContentLength
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

// validateDocumentDownloadURL ensures downloadURL is an ODP patent-application
// document download URL (a DownloadOptionBag.DownloadUrl from GetPatentDocuments,
// e.g. {BaseURL}/api/v1/download/applications/{appNum}/{id}.pdf). Restricting to
// the configured host with the document download path keeps the authenticated
// request from being pointed at an arbitrary host.
func (c *Client) validateDocumentDownloadURL(downloadURL string) error {
	if downloadURL == "" {
		return fmt.Errorf("downloadURL cannot be empty")
	}
	expectedPrefix := c.config.BaseURL + "/api/v1/download/"
	if !strings.HasPrefix(downloadURL, expectedPrefix) {
		return fmt.Errorf("invalid document downloadURL: must start with %s (got: %s)", expectedPrefix, downloadURL)
	}
	return nil
}

// DownloadPatentDocument streams a patent file-wrapper document to w. downloadURL
// is a DownloadOptionBag.DownloadUrl returned by GetPatentDocuments (the documents
// are served as PDF). Covers applicant remarks, claim amendments, IDS, examiner's
// amendments, and the rest of the file wrapper.
func (c *Client) DownloadPatentDocument(ctx context.Context, downloadURL string, w io.Writer) error {
	if err := c.validateDocumentDownloadURL(downloadURL); err != nil {
		return err
	}
	return c.streamDownload(ctx, downloadURL, w, nil)
}

// SearchPetitions searches for petition decisions
func (c *Client) SearchPetitions(ctx context.Context, query string, offset, limit int) (*generated.PetitionDecisionResponseBag, error) {
	if err := validatePagination(offset, limit); err != nil {
		return nil, err
	}
	req := generated.PetitionDecisionSearchRequest{
		Q: StringPtr(query),
		Pagination: &generated.Pagination{
			Offset: Int32Ptr(int32(offset)),
			Limit:  Int32Ptr(int32(limit)),
		},
	}

	var resp *generated.PostApiV1PetitionDecisionsSearchResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PetitionDecisionsSearchWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
					RecordedDate:        derefStr(a.AssignmentRecordedDate),
					Conveyance:          derefStr(a.ConveyanceText),
					ReelFrame:           derefStr(a.ReelAndFrameNumber),
					MailedDate:          derefStr(a.AssignmentMailedDate),
					ReceivedDate:        derefStr(a.AssignmentReceivedDate),
					DocumentLocationURI: derefStr(a.AssignmentDocumentLocationURI),
				}
				if a.AssignorBag != nil {
					for _, assignor := range *a.AssignorBag {
						entry.Assignors = append(entry.Assignors, Assignor{
							Name:          derefStr(assignor.AssignorName),
							ExecutionDate: derefStr(assignor.ExecutionDate),
						})
					}
				}
				if a.AssigneeBag != nil {
					for _, assignee := range *a.AssigneeBag {
						p := Assignee{Name: derefStr(assignee.AssigneeNameText)}
						if assignee.AssigneeAddress != nil {
							addr := assignee.AssigneeAddress
							p.City = derefStr(addr.CityName)
							// 3.6 consolidated location into geographicRegionCode;
							// fall back to the deprecated countryOrStateCode for
							// older records.
							p.GeographicRegion = derefStr(addr.GeographicRegionCode)
							if p.GeographicRegion == "" {
								p.GeographicRegion = derefStr(addr.CountryOrStateCode)
							}
							p.PostalCode = derefStr(addr.PostalCode)
							p.CountryName = derefStr(addr.CountryName)
						}
						entry.Assignees = append(entry.Assignees, p)
					}
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
func (c *Client) GetPatentForeignPriority(ctx context.Context, applicationNumber string) (*ForeignPriorityResponse, error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextForeignPriorityResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextForeignPriorityWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	result := &ForeignPriorityResponse{
		ApplicationNumber: applicationNumber,
		Claims:            []ForeignPriorityClaim{},
	}
	if resp.JSON200 != nil && resp.JSON200.PatentFileWrapperDataBag != nil && len(*resp.JSON200.PatentFileWrapperDataBag) > 0 {
		bag := (*resp.JSON200.PatentFileWrapperDataBag)[0]
		if bag.ForeignPriorityBag != nil {
			for _, fp := range *bag.ForeignPriorityBag {
				result.Claims = append(result.Claims, ForeignPriorityClaim{
					ApplicationNumber: derefStr(fp.ApplicationNumberText),
					FilingDate:        derefStr(fp.FilingDate),
					IPOfficeName:      derefStr(fp.IpOfficeName),
				})
			}
		}
	}
	return result, nil
}

// GetPatentMetaData retrieves patent metadata (status, filing date, examiner, classification, etc.).
func (c *Client) GetPatentMetaData(ctx context.Context, applicationNumber string) (*MetaDataResponse, error) {
	var resp *generated.GetApiV1PatentApplicationsApplicationNumberTextMetaDataResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.GetApiV1PatentApplicationsApplicationNumberTextMetaDataWithResponse(ctx, applicationNumber)
		if err != nil {
			return err
		}
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil || resp.JSON200.PatentFileWrapperDataBag == nil || len(*resp.JSON200.PatentFileWrapperDataBag) == 0 {
		return nil, nil
	}
	bag := (*resp.JSON200.PatentFileWrapperDataBag)[0]
	if bag.ApplicationMetaData == nil {
		return nil, nil
	}
	m := bag.ApplicationMetaData
	out := &MetaDataResponse{
		ApplicationNumber:                        applicationNumber,
		ApplicationConfirmationNumber:            float32ToIntPtr(m.ApplicationConfirmationNumber),
		PatentNumber:                             derefStr(m.PatentNumber),
		InventionTitle:                           derefStr(m.InventionTitle),
		FilingDate:                               derefStr(m.FilingDate),
		GrantDate:                                derefStr(m.GrantDate),
		EffectiveFilingDate:                      derefStr(m.EffectiveFilingDate),
		EarliestPublicationNumber:                derefStr(m.EarliestPublicationNumber),
		EarliestPublicationDate:                  derefStr(m.EarliestPublicationDate),
		ApplicationStatusCode:                    m.ApplicationStatusCode,
		ApplicationStatusDescriptionText:         derefStr(m.ApplicationStatusDescriptionText),
		ApplicationStatusDate:                    derefStr(m.ApplicationStatusDate),
		ApplicationTypeCategory:                  derefStr(m.ApplicationTypeCategory),
		ApplicationTypeCode:                      derefStr(m.ApplicationTypeCode),
		ApplicationTypeLabelName:                 derefStr(m.ApplicationTypeLabelName),
		ExaminerNameText:                         derefStr(m.ExaminerNameText),
		GroupArtUnitNumber:                       derefStr(m.GroupArtUnitNumber),
		DocketNumber:                             derefStr(m.DocketNumber),
		CustomerNumber:                           m.CustomerNumber,
		FirstApplicantName:                       derefStr(m.FirstApplicantName),
		FirstInventorName:                        derefStr(m.FirstInventorName),
		FirstInventorToFileIndicator:             derefStr(m.FirstInventorToFileIndicator),
		NationalStageIndicator:                   m.NationalStageIndicator,
		PctPublicationNumber:                     derefStr(m.PctPublicationNumber),
		PctPublicationDate:                       derefStr(m.PctPublicationDate),
		InternationalRegistrationNumber:          derefStr(m.InternationalRegistrationNumber),
		InternationalRegistrationPublicationDate: derefStr(m.InternationalRegistrationPublicationDate),
		UspcSymbolText:                           derefStr(m.UspcSymbolText),
		Class:                                    derefStr(m.Class),
		Subclass:                                 derefStr(m.Subclass),
	}
	if m.CpcClassificationBag != nil {
		out.CpcClassificationBag = append(out.CpcClassificationBag, *m.CpcClassificationBag...)
	}
	if m.PublicationCategoryBag != nil {
		out.PublicationCategoryBag = append(out.PublicationCategoryBag, *m.PublicationCategoryBag...)
	}
	if m.PublicationDateBag != nil {
		out.PublicationDateBag = append(out.PublicationDateBag, *m.PublicationDateBag...)
	}
	if m.PublicationSequenceNumberBag != nil {
		out.PublicationSequenceNumberBag = append(out.PublicationSequenceNumberBag, *m.PublicationSequenceNumberBag...)
	}
	if m.EntityStatusData != nil {
		out.EntityStatus = &EntityStatus{
			BusinessEntityStatusCategory: derefStr(m.EntityStatusData.BusinessEntityStatusCategory),
			SmallEntityStatusIndicator:   m.EntityStatusData.SmallEntityStatusIndicator,
		}
	}
	if m.ApplicantBag != nil {
		for _, a := range *m.ApplicantBag {
			ap := Applicant{
				ApplicantNameText: derefStr(a.ApplicantNameText),
				FirstName:         derefStr(a.FirstName),
				MiddleName:        derefStr(a.MiddleName),
				LastName:          derefStr(a.LastName),
				NamePrefix:        derefStr(a.NamePrefix),
				NameSuffix:        derefStr(a.NameSuffix),
				PreferredName:     derefStr(a.PreferredName),
				CountryCode:       derefStr(a.CountryCode),
			}
			if a.CorrespondenceAddressBag != nil {
				for _, c := range *a.CorrespondenceAddressBag {
					ap.CorrespondenceAddressBag = append(ap.CorrespondenceAddressBag, CorrespondenceAddress{
						NameLineOne:           derefStr(c.NameLineOneText),
						NameLineTwo:           derefStr(c.NameLineTwoText),
						CityName:              derefStr(c.CityName),
						CountryCode:           derefStr(c.CountryCode),
						CountryName:           derefStr(c.CountryName),
						GeographicRegionCode:  derefStr(c.GeographicRegionCode),
						GeographicRegionName:  derefStr(c.GeographicRegionName),
						PostalAddressCategory: derefStr(c.PostalAddressCategory),
					})
				}
			}
			out.Applicants = append(out.Applicants, ap)
		}
	}
	if m.InventorBag != nil {
		for _, i := range *m.InventorBag {
			inv := Inventor{
				InventorNameText: derefStr(i.InventorNameText),
				FirstName:        derefStr(i.FirstName),
				MiddleName:       derefStr(i.MiddleName),
				LastName:         derefStr(i.LastName),
				NamePrefix:       derefStr(i.NamePrefix),
				NameSuffix:       derefStr(i.NameSuffix),
				PreferredName:    derefStr(i.PreferredName),
				CountryCode:      derefStr(i.CountryCode),
			}
			if i.CorrespondenceAddressBag != nil {
				for _, c := range *i.CorrespondenceAddressBag {
					inv.CorrespondenceAddressBag = append(inv.CorrespondenceAddressBag, CorrespondenceAddress{
						NameLineOne:           derefStr(c.NameLineOneText),
						NameLineTwo:           derefStr(c.NameLineTwoText),
						CityName:              derefStr(c.CityName),
						CountryCode:           derefStr(c.CountryCode),
						CountryName:           derefStr(c.CountryName),
						GeographicRegionCode:  derefStr(c.GeographicRegionCode),
						GeographicRegionName:  derefStr(c.GeographicRegionName),
						PostalAddressCategory: derefStr(c.PostalAddressCategory),
					})
				}
			}
			out.Inventors = append(out.Inventors, inv)
		}
	}
	return out, nil
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
func (c *Client) SearchTrialProceedings(ctx context.Context, query string, offset, limit int) (*generated.ProceedingDataResponse, error) {
	if err := validatePagination(offset, limit); err != nil {
		return nil, err
	}
	req := generated.PostApiV1PatentTrialsProceedingsSearchJSONRequestBody{
		Q: StringPtr(query),
		Pagination: &generated.Pagination{
			Offset: Int32Ptr(int32(offset)),
			Limit:  Int32Ptr(int32(limit)),
		},
	}

	var resp *generated.PostApiV1PatentTrialsProceedingsSearchResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentTrialsProceedingsSearchWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
func (c *Client) SearchTrialDecisions(ctx context.Context, query string, offset, limit int) (*generated.DecisionDataResponse, error) {
	if err := validatePagination(offset, limit); err != nil {
		return nil, err
	}
	req := generated.PostApiV1PatentTrialsDecisionsSearchJSONRequestBody{
		Q: StringPtr(query),
		Pagination: &generated.Pagination{
			Offset: Int32Ptr(int32(offset)),
			Limit:  Int32Ptr(int32(limit)),
		},
	}

	var resp *generated.PostApiV1PatentTrialsDecisionsSearchResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentTrialsDecisionsSearchWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
func (c *Client) SearchTrialDocuments(ctx context.Context, query string, offset, limit int) (*generated.DocumentDataResponse, error) {
	if err := validatePagination(offset, limit); err != nil {
		return nil, err
	}
	req := generated.PostApiV1PatentTrialsDocumentsSearchJSONRequestBody{
		Q: StringPtr(query),
		Pagination: &generated.Pagination{
			Offset: Int32Ptr(int32(offset)),
			Limit:  Int32Ptr(int32(limit)),
		},
	}

	var resp *generated.PostApiV1PatentTrialsDocumentsSearchResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentTrialsDocumentsSearchWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
func (c *Client) SearchAppealDecisions(ctx context.Context, query string, offset, limit int) (*generated.AppealDecisionDataResponse, error) {
	if err := validatePagination(offset, limit); err != nil {
		return nil, err
	}
	req := generated.PostApiV1PatentAppealsDecisionsSearchJSONRequestBody{
		Q: StringPtr(query),
		Pagination: &generated.Pagination{
			Offset: Int32Ptr(int32(offset)),
			Limit:  Int32Ptr(int32(limit)),
		},
	}

	var resp *generated.PostApiV1PatentAppealsDecisionsSearchResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentAppealsDecisionsSearchWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
func (c *Client) SearchInterferenceDecisions(ctx context.Context, query string, offset, limit int) (*generated.InterferenceDecisionDataResponse, error) {
	if err := validatePagination(offset, limit); err != nil {
		return nil, err
	}
	req := generated.PostApiV1PatentInterferencesDecisionsSearchJSONRequestBody{
		Q: StringPtr(query),
		Pagination: &generated.Pagination{
			Offset: Int32Ptr(int32(offset)),
			Limit:  Int32Ptr(int32(limit)),
		},
	}

	var resp *generated.PostApiV1PatentInterferencesDecisionsSearchResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.generated.PostApiV1PatentInterferencesDecisionsSearchWithResponse(ctx, req)
		if err != nil {
			return err
		}
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
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
		if err := checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse)); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
