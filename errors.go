package odp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// APIError represents an error returned by the USPTO API with status code
type APIError struct {
	StatusCode int
	Message    string
	Body       string // server response body for debugging
	// RetryAfter, when non-zero, is the duration the server asked the client
	// to wait before retrying (parsed from the Retry-After header).
	RetryAfter time.Duration
	// Empty is set when the server returned a success status with no body. USPTO
	// services (notably TSDR) do this when degraded; it is treated as transient.
	Empty bool
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

// IsRetryable returns true for transient conditions: HTTP 429, 5xx, or an empty
// body on a success status (a degraded-service symptom that often clears on a
// retry). The Retry-After cap is a *client* policy, enforced by retryableRequest,
// not by the error itself.
func (e *APIError) IsRetryable() bool {
	return e.Empty || e.StatusCode == http.StatusTooManyRequests || e.StatusCode >= 500
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.IsRetryable()
	}
	// Context cancellation and deadline-exceeded are caller intent, not a
	// transient network condition. context.DeadlineExceeded satisfies
	// net.Error.Timeout(), so short-circuit before the timeout check below.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
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

// parseRetryAfter parses a Retry-After header value (RFC 7231): either delta
// seconds or an HTTP-date. Returns 0 if absent or unparseable. RFC 7231 also
// permits "0" to mean "retry immediately" -- that maps to 0 here, which causes
// retryableRequest to use its exponential backoff floor.
func parseRetryAfter(headers http.Header) time.Duration {
	if headers == nil {
		return 0
	}
	v := strings.TrimSpace(headers.Get("Retry-After"))
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil && secs >= 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		d := time.Until(t)
		if d <= 0 {
			return 0
		}
		return d
	}
	return 0
}

// checkResponseStatus returns an APIError for non-2xx responses, including
// the response body for debugging. If headers is non-nil, the Retry-After
// value (if present) is parsed onto the APIError.
func checkResponseStatus(statusCode int, body []byte, headers http.Header) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}
	apiErr := &APIError{
		StatusCode: statusCode,
		Message:    fmt.Sprintf("API returned status %d", statusCode),
	}
	if len(body) > 0 {
		// 4 KiB keeps debug payloads (USPTO often echoes the request body
		// in 4xx responses) without exposing arbitrarily large blobs.
		apiErr.Body = truncatePreview(string(body), 4096)
	}
	apiErr.RetryAfter = parseRetryAfter(headers)
	return apiErr
}

// checkEmptyBody reports a clear, retryable error when a success response carries
// no body. USPTO services (notably TSDR) occasionally return an empty 200/204 when
// degraded; without this, callers fail later with an opaque "unexpected end of
// JSON input" or "XML unmarshal: EOF" that hides the real cause. Call it only
// after checkResponseStatus has confirmed a 2xx.
func checkEmptyBody(statusCode int, body []byte) error {
	if len(bytes.TrimSpace(body)) > 0 {
		return nil
	}
	return &APIError{
		StatusCode: statusCode,
		Empty:      true,
		Message:    fmt.Sprintf("USPTO returned an empty response body with HTTP %d", statusCode),
	}
}
