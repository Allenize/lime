<div align="center">

<img src="docs/assets/logo.svg" alt="Lime logo" width="140" />

# Lime

**A reverse proxy & load balancer, built from scratch in Go.**

Built by [Allen](https://github.com/Allenize) · pure standard library · zero dependencies

[![CI](https://github.com/Allenize/lime/actions/workflows/ci.yml/badge.svg)](https://github.com/Allenize/lime/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Allenize/lime)](https://goreportcard.com/report/github.com/Allenize/lime)
[![Go Version](https://img.shields.io/badge/Go-1.22-408A71?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-285A48?logoColor=white)](./LICENSE)

</div>

---

> **Status:** feature-complete — round-robin and least-connections load balancing, automatic health-check failover, a live admin dashboard, and optional TLS. Ready to deploy.

## <img src="docs/assets/icons/layers.svg" width="18" align="center"/> Overview

Lime is a hand-built reverse proxy and load balancer, made to understand — and eventually match — what tools like Traefik and nginx do under the hood. No frameworks, no third-party routers: just Go's standard library.

## <img src="docs/assets/icons/rocket.svg" width="18" align="center"/> Features / Roadmap

- [x] Health check endpoint
- [x] Reverse proxy request forwarding
- [x] Load balancing (round robin)
- [x] Backend health checks (automatic failover)
- [x] Least-connections balancing strategy
- [x] Simple admin dashboard
- [x] TLS support

## <img src="docs/assets/icons/box.svg" width="18" align="center"/> Project layout

```
lime/
├── .github/workflows/ci.yml   # CI: format, vet, build, test on every push
├── cmd/lime/main.go           # entrypoint (the binary you run)
├── internal/balancer/         # load balancing strategies (round robin, least connections)
├── internal/proxy/            # reverse proxy / request forwarding logic
├── internal/admin/            # live admin dashboard + JSON status endpoint
├── docs/assets/               # logo + icons used in this README
├── Dockerfile                 # containerized build, for free deployment anywhere
├── go.mod
└── README.md
```

## <img src="docs/assets/icons/play.svg" width="18" align="center"/> Run locally

Lime needs at least one backend server to forward traffic to. Configure it via environment variables:

| Variable        | Description                                              | Default                    |
|-----------------|------------------------------------------------------------|-----------------------------|
| `BACKENDS`      | Comma-separated backend URLs                              | `http://localhost:9001`    |
| `STRATEGY`      | `round_robin` or `least_conn`                              | `round_robin`               |
| `PORT`          | Port Lime listens on                                       | `8080`                      |
| `TLS_CERT_FILE` | Path to a TLS certificate (enables HTTPS if both are set)  | unset                        |
| `TLS_KEY_FILE`  | Path to the matching TLS private key                       | unset                        |

```bash
BACKENDS="http://localhost:9001,http://localhost:9002" STRATEGY=least_conn go run ./cmd/lime
```

- Visit `http://localhost:8080/health` to confirm Lime itself is up.
- Visit `http://localhost:8080/admin` for a live dashboard of backend status and active connection counts (auto-refreshes every 5 seconds). A JSON version is available at `/admin/status`.
- Requests to `http://localhost:8080/` are load-balanced across your configured backends, and unhealthy backends are automatically removed from rotation.

To enable HTTPS, set `TLS_CERT_FILE` and `TLS_KEY_FILE` to a certificate/key pair — Lime then serves exclusively over TLS.

## <img src="docs/assets/icons/container.svg" width="18" align="center"/> Run with Docker

```bash
docker build -t lime .
docker run -p 8080:8080 lime
```

## <img src="docs/assets/icons/cloud.svg" width="18" align="center"/> Deploying (Render — free, no expiration)

Lime reads the `PORT` environment variable (falls back to `8080` locally), which is what Render and most hosts require.

1. Push this repo to GitHub
2. Go to [render.com](https://render.com) → sign up / log in → connect your GitHub account
3. **New** → **Web Service** → select this repo
4. Render auto-detects the `Dockerfile` — leave build/start commands as-is
5. Choose the **Free** instance type
6. Under **Advanced**, set:
   - **Health Check Path**: `/health`
   - **Environment Variable**: `BACKENDS` = your comma-separated backend URLs
   - Optionally: `STRATEGY` = `least_conn` (defaults to `round_robin`)
7. Click **Create Web Service**

From then on, every `git push` to `main` auto-deploys. Note: free instances spin down after 15 minutes of no traffic and take ~30-60 seconds to wake back up on the next request — fine for a personal project, not for production traffic.

Render automatically provides HTTPS at its edge, so you don't need to set `TLS_CERT_FILE`/`TLS_KEY_FILE` when deploying there — Lime's built-in TLS support is for running it directly on your own server or behind a different host that doesn't terminate TLS for you.

## <img src="docs/assets/icons/wrench.svg" width="18" align="center"/> Development

```bash
go build ./...   # build
go vet ./...      # static analysis
gofmt -l .          # check formatting
go test ./...      # run tests
```

These are the exact same checks CI runs on every push — running them locally before pushing keeps CI green.

## Brand palette

<div align="center">

![#091413](https://img.shields.io/badge/091413-091413?style=flat-square&labelColor=091413&color=091413) &nbsp;
![#285A48](https://img.shields.io/badge/285A48-285A48?style=flat-square&labelColor=285A48&color=285A48) &nbsp;
![#408A71](https://img.shields.io/badge/408A71-408A71?style=flat-square&labelColor=408A71&color=408A71) &nbsp;
![#B0E4CC](https://img.shields.io/badge/B0E4CC-B0E4CC?style=flat-square&labelColor=B0E4CC&color=B0E4CC)

</div>

---

<div align="center">

**Built by Allen**

[![GitHub](https://img.shields.io/badge/GitHub-YOUR__USERNAME-091413?logo=github&logoColor=white)](https://github.com/Allenize)

</div>