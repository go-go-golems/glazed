package publish

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const requestIDHeader = "X-Request-ID"

type contextKey string

const requestIDContextKey contextKey = "docs-registry-request-id"

func withRequestIDValue(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

func requestIDFromContext(ctx context.Context) string {
	if value, ok := ctx.Value(requestIDContextKey).(string); ok {
		return value
	}
	return ""
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
		r.ResponseWriter.WriteHeader(status)
	}
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(data)
	r.bytes += n
	return n, err
}

func (r *statusRecorder) statusCode() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := strings.TrimSpace(r.Header.Get(requestIDHeader))
		if requestID == "" {
			requestID = newRequestID()
		}
		w.Header().Set(requestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(withRequestIDValue(r.Context(), requestID)))
	})
}

func withAccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		slog.Info("docs registry request",
			"request_id", requestIDFromContext(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"route_class", registryRouteClass(r),
			"status", recorder.statusCode(),
			"response_bytes", recorder.bytes,
			"duration_ms", time.Since(start).Milliseconds(),
			"client_ip", clientIP(r),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(b[:])
}

func clientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if first := strings.TrimSpace(parts[0]); first != "" {
			return first
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

func registryRouteClass(r *http.Request) string {
	if r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/v1/packages/") && strings.HasSuffix(r.URL.Path, "/sqlite") {
		return "publish"
	}
	if r.Method == http.MethodGet && r.URL.Path == "/v1/packages" {
		return "list"
	}
	if r.Method == http.MethodGet && r.URL.Path == "/healthz" {
		return "health"
	}
	return "other"
}

type rateBucket struct {
	tokens float64
	last   time.Time
}

type SimpleRateLimiter struct {
	mu       sync.Mutex
	rate     float64
	burst    float64
	now      func() time.Time
	buckets  map[string]*rateBucket
	lastSeen map[string]time.Time
}

func NewSimpleRateLimiter(requestsPerMinute, burst int) *SimpleRateLimiter {
	if requestsPerMinute <= 0 || burst <= 0 {
		return nil
	}
	return &SimpleRateLimiter{
		rate:     float64(requestsPerMinute) / 60.0,
		burst:    float64(burst),
		now:      time.Now,
		buckets:  map[string]*rateBucket{},
		lastSeen: map[string]time.Time{},
	}
}

func (l *SimpleRateLimiter) Allow(key string) bool {
	if l == nil {
		return true
	}
	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()

	bucket := l.buckets[key]
	if bucket == nil {
		bucket = &rateBucket{tokens: l.burst, last: now}
		l.buckets[key] = bucket
	}
	elapsed := now.Sub(bucket.last).Seconds()
	bucket.tokens += elapsed * l.rate
	if bucket.tokens > l.burst {
		bucket.tokens = l.burst
	}
	bucket.last = now
	l.lastSeen[key] = now
	if bucket.tokens < 1 {
		return false
	}
	bucket.tokens--
	return true
}

func withRateLimit(next http.Handler, limiter *SimpleRateLimiter) http.Handler {
	if limiter == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := clientIP(r) + ":" + registryRouteClass(r)
		if !limiter.Allow(key) {
			writeRegistryError(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
			return
		}
		next.ServeHTTP(w, r)
	})
}
