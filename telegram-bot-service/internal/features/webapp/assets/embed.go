package assets

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed index.html app.js styles.css
var content embed.FS

func Handler() http.Handler {
	sub, err := fs.Sub(content, ".")
	if err != nil {
		return http.NotFoundHandler()
	}
	return http.FileServer(http.FS(sub))
}

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	data, err := content.ReadFile("index.html")
	if err != nil {
		http.Error(w, "web app is unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if _, err := w.Write(data); err != nil {
		return
	}
}
