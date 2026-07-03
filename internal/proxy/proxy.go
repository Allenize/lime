// Package proxy implements the HTTP reverse proxy that forwards incoming
// requests to a backend chosen by the load balancer.
package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/YOUR_USERNAME/lime/internal/balancer"
)

// Handler is an http.Handler that load-balances requests across backends.
type Handler struct {
	rr *balancer.RoundRobin
}

// New creates a proxy Handler backed by the given round-robin balancer.
func New(rr *balancer.RoundRobin) *Handler {
	return &Handler{rr: rr}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := h.rr.Next()
	if backend == nil {
		http.Error(w, "no healthy backends available", http.StatusBadGateway)
		return
	}

	rp := httputil.NewSingleHostReverseProxy(backend.URL)

	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error forwarding to %s: %v", backend.URL, err)
		backend.SetAlive(false)
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}

	log.Printf("forwarding %s %s -> %s", r.Method, r.URL.Path, backend.URL)
	rp.ServeHTTP(w, r)
}

// StartHealthChecks periodically pings each backend's /health endpoint and
// updates its alive status. It runs until the provided stop channel is
// closed, so call it with `go StartHealthChecks(...)`.
func StartHealthChecks(rr *balancer.RoundRobin, interval time.Duration, stop <-chan struct{}) {
	client := &http.Client{Timeout: 2 * time.Second}

	check := func() {
		for _, b := range rr.Backends() {
			healthURL := b.URL.String() + "/health"
			resp, err := client.Get(healthURL)
			alive := err == nil && resp.StatusCode == http.StatusOK
			if resp != nil {
				resp.Body.Close()
			}
			if alive != b.IsAlive() {
				log.Printf("backend %s alive=%v", b.URL, alive)
			}
			b.SetAlive(alive)
		}
	}

	// Check once immediately so we don't start with false assumptions.
	check()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			check()
		case <-stop:
			return
		}
	}
}
