package service_web

import (
	"fmt"
)

func (s *WebService) GetMainPage() ([]byte, error) {
	htmlFilePath := "./public/test.html"
	html, err := s.webRepository.GetFile(htmlFilePath)
	if err != nil {
		return nil, fmt.Errorf("GetMainPage: %w", err)
	}
	return html, nil
}
