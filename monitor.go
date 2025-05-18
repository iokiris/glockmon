package glockmon

import (
	"github.com/iokiris/glockmon/config"
	"sync"
	"time"
)

// LockInfo holds information about a mutex lock event.
// It tracks when the lock was acquired, how long we waited,
// the stack trace at that moment, and the category for grouping.
//
// Timestamp: when the lock was taken.
// Wait: how long it took to get the lock.
// Stack: the call stack as a string.
// Category: a label to group similar locks.
type LockInfo struct {
	Timestamp time.Time
	Wait      time.Duration
	Stack     string
	Category  string
}

// CategoryStats keeps summary data about locks within a category.
//
// Count: number of recorded locks.
// TotalWait: total wait time summed across all locks.
// AverageWait: average wait time per lock.
type CategoryStats struct {
	Count       int
	TotalWait   time.Duration
	AverageWait time.Duration
}

// Monitor tracks long mutex locks, saving their info,
// including stack traces and wait times, and maintains stats by category.
type Monitor struct {
	mu              sync.Mutex
	longLocks       map[uint64]LockInfo       // key: stack hash, value: lock info
	stackCache      map[uint64]string         // key: stack hash, value: stack string
	keepRecords     bool                      // whether to keep records or not
	defaultCategory string                    // fallback category name
	categoryStats   map[string]*CategoryStats // stats per category
}

// NewMonitor creates and returns a new Monitor instance configured by cfg.
// It also starts an HTTP server for monitoring endpoints based on cfg settings.
//
// cfg: configuration options including whether to keep records, default category,
//
//	HTTP server address, and HTTP endpoint paths.
//
// The returned Monitor is ready to track lock statistics and serves HTTP requests
// providing information about lock events and categories.
//
// The HTTP server runs in a separate goroutine and listens on the configured address.
func NewMonitor(cfg *config.MonitorConfig) *Monitor {
	monitor := &Monitor{
		longLocks:       make(map[uint64]LockInfo),
		stackCache:      make(map[uint64]string),
		categoryStats:   make(map[string]*CategoryStats),
		keepRecords:     cfg.KeepRecords,
		defaultCategory: cfg.DefaultCategory,
	}

	server := NewHTTPServer(cfg, monitor)
	server.Start()

	return monitor
}

// Add records a new long lock event.
//
// stack: the call stack at lock time.
// info: details about the lock (timestamp, wait duration, category).
//
// Updates internal stats and caches the stack for later reference.
func (m *Monitor) Add(stack string, info LockInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	h := hashStack(stack)
	info.Stack = stack
	m.longLocks[h] = info
	m.stackCache[h] = stack

	stats, exists := m.categoryStats[info.Category]
	if !exists {
		stats = &CategoryStats{}
		m.categoryStats[info.Category] = stats
	}

	stats.Count++
	stats.TotalWait += info.Wait
	stats.AverageWait = stats.TotalWait / time.Duration(stats.Count)
}

// RemoveByStack deletes a lock record given its stack trace.
//
// stack: the stack trace string to identify the record.
//
// If found, removes it from the internal maps.
func (m *Monitor) RemoveByStack(stack string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	h := hashStack(stack)
	delete(m.longLocks, h)
	delete(m.stackCache, h)
}

// Snapshot returns a copy of all current long lock records.
//
// Returns a map keyed by stack hash, with LockInfo as values.
func (m *Monitor) Snapshot() map[uint64]LockInfo {
	m.mu.Lock()
	defer m.mu.Unlock()

	copyMap := make(map[uint64]LockInfo, len(m.longLocks))
	for k, v := range m.longLocks {
		copyMap[k] = v
	}
	return copyMap
}

// GetStack fetches the stack trace string for a given hash.
//
// hash: the hash key of the stack trace.
//
// Returns the stack string and true if found, or empty string and false if not.
func (m *Monitor) GetStack(hash uint64) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.stackCache[hash]
	return s, ok
}

// GetStackCache returns a copy of the full stack trace cache.
//
// Useful if you want to inspect or export all cached stacks.
//
// Returns a map of stack hashes to stack strings.
func (m *Monitor) GetStackCache() map[uint64]string {
	m.mu.Lock()
	defer m.mu.Unlock()

	copyCache := make(map[uint64]string, len(m.stackCache))
	for k, v := range m.stackCache {
		copyCache[k] = v
	}
	return copyCache
}

// GetCategoryStats returns a snapshot of current lock statistics by category.
//
// Returns a map with category names as keys and their aggregated stats.
func (m *Monitor) GetCategoryStats() map[string]CategoryStats {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make(map[string]CategoryStats, len(m.categoryStats))
	for cat, stats := range m.categoryStats {
		result[cat] = *stats
	}
	return result
}
