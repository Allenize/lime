package admin

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/Allenize/lime/internal/balancer"
)

type backendView struct {
	URL           string
	Alive         bool
	Conns         int64
	TotalRequests int64
	LastCheckedMs int64
}

type pageData struct {
	Backends   []backendView
	Total      int
	Healthy    int
	Unhealthy  int
	TotalConns int64
	FaviconB64 string
}

func snapshot(b balancer.Balancer) []backendView {
	backends := b.Backends()
	views := make([]backendView, 0, len(backends))
	for _, be := range backends {
		views = append(views, backendView{
			URL:           be.URL.String(),
			Alive:         be.IsAlive(),
			Conns:         be.ConnCount(),
			TotalRequests: be.TotalRequests(),
			LastCheckedMs: be.LastCheckedMillis(),
		})
	}
	return views
}

func buildPageData(b balancer.Balancer) pageData {
	views := snapshot(b)
	pd := pageData{Backends: views, Total: len(views), FaviconB64: faviconB64}
	for _, v := range views {
		if v.Alive {
			pd.Healthy++
		} else {
			pd.Unhealthy++
		}
		pd.TotalConns += v.Conns
	}
	return pd
}

const faviconB64 = "PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCA2NCA2NCI+CiAgPGNpcmNsZSBjeD0iMzIiIGN5PSIzMiIgcj0iMzAiIGZpbGw9IiMyODVBNDgiLz4KICA8Y2lyY2xlIGN4PSIzMiIgY3k9IjMyIiByPSIyNSIgZmlsbD0iI0VBRjdFRiIvPgogIDxjaXJjbGUgY3g9IjMyIiBjeT0iMzIiIHI9IjIyIiBmaWxsPSIjNDA4QTcxIi8+CiAgPGcgc3Ryb2tlPSIjRUFGN0VGIiBzdHJva2Utd2lkdGg9IjIiIHN0cm9rZS1saW5lY2FwPSJyb3VuZCI+CiAgICA8bGluZSB4MT0iMzIiIHkxPSIzMiIgeDI9IjMyIiB5Mj0iMTAiLz4KICAgIDxsaW5lIHgxPSIzMiIgeTE9IjMyIiB4Mj0iNTIiIHkyPSIyMCIvPgogICAgPGxpbmUgeDE9IjMyIiB5MT0iMzIiIHgyPSI1MiIgeTI9IjQ0Ii8+CiAgICA8bGluZSB4MT0iMzIiIHkxPSIzMiIgeDI9IjMyIiB5Mj0iNTQiLz4KICAgIDxsaW5lIHgxPSIzMiIgeTE9IjMyIiB4Mj0iMTIiIHkyPSI0NCIvPgogICAgPGxpbmUgeDE9IjMyIiB5MT0iMzIiIHgyPSIxMiIgeTI9IjIwIi8+CiAgPC9nPgogIDxjaXJjbGUgY3g9IjMyIiBjeT0iMzIiIHI9IjMiIGZpbGw9IiNCMEU0Q0MiLz4KPC9zdmc+Cg=="

var pageTemplate = template.Must(template.New("admin").Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<link rel="icon" type="image/svg+xml" href="data:image/svg+xml;base64,{{.FaviconB64}}">
<title>Lime admin</title>
<style>
  :root {
    --bg: #091413;
    --surface: #0f201e;
    --surface-2: #142c29;
    --border: #285A48;
    --accent: #408A71;
    --accent-bright: #B0E4CC;
    --text: #EAF7EF;
    --text-dim: #7fa896;
    --down: #e8836b;
  }
  * { box-sizing: border-box; }
  html, body { height: 100%; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    background: var(--bg);
    color: var(--text);
    margin: 0;
    padding: clamp(24px, 4vw, 56px);
    width: 100%;
  }
  .header { display: flex; align-items: center; gap: 14px; margin-bottom: 28px; }
  .header svg { width: 34px; height: 34px; flex-shrink: 0; }
  h1 { margin: 0; font-size: clamp(22px, 2.4vw, 30px); font-weight: 600; letter-spacing: -0.02em; }

  .stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 16px;
    width: 100%;
    margin-bottom: 32px;
  }
  .stat-card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 14px;
    padding: 22px 24px;
  }
  .stat-value { font-size: clamp(28px, 3vw, 38px); font-weight: 700; letter-spacing: -0.02em; }
  .stat-label { font-size: 13px; color: var(--text-dim); margin-top: 6px; text-transform: uppercase; letter-spacing: 0.05em; }
  .stat-card.healthy .stat-value { color: var(--accent-bright); }
  .stat-card.unhealthy .stat-value { color: var(--down); }

  .backend-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 16px;
    width: 100%;
  }
  .backend-card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 14px;
    padding: 22px 24px;
  }
  .backend-url { font-size: 15px; font-weight: 600; word-break: break-all; margin-bottom: 14px; }
  .backend-status { display: flex; align-items: center; gap: 8px; font-size: 14px; margin-bottom: 16px; }
  .dot { display: inline-block; width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0; }
  .up { background: var(--accent-bright); box-shadow: 0 0 8px rgba(176,228,204,0.5); }
  .down { background: var(--down); }
  .backend-metrics { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; font-size: 13px; color: var(--text-dim); }
  .metric-value { font-size: 18px; font-weight: 600; color: var(--text); display: block; margin-bottom: 2px; }
  .empty { color: var(--text-dim); padding: 40px; text-align: center; border: 1px dashed var(--border); border-radius: 14px; width: 100%; }
