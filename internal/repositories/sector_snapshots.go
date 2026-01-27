package repositories

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/equitywala/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SectorSnapshotRepo defines the interface for sector snapshot data persistence
type SectorSnapshotRepo interface {
	// Create creates a new sector snapshot
	Create(ctx context.Context, snapshot *models.SectorSnapshot) error

	// FindByID finds a sector snapshot by ID
	FindByID(ctx context.Context, id uuid.UUID) (*models.SectorSnapshot, error)

	// FindBySectorName finds a sector snapshot by sector name
	FindBySectorName(ctx context.Context, sectorName string) (*models.SectorSnapshot, error)

	// FindAll finds all sector snapshots
	FindAll(ctx context.Context, limit int) ([]*models.SectorSnapshot, error)

	// Update updates an existing sector snapshot
	Update(ctx context.Context, snapshot *models.SectorSnapshot) error

	// Delete deletes a sector snapshot
	Delete(ctx context.Context, id uuid.UUID) error
}

// sectorSnapshotRepo implements the sector snapshot repository interface
type sectorSnapshotRepo struct {
	*RepoContext
}

// NewSectorSnapshotRepo creates a new sector snapshot repository
func NewSectorSnapshotRepo(ctx *RepoContext) SectorSnapshotRepo {
	return &sectorSnapshotRepo{RepoContext: ctx}
}

// Create creates a new sector snapshot
func (r *sectorSnapshotRepo) Create(ctx context.Context, snapshot *models.SectorSnapshot) error {
	if err := r.db.WithContext(ctx).Create(snapshot).Error; err != nil {
		return fmt.Errorf("failed to create sector snapshot: %w", err)
	}
	return nil
}

// FindByID finds a sector snapshot by ID
func (r *sectorSnapshotRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.SectorSnapshot, error) {
	var snapshot models.SectorSnapshot
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&snapshot).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("SECTOR_SNAPSHOT_NOT_FOUND", "sector snapshot not found")
		}
		return nil, fmt.Errorf("failed to find sector snapshot: %w", err)
	}
	return &snapshot, nil
}

// FindBySectorName finds a sector snapshot by sector name
func (r *sectorSnapshotRepo) FindBySectorName(ctx context.Context, sectorName string) (*models.SectorSnapshot, error) {
	var snapshot models.SectorSnapshot
	if err := r.db.WithContext(ctx).Where("sector_name = ?", sectorName).First(&snapshot).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("SECTOR_SNAPSHOT_NOT_FOUND", "sector snapshot not found")
		}
		return nil, fmt.Errorf("failed to find sector snapshot: %w", err)
	}
	return &snapshot, nil
}

// FindAll finds all sector snapshots
func (r *sectorSnapshotRepo) FindAll(ctx context.Context, limit int) ([]*models.SectorSnapshot, error) {
	var snapshots []*models.SectorSnapshot
	query := r.db.WithContext(ctx)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Order("sector_name ASC").Find(&snapshots).Error; err != nil {
		return nil, fmt.Errorf("failed to find sector snapshots: %w", err)
	}
	return snapshots, nil
}

// Update updates an existing sector snapshot
func (r *sectorSnapshotRepo) Update(ctx context.Context, snapshot *models.SectorSnapshot) error {
	if err := r.db.WithContext(ctx).Save(snapshot).Error; err != nil {
		return fmt.Errorf("failed to update sector snapshot: %w", err)
	}
	return nil
}

// Delete deletes a sector snapshot
func (r *sectorSnapshotRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.SectorSnapshot{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete sector snapshot: %w", err)
	}
	return nil
}
