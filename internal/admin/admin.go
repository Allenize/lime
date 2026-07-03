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

var pageTemplate = template.Must(template.New("admin").Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta http-equiv="refresh" content="5">
<title>Lime admin</title>
<style>
body { font-family: -apple-system, sans-serif; background: #EAF7EF; color: #091413; margin: 0; padding: 40px; }
h1 { color: #285A48; margin-bottom: 4px; }
p.sub { color: #408A71; margin-top: 0; }
table { width: 100%; max-width: 720px; border-collapse: collapse; background: white; border-radius: 8px; overflow: hidden; }
th, td { text-align: left; padding: 12px 16px; }
th { background: #285A48; color: white; font-weight: 500; }
tr:nth-child(even) { background: #F5FBF7; }
.dot { display: inline-block; width: 10px; height: 10px; border-radius: 50%; margin-right: 8px; }
.up { background: #408A71; }
.down { background: #091413; }
</style>
</head>
<body>
<h1>Lime admin</h1>
<p class="sub">Auto-refreshes every 5 seconds</p>
<table>
<tr><th>Backend</th><th>Status</th><th>Active connections</th></tr>
{{range .}}
<tr>
<td>{{.URL}}</td>
<td><span class="dot {{if .Alive}}up{{else}}down{{end}}"></span>{{if .Alive}}up{{else}}down{{end}}</td>
<td>{{.Conns}}</td>
</tr>
{{end}}
</table>
</body>
</html>`))

func Handler(b balancer.Balancer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		views := snapshot(b)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		pageTemplate.Execute(w, views)
	}
}

func StatusHandler(b balancer.Balancer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		views := snapshot(b)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(views)
	}
}