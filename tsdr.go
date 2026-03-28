package odp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	tsdrgen "github.com/patent-dev/uspto-odp/generated/tsdr"
)

var (
	errTSDRNotConfigured = fmt.Errorf("TSDR client not configured: set TSDRAPIKey in Config")
	tsdrSerialNumberRe   = regexp.MustCompile(`^\d{8}$`)
)

// validateSerialNumber checks that a trademark serial number is 8 digits.
func validateSerialNumber(serialNumber string) error {
	if !tsdrSerialNumberRe.MatchString(serialNumber) {
		return fmt.Errorf("invalid serial number %q: must be exactly 8 digits", serialNumber)
	}
	return nil
}

// TSDRStatusResponse is the JSON response from the TSDR case status endpoint.
// The status endpoint uses content negotiation (Accept: application/json) to return
// JSON instead of the default XML. The response structure does not match the generated
// XML-based types, so this wrapper provides a typed envelope while the inner trademark
// data uses dynamic maps matching the API's JSON schema.
type TSDRStatusResponse struct {
	Trademarks []map[string]any `json:"trademarks"`
}

// GetTrademarkStatus retrieves the full trademark case status for a serial number.
// The serial number must be 8 digits (e.g., "97123456").
// Returns the raw XML response wrapper from the generated client.
func (c *Client) GetTrademarkStatus(ctx context.Context, serialNumber string) (*tsdrgen.LoadXMLResponse, error) {
	if c.tsdr == nil {
		return nil, errTSDRNotConfigured
	}
	if err := validateSerialNumber(serialNumber); err != nil {
		return nil, err
	}

	var resp *tsdrgen.LoadXMLResponse
	err := c.retryableRequest(ctx, func() error {
		resp = nil
		var err error
		resp, err = c.tsdr.LoadXMLWithResponse(ctx, serialNumber)
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
	return resp, nil
}

// tsdrJSONEditor sets Accept: application/json for TSDR content negotiation.
// TSDR endpoints use content negotiation - without this, they return XML.
func tsdrJSONEditor(_ context.Context, req *http.Request) error {
	req.Header.Set("Accept", "application/json")
	return nil
}

// GetTrademarkStatusJSON retrieves trademark case status as parsed JSON.
// Uses content negotiation to request JSON instead of the default XML.
func (c *Client) GetTrademarkStatusJSON(ctx context.Context, serialNumber string) (*TSDRStatusResponse, error) {
	if c.tsdr == nil {
		return nil, errTSDRNotConfigured
	}
	if err := validateSerialNumber(serialNumber); err != nil {
		return nil, err
	}

	var result TSDRStatusResponse
	err := c.retryableRequest(ctx, func() error {
		result = TSDRStatusResponse{}
		resp, err := c.tsdr.LoadXML(ctx, serialNumber, tsdrJSONEditor)
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

// GetTrademarkDocumentsXML retrieves the document list for a trademark case as raw XML.
func (c *Client) GetTrademarkDocumentsXML(ctx context.Context, serialNumber string) ([]byte, error) {
	if c.tsdr == nil {
		return nil, errTSDRNotConfigured
	}
	if err := validateSerialNumber(serialNumber); err != nil {
		return nil, err
	}

	var result []byte
	err := c.retryableRequest(ctx, func() error {
		resp, err := c.tsdr.GetCaseDocsInfoXml(ctx, serialNumber)
		if err != nil {
			return err
		}
		defer drainClose(resp.Body)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response: %w", err)
		}
		if err := checkStatusWithBody(resp.StatusCode, body); err != nil {
			return err
		}
		result = body
		return nil
	})
	return result, err
}

// GetTrademarkDocumentInfo retrieves info about a specific document within a trademark case.
// The docID format is {DocumentTypeCode}{YYYYMMDD}, e.g., "NOA20230322".
// Obtain valid type codes and dates from the GetTrademarkDocumentsXML listing.
// Returns raw XML because this endpoint does not support JSON content negotiation.
func (c *Client) GetTrademarkDocumentInfo(ctx context.Context, serialNumber, docID string) ([]byte, error) {
	if c.tsdr == nil {
		return nil, errTSDRNotConfigured
	}
	if err := validateSerialNumber(serialNumber); err != nil {
		return nil, err
	}

	var result []byte
	err := c.retryableRequest(ctx, func() error {
		resp, err := c.tsdr.GetDocumentInfoXml(ctx, serialNumber, docID)
		if err != nil {
			return err
		}
		defer drainClose(resp.Body)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response: %w", err)
		}
		if err := checkStatusWithBody(resp.StatusCode, body); err != nil {
			return err
		}
		result = body
		return nil
	})
	return result, err
}

// DownloadTrademarkDocument downloads a trademark document as PDF.
// When retries are enabled, the entire document is buffered in memory before
// writing to w, so partial writes from transient failures don't corrupt the output.
// The final copy to w is not retried - if w fails (e.g., closed file), the data is lost.
func (c *Client) DownloadTrademarkDocument(ctx context.Context, serialNumber, docID string, w io.Writer) error {
	if c.tsdr == nil {
		return errTSDRNotConfigured
	}
	if err := validateSerialNumber(serialNumber); err != nil {
		return err
	}

	checkPDFResponse := func(resp *http.Response) error {
		if err := checkStatus(resp.StatusCode); err != nil {
			drainClose(resp.Body)
			return err
		}
		ct := resp.Header.Get("Content-Type")
		if ct != "" && !strings.HasPrefix(ct, "application/pdf") {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("expected application/pdf but got %s: %s", ct, truncatePreview(string(body), 256))
		}
		return nil
	}

	if c.config.MaxRetries > 0 {
		// Buffer the download to prevent partial writes on retry
		var buf bytes.Buffer
		err := c.retryableRequest(ctx, func() error {
			buf.Reset()
			resp, err := c.tsdr.GetDocumentContentPdf(ctx, serialNumber, docID)
			if err != nil {
				return err
			}
			defer drainClose(resp.Body)
			if err := checkPDFResponse(resp); err != nil {
				return err
			}
			_, err = io.Copy(&buf, resp.Body)
			return err
		})
		if err != nil {
			return err
		}
		_, err = io.Copy(w, &buf)
		return err
	}

	// No retries: stream directly to writer
	resp, err := c.tsdr.GetDocumentContentPdf(ctx, serialNumber, docID)
	if err != nil {
		return err
	}
	defer drainClose(resp.Body)
	if err := checkPDFResponse(resp); err != nil {
		return err
	}
	_, err = io.Copy(w, resp.Body)
	return err
}

// GetTrademarkLastUpdate retrieves the last update time for a trademark case.
func (c *Client) GetTrademarkLastUpdate(ctx context.Context, serialNumber string) (*tsdrgen.CaseUpdateInfoList, error) {
	if c.tsdr == nil {
		return nil, errTSDRNotConfigured
	}

	var resp *tsdrgen.GetcaseUpdateInfoResponse
	err := c.retryableRequest(ctx, func() error {
		resp = nil
		var err error
		params := &tsdrgen.GetcaseUpdateInfoParams{Sn: serialNumber}
		resp, err = c.tsdr.GetcaseUpdateInfoWithResponse(ctx, params)
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

// GetTrademarkMultiStatus retrieves status for multiple trademarks by type.
// Type can be "sn" (serial number), "rn" (registration number), or "in" (international number).
func (c *Client) GetTrademarkMultiStatus(ctx context.Context, pType string, numbers []string) (*tsdrgen.TransactionBag, error) {
	if c.tsdr == nil {
		return nil, errTSDRNotConfigured
	}

	validTSDRTypes := map[string]bool{"sn": true, "rn": true, "in": true}
	if !validTSDRTypes[pType] {
		return nil, fmt.Errorf("invalid type %q: must be one of sn, rn, in", pType)
	}

	var resp *tsdrgen.GetListResponse
	err := c.retryableRequest(ctx, func() error {
		resp = nil
		var err error
		params := &tsdrgen.GetListParams{Ids: numbers}
		resp, err = c.tsdr.GetListWithResponse(ctx, pType, params)
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
