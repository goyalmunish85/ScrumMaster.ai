package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiter(t *testing.T) {
	// Create a rate limiter allowing 1 request per second with a burst of 2.
	rl := NewRateLimiter(1, 2)

	// Create a dummy handler that returns 200 OK.
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the dummy handler with the rate limiter.
	handler := rl.Handler(dummyHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	// Request 1: Should be allowed (uses 1 burst token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Request 2: Should be allowed (uses 2nd burst token)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Request 3: Should be blocked (exceeded burst)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected status %d, got %d", http.StatusTooManyRequests, rr.Code)
	}

	// Test another IP
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "10.0.0.1:4321"

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req2)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d for new IP, got %d", http.StatusOK, rr.Code)
	}
}
