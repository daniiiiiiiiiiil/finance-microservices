package web

import (
	"fmt"
	"os"
)

type WebRepository struct{}

func NewWebRepository() *WebRepository {
	return &WebRepository{}
}

func (r *WebRepository) GetFile(filePath string) ([]byte, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", filePath)
		}
		return nil, err
	}
	return file, nil
}
