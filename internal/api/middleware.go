package api

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// secureHeaders applies OWASP-recommended response headers.
// The CSP allows only same-origin scripts/styles/connections — the app is
// fully self-contained, so nothing external is ever legitimate.
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data:; connect-src 'self'; font-src 'self'; "+
				"object-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		// HSTS is meaningful only over TLS; hosts terminate TLS in front of us.
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

// rateLimit is a per-IP token bucket: burst of `burst`, refilling at `rps`
// tokens per second. Buckets idle for 10 minutes are evicted.
func rateLimit(rps float64, burst float64, next http.Handler) http.Handler {
	type bucket struct {
		tokens float64
		last   time.Time
	}
	var (
		mu      sync.Mutex
		buckets = map[string]*bucket{}
	)

	// periodic eviction so the map cannot grow unboundedly
	go func() {
		for range time.Tick(time.Minute) {
			mu.Lock()
			for ip, b := range buckets {
				if time.Since(b.last) > 10*time.Minute {
					delete(buckets, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		now := time.Now()

		mu.Lock()
		b, ok := buckets[ip]
		if !ok {
			b = &bucket{tokens: burst, last: now}
			buckets[ip] = b
		}
		b.tokens = min(burst, b.tokens+now.Sub(b.last).Seconds()*rps)
		b.last = now
		allowed := b.tokens >= 1
		if allowed {
			b.tokens--
		}
		mu.Unlock()

		if !allowed {
			w.Header().Set("Retry-After", "10")
			writeJSON(w, http.StatusTooManyRequests, map[string]string{
				"error": "Too many requests — please slow down.",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP prefers the first X-Forwarded-For hop (set by the hosting proxy),
// falling back to the socket address.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if first, _, ok := strings.Cut(xff, ","); ok {
			return strings.TrimSpace(first)
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// bodyLimit caps request bodies (the only body-accepting endpoint is a tiny
// JSON skill list, so 8 KiB is generous).
func bodyLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
		}
		next.ServeHTTP(w, r)
	})
}

// requestTimeout bounds each request's total time so a slow database cannot
// pile up goroutines under load.
func requestTimeout(d time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), d)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
