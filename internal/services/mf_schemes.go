package services

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/repositories"
	"github.com/google/uuid"
)

// GetMFSchemeQuery represents the query for getting an MF scheme
type GetMFSchemeQuery struct {
	SchemeID uuid.UUID
}

// GetMFSchemeResult represents the output
type GetMFSchemeResult struct {
	Scheme *models.MFScheme
}

// GetMFSchemeService handles MF scheme retrieval
type GetMFSchemeService struct {
	mfSchemeRepo repositories.MFSchemeRepo
}

// NewGetMFSchemeService creates a new get MF scheme service
func NewGetMFSchemeService(mfSchemeRepo repositories.MFSchemeRepo) *GetMFSchemeService {
	return &GetMFSchemeService{
		mfSchemeRepo: mfSchemeRepo,
	}
}

// Execute executes the GetMFScheme query
func (s *GetMFSchemeService) Execute(ctx context.Context, query GetMFSchemeQuery) (*GetMFSchemeResult, error) {
	scheme, err := s.mfSchemeRepo.FindByID(ctx, query.SchemeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get MF scheme: %w", err)
	}

	return &GetMFSchemeResult{
		Scheme: scheme,
	}, nil
}

// ListMFSchemesQuery represents the query for listing MF schemes
type ListMFSchemesQuery struct {
	Status string
	Limit  int
	Offset int
}

// ListMFSchemesResult represents the output
type ListMFSchemesResult struct {
	Schemes []*models.MFScheme
}

// ListMFSchemesService handles MF scheme listing
type ListMFSchemesService struct {
	mfSchemeRepo repositories.MFSchemeRepo
}

// NewListMFSchemesService creates a new list MF schemes service
func NewListMFSchemesService(mfSchemeRepo repositories.MFSchemeRepo) *ListMFSchemesService {
	return &ListMFSchemesService{
		mfSchemeRepo: mfSchemeRepo,
	}
}

// Execute executes the ListMFSchemes query
func (s *ListMFSchemesService) Execute(ctx context.Context, query ListMFSchemesQuery) (*ListMFSchemesResult, error) {
	limit := query.Limit
	if limit == 0 {
		limit = 10
	}

	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	schemes, err := s.mfSchemeRepo.FindByStatus(ctx, query.Status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list MF schemes: %w", err)
	}

	return &ListMFSchemesResult{
		Schemes: schemes,
	}, nil
}