</style>
</head>
<body>
  <div class="header">
    <svg viewBox="0 0 240 240" xmlns="http://www.w3.org/2000/svg">
      <circle cx="120" cy="120" r="110" fill="#285A48"/>
      <circle cx="120" cy="120" r="96" fill="#EAF7EF"/>
      <circle cx="120" cy="120" r="88" fill="#408A71"/>
      <g stroke="#EAF7EF" stroke-width="2.5" stroke-linecap="round">
        <line x1="120" y1="120" x2="120" y2="32"/>
        <line x1="120" y1="120" x2="182" y2="58"/>
        <line x1="120" y1="120" x2="208" y2="120"/>
        <line x1="120" y1="120" x2="182" y2="182"/>
        <line x1="120" y1="120" x2="120" y2="208"/>
        <line x1="120" y1="120" x2="58" y2="182"/>
        <line x1="120" y1="120" x2="32" y2="120"/>
        <line x1="120" y1="120" x2="58" y2="58"/>
      </g>
      <circle cx="120" cy="120" r="8" fill="#B0E4CC"/>
    </svg>
    <h1>Lime admin</h1>
  </div>

  <div class="stats">
    <div class="stat-card">
      <div class="stat-value" id="stat-total">{{.Total}}</div>
      <div class="stat-label">Backends</div>
    </div>
    <div class="stat-card healthy">
      <div class="stat-value" id="stat-healthy">{{.Healthy}}</div>
      <div class="stat-label">Healthy</div>
    </div>
    <div class="stat-card unhealthy">
      <div class="stat-value" id="stat-unhealthy">{{.Unhealthy}}</div>
      <div class="stat-label">Unhealthy</div>
    </div>
    <div class="stat-card">
      <div class="stat-value" id="stat-conns">{{.TotalConns}}</div>
      <div class="stat-label">Active connections</div>
    </div>
  </div>

  <div class="backend-grid" id="backend-grid">
    {{if not .Backends}}
    <div class="empty">No backends configured</div>
    {{end}}
    {{range .Backends}}
    <div class="backend-card">
      <div class="backend-url">{{.URL}}</div>
      <div class="backend-status"><span class="dot {{if .Alive}}up{{else}}down{{end}}"></span>{{if .Alive}}Healthy{{else}}Unhealthy{{end}}</div>
      <div class="backend-metrics">
        <div><span class="metric-value">{{.Conns}}</span>Active connections</div>
        <div><span class="metric-value">{{.TotalRequests}}</span>Total requests</div>
        <div><span class="metric-value last-checked" data-last="{{.LastCheckedMs}}">-</span>Last checked</div>
      </div>
    </div>
    {{end}}
  </div>

<script>
function timeAgo(ms) {
  if (!ms) return 'never';
  const s = Math.floor((Date.now() - ms) / 1000);
  if (s < 5) return 'just now';
  if (s < 60) return s + 's ago';
  const m = Math.floor(s / 60);
  if (m < 60) return m + 'm ago';
  const h = Math.floor(m / 60);
  return h + 'h ago';
}

async function refresh() {
  try {
    const res = await fetch('/admin/status');
    const backends = await res.json();
    let healthy = 0, unhealthy = 0, totalConns = 0;
    const cards = backends.map(b => {
      if (b.Alive) healthy++; else unhealthy++;
      totalConns += b.Conns;
      return '<div class="backend-card">' +
        '<div class="backend-url">' + b.URL + '</div>' +
        '<div class="backend-status"><span class="dot ' + (b.Alive ? 'up' : 'down') + '"></span>' +
        (b.Alive ? 'Healthy' : 'Unhealthy') + '</div>' +
        '<div class="backend-metrics">' +
        '<div><span class="metric-value">' + b.Conns + '</span>Active connections</div>' +
        '<div><span class="metric-value">' + b.TotalRequests + '</span>Total requests</div>' +
        '<div><span class="metric-value last-checked" data-last="' + b.LastCheckedMs + '">-</span>Last checked</div>' +
        '</div></div>';
    }).join('') || '<div class="empty">No backends configured</div>';

    document.getElementById('backend-grid').innerHTML = cards;
    document.getElementById('stat-total').textContent = backends.length;
    document.getElementById('stat-healthy').textContent = healthy;
    document.getElementById('stat-unhealthy').textContent = unhealthy;
    document.getElementById('stat-conns').textContent = totalConns;
    updateTimeAgo();
  } catch (e) {}
}

function updateTimeAgo() {
  document.querySelectorAll('.last-checked').forEach(el => {
    el.textContent = timeAgo(parseInt(el.dataset.last, 10));
  });
}

setInterval(refresh, 3000);
setInterval(updateTimeAgo, 1000);
updateTimeAgo();
</script>
</body>
</html>`))

func Handler(b balancer.Balancer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		pageTemplate.Execute(w, buildPageData(b))
	}
}

func StatusHandler(b balancer.Balancer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		views := snapshot(b)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(views)
	}
}
