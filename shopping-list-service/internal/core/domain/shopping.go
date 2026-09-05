package domain

import (
	"fmt"
	"time"
)

type Shopping struct {
	ID             uint
	Version        uint
	Title          string
	Description    *string
	AmountNow      float64
	AmountFinish   float64
	ImageKey       *string
	Completed      bool
	CreatedAt      time.Time
	UpdatedAt      *time.Time
	CompletedAt    *time.Time
	CompletionDate *time.Time
}

func NewShopping(
	id uint,
	version uint,
	title string,
	description *string,
	amountNow float64,
	amountFinish float64,
	imageKey *string,
	completed bool,
	createdAt time.Time,
	updatedAt *time.Time,
	completedAt *time.Time,
	completionDate *time.Time,
) *Shopping {
	return &Shopping{
		ID:             id,
		Version:        version,
		Title:          title,
		Description:    description,
		AmountNow:      amountNow,
		AmountFinish:   amountFinish,
		ImageKey:       imageKey,
		Completed:      completed,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		CompletedAt:    completedAt,
		CompletionDate: completionDate,
	}
}

func (s *Shopping) Validate() error {
	titleRun := []rune(s.Title)
	descRun := []rune(*s.Description)
	imageKeyRun := []rune(*s.ImageKey)
	if s.Title == " " || len(titleRun) >= 200 || len(titleRun) <= 0 {
		return fmt.Errorf("title is required")
	}
	if s.Description == nil || len(descRun) >= 1000 {
		return fmt.Errorf("description is required")
	}
	if s.AmountNow < 0 {
		return fmt.Errorf("amountNow is required amount now < 0")
	}
	if s.AmountFinish < 0 {
		return fmt.Errorf("amountFinish is required amount finish < 0")
	}
	if len(imageKeyRun) > 1000 {
		return fmt.Errorf("imageKey is required len > 1000")
	}
	return nil
}
