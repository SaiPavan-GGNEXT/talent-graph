package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"talentgraph/internal/graph"
)

type Server struct {
	store     *graph.Store
	staticDir string
}

func NewServer(store *graph.Store, staticDir string) *Server {
	return &Server{store: store, staticDir: staticDir}
}

// Routes wires every endpoint using Go 1.22+ pattern routing, behind a
// hardening chain: security headers → rate limit → body cap → timeout.
// Read-heavy endpoints sit behind a short-TTL cache with request coalescing
// so bursts of traffic don't overwhelm the small free-tier database.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	// The dataset only changes when the seed script runs, so a long TTL is
	// safe; the prewarmer below renews entries before they expire, keeping
	// the hot endpoints at cache-hit latency permanently.
	cache := newTTLCache(15 * time.Minute)

	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/stats", cache.wrap(s.handleStats))
	mux.HandleFunc("GET /api/people", cache.wrap(s.handlePeople))
	mux.HandleFunc("GET /api/people/{id}", cache.wrap(s.handlePerson))
	mux.HandleFunc("GET /api/skills", cache.wrap(s.handleSkills))
	mux.HandleFunc("GET /api/graph", cache.wrap(s.handleGraph))
	mux.HandleFunc("GET /api/experts", cache.wrap(s.handleExperts))
	mux.HandleFunc("GET /api/skills/adjacent", cache.wrap(s.handleAdjacentSkills))
	mux.HandleFunc("GET /api/path", cache.wrap(s.handleIntroPath))
	mux.HandleFunc("POST /api/team-plan", s.handleTeamPlan)

	mux.HandleFunc("/", s.handleStatic)

	cache.prewarm(mux, []string{"/api/stats", "/api/people", "/api/skills", "/api/graph"}, 10*time.Minute)

	var h http.Handler = mux
	h = requestTimeout(15*time.Second, h)
	h = bodyLimit(h)
	h = rateLimit(2, 60, h) // 2 req/s sustained, burst of 60 per IP
	h = secureHeaders(h)
	return logging(h)
}

// maxParam guards against absurd query parameter values.
const maxParam = 100

func validParam(v string) bool {
	return len(v) <= maxParam
}

// ---- handlers ----------------------------------------------------------------

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Ping(r.Context()); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.store.Stats(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) handlePeople(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if !validParam(q) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query too long"})
		return
	}
	var (
		people []graph.PersonSummary
		err    error
	)
	if q == "" {
		people, err = s.store.ListPeople(r.Context())
	} else {
		people, err = s.store.SearchPeople(r.Context(), q)
	}
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, people)
}

func (s *Server) handlePerson(w http.ResponseWriter, r *http.Request) {
	detail, err := s.store.GetPerson(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	if detail == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "person not found"})
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	skills, err := s.store.ListSkills(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, skills)
}

func (s *Server) handleExperts(w http.ResponseWriter, r *http.Request) {
	skill := strings.TrimSpace(r.URL.Query().Get("skill"))
	from := strings.TrimSpace(r.URL.Query().Get("from"))
	if skill == "" || !validParam(skill) || !validParam(from) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query parameter 'skill' is required"})
		return
	}
	experts, err := s.store.Experts(r.Context(), skill, from)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, experts)
}

func (s *Server) handleAdjacentSkills(w http.ResponseWriter, r *http.Request) {
	skill := strings.TrimSpace(r.URL.Query().Get("skill"))
	if skill == "" || !validParam(skill) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query parameter 'skill' is required"})
		return
	}
	adjacent, err := s.store.AdjacentSkills(r.Context(), skill)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, adjacent)
}

func (s *Server) handleIntroPath(w http.ResponseWriter, r *http.Request) {
	from := strings.TrimSpace(r.URL.Query().Get("from"))
	to := strings.TrimSpace(r.URL.Query().Get("to"))
	if from == "" || to == "" || !validParam(from) || !validParam(to) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query parameters 'from' and 'to' are required"})
		return
	}
	path, err := s.store.IntroPath(r.Context(), from, to)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"path": path, "connected": len(path) > 0})
}

func (s *Server) handleTeamPlan(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Skills []string `json:"skills"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Skills) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "body must be {\"skills\": [\"…\"]} with at least one skill"})
		return
	}
	if len(body.Skills) > 10 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "at most 10 skills"})
		return
	}
	for _, sk := range body.Skills {
		if !validParam(sk) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "skill name too long"})
			return
		}
	}
	plan, err := s.store.TeamPlan(r.Context(), body.Skills)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, plan)
}

func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	view, err := s.store.GraphView(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

// handleStatic serves the built React app with an SPA fallback to index.html.
// Vite emits content-hashed filenames under /assets/, so those are immutable
// and cacheable forever; index.html must always revalidate.
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	path := filepath.Join(s.staticDir, filepath.Clean("/"+r.URL.Path))
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		http.ServeFile(w, r, path)
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, filepath.Join(s.staticDir, "index.html"))
}

// ---- helpers -------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

// writeErr maps database-unreachable errors to 503 so the frontend can show
// a dedicated "database offline" state instead of a generic failure.
func writeErr(w http.ResponseWriter, err error) {
	log.Printf("request failed: %v", err)
	if errors.Is(err, graph.ErrUnavailable) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "The graph database is unreachable right now. Please try again shortly.",
		})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{
		"error": "Something went wrong on our side.",
	})
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		if strings.HasPrefix(r.URL.Path, "/api/") {
			log.Printf("%s %s", r.Method, r.URL.RequestURI())
		}
	})
}
