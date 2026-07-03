package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Allenize/lime/internal/admin"
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
	strategy := os.Getenv("STRATEGY")
	log.Printf("configured backends: %v (strategy: %s)", backends, strategyLabel(strategy))

	b, err := balancer.New(strategy, backends)
	if err != nil {
		log.Fatalf("invalid balancer config: %v", err)
	}

	stop := make(chan struct{})
	go proxy.StartHealthChecks(b, 5*time.Second, stop)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/admin", admin.Handler(b))
	mux.HandleFunc("/admin/status", admin.StatusHandler(b))
	mux.Handle("/", proxy.New(b))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")

	if certFile != "" && keyFile != "" {
		log.Printf("lime listening on %s (TLS enabled)", addr)
		if err := http.ListenAndServeTLS(addr, certFile, keyFile, mux); err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Printf("lime listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func strategyLabel(s string) string {
	if s == "" {
		return "round_robin"
	}
	return s
}