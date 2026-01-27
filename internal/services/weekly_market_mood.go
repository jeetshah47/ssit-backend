package services

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/repositories"
	"github.com/google/uuid"
)

// GetWeeklyMarketMoodQuery represents the query for getting weekly market mood
type GetWeeklyMarketMoodQuery struct {
	MoodID uuid.UUID
}

// GetWeeklyMarketMoodResult represents the output
type GetWeeklyMarketMoodResult struct {
	Mood *models.WeeklyMarketMood
}

// GetWeeklyMarketMoodService handles weekly market mood retrieval
type GetWeeklyMarketMoodService struct {
	moodRepo repositories.WeeklyMarketMoodRepo
}

// NewGetWeeklyMarketMoodService creates a new get weekly market mood service
func NewGetWeeklyMarketMoodService(moodRepo repositories.WeeklyMarketMoodRepo) *GetWeeklyMarketMoodService {
	return &GetWeeklyMarketMoodService{
		moodRepo: moodRepo,
	}
}

// Execute executes the GetWeeklyMarketMood query
func (s *GetWeeklyMarketMoodService) Execute(ctx context.Context, query GetWeeklyMarketMoodQuery) (*GetWeeklyMarketMoodResult, error) {
	var mood *models.WeeklyMarketMood
	var err error

	if query.MoodID != uuid.Nil {
		mood, err = s.moodRepo.FindByID(ctx, query.MoodID)
	} else {
		mood, err = s.moodRepo.FindLatest(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get weekly market mood: %w", err)
	}

	return &GetWeeklyMarketMoodResult{
		Mood: mood,
	}, nil
}
