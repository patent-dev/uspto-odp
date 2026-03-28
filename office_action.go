package odp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	oa "github.com/patent-dev/uspto-odp/generated/oa"
)

// DSAPIResponse is the response envelope from Office Action DSAPI endpoints.
// The DSAPI uses Solr/Lucene with dynamic schemas - each dataset (text, citations,
// rejections, enriched) has different fields per document. The Docs field uses
// []map[string]any because the field set is not fixed at compile time. Use the
// /fields endpoints (e.g., GetOfficeActionFields) to discover available fields.
type DSAPIResponse struct {
	Response DSAPIResult `json:"response"`
}

// DSAPIResult holds the search results from a DSAPI query.
type DSAPIResult struct {
	NumFound int              `json:"numFound"`
	Start    int              `json:"start"`
	Docs     []map[string]any `json:"docs"`
}

// DSAPIFieldsResponse is the response from a DSAPI /fields endpoint.
type DSAPIFieldsResponse struct {
	APIKey              string   `json:"apiKey"`
	APIVersionNumber    string   `json:"apiVersionNumber"`
	APIURL              string   `json:"apiUrl"`
	APIDocumentationURL string   `json:"apiDocumentationUrl"`
	APIStatus           string   `json:"apiStatus"`
	FieldCount          int      `json:"fieldCount"`
	Fields              []string `json:"fields"`
}

// readJSONResponse reads an HTTP response body, checks status, and unmarshals JSON into result.
func readJSONResponse(resp *http.Response, result any) error {
	defer drainClose(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}
	if err := checkStatusWithBody(resp.StatusCode, body); err != nil {
		return err
	}
	return json.Unmarshal(body, result)
}

