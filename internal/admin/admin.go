package admin

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/Allenize/lime/internal/balancer"
)

type backendView struct {
	URL   string
	Alive bool
	Conns int64
}

type pageData struct {
	Backends   []backendView
	Total      int
	Healthy    int
	Unhealthy  int
	TotalConns int64
}

func snapshot(b balancer.Balancer) []backendView {
	backends := b.Backends()
	views := make([]backendView, 0, len(backends))
	for _, be := range backends {
		views = append(views, backendView{
			URL:   be.URL.String(),
			Alive: be.IsAlive(),
			Conns: be.ConnCount(),
		})
	}
	return views
}

func buildPageData(b balancer.Balancer) pageData {
	views := snapshot(b)
	pd := pageData{Backends: views, Total: len(views)}
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

var pageTemplate = template.Must(template.New("admin").Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
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
  }
  * { box-sizing: border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    background: var(--bg);
    color: var(--text);
    margin: 0;
    padding: 48px 40px;
  }
  .header { display: flex; align-items: center; gap: 14px; margin-bottom: 4px; }
  .header svg { width: 32px; height: 32px; }
  h1 { margin: 0; font-size: 26px; font-weight: 600; letter-spacing: -0.02em; }
  p.sub { color: var(--text-dim); margin: 6px 0 32px 0; font-size: 14px; display: flex; align-items: center; gap: 8px; }
  .live-dot {
    width: 7px; height: 7px; border-radius: 50%;
    background: var(--accent-bright);
    box-shadow: 0 0 0 0 rgba(176,228,204,0.6);
    animation: pulse 2s infinite;
  }
  @keyframes pulse {
    0% { box-shadow: 0 0 0 0 rgba(176,228,204,0.5); }
    70% { box-shadow: 0 0 0 8px rgba(176,228,204,0); }
    100% { box-shadow: 0 0 0 0 rgba(176,228,204,0); }
  }
  .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 16px; max-width: 900px; margin-bottom: 36px; }
  .stat-card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 20px;
  }
  .stat-value { font-size: 32px; font-weight: 700; letter-spacing: -0.02em; }
  .stat-label { font-size: 13px; color: var(--text-dim); margin-top: 4px; text-transform: uppercase; letter-spacing: 0.04em; }
  .stat-card.healthy .stat-value { color: var(--accent-bright); }
  .stat-card.unhealthy .stat-value { color: #e8836b; }
  table { width: 100%; max-width: 900px; border-collapse: collapse; background: var(--surface); border: 1px solid var(--border); border-radius: 12px; overflow: hidden; }
  th, td { text-align: left; padding: 14px 18px; font-size: 14px; }
  th { background: var(--surface-2); color: var(--text-dim); font-weight: 500; font-size: 12px; text-transform: uppercase; letter-spacing: 0.04em; border-bottom: 1px solid var(--border); }
  tr:not(:last-child) td { border-bottom: 1px solid var(--border); }
  .dot { display: inline-block; width: 9px; height: 9px; border-radius: 50%; margin-right: 8px; vertical-align: middle; }
  .up { background: var(--accent-bright); box-shadow: 0 0 8px rgba(176,228,204,0.5); }
  .down { background: #e8836b; }
  .conn-bar-track { display: inline-block; width: 60px; height: 6px; background: var(--surface-2); border-radius: 3px; margin-right: 10px; vertical-align: middle; overflow: hidden; }
  .conn-bar-fill { height: 100%; background: var(--accent); border-radius: 3px; transition: width 0.4s ease; }
  .empty { color: var(--text-dim); padding: 32px 18px; text-align: center; }
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
  <p class="sub"><span class="live-dot"></span>Live &mdash; updates automatically, no reload needed</p>

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

  <table>
    <thead>
      <tr><th>Backend</th><th>Status</th><th>Active connections</th></tr>
    </thead>
    <tbody id="backend-rows">
      {{if not .Backends}}
      <tr><td colspan="3" class="empty">No backends configured</td></tr>
      {{end}}
      {{range .Backends}}
      <tr>
        <td>{{.URL}}</td>
        <td><span class="dot {{if .Alive}}up{{else}}down{{end}}"></span>{{if .Alive}}up{{else}}down{{end}}</td>
        <td><span class="conn-bar-track"><span class="conn-bar-fill" style="width: {{if gt .Conns 0}}60{{else}}0{{end}}%"></span></span>{{.Conns}}</td>
      </tr>
      {{end}}
    </tbody>
  </table>

<script>
async function refresh() {
  try {
    const res = await fetch('/admin/status');
    const backends = await res.json();
    let healthy = 0, unhealthy = 0, totalConns = 0;
    const rows = backends.map(b => {
      if (b.Alive) healthy++; else unhealthy++;
      totalConns += b.Conns;
      const barWidth = b.Conns > 0 ? Math.min(100, b.Conns * 20) : 0;
      return '<tr><td>' + b.URL + '</td><td><span class="dot ' + (b.Alive ? 'up' : 'down') + '"></span>' +
        (b.Alive ? 'up' : 'down') + '</td><td><span class="conn-bar-track"><span class="conn-bar-fill" style="width:' +
        barWidth + '%"></span></span>' + b.Conns + '</td></tr>';
    }).join('') || '<tr><td colspan="3" class="empty">No backends configured</td></tr>';

    document.getElementById('backend-rows').innerHTML = rows;
    document.getElementById('stat-total').textContent = backends.length;
    document.getElementById('stat-healthy').textContent = healthy;
    document.getElementById('stat-unhealthy').textContent = unhealthy;
    document.getElementById('stat-conns').textContent = totalConns;
  } catch (e) {}
}
setInterval(refresh, 3000);
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
