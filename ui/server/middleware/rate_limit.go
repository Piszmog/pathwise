package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimitMiddleware struct {
	limiters      sync.Map
	authLimit     rate.Limit
	writeLimit    rate.Limit
	readLimit     rate.Limit
	authBurst     int
	writeBurst    int
	readBurst     int
	cleanupTicker *time.Ticker
	done          chan struct{}
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewRateLimitMiddleware() *RateLimitMiddleware {
	rl := &RateLimitMiddleware{
		authLimit:  rate.Limit(5) / 60,
		writeLimit: rate.Limit(30) / 60,
		readLimit:  rate.Limit(100) / 60,
		authBurst:  1,
		writeBurst: 5,
		readBurst:  10,
		done:       make(chan struct{}),
	}

	rl.cleanupTicker = time.NewTicker(10 * time.Minute)
	go rl.cleanup()

	return rl
}

func (rl *RateLimitMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			next.ServeHTTP(w, r)
			return
		}

		ip := getClientIP(r)
		limiter := rl.getLimiter(ip, r.Method, r.URL.Path)

		if !limiter.Allow() {
			retryAfter := time.Second / time.Duration(rl.getLimitForRequest(r.Method, r.URL.Path))
			w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			w.Header().Set("X-RateLimit-Limit", rl.getLimitHeader(r.Method, r.URL.Path))
			w.Header().Set("X-RateLimit-Remaining", "0")
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimitMiddleware) getLimiter(ip, method, path string) *rate.Limiter {
	key := fmt.Sprintf("%s:%s:%s", ip, method, rl.getEndpointType(method, path))

	now := time.Now()
	if entry, exists := rl.limiters.Load(key); exists {
		rateLimiterEntry := entry.(*rateLimiterEntry)
		rateLimiterEntry.lastSeen = now
		return rateLimiterEntry.limiter
	}

	limit, burst := rl.getLimitAndBurst(method, path)
	limiter := rate.NewLimiter(limit, burst)
	rl.limiters.Store(key, &rateLimiterEntry{
		limiter:  limiter,
		lastSeen: now,
	})

	return limiter
}

func (rl *RateLimitMiddleware) getLimitAndBurst(method, path string) (rate.Limit, int) {
	if rl.isAuthEndpoint(path) {
		return rl.authLimit, rl.authBurst
	}
	if method == http.MethodPost || method == http.MethodPatch || method == http.MethodDelete {
		return rl.writeLimit, rl.writeBurst
	}
	return rl.readLimit, rl.readBurst
}

func (rl *RateLimitMiddleware) getLimitForRequest(method, path string) rate.Limit {
	if rl.isAuthEndpoint(path) {
		return rl.authLimit * 60
	}
	if method == http.MethodPost || method == http.MethodPatch || method == http.MethodDelete {
		return rl.writeLimit * 60
	}
	return rl.readLimit * 60
}

func (rl *RateLimitMiddleware) getLimitHeader(method, path string) string {
	limit := rl.getLimitForRequest(method, path)
	return strconv.Itoa(int(limit))
}

func (rl *RateLimitMiddleware) getEndpointType(method, path string) string {
	if rl.isAuthEndpoint(path) {
		return "auth"
	}
	if method == http.MethodPost || method == http.MethodPatch || method == http.MethodDelete {
		return "write"
	}
	return "read"
}

func (rl *RateLimitMiddleware) isAuthEndpoint(path string) bool {
	return path == "/signin" || path == "/signup"
}

func (rl *RateLimitMiddleware) cleanup() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.cleanupExpiredLimiters()
		case <-rl.done:
			rl.cleanupTicker.Stop()
			return
		}
	}
}

func (rl *RateLimitMiddleware) cleanupExpiredLimiters() {
	cutoff := time.Now().Add(-30 * time.Minute)
	rl.limiters.Range(func(key, value interface{}) bool {
		entry := value.(*rateLimiterEntry)
		if entry.lastSeen.Before(cutoff) {
			rl.limiters.Delete(key)
		}
		return true
	})
}

func (rl *RateLimitMiddleware) Close() {
	close(rl.done)
}

func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	xri := r.Header.Get("X-Real-Ip")
	if xri != "" {
		return xri
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
