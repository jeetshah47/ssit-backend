package repositories

import (
	"context"
	"fmt"

	"github.com/equitywala/backend/internal/common/errors"
	"github.com/equitywala/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WeeklyMarketMoodRepo defines the interface for weekly market mood data persistence
type WeeklyMarketMoodRepo interface {
	Create(ctx context.Context, mood *models.WeeklyMarketMood) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.WeeklyMarketMood, error)
	FindLatest(ctx context.Context) (*models.WeeklyMarketMood, error)
	FindAll(ctx context.Context, limit, offset int) ([]*models.WeeklyMarketMood, error)
	Update(ctx context.Context, mood *models.WeeklyMarketMood) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// weeklyMarketMoodRepo implements the weekly market mood repository interface
type weeklyMarketMoodRepo struct {
	*RepoContext
}

// NewWeeklyMarketMoodRepo creates a new weekly market mood repository
func NewWeeklyMarketMoodRepo(ctx *RepoContext) WeeklyMarketMoodRepo {
	return &weeklyMarketMoodRepo{RepoContext: ctx}
}

// Create creates a new weekly market mood
func (r *weeklyMarketMoodRepo) Create(ctx context.Context, mood *models.WeeklyMarketMood) error {
	if err := r.db.WithContext(ctx).Create(mood).Error; err != nil {
		return fmt.Errorf("failed to create weekly market mood: %w", err)
	}
	return nil
}

// FindByID finds a weekly market mood by ID
func (r *weeklyMarketMoodRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.WeeklyMarketMood, error) {
	var mood models.WeeklyMarketMood
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&mood).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("WEEKLY_MARKET_MOOD_NOT_FOUND", "weekly market mood not found")
		}
		return nil, fmt.Errorf("failed to find weekly market mood: %w", err)
	}
	return &mood, nil
}

// FindLatest finds the latest weekly market mood
func (r *weeklyMarketMoodRepo) FindLatest(ctx context.Context) (*models.WeeklyMarketMood, error) {
	var mood models.WeeklyMarketMood
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		First(&mood).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewDomainError("WEEKLY_MARKET_MOOD_NOT_FOUND", "no weekly market mood found")
		}
		return nil, fmt.Errorf("failed to find latest weekly market mood: %w", err)
	}
	return &mood, nil
}

// FindAll finds all weekly market moods
func (r *weeklyMarketMoodRepo) FindAll(ctx context.Context, limit, offset int) ([]*models.WeeklyMarketMood, error) {
	var moods []*models.WeeklyMarketMood
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&moods).Error; err != nil {
		return nil, fmt.Errorf("failed to find weekly market moods: %w", err)
	}
	return moods, nil
}

// Update updates an existing weekly market mood
func (r *weeklyMarketMoodRepo) Update(ctx context.Context, mood *models.WeeklyMarketMood) error {
	if err := r.db.WithContext(ctx).Save(mood).Error; err != nil {
		return fmt.Errorf("failed to update weekly market mood: %w", err)
	}
	return nil
}

// Delete deletes a weekly market mood
func (r *weeklyMarketMoodRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.WeeklyMarketMood{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete weekly market mood: %w", err)
	}
	return nil
}
