package pingdom

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"time"
)

// rateLimitedTransport wraps an http.RoundTripper with two safeguards for the
// Pingdom API:
//  1. Enforces a minimum interval between consecutive requests, smoothing bursts
//     toward Pingdom's documented 1 req/sec ceiling.
//  2. Retries HTTP 429 responses with exponential backoff. Without this, large
//     workspaces with many checks/contacts trip Pingdom's server-side burst
//     protection during refresh.
//
// Single-process only — Terraform plan/apply run a fresh provider process per
// invocation, so cross-process coordination is not needed.
type rateLimitedTransport struct {
	inner       http.RoundTripper
	minInterval time.Duration
	maxRetries  int

	mu          sync.Mutex
	lastRequest time.Time
}

// newRateLimitedTransport returns a transport that spaces requests at least
// minInterval apart and retries 429 up to maxRetries times. minInterval values
// below 1ms are clamped to 1ms; pass 0 to effectively disable spacing.
func newRateLimitedTransport(minInterval time.Duration, maxRetries int) *rateLimitedTransport {
	if minInterval < time.Millisecond {
		minInterval = time.Millisecond
	}
	if maxRetries < 0 {
		maxRetries = 0
	}
	return &rateLimitedTransport{
		inner:       http.DefaultTransport,
		minInterval: minInterval,
		maxRetries:  maxRetries,
	}
}

// RoundTrip implements http.RoundTripper. Holds the mutex for the duration of
// the call to keep request spacing accurate; this serialises concurrent
// callers, which is acceptable for a Terraform provider (parallelism is
// already user-controlled via -parallelism).
func (t *rateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Buffer body for retry replay. GET requests have no body; PUT/POST may.
	var bodyBytes []byte
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close()
		bodyBytes = b
	}

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		if !t.lastRequest.IsZero() {
			elapsed := time.Since(t.lastRequest)
			if elapsed < t.minInterval {
				time.Sleep(t.minInterval - elapsed)
			}
		}

		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		t.lastRequest = time.Now()
		resp, err = t.inner.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		// Drain so the connection can be reused.
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if attempt == t.maxRetries {
			return resp, nil
		}

		backoff := t.minInterval * time.Duration(1<<attempt)
		time.Sleep(backoff)
	}

	return resp, nil
}
