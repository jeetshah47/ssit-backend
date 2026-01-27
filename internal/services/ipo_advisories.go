package services

import (
	"context"
	"fmt"
	"time"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/equitywala/backend/internal/models"
	"github.com/equitywala/backend/internal/repositories"
	"github.com/google/uuid"
)

// CreateIPOAdvisoryCmd represents the command for creating an IPO advisory
type CreateIPOAdvisoryCmd struct {
	IPOName      string
	IPOSymbol    *string
	GMP          *float64
	Suggestion   *string
	LotSize      int
	PriceBandMin float64
	PriceBandMax float64
	IssueDate    *string
	IssueSize    *string
	IPOTimetable *string
	PublishedBy  uuid.UUID
}

// CreateIPOAdvisoryCmdOutputData represents the output
type CreateIPOAdvisoryCmdOutputData struct {
	Advisory *models.IPOAdvisory
}

// CreateIPOAdvisoryService handles IPO advisory creation
type CreateIPOAdvisoryService struct {
	ipoAdvisoryRepo  repositories.IPOAdvisoryRepo
	advisoryTypeRepo repositories.AdvisoryTypeRepo
}

// NewCreateIPOAdvisoryService creates a new create IPO advisory service
func NewCreateIPOAdvisoryService(
	ipoAdvisoryRepo repositories.IPOAdvisoryRepo,
	advisoryTypeRepo repositories.AdvisoryTypeRepo,
) *CreateIPOAdvisoryService {
	return &CreateIPOAdvisoryService{
		ipoAdvisoryRepo:  ipoAdvisoryRepo,
		advisoryTypeRepo: advisoryTypeRepo,
	}
}

