package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Piszmog/pathwise/ui/server/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		requestCount   int
		waitBetween    time.Duration
	}{
		{
			name:           "auth endpoint allows requests within limit",
			method:         http.MethodPost,
			path:           "/signin",
			expectedStatus: http.StatusOK,
			requestCount:   1,
		},
		{
			name:           "auth endpoint blocks requests over limit",
			method:         http.MethodPost,
			path:           "/signin",
			expectedStatus: http.StatusTooManyRequests,
			requestCount:   10,
		},
		{
			name:           "write endpoint allows requests within limit",
			method:         http.MethodPost,
			path:           "/jobs",
			expectedStatus: http.StatusOK,
			requestCount:   5,
		},
		{
			name:           "read endpoint allows requests within limit",
			method:         http.MethodGet,
			path:           "/jobs",
			expectedStatus: http.StatusOK,
			requestCount:   10,
		},
		{
			name:           "assets are not rate limited",
			method:         http.MethodGet,
			path:           "/assets/css/style.css",
			expectedStatus: http.StatusOK,
			requestCount:   50,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rateLimiter := middleware.NewRateLimitMiddleware()
			defer rateLimiter.Close()

			handler := rateLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			var lastStatus int
			for i := 0; i < test.requestCount; i++ {
				req := httptest.NewRequest(test.method, test.path, nil)
				req.RemoteAddr = "192.168.1.1:8080"
				w := httptest.NewRecorder()

				handler.ServeHTTP(w, req)
				lastStatus = w.Code

				if test.waitBetween > 0 {
					time.Sleep(test.waitBetween)
				}
			}

			assert.Equal(t, test.expectedStatus, lastStatus)
		})
	}
}

func TestRateLimitHeaders(t *testing.T) {
	t.Parallel()

	rateLimiter := middleware.NewRateLimitMiddleware()
	defer rateLimiter.Close()

	handler := rateLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/signin", nil)
	req.RemoteAddr = "192.168.1.1:8080"

	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code == http.StatusTooManyRequests {
			assert.NotEmpty(t, w.Header().Get("Retry-After"))
			assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
			assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
			break
		}
	}
}

func TestRateLimitDifferentIPs(t *testing.T) {
	t.Parallel()

	rateLimiter := middleware.NewRateLimitMiddleware()
	defer rateLimiter.Close()

	handler := rateLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ips := []string{"192.168.1.1:8080", "192.168.1.2:8080", "192.168.1.3:8080"}

	for _, ip := range ips {
		req := httptest.NewRequest(http.MethodPost, "/signin", nil)
		req.RemoteAddr = ip
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestRateLimitXForwardedFor(t *testing.T) {
	t.Parallel()

	rateLimiter := middleware.NewRateLimitMiddleware()
	defer rateLimiter.Close()

	handler := rateLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/signin", nil)
	req.RemoteAddr = "10.0.0.1:8080"
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")

	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code == http.StatusTooManyRequests {
			break
		}
	}

	newReq := httptest.NewRequest(http.MethodPost, "/signin", nil)
	newReq.RemoteAddr = "10.0.0.1:8080"

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, newReq)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimitConcurrentRequests(t *testing.T) {
	t.Parallel()

	rateLimiter := middleware.NewRateLimitMiddleware()
	defer rateLimiter.Close()

	handler := rateLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	var wg sync.WaitGroup
	results := make([]int, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/signin", nil)
			req.RemoteAddr = "192.168.1.1:8080"
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)
			results[index] = w.Code
		}(i)
	}

	wg.Wait()

	okCount := 0
	rateLimitedCount := 0
	for _, code := range results {
		if code == http.StatusOK {
			okCount++
		} else if code == http.StatusTooManyRequests {
			rateLimitedCount++
		}
	}

	assert.True(t, okCount >= 1, "should allow at least burst requests")
	assert.True(t, rateLimitedCount > 0, "should rate limit excess requests")
}

func TestRateLimitEndpointTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		method          string
		path            string
		burstRequests   int
		shouldBeBlocked bool
	}{
		{
			name:            "auth endpoint has low burst",
			method:          http.MethodPost,
			path:            "/signin",
			burstRequests:   2,
			shouldBeBlocked: true,
		},
		{
			name:            "write endpoint has medium burst",
			method:          http.MethodPost,
			path:            "/jobs",
			burstRequests:   6,
			shouldBeBlocked: true,
		},
		{
			name:            "read endpoint has high burst",
			method:          http.MethodGet,
			path:            "/jobs",
			burstRequests:   11,
			shouldBeBlocked: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rateLimiter := middleware.NewRateLimitMiddleware()
			defer rateLimiter.Close()

			handler := rateLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			var lastStatus int
			for i := 0; i < test.burstRequests; i++ {
				req := httptest.NewRequest(test.method, test.path, nil)
				req.RemoteAddr = "192.168.1.1:8080"
				w := httptest.NewRecorder()

				handler.ServeHTTP(w, req)
				lastStatus = w.Code
			}

			if test.shouldBeBlocked {
				assert.Equal(t, http.StatusTooManyRequests, lastStatus)
			} else {
				assert.Equal(t, http.StatusOK, lastStatus)
			}
		})
	}
}

func TestRateLimitCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cleanup test in short mode")
	}

	rateLimiter := middleware.NewRateLimitMiddleware()
	defer rateLimiter.Close()

	handler := rateLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	req.RemoteAddr = "192.168.1.1:8080"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	time.Sleep(100 * time.Millisecond)

	req2 := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	req2.RemoteAddr = "192.168.1.1:8080"
	w2 := httptest.NewRecorder()

	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}
