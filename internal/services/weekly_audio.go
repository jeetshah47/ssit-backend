package services

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/repositories"
	"github.com/google/uuid"
)

// GetWeeklyAudioQuery represents the query for getting weekly audio
type GetWeeklyAudioQuery struct {
	AudioID uuid.UUID
}

// GetWeeklyAudioResult represents the output
type GetWeeklyAudioResult struct {
	Audio *models.WeeklyAudio
}

// GetWeeklyAudioService handles weekly audio retrieval
type GetWeeklyAudioService struct {
	audioRepo repositories.WeeklyAudioRepo
}

// NewGetWeeklyAudioService creates a new get weekly audio service
func NewGetWeeklyAudioService(audioRepo repositories.WeeklyAudioRepo) *GetWeeklyAudioService {
	return &GetWeeklyAudioService{
		audioRepo: audioRepo,
	}
}

// Execute executes the GetWeeklyAudio query
func (s *GetWeeklyAudioService) Execute(ctx context.Context, query GetWeeklyAudioQuery) (*GetWeeklyAudioResult, error) {
	var audio *models.WeeklyAudio
	var err error

	if query.AudioID != uuid.Nil {
		audio, err = s.audioRepo.FindByID(ctx, query.AudioID)
	} else {
		audio, err = s.audioRepo.FindLatest(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get weekly audio: %w", err)
	}

	return &GetWeeklyAudioResult{
		Audio: audio,
	}, nil
}

// ListWeeklyAudioQuery represents the query for listing weekly audio
type ListWeeklyAudioQuery struct {
	Limit  int
	Offset int
}

// ListWeeklyAudioResult represents the output
type ListWeeklyAudioResult struct {
	Audios []*models.WeeklyAudio
}

// ListWeeklyAudioService handles weekly audio listing
type ListWeeklyAudioService struct {
	audioRepo repositories.WeeklyAudioRepo
}

// NewListWeeklyAudioService creates a new list weekly audio service
func NewListWeeklyAudioService(audioRepo repositories.WeeklyAudioRepo) *ListWeeklyAudioService {
	return &ListWeeklyAudioService{
		audioRepo: audioRepo,
	}
}

// Execute executes the ListWeeklyAudio query
func (s *ListWeeklyAudioService) Execute(ctx context.Context, query ListWeeklyAudioQuery) (*ListWeeklyAudioResult, error) {
	limit := query.Limit
	if limit == 0 {
		limit = 10
	}

	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	audios, err := s.audioRepo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list weekly audio: %w", err)
	}

	return &ListWeeklyAudioResult{
		Audios: audios,
	}, nil
}