// dsapiSearch is a shared helper for all Office Action DSAPI search endpoints.
// Callers must normalize criteria (e.g., default to "*:*") before calling.
func (c *Client) dsapiSearch(ctx context.Context,
	doRequest func(context.Context) (*http.Response, error),
) (*DSAPIResponse, error) {
	var result DSAPIResponse
	err := c.retryableRequest(ctx, func() error {
		result = DSAPIResponse{}
		resp, err := doRequest(ctx)
		if err != nil {
			return err
		}
		return readJSONResponse(resp, &result)
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// dsapiFields is a shared helper for all Office Action DSAPI /fields endpoints.
func (c *Client) dsapiFields(ctx context.Context,
	doRequest func(context.Context) (*http.Response, error),
) (*DSAPIFieldsResponse, error) {
	var result DSAPIFieldsResponse
	err := c.retryableRequest(ctx, func() error {
		result = DSAPIFieldsResponse{}
		resp, err := doRequest(ctx)
		if err != nil {
			return err
		}
		return readJSONResponse(resp, &result)
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// NormalizeDSAPIApplicationNumber normalizes a US application number for use
// in DSAPI Lucene queries. The DSAPI expects digits only (e.g., "17248024").
// Accepts formats like "17/248,024", "US17/248,024", "US 17248024", "17248024".
// Returns empty string for empty input.
func NormalizeDSAPIApplicationNumber(appNum string) string {
	num := strings.TrimSpace(appNum)
	if num == "" {
		return ""
	}
	// Strip "US" prefix (case-insensitive)
	upper := strings.ToUpper(num)
	if strings.HasPrefix(upper, "US") {
		num = num[2:]
	}
	// Strip slash, commas, spaces
	num = strings.NewReplacer("/", "", ",", "", " ", "").Replace(num)
	return num
}

// DSAPIApplicationCriteria builds a Lucene criteria string for querying by
// application number. Returns "*:*" (match all) for empty input.
func DSAPIApplicationCriteria(appNum string) string {
	normalized := NormalizeDSAPIApplicationNumber(appNum)
	if normalized == "" {
		return "*:*"
	}
	return "patentApplicationNumber:" + normalized
}

// --- Office Action Text Retrieval API ---

// SearchOfficeActions searches office action full-text records using Lucene query syntax.
// Criteria uses Lucene syntax (e.g., "patentApplicationNumber:16123456").
// Use "*:*" or empty string to match all records.
func (c *Client) SearchOfficeActions(ctx context.Context, criteria string, start, rows int32) (*DSAPIResponse, error) {
	if criteria == "" {
		criteria = "*:*"
	}
	return c.dsapiSearch(ctx, func(ctx context.Context) (*http.Response, error) {
		req := oa.OaActionsSearchFormdataRequestBody{
			Criteria: criteria,
			Start:    Int32Ptr(start),
			Rows:     Int32Ptr(rows),
		}
		return c.oa.OaActionsSearchWithFormdataBody(ctx, req)
	})
}

// GetOfficeActionFields returns the list of searchable fields for the Office Action Text API.
func (c *Client) GetOfficeActionFields(ctx context.Context) (*DSAPIFieldsResponse, error) {
	return c.dsapiFields(ctx, func(ctx context.Context) (*http.Response, error) {
		return c.oa.OaActionsListFields(ctx)
	})
}

// --- Office Action Citations API ---

// SearchOfficeActionCitations searches office action citation records.
// Data covers Office Actions mailed from June 1, 2018 to 180 days prior to current date.
func (c *Client) SearchOfficeActionCitations(ctx context.Context, criteria string, start, rows int32) (*DSAPIResponse, error) {
	if criteria == "" {
		criteria = "*:*"
	}
	return c.dsapiSearch(ctx, func(ctx context.Context) (*http.Response, error) {
		req := oa.OaCitationsSearchFormdataRequestBody{
			Criteria: criteria,
			Start:    Int32Ptr(start),
			Rows:     Int32Ptr(rows),
		}
		return c.oa.OaCitationsSearchWithFormdataBody(ctx, req)
	})
}

// GetOfficeActionCitationFields returns the list of searchable fields for the OA Citations API.
func (c *Client) GetOfficeActionCitationFields(ctx context.Context) (*DSAPIFieldsResponse, error) {
	return c.dsapiFields(ctx, func(ctx context.Context) (*http.Response, error) {
		return c.oa.OaCitationsListFields(ctx)
	})
}

// --- Office Action Rejections API ---

// SearchOfficeActionRejections searches office action rejection records.
// Includes rejection types (101, 102, 103, 112, double patenting) and patent eligibility indicators.
func (c *Client) SearchOfficeActionRejections(ctx context.Context, criteria string, start, rows int32) (*DSAPIResponse, error) {
	if criteria == "" {
		criteria = "*:*"
	}
	return c.dsapiSearch(ctx, func(ctx context.Context) (*http.Response, error) {
		req := oa.OaRejectionsSearchFormdataRequestBody{
			Criteria: criteria,
			Start:    Int32Ptr(start),
			Rows:     Int32Ptr(rows),
		}
		return c.oa.OaRejectionsSearchWithFormdataBody(ctx, req)
	})
}

// GetOfficeActionRejectionFields returns the list of searchable fields for the OA Rejections API.
func (c *Client) GetOfficeActionRejectionFields(ctx context.Context) (*DSAPIFieldsResponse, error) {
	return c.dsapiFields(ctx, func(ctx context.Context) (*http.Response, error) {
		return c.oa.OaRejectionsListFields(ctx)
	})
}

// --- Enriched Citation API ---

// SearchEnrichedCitations searches enriched citation metadata records.
// Uses AI/ML to extract statutes, rejected claims, prior art references from office actions.
func (c *Client) SearchEnrichedCitations(ctx context.Context, criteria string, start, rows int32) (*DSAPIResponse, error) {
	if criteria == "" {
		criteria = "*:*"
	}
	return c.dsapiSearch(ctx, func(ctx context.Context) (*http.Response, error) {
		req := oa.EnrichedCitationsSearchFormdataRequestBody{
			Criteria: criteria,
			Start:    Int32Ptr(start),
			Rows:     Int32Ptr(rows),
		}
		return c.oa.EnrichedCitationsSearchWithFormdataBody(ctx, req)
	})
}

// GetEnrichedCitationFields returns the list of searchable fields for the Enriched Citation API.
func (c *Client) GetEnrichedCitationFields(ctx context.Context) (*DSAPIFieldsResponse, error) {
	return c.dsapiFields(ctx, func(ctx context.Context) (*http.Response, error) {
		return c.oa.EnrichedCitationsListFields(ctx)
	})
}
