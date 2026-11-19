package odp

import (
	"context"
	"fmt"

	"github.com/patent-dev/uspto-odp/generated/oasearch"
)

// ODP Office Action Search API (ODP release 5.0): office actions on patent
// applications filed in 2001 or later, searchable by metadata and full text,
// on the ODP host under /api/v1/patent/office-action. These are separate from
// the legacy DSAPI Office Action wrappers (SearchOfficeActions and friends).

// SearchOfficeActionRecords searches office actions with an OpenSearch query.
// An empty query returns the most recent office actions (by submissionDate).
func (c *Client) SearchOfficeActionRecords(ctx context.Context, query string, offset, limit int) (*oasearch.OfficeActionSearchResponse, error) {
	if err := validatePagination(offset, limit); err != nil {
		return nil, err
	}
	o, l := int32(offset), int32(limit)
	req := oasearch.OfficeActionSearchRequest{Pagination: &oasearch.Pagination{Offset: &o, Limit: &l}}
	if query != "" {
		req.Q = &query
	}
	return c.SearchOfficeActionRecordsWithRequest(ctx, req)
}

// SearchOfficeActionRecordsWithRequest runs an office action search with the
// full request body: filters, range filters, sort, field projection, facets.
func (c *Client) SearchOfficeActionRecordsWithRequest(ctx context.Context, req oasearch.OfficeActionSearchRequest) (*oasearch.OfficeActionSearchResponse, error) {
	var resp *oasearch.SearchOfficeActionsPostResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.oasearch.SearchOfficeActionsPostWithResponse(ctx, req)
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

// GetOfficeActionRecord retrieves one office action by its
// officeActionRecordIdentifier; the record is the single entry of
// OfficeActionDataBag. Its officeActionText is an excerpt; use
// GetOfficeActionText for the full text.
func (c *Client) GetOfficeActionRecord(ctx context.Context, recordID string) (*oasearch.OfficeActionResponse, error) {
	if recordID == "" {
		return nil, fmt.Errorf("recordID cannot be empty")
	}
	var resp *oasearch.GetOfficeActionByIdResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.oasearch.GetOfficeActionByIdWithResponse(ctx, recordID)
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

// SearchOfficeActionRecordsDownload downloads office action search results in
// the requested format (json or csv). Facets are not supported here.
func (c *Client) SearchOfficeActionRecordsDownload(ctx context.Context, req oasearch.OfficeActionDownloadRequest) ([]byte, error) {
	var resp *oasearch.DownloadOfficeActionsPostResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.oasearch.DownloadOfficeActionsPostWithResponse(ctx, req)
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

// GetOfficeActionText returns the full raw text of one office action (plain
// text, as USPTO extracted it from the document).
func (c *Client) GetOfficeActionText(ctx context.Context, recordID string) ([]byte, error) {
	if recordID == "" {
		return nil, fmt.Errorf("recordID cannot be empty")
	}
	var resp *oasearch.DownloadOfficeActionTextResponse
	err := c.retryableRequest(ctx, func() error {
		var err error
		resp, err = c.oasearch.DownloadOfficeActionTextWithResponse(ctx, recordID)
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
