package web

import (
	"fmt"
	"os"
	"path/filepath"
)

type WebService struct {
	repo *WebRepository
}

func NewWebService(repo *WebRepository) *WebService {
	return &WebService{repo: repo}
}

func (s *WebService) GetHTMLFile(filename string) ([]byte, error) {
	// Ищем файл в разных местах
	paths := []string{
		filepath.Join(".", "public", filename),
		filepath.Join("..", "public", filename),
	}

	if projectRoot := os.Getenv("PROJECT_ROOT"); projectRoot != "" {
		paths = append([]string{filepath.Join(projectRoot, "public", filename)}, paths...)
	}

	var lastErr error
	for _, path := range paths {
		html, err := s.repo.GetFile(path)
		if err == nil {
			return html, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("get file from repository: %w", lastErr)
}

func (s *WebService) GetMainPage() ([]byte, error) {
	return s.GetHTMLFile("/index.html")
}
