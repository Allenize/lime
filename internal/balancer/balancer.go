package balancer

import (
	"net/url"
	"sync"
	"sync/atomic"
)

type Backend struct {
	URL   *url.URL
	Alive atomic.Bool
}

func (b *Backend) SetAlive(alive bool) {
	b.Alive.Store(alive)
}

func (b *Backend) IsAlive() bool {
	return b.Alive.Load()
}

type RoundRobin struct {
	mu       sync.Mutex
	backends []*Backend
	current  int
}

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
