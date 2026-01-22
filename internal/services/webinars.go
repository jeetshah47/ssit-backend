package services

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/repositories"
	"github.com/google/uuid"
)

// GetWebinarQuery represents the query for getting a webinar
type GetWebinarQuery struct {
	WebinarID uuid.UUID
}

// GetWebinarResult represents the output
type GetWebinarResult struct {
	Webinar *models.Webinar
}

// GetWebinarService handles webinar retrieval
type GetWebinarService struct {
	webinarRepo repositories.WebinarRepo
}

// NewGetWebinarService creates a new get webinar service
func NewGetWebinarService(webinarRepo repositories.WebinarRepo) *GetWebinarService {
	return &GetWebinarService{
		webinarRepo: webinarRepo,
	}
}

// Execute executes the GetWebinar query
func (s *GetWebinarService) Execute(ctx context.Context, query GetWebinarQuery) (*GetWebinarResult, error) {
	webinar, err := s.webinarRepo.FindByID(ctx, query.WebinarID)
	if err != nil {
		return nil, fmt.Errorf("failed to get webinar: %w", err)
	}

	return &GetWebinarResult{
		Webinar: webinar,
	}, nil
}

// ListWebinarsQuery represents the query for listing webinars
type ListWebinarsQuery struct {
	Status string
	Limit  int
	Offset int
}

// ListWebinarsResult represents the output
type ListWebinarsResult struct {
	Webinars []*models.Webinar
}

// ListWebinarsService handles webinar listing
type ListWebinarsService struct {
	webinarRepo repositories.WebinarRepo
}

// NewListWebinarsService creates a new list webinars service
func NewListWebinarsService(webinarRepo repositories.WebinarRepo) *ListWebinarsService {
	return &ListWebinarsService{
		webinarRepo: webinarRepo,
	}
}

// Execute executes the ListWebinars query
func (s *ListWebinarsService) Execute(ctx context.Context, query ListWebinarsQuery) (*ListWebinarsResult, error) {
	limit := query.Limit
	if limit == 0 {
		limit = 10
	}

	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	webinars, err := s.webinarRepo.FindByStatus(ctx, query.Status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list webinars: %w", err)
	}

	return &ListWebinarsResult{
		Webinars: webinars,
	}, nil
}
