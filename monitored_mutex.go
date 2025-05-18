package glockmon

import (
	"sync"
	"time"
)

// MonitoredMutex is a mutex wrapper that tracks lock wait times and reports
// long locks to a Monitor if the lock duration exceeds a configured threshold.
//
// It stores the stack trace of where the lock happened and associates it with a category label.
//
// Use New to create an instance.
type MonitoredMutex struct {
	mu        sync.Mutex
	threshold time.Duration
	monitor   *Monitor
	lockedKey uint64
	category  string
}

// New creates a new MonitoredMutex that reports to the given Monitor.
// threshold sets the minimum duration a Lock must wait before it is considered "long".
//
// monitor: the Monitor instance to report long lock info to.
// threshold: minimum lock wait time to trigger monitoring.
func New(monitor *Monitor, threshold time.Duration) *MonitoredMutex {
	return &MonitoredMutex{
		threshold: threshold,
		monitor:   monitor,
		category:  monitor.defaultCategory,
	}
}

// SetCategory changes the category label for this mutex's lock events.
// Use this to group locks by logical categories.
//
// If an empty string is passed, the default category from the Monitor is used.
func (m *MonitoredMutex) SetCategory(category string) {
	if category == "" {
		m.category = m.monitor.defaultCategory
	} else {
		m.category = category
	}
}

// Lock attempts to acquire the mutex and measures the wait time.
//
// If the wait exceeds the configured threshold, the lock event with
// stack trace and category info is reported to the Monitor.
//
// This helps to identify and track long lock occurrences.
func (m *MonitoredMutex) Lock() {
	start := time.Now()

	m.mu.Lock()

	wait := time.Since(start)
	if wait > m.threshold && m.monitor != nil {
		stack := getStackTrace()

		key := hashStack(stack)
		m.lockedKey = key

		m.monitor.Add(stack, LockInfo{
			Timestamp: time.Now(),
			Wait:      wait,
			Stack:     stack,
			Category:  m.category,
		})
	} else {
		m.lockedKey = 0
	}
}

// Unlock releases the mutex.
//
// If the Monitor does not keep records, it removes the tracked lock info for this mutex
// after unlocking, cleaning up memory for short-lived lock records.
func (m *MonitoredMutex) Unlock() {
	if m.monitor != nil && !m.monitor.keepRecords && m.lockedKey != 0 {
		if stack, ok := m.monitor.GetStack(m.lockedKey); ok {
			m.monitor.RemoveByStack(stack)
		}
		m.lockedKey = 0
	}
	m.mu.Unlock()
}
