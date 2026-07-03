// Package balancer implements backend selection strategies for the proxy.
package balancer

import (
	"net/url"
	"sync"
	"sync/atomic"
)

// Backend represents a single upstream server the proxy can forward to.
type Backend struct {
	URL   *url.URL
	Alive atomic.Bool
}

// SetAlive updates whether this backend is currently considered healthy.
func (b *Backend) SetAlive(alive bool) {
	b.Alive.Store(alive)
}

// IsAlive reports whether this backend is currently considered healthy.
func (b *Backend) IsAlive() bool {
	return b.Alive.Load()
}

// RoundRobin is a thread-safe round-robin load balancer over a fixed set
// of backends. It skips backends currently marked unhealthy.
type RoundRobin struct {
	mu       sync.Mutex
	backends []*Backend
	current  int
}

// NewRoundRobin builds a round-robin balancer from a list of backend URLs.
// All backends start out marked alive; a health checker can update that
// later via Backends().
func NewRoundRobin(rawURLs []string) (*RoundRobin, error) {
	backends := make([]*Backend, 0, len(rawURLs))
	for _, raw := range rawURLs {
		u, err := url.Parse(raw)
		if err != nil {
			return nil, err
		}
		b := &Backend{URL: u}
		b.SetAlive(true)
		backends = append(backends, b)
	}
	return &RoundRobin{backends: backends}, nil
}

// Backends returns the full backend list, e.g. for health checking.
func (rr *RoundRobin) Backends() []*Backend {
	return rr.backends
}

// Next returns the next healthy backend in rotation, or nil if none are
// currently alive.
func (rr *RoundRobin) Next() *Backend {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	n := len(rr.backends)
	if n == 0 {
		return nil
	}

	for i := 0; i < n; i++ {
		rr.current = (rr.current + 1) % n
		b := rr.backends[rr.current]
		if b.IsAlive() {
			return b
		}
	}
	// No backend is currently alive.
	return nil
}