// Execute executes the CreateIPOAdvisory command
func (s *CreateIPOAdvisoryService) Execute(ctx context.Context, cmd CreateIPOAdvisoryCmd) (*CreateIPOAdvisoryCmdOutputData, error) {
	// Validate command
	if err := s.validateCommand(cmd); err != nil {
		return nil, err
	}

	// Get ipo advisory type
	advisoryType, err := s.advisoryTypeRepo.FindByName(ctx, "ipo")
	if err != nil {
		return nil, fmt.Errorf("failed to find advisory type: %w", err)
	}
	advisoryTypeID := advisoryType.ID

	now := time.Now()
	advisory := &models.IPOAdvisory{
		ID:             uuid.New(),
		AdvisoryTypeID: advisoryTypeID,
		IPOName:        cmd.IPOName,
		IPOSymbol:      cmd.IPOSymbol,
		GMP:            cmd.GMP,
		Suggestion:     cmd.Suggestion,
		LotSize:        cmd.LotSize,
		PriceBandMin:   cmd.PriceBandMin,
		PriceBandMax:   cmd.PriceBandMax,
		IssueDate:      cmd.IssueDate,
		IssueSize:      cmd.IssueSize,
		IPOTimetable:   cmd.IPOTimetable,
		Status:         "upcoming",
		PublishedBy:    &cmd.PublishedBy,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Persist via repository
	if err := s.ipoAdvisoryRepo.Create(ctx, advisory); err != nil {
		return nil, fmt.Errorf("failed to create IPO advisory: %w", err)
	}

	return &CreateIPOAdvisoryCmdOutputData{Advisory: advisory}, nil
}

// validateCommand validates the command input
func (s *CreateIPOAdvisoryService) validateCommand(cmd CreateIPOAdvisoryCmd) error {
	if cmd.IPOName == "" {
		return errors.NewDomainError("INVALID_INPUT", "IPO name is required")
	}
	if cmd.LotSize <= 0 {
		return errors.NewDomainError("INVALID_INPUT", "lot size must be greater than 0")
	}
	if cmd.PriceBandMin <= 0 || cmd.PriceBandMax <= 0 {
		return errors.NewDomainError("INVALID_INPUT", "price band must be greater than 0")
	}
	if cmd.PriceBandMax < cmd.PriceBandMin {
		return errors.NewDomainError("INVALID_INPUT", "price band max must be greater than or equal to min")
	}
	return nil
}

// GetIPOAdvisoryQuery represents the query for getting an IPO advisory
type GetIPOAdvisoryQuery struct {
	AdvisoryID uuid.UUID
}

// GetIPOAdvisoryResult represents the output
type GetIPOAdvisoryResult struct {
	Advisory *models.IPOAdvisory
}

// GetIPOAdvisoryService handles IPO advisory retrieval
type GetIPOAdvisoryService struct {
	ipoAdvisoryRepo repositories.IPOAdvisoryRepo
}

// NewGetIPOAdvisoryService creates a new get IPO advisory service
func NewGetIPOAdvisoryService(ipoAdvisoryRepo repositories.IPOAdvisoryRepo) *GetIPOAdvisoryService {
	return &GetIPOAdvisoryService{
		ipoAdvisoryRepo: ipoAdvisoryRepo,
	}
}

// Execute executes the GetIPOAdvisory query
func (s *GetIPOAdvisoryService) Execute(ctx context.Context, query GetIPOAdvisoryQuery) (*GetIPOAdvisoryResult, error) {
	advisory, err := s.ipoAdvisoryRepo.FindByID(ctx, query.AdvisoryID)
	if err != nil {
		return nil, err
	}

	return &GetIPOAdvisoryResult{Advisory: advisory}, nil
}

// ListIPOAdvisoriesQuery represents the query for listing IPO advisories
type ListIPOAdvisoriesQuery struct {
	Status string
	Limit  int
	Offset int
}

// ListIPOAdvisoriesResult represents the output
type ListIPOAdvisoriesResult struct {
	Advisories []*models.IPOAdvisory
}

// ListIPOAdvisoriesService handles IPO advisory listing
type ListIPOAdvisoriesService struct {
	ipoAdvisoryRepo repositories.IPOAdvisoryRepo
}

// NewListIPOAdvisoriesService creates a new list IPO advisories service
func NewListIPOAdvisoriesService(ipoAdvisoryRepo repositories.IPOAdvisoryRepo) *ListIPOAdvisoriesService {
	return &ListIPOAdvisoriesService{
		ipoAdvisoryRepo: ipoAdvisoryRepo,
	}
}

// Execute executes the ListIPOAdvisories query
func (s *ListIPOAdvisoriesService) Execute(ctx context.Context, query ListIPOAdvisoriesQuery) (*ListIPOAdvisoriesResult, error) {
	if query.Limit == 0 {
		query.Limit = 10
	}

	advisories, err := s.ipoAdvisoryRepo.FindByStatus(ctx, query.Status, query.Limit, query.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list IPO advisories: %w", err)
	}

	return &ListIPOAdvisoriesResult{Advisories: advisories}, nil
}

// UpdateIPOAdvisoryCmd represents the command for updating an IPO advisory
type UpdateIPOAdvisoryCmd struct {
	AdvisoryID   uuid.UUID
	IPOName      *string
	GMP          *float64
	Suggestion   *string
	Status       *string
	IssueDate    *string
	IPOTimetable *string
}

// UpdateIPOAdvisoryCmdOutputData represents the output
type UpdateIPOAdvisoryCmdOutputData struct {
	Advisory *models.IPOAdvisory
}

// UpdateIPOAdvisoryService handles IPO advisory updates
type UpdateIPOAdvisoryService struct {
	ipoAdvisoryRepo repositories.IPOAdvisoryRepo
}

// NewUpdateIPOAdvisoryService creates a new update IPO advisory service
func NewUpdateIPOAdvisoryService(ipoAdvisoryRepo repositories.IPOAdvisoryRepo) *UpdateIPOAdvisoryService {
	return &UpdateIPOAdvisoryService{
		ipoAdvisoryRepo: ipoAdvisoryRepo,
	}
}

// Execute executes the UpdateIPOAdvisory command
func (s *UpdateIPOAdvisoryService) Execute(ctx context.Context, cmd UpdateIPOAdvisoryCmd) (*UpdateIPOAdvisoryCmdOutputData, error) {
	// Get existing advisory
	advisory, err := s.ipoAdvisoryRepo.FindByID(ctx, cmd.AdvisoryID)
	if err != nil {
		return nil, err
	}

	// Check if published - core fields immutable
	if advisory.IsPublished() {
		// Only allow status and GMP updates for published advisories
		if cmd.Status != nil {
			advisory.Status = *cmd.Status
		}
		if cmd.GMP != nil {
			advisory.GMP = cmd.GMP
		}
	} else {
		// Allow all updates for draft/upcoming advisories
		if cmd.IPOName != nil {
			advisory.IPOName = *cmd.IPOName
		}
		if cmd.GMP != nil {
			advisory.GMP = cmd.GMP
		}
		if cmd.Suggestion != nil {
			advisory.Suggestion = cmd.Suggestion
		}
		if cmd.Status != nil {
			advisory.Status = *cmd.Status
			if *cmd.Status != "upcoming" {
				now := time.Now()
				advisory.PublishedAt = &now
			}
		}
		if cmd.IssueDate != nil {
			advisory.IssueDate = cmd.IssueDate
		}
		if cmd.IPOTimetable != nil {
			advisory.IPOTimetable = cmd.IPOTimetable
		}
	}

	// Update
	if err := s.ipoAdvisoryRepo.Update(ctx, advisory); err != nil {
		return nil, fmt.Errorf("failed to update IPO advisory: %w", err)
	}

	return &UpdateIPOAdvisoryCmdOutputData{Advisory: advisory}, nil
}
