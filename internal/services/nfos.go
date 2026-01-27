package services

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/repositories"
	"github.com/google/uuid"
)

// GetNFOQuery represents the query for getting an NFO
type GetNFOQuery struct {
	NFOID uuid.UUID
}

// GetNFOResult represents the output
type GetNFOResult struct {
	NFO *models.NFO
}

// GetNFOService handles NFO retrieval
type GetNFOService struct {
	nfoRepo repositories.NFORepo
}

// NewGetNFOService creates a new get NFO service
func NewGetNFOService(nfoRepo repositories.NFORepo) *GetNFOService {
	return &GetNFOService{
		nfoRepo: nfoRepo,
	}
}

// Execute executes the GetNFO query
func (s *GetNFOService) Execute(ctx context.Context, query GetNFOQuery) (*GetNFOResult, error) {
	nfo, err := s.nfoRepo.FindByID(ctx, query.NFOID)
	if err != nil {
		return nil, fmt.Errorf("failed to get NFO: %w", err)
	}

	return &GetNFOResult{
		NFO: nfo,
	}, nil
}

// ListNFOsQuery represents the query for listing NFOs
type ListNFOsQuery struct {
	Status string
	Limit  int
	Offset int
}

// ListNFOsResult represents the output
type ListNFOsResult struct {
	NFOs []*models.NFO
}

// ListNFOsService handles NFO listing
type ListNFOsService struct {
	nfoRepo repositories.NFORepo
}

// NewListNFOsService creates a new list NFOs service
func NewListNFOsService(nfoRepo repositories.NFORepo) *ListNFOsService {
	return &ListNFOsService{
		nfoRepo: nfoRepo,
	}
}

// Execute executes the ListNFOs query
func (s *ListNFOsService) Execute(ctx context.Context, query ListNFOsQuery) (*ListNFOsResult, error) {
	limit := query.Limit
	if limit == 0 {
		limit = 10
	}

	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	var nfos []*models.NFO
	var err error

	if query.Status == "open" {
		nfos, err = s.nfoRepo.FindOpen(ctx)
	} else {
		nfos, err = s.nfoRepo.FindByStatus(ctx, query.Status, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list NFOs: %w", err)
	}

	return &ListNFOsResult{
		NFOs: nfos,
	}, nil
}
