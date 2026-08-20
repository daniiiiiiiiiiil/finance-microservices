package service_user

import (
	"backend/internal/core/kafka"
	"fmt"

	"golang.org/x/net/context"
)

func (s *UsersService) sendUserEvent(ctx context.Context, eventType string, userID int, email, fullName string, isAdmin bool, status string) {
	eventData := kafka.UserEvent{
		UserID:   userID,
		Email:    email,
		FullName: fullName,
		IsAdmin:  isAdmin,
		Status:   status}
	if err := s.producer.SendEvent(ctx, eventType, eventData); err != nil {
		fmt.Errorf("Error sending user event: %v", err)
		return
	}
}
