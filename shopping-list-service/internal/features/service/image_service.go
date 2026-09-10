package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	MaxImageSize = 10 * 1024 * 1024
)

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

func (s *ShoppingService) UploadImage(ctx context.Context, shoppingID, userID int, fileData []byte, filename string) (string, error) {
	if err := validateImage(fileData, filename); err != nil {
		s.logger.Error("image validation failed",
			zap.Int("shopping_id", shoppingID),
			zap.Int("user_id", userID),
			zap.Error(err))
		return "", fmt.Errorf("image validation failed: %w", err)
	}

	key := generateImageKey(userID, shoppingID, filename)

	if err := s.storage.Put(ctx, key, fileData); err != nil {
		s.logger.Error("failed to upload image to S3",
			zap.Int("shopping_id", shoppingID),
			zap.Int("user_id", userID),
			zap.String("key", key),
			zap.Error(err))
		return "", fmt.Errorf("failed to upload image: %w", err)
	}
	s.logger.Info("image uploaded successfully", zap.Int("shopping_id", shoppingID), zap.Int("user_id", userID), zap.String("key", key), zap.Int("size", len(fileData)))

	return key, nil
}

func (s *ShoppingService) GetImageURL(ctx context.Context, imageKey string) ([]byte, error) {
	imageURL, err := s.storage.Get(ctx, imageKey)
	if err != nil {
		s.logger.Error("failed to get image URL",
			zap.String("key", imageKey),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get image URL: %w", err)
	}
	return imageURL, nil
}

func (s *ShoppingService) DeleteImage(ctx context.Context, imageKey string) error {
	if imageKey == "" {
		s.logger.Debug("empty image key, skipping deletion")
		return nil
	}

	if err := s.storage.Delete(ctx, imageKey); err != nil {
		s.logger.Error("failed to delete image from S3",
			zap.String("key", imageKey),
			zap.Error(err))
		return fmt.Errorf("failed to delete image: %w", err)
	}

	s.logger.Debug("image deleted successfully",
		zap.String("key", imageKey))
	return nil
}

func validateImage(fileData []byte, filename string) error {
	if len(fileData) == 0 {
		return fmt.Errorf("empty file")
	}
	if len(fileData) > MaxImageSize {
		return fmt.Errorf("file size %d bytes exceeds limit %d bytes",
			len(fileData), MaxImageSize)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedExtensions[ext] {
		return fmt.Errorf("unsupported file extension: %s (allowed: .jpg, .jpeg, .png, .webp)", ext)
	}

	return nil
}
func generateImageKey(userID, shoppingID int, filename string) string {
	ext := filepath.Ext(filename)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("shopping/%d/%d/%d_%s%s",
		userID,
		shoppingID,
		timestamp,
		strings.TrimSuffix(filepath.Base(filename), ext),
		ext)
}
