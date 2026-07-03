package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Allenize/lime/internal/balancer"
	"github.com/Allenize/lime/internal/proxy"
)

func backendList() []string {
	raw := os.Getenv("BACKENDS")
	if raw == "" {
		return []string{"http://localhost:9001"}
	}
	parts := strings.Split(raw, ",")
	backends := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			backends = append(backends, p)
		}
	}
	return backends
}

func main() {
	backends := backendList()
	log.Printf("configured backends: %v", backends)

	rr, err := balancer.NewRoundRobin(backends)
	if err != nil {
		log.Fatalf("invalid backend URL: %v", err)
	}

	stop := make(chan struct{})
	go proxy.StartHealthChecks(rr, 5*time.Second, stop)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.Handle("/", proxy.New(rr))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("reverse proxy listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
