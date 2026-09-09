package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func (s *ShoppingService) GetImage(ctx context.Context, imageKey string) ([]byte, error) {
	if imageKey == "" {
		return nil, fmt.Errorf("imageID is empty")
	}
	imageData, err := s.storage.Get(ctx, imageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get image data from storage: %w", err)
	}

	s.logger.Debug("image retrieved from S3",
		zap.String("image_key", imageKey),
		zap.Int("size", len(imageData)))

	return imageData, nil
}

func (s *ShoppingService) GetImageByShoppingID(ctx context.Context, shoppingID, userID int) ([]byte, string, error) {
	shopping, err := s.shoppingRepository.GetShopping(ctx, shoppingID, userID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get shopping data from storage: %w", err)
	}

	if shopping.ImageKey == nil || *shopping.ImageKey == "" {
		return nil, "", fmt.Errorf("image key is empty")
	}

	imageData, err := s.storage.Get(ctx, *shopping.ImageKey)
	if err != nil {
		s.logger.Error("failed to get image from S3",
			zap.Int("shopping_id", shoppingID),
			zap.String("image_key", *shopping.ImageKey),
			zap.Error(err))
		return nil, "", fmt.Errorf("failed to get image: %w", err)
	}

	s.logger.Debug("image retrieved by shopping ID",
		zap.Int("shopping_id", shoppingID),
		zap.String("image_key", *shopping.ImageKey),
		zap.Int("size", len(imageData)))

	return imageData, *shopping.ImageKey, nil
}
