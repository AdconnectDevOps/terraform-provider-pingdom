package pingdom

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRateLimitedTransport_SpacesRequests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	tr := newRateLimitedTransport(200*time.Millisecond, 0)
	client := &http.Client{Transport: tr}

	start := time.Now()
	for i := 0; i < 3; i++ {
		resp, err := client.Get(srv.URL)
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		resp.Body.Close()
	}
	elapsed := time.Since(start)

	// Three requests with 200ms min spacing: 1st immediate, 2nd ≥200ms, 3rd ≥400ms.
	if elapsed < 400*time.Millisecond {
		t.Errorf("expected at least 400ms between three requests, got %v", elapsed)
	}
}

func TestRateLimitedTransport_Retries429(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	tr := newRateLimitedTransport(50*time.Millisecond, 3)
	client := &http.Client{Transport: tr}

	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 after retries, got %d", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("expected 3 server calls (2 × 429 then 200), got %d", got)
	}
}

func TestRateLimitedTransport_GivesUpAfterMaxRetries(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	tr := newRateLimitedTransport(20*time.Millisecond, 2)
	client := &http.Client{Transport: tr}

	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected 429 after exhausting retries, got %d", resp.StatusCode)
	}
	// maxRetries=2 means the request is attempted up to 3 times (initial + 2 retries).
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("expected 3 server calls (initial + 2 retries), got %d", got)
	}
}

func TestRateLimitedTransport_ReplaysBodyOnRetry(t *testing.T) {
	var bodies []string
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		buf := make([]byte, 32)
		n2, _ := r.Body.Read(buf)
		bodies = append(bodies, string(buf[:n2]))
		if n < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	tr := newRateLimitedTransport(20*time.Millisecond, 2)
	client := &http.Client{Transport: tr}

	resp, err := client.Post(srv.URL, "text/plain", strings.NewReader("payload"))
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	defer resp.Body.Close()

	if len(bodies) != 2 {
		t.Fatalf("expected 2 server calls capturing body, got %d", len(bodies))
	}
	for i, b := range bodies {
		if b != "payload" {
			t.Errorf("attempt %d: body = %q, want %q", i, b, "payload")
		}
	}
}

func TestRateLimitedTransport_ClampsMinInterval(t *testing.T) {
	tr := newRateLimitedTransport(0, 0)
	if tr.minInterval < time.Millisecond {
		t.Errorf("minInterval = %v, expected clamp to ≥1ms", tr.minInterval)
	}
}
