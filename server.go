package glockmon

import (
	"encoding/json"
	"errors"
	"github.com/iokiris/glockmon/config"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// HTTPServer manages the lifecycle of an HTTP server exposing monitoring data.
//
// It serves three main endpoints:
//
//	GET /blocked    - returns JSON list of currently tracked long locks
//	GET /stacks/{id} - returns the stack trace text for a given lock id
//	GET /categories - returns JSON statistics aggregated by categories
//
// This server is intended for internal monitoring and debugging purposes.
//
// Create an instance with NewHTTPServer and call Start() to run it in background.
type HTTPServer struct {
	addr    string
	monitor *Monitor
	server  *http.Server
	mu      sync.Mutex
	running bool

	endpoints struct {
		Blocked    string
		Categories string
		Stack      string
	}
}

// BlockedEntry represents a single long lock event returned by the config.HTTPConfig endpoint.
type BlockedEntry struct {
	ID        uint64 `json:"id,string"`
	Category  string `json:"category"`
	WaitMs    int64  `json:"wait_ms"`
	Timestamp int64  `json:"timestamp,string"` // UnixNano
}

// CategoryStatsResponse represents aggregated lock stats by category returned by the config.HTTPConfig endpoint.
type CategoryStatsResponse struct {
	Category    string  `json:"category"`
	Count       int     `json:"count"`
	AverageWait float64 `json:"average_wait_ms"`
	TotalWait   int64   `json:"total_wait_ms"`
}

// NewHTTPServer creates a new HTTPServer from the Monitor instance.
func NewHTTPServer(cfg *config.MonitorConfig, monitor *Monitor) *HTTPServer {
	s := &HTTPServer{
		addr:    cfg.HTTPServerAddr,
		monitor: monitor,
	}

	s.endpoints.Blocked = cfg.HTTPEndpoints.Blocked
	s.endpoints.Categories = cfg.HTTPEndpoints.Categories
	s.endpoints.Stack = cfg.HTTPEndpoints.Stack

	return s

}

// Start runs http server from the Monitor instance.
func (s *HTTPServer) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc(s.endpoints.Blocked, s.handleBlocked)
	mux.HandleFunc(s.endpoints.Categories, s.handleCategories)
	mux.HandleFunc(s.endpoints.Stack, s.handleStack)

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	s.running = true
	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
		}
	}()
	log.Printf("HTTP monitoring server started on %s", s.addr)
}

// handleBlocked returns a JSON list of all currently tracked long lock events.
//
// Each entry contains ID, category, wait time in milliseconds, timestamp (UnixNano), and stack ID.
func (s *HTTPServer) handleBlocked(w http.ResponseWriter, r *http.Request) {
	locksMap := s.monitor.Snapshot()

	var entries []BlockedEntry
	for id, info := range locksMap {
		entries = append(entries, BlockedEntry{
			ID:        id,
			Category:  info.Category,
			WaitMs:    info.Wait.Milliseconds(),
			Timestamp: info.Timestamp.UnixNano(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		http.Error(w, "failed to encode JSON", http.StatusInternalServerError)
	}
}

// handleStack returns the raw stack trace text for a given stack ID.
//
// The stack ID must be specified in the URL path as /stacks/{id}.
//
// Responds with 400 Bad Request if the ID is invalid,
// or 404 Not Found if no stack trace exists for that ID.
func (s *HTTPServer) handleStack(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/stacks/")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid stack id", http.StatusBadRequest)
		return
	}

	stackCache := s.monitor.GetStackCache()
	stack, ok := stackCache[id]
	if !ok {
		http.Error(w, "stack not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(stack))
}

// handleCategories returns JSON aggregated statistics by lock category.
//
// Each item contains the category name, total count of lock events,
// average wait time (in milliseconds), and total wait time (in milliseconds).
func (s *HTTPServer) handleCategories(w http.ResponseWriter, r *http.Request) {
	statsMap := s.monitor.GetCategoryStats()

	var stats []CategoryStatsResponse
	for cat, cs := range statsMap {
		stats = append(stats, CategoryStatsResponse{
			Category:    cat,
			Count:       cs.Count,
			AverageWait: float64(cs.AverageWait.Milliseconds()),
			TotalWait:   cs.TotalWait.Milliseconds(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "failed to encode JSON", http.StatusInternalServerError)
	}
}
