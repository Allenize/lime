package balancer

import (
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
)

type Backend struct {
	URL   *url.URL
	Alive atomic.Bool
	Conns atomic.Int64
}

func (b *Backend) SetAlive(alive bool) {
	b.Alive.Store(alive)
}

func (b *Backend) IsAlive() bool {
	return b.Alive.Load()
}

func (b *Backend) IncConn() {
	b.Conns.Add(1)
}

func (b *Backend) DecConn() {
	b.Conns.Add(-1)
}

func (b *Backend) ConnCount() int64 {
	return b.Conns.Load()
}

type Balancer interface {
	Next() *Backend
	Backends() []*Backend
}

func parseBackends(rawURLs []string) ([]*Backend, error) {
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
	return backends, nil
}

type RoundRobin struct {
	mu       sync.Mutex
	backends []*Backend
	current  int
}

func NewRoundRobin(rawURLs []string) (*RoundRobin, error) {
	backends, err := parseBackends(rawURLs)
	if err != nil {
		return nil, err
	}
	return &RoundRobin{backends: backends}, nil
}

func (rr *RoundRobin) Backends() []*Backend {
	return rr.backends
}

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
	return nil
}

type LeastConnections struct {
	mu       sync.Mutex
	backends []*Backend
}

func NewLeastConnections(rawURLs []string) (*LeastConnections, error) {
	backends, err := parseBackends(rawURLs)
	if err != nil {
		return nil, err
	}
	return &LeastConnections{backends: backends}, nil
}

func (lc *LeastConnections) Backends() []*Backend {
	return lc.backends
}

func (lc *LeastConnections) Next() *Backend {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	var best *Backend
	for _, b := range lc.backends {
		if !b.IsAlive() {
			continue
		}
		if best == nil || b.ConnCount() < best.ConnCount() {
			best = b
		}
	}
	return best
}

func New(strategy string, rawURLs []string) (Balancer, error) {
	switch strategy {
	case "", "round_robin":
		return NewRoundRobin(rawURLs)
	case "least_conn":
		return NewLeastConnections(rawURLs)
	default:
		return nil, fmt.Errorf("unknown balancing strategy %q", strategy)
	}
}
