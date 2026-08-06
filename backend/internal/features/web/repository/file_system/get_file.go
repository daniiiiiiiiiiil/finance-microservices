package file_system_web

import (
	errors_core "backend/internal/core/errors"
	"fmt"
	"os"
)

func (r *WebRepository) GetFile(filePath string) ([]byte, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file %s:%w", filePath, errors_core.ErrNotFound)
		}
		return nil, fmt.Errorf("file %s:%w", filePath, err)
	}
	return file, err
}
