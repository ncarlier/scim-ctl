package common

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// RetryTransport is an http.RoundTripper that retries failed requests.
type RetryTransport struct {
	Transport  http.RoundTripper
	MaxRetries int
	Verbose    bool
}

// RoundTrip executes a single HTTP transaction, returning a Response for the provided Request.
// It retries the request according to MaxRetries on network errors or 5xx/429 status codes.
func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	var err error

	// Read the body if it exists so we can replay it on retries
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body for retry: %w", err)
		}
		req.Body.Close()
	}

	var resp *http.Response
	var reqErr error

	for attempt := 0; attempt <= t.MaxRetries; attempt++ {
		// Prepare a new body reader for each attempt if we had a body
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		resp, reqErr = t.Transport.RoundTrip(req)
		
		// If we succeed and don't get a retryable status code, break and return
		if reqErr == nil && !isRetryableStatusCode(resp.StatusCode) {
			return resp, nil
		}

		// Don't retry context errors (canceled or deadline exceeded)
		if req.Context().Err() != nil {
			return resp, reqErr
		}

		if attempt < t.MaxRetries {
			// Backoff logic (1s, 2s, 4s, etc.)
			backoff := time.Duration(1<<attempt) * time.Second
			if t.Verbose {
				errMsg := ""
				if reqErr != nil {
					errMsg = reqErr.Error()
				} else if resp != nil {
					errMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
				}
				fmt.Fprintf(os.Stderr, "Request to %s failed (%s), retrying in %s (attempt %d/%d)...\n", req.URL.String(), errMsg, backoff, attempt+1, t.MaxRetries)
			}

			// Clean up the previous response body if we got one before retrying
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}

			select {
			case <-time.After(backoff):
			case <-req.Context().Done():
				return resp, req.Context().Err()
			}
		}
	}

	return resp, reqErr
}

func isRetryableStatusCode(code int) bool {
	return code == http.StatusTooManyRequests || 
	       code == http.StatusInternalServerError || 
	       code == http.StatusBadGateway || 
	       code == http.StatusServiceUnavailable || 
	       code == http.StatusGatewayTimeout
}
