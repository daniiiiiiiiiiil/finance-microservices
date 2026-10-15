package web

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type WebController struct {
	service *WebService
}

func NewWebController(service *WebService) *WebController {
	return &WebController{service: service}
}

func (c *WebController) GetMainPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	html, err := c.service.GetMainPage()
	if err != nil {
		http.Error(w, "Frontend entry point not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(html)
}

func (c *WebController) ServeAssets(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/assets/") {
		http.NotFound(w, r)
		return
	}
	filename := strings.TrimPrefix(r.URL.Path, "/assets/")
	if filename == "" || strings.Contains(filename, "\\") {
		http.NotFound(w, r)
		return
	}
	clean := filepath.Clean(filepath.FromSlash(filename))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || filepath.IsAbs(clean) {
		http.NotFound(w, r)
		return
	}

	roots := []string{}
	if projectRoot := os.Getenv("PROJECT_ROOT"); projectRoot != "" {
		roots = append(roots, filepath.Join(projectRoot, "public", "assets"))
	}
	roots = append(roots, filepath.Join("public", "assets"), filepath.Join("..", "public", "assets"))

	for _, root := range roots {
		rootAbs, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		assetPath := filepath.Join(rootAbs, clean)
		rel, err := filepath.Rel(rootAbs, assetPath)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		info, err := os.Stat(assetPath)
		if err != nil || info.IsDir() {
			continue
		}
		if contentType := mime.TypeByExtension(filepath.Ext(assetPath)); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.ServeFile(w, r, assetPath)
		return
	}
	http.NotFound(w, r)
}

func (c *WebController) GetAdminPage(w http.ResponseWriter, r *http.Request) {
	html, err := c.service.GetHTMLFile("admin.html")
	if err != nil {
		http.Error(w, "Admin page not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-cache")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(html)
}
