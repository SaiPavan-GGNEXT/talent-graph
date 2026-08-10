package api

import (
	"bytes"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// ttlCache memoises hot read-only endpoints. The dataset changes only when
// the seed script runs, so a short TTL keeps the app fresh while shielding
// the small free-tier database from request bursts. singleflight collapses
// concurrent misses for the same key into one upstream call (no stampedes).
type ttlCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	group   singleflight.Group
	ttl     time.Duration
}

type cacheEntry struct {
	status  int
	body    []byte
	expires time.Time
}

func newTTLCache(ttl time.Duration) *ttlCache {
	c := &ttlCache{entries: map[string]cacheEntry{}, ttl: ttl}
	go func() {
		for range time.Tick(ttl) {
			now := time.Now()
			c.mu.Lock()
			for k, e := range c.entries {
				if now.After(e.expires) {
					delete(c.entries, k)
				}
			}
			c.mu.Unlock()
		}
	}()
	return c
}

// recorder captures a handler's response so it can be stored.
type recorder struct {
	header http.Header
	status int
	buf    bytes.Buffer
}

func (r *recorder) Header() http.Header        { return r.header }
func (r *recorder) WriteHeader(status int)     { r.status = status }
func (r *recorder) Write(p []byte) (int, error) { return r.buf.Write(p) }

// wrap serves GET responses from cache, recording on miss. Only 200s are
// cached, so errors and 503s always reflect live state.
func (c *ttlCache) wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.RequestURI()

		c.mu.RLock()
		entry, ok := c.entries[key]
		c.mu.RUnlock()
		if ok && time.Now().Before(entry.expires) {
			serveCached(w, entry, "HIT")
			return
		}

		result, _, _ := c.group.Do(key, func() (any, error) {
			rec := &recorder{header: http.Header{}, status: http.StatusOK}
			next(rec, r)
			e := cacheEntry{status: rec.status, body: rec.buf.Bytes(), expires: time.Now().Add(c.ttl)}
			if rec.status == http.StatusOK {
				c.mu.Lock()
				c.entries[key] = e
				c.mu.Unlock()
			}
			return e, nil
		})
		serveCached(w, result.(cacheEntry), "MISS")
	}
}

func serveCached(w http.ResponseWriter, e cacheEntry, state string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", state)
	w.WriteHeader(e.status)
	w.Write(e.body)
}
