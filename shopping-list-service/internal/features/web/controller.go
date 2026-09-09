package web

import (
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
	html, err := c.service.GetMainPage()
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(html)
}

// ServeAssets отдает статические файлы (CSS, JS, изображения)
func (c *WebController) ServeAssets(w http.ResponseWriter, r *http.Request) {
	// Убираем /assets/ из пути
	filename := strings.TrimPrefix(r.URL.Path, "/assets/")
	if filename == "" {
		http.NotFound(w, r)
		return
	}

	// Ищем файл
	paths := []string{
		filepath.Join(".", "public", "assets", filename),
		filepath.Join("..", "public", "assets", filename),
	}

	if projectRoot := os.Getenv("PROJECT_ROOT"); projectRoot != "" {
		paths = append([]string{filepath.Join(projectRoot, "public", "assets", filename)}, paths...)
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			// Определяем Content-Type
			ext := filepath.Ext(filename)
			switch ext {
			case ".css":
				w.Header().Set("Content-Type", "text/css")
			case ".js":
				w.Header().Set("Content-Type", "application/javascript")
			case ".png":
				w.Header().Set("Content-Type", "image/png")
			case ".jpg", ".jpeg":
				w.Header().Set("Content-Type", "image/jpeg")
			case ".svg":
				w.Header().Set("Content-Type", "image/svg+xml")
			}
			http.ServeFile(w, r, path)
			return
		}
	}

	http.NotFound(w, r)
}
