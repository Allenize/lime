package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/YOUR_USERNAME/lime/internal/balancer"
)

type Handler struct {
	rr *balancer.RoundRobin
}

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
