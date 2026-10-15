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
	filename = filepath.Clean(filepath.FromSlash(filename))
	if filename == "." || filepath.IsAbs(filename) || filename == ".." || len(filename) >= 3 && filename[:3] == ".."+string(filepath.Separator) {
		return nil, fmt.Errorf("invalid public file path")
	}
	paths := []string{}
	if projectRoot := os.Getenv("PROJECT_ROOT"); projectRoot != "" {
		paths = append(paths, filepath.Join(projectRoot, "public", filename))
	}
	paths = append(paths, filepath.Join("public", filename), filepath.Join("..", "public", filename))
	var lastErr error
	for _, path := range paths {
		content, err := s.repo.GetFile(path)
		if err == nil {
			return content, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("get file from repository: %w", lastErr)
}

func (s *WebService) GetMainPage() ([]byte, error) {
	return s.GetHTMLFile("index.html")
}
