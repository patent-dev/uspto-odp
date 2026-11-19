package odp

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/patent-dev/uspto-odp/generated/tmd"
)

// Trademark Decisions and Proceedings API (ODP release 5.1). One dataset
// covering three decisionProceedingCategory values: ExpungReexam (expungement
// and reexamination proceedings), PetitionsToDirector and Administrative
// (sanctions, show-cause orders and similar). Records link to the TSDR case
// (serialNumber, registrationNumber, documentURI, markImageURI); Administrative
// records carry a fileDownloadURI for the decision PDF instead.

// TrademarkDecisionCategory values for the decisionProceedingCategory filter.
const (
	TrademarkDecisionCategoryExpungReexam        = "ExpungReexam"
	TrademarkDecisionCategoryPetitionsToDirector = "PetitionsToDirector"
	TrademarkDecisionCategoryAdministrative      = "Administrative"
)

// trademarkDecisionFilePath is the path prefix of fileDownloadURI values.
const trademarkDecisionFilePath = "/api/v1/trademark/decision-proceeding/file/"

// SearchTrademarkDecisions searches trademark decisions and proceedings with an
// OpenSearch query. An empty query returns the most recent records (sorted by
// decisionActionDate descending). The API accepts limit 1-100.
func (c *Client) SearchTrademarkDecisions(ctx context.Context, query string, offset, limit int) (*tmd.SearchResponse, error) {
	if err := validatePagination(offset, limit); err != nil {
		return nil, err
	}
	req := tmd.SearchRequest{
		Pagination: &tmd.Pagination{Offset: &offset, Limit: &limit},
	}
	if query != "" {
		req.Q = &query
	}
	return c.SearchTrademarkDecisionsWithRequest(ctx, req)
}

// SearchTrademarkDecisionsWithRequest runs a trademark decisions search with
// the full request body: filters (e.g. decisionProceedingCategory), range
// filters, sort, field projection and facets.
func (c *Client) SearchTrademarkDecisionsWithRequest(ctx context.Context, req tmd.SearchRequest) (*tmd.SearchResponse, error) {
	var resp *tmd.SearchPostResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.tmd.SearchPostWithResponse(ctx, req)
		if err != nil {
			return err
		}
		return checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse))
	})
	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// GetTrademarkDecision retrieves one trademark decision by its
// documentIdentifier. The record is TrademarkDecisionsDataBag (a single
// object, not a list).
func (c *Client) GetTrademarkDecision(ctx context.Context, documentIdentifier string) (*tmd.TrademarkDecisionResponse, error) {
	if documentIdentifier == "" {
		return nil, fmt.Errorf("documentIdentifier cannot be empty")
	}
	var resp *tmd.GetByDocumentIdentifierResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.tmd.GetByDocumentIdentifierWithResponse(ctx, documentIdentifier)
		if err != nil {
			return err
		}
		return checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse))
	})
	if err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// SearchTrademarkDecisionsDownload downloads trademark decision search results
// in the requested format (json or csv). Facets are not supported here.
func (c *Client) SearchTrademarkDecisionsDownload(ctx context.Context, req tmd.DownloadRequest) ([]byte, error) {
	var resp *tmd.SearchDownloadPostResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.tmd.SearchDownloadPostWithResponse(ctx, req)
		if err != nil {
			return err
		}
		return checkResponseStatus(resp.StatusCode(), resp.Body, headerOf(resp.HTTPResponse))
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// DownloadTrademarkDecisionFile streams the decision document behind a
// record's FileDownloadURI (a PDF) to w. The endpoint is not in the published
// spec; the URI must point at this client's base URL.
func (c *Client) DownloadTrademarkDecisionFile(ctx context.Context, fileDownloadURI string, w io.Writer) error {
	expectedPrefix := strings.TrimRight(c.config.BaseURL, "/") + trademarkDecisionFilePath
	if !strings.HasPrefix(fileDownloadURI, expectedPrefix) {
		return fmt.Errorf("invalid trademark decision fileDownloadURI: must start with %s (got: %s)", expectedPrefix, fileDownloadURI)
	}
	return c.streamDownload(ctx, fileDownloadURI, w, nil)
}
